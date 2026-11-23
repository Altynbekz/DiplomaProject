package service

import (
	"context"
	"diploma/internal/domain"
	"diploma/internal/store"
	db "diploma/internal/store/sqlc"
	"diploma/internal/zap"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"reflect"
	"strings"
	"time"
)

type Service struct {
	Store                    *store.Store
	Origin, Actor, CursorKey string
	Now                      func() time.Time
}

func (s *Service) audit(ctx context.Context, q *db.Queries, request, action, kind, id string) error {
	return q.Audit(ctx, db.AuditParams{ID: uuid.NewString(), RequestID: request, Actor: s.Actor, Action: action, EntityType: kind, EntityID: id, CreatedAt: s.Now().UTC()})
}
func target(v db.Target) Target {
	return Target{v.ID, v.Name, v.CanonicalOrigin, v.LabTargetKey, v.CreatedAt.UTC()}
}
func (s *Service) Ready(ctx context.Context) error {
	v, e := s.Store.Q.Ready(ctx)
	if e != nil {
		return e
	}
	if !v.Bool {
		return errors.New("schema unavailable")
	}
	return nil
}
func (s *Service) CreateTarget(ctx context.Context, name, key, request string) (Target, error) {
	if strings.TrimSpace(name) == "" || len(name) > 200 || key != "juice-shop" {
		return Target{}, domain.Fail(422, "validation_failed", "", "Name and registered juice-shop lab key required")
	}
	tx, q, e := s.Store.Begin(ctx)
	if e != nil {
		return Target{}, e
	}
	defer tx.Rollback(ctx)
	v, e := q.CreateTarget(ctx, db.CreateTargetParams{ID: uuid.NewString(), Name: name, CanonicalOrigin: s.Origin, LabTargetKey: key, CreatedAt: s.Now().UTC()})
	if e != nil {
		return Target{}, e
	}
	if e = s.audit(ctx, q, request, "target.create", "target", v.ID); e != nil {
		return Target{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Target{}, e
	}
	return target(v), nil
}
func (s *Service) Target(ctx context.Context, id string) (Target, error) {
	v, e := s.Store.Q.GetTarget(ctx, id)
	return target(v), e
}
func (s *Service) Targets(ctx context.Context, limit int, raw string) (Page[Target], error) {
	p := Page[Target]{Items: []Target{}}
	filter := Filter("targets")
	c, e := s.Cursor(raw, filter)
	if e != nil {
		return p, e
	}
	vs, e := s.Store.Q.ListTargets(ctx, db.ListTargetsParams{ID: c.ID, Limit: int32(limit + 1)})
	if e != nil {
		return p, e
	}
	for i, v := range vs {
		if i == limit {
			p.Next = s.Next(Cursor{ID: vs[i-1].ID, Filter: filter})
			break
		}
		p.Items = append(p.Items, target(v))
	}
	return p, nil
}
func (s *Service) Import(ctx context.Context, targetID, request string, raw []byte, submitted *Provenance) (Import, error) {
	t, e := s.Target(ctx, targetID)
	if e != nil {
		return Import{}, e
	}
	r, e := zap.Parse(raw, targetID, t.Origin)
	if e != nil {
		return Import{}, e
	}
	var prov Provenance
	if submitted != nil {
		prov = *submitted
	}
	prov.SchemaVersion = "report-provenance/v2"
	if prov.SanitizationSteps == nil {
		prov.SanitizationSteps = []string{}
	}
	var declaredVersion *string
	if r.Version != "" {
		declaredVersion = &r.Version
	}
	for _, h := range []*string{prov.PlanSHA256, prov.RawSHA256, prov.SanitizedSHA256} {
		if h != nil {
			if len(*h) != 64 || strings.Trim(*h, "0123456789abcdef") != "" {
				return Import{}, domain.Fail(422, "validation_failed", "/provenance", "Digests must be lowercase SHA-256")
			}
		}
	}
	if prov.ScanTime != nil {
		if _, e = time.Parse(time.RFC3339, *prov.ScanTime); e != nil {
			return Import{}, domain.Fail(422, "validation_failed", "/provenance/scan_time", "Expected RFC3339 scan time")
		}
	}
	if prov.SanitizedSHA256 != nil && *prov.SanitizedSHA256 != r.Hash {
		return Import{}, domain.Fail(422, "validation_failed", "/provenance/sanitized_sha256", "Sanitized digest differs from uploaded bytes")
	}
	if prov.AuthorizationRef == nil || prov.ScanTime == nil {
		r.Warnings = append(r.Warnings, zap.Warning{Code: "incomplete_provenance", Pointer: "/provenance", Message: "Authorization reference or scan time unknown; genuine-report acceptance is not established"})
	}
	meta, e := json.Marshal(prov)
	if e != nil {
		return Import{}, e
	}
	if len(meta) > 65536 {
		return Import{}, domain.Fail(413, "limit_exceeded", "/provenance", "Provenance exceeds 64 KiB")
	}
	// Validate the exact shape persisted as JSONB, including nested keys.
	var persistedMetadata map[string]any
	if err := json.Unmarshal(meta, &persistedMetadata); err != nil {
		return Import{}, err
	}
	if err := domain.ValidateStorageJSONText(persistedMetadata, "/provenance"); err != nil {
		return Import{}, err
	}
	tx, q, e := s.Store.Begin(ctx)
	if e != nil {
		return Import{}, e
	}
	defer tx.Rollback(ctx)
	// The target row serializes imports, including concurrent identical report bytes.
	if _, e = q.LockTarget(ctx, targetID); e != nil {
		return Import{}, e
	}
	old, e := q.FindImport(ctx, db.FindImportParams{TargetID: targetID, ReportSha256: r.Hash, ParserVersion: zap.Version})
	if e == nil {
		existing, err := importDTO(db.GetImportRow(old))
		if err != nil {
			return Import{}, err
		}
		// Omitting provenance asks for the existing record without a change.
		if submitted != nil {
			prov.SchemaVersion = existing.Provenance.SchemaVersion
			if !reflect.DeepEqual(prov, existing.Provenance) {
				return Import{}, domain.Fail(409, "conflict", "/provenance", "This report was already imported with different provenance. The stored provenance is immutable.")
			}
		}
		existing.Duplicate = true
		return existing, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return Import{}, e
	}
	id := uuid.NewString()
	now := s.Now().UTC().Truncate(time.Microsecond)
	unique := map[string]bool{}
	for _, o := range r.Observations {
		unique[o.Fingerprint] = true
	}
	warnings, _ := json.Marshal(r.Warnings)
	unlocated, _ := json.Marshal(r.Unlocated)
	e = q.CreateImport(ctx, db.CreateImportParams{ID: id, TargetID: targetID, ReportSha256: r.Hash, ParserVersion: zap.Version, RawBytes: raw, SourceMetadataJson: meta, WarningsJson: warnings, UnlocatedAlertsJson: unlocated, FindingCount: int32(len(unique)), ObservationCount: int32(len(r.Observations)), CreatedAt: now, ReportDeclaredVersion: pgtype.Text{String: r.Version, Valid: declaredVersion != nil}})
	if e != nil {
		// The stored-size constraint uses PostgreSQL's JSONB text rendering,
		// which may be larger than the compact client-side encoding.
		var violation *pgconn.PgError
		if errors.As(e, &violation) && violation.Code == "23514" &&
			violation.TableName == "report_imports" &&
			violation.ConstraintName == "report_imports_source_metadata_json_check" {
			return Import{}, domain.Fail(413, "limit_exceeded", "/provenance", "Stored provenance exceeds 64 KiB")
		}
		return Import{}, e
	}
	// Hold the target lock while assigning receipt times. Preserve source
	// order with microsecond steps and keep subsequent imports later even
	// when the application clock has not advanced.
	observationTime := now
	latest, err := q.LatestTargetObservationTime(ctx, targetID)
	if err == nil && !observationTime.After(latest) {
		observationTime = latest.Add(time.Microsecond)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Import{}, err
	}
	for _, o := range r.Observations {
		fid, e := q.FindFinding(ctx, db.FindFindingParams{TargetID: targetID, Fingerprint: o.Fingerprint})
		if errors.Is(e, pgx.ErrNoRows) {
			fid = uuid.NewString()
			e = q.CreateFinding(ctx, db.CreateFindingParams{ID: fid, TargetID: targetID, Fingerprint: o.Fingerprint, PluginID: o.PluginID, Method: o.Method, CanonicalOrigin: o.Origin, Path: o.Path, QueryParameterNames: o.QueryNames, ParameterName: o.Parameter, Title: o.Title, CreatedAt: now})
		}
		if e != nil {
			return Import{}, e
		}
		e = q.CreateObservation(ctx, db.CreateObservationParams{ID: uuid.NewString(), TargetID: targetID, FindingID: fid, ReportImportID: id, SourcePointer: o.Pointer, InstanceIndex: o.Index, RawUri: o.RawURI, ScannerSeverityRaw: o.SeverityRaw, ScannerSeverity: o.Severity, ConfidenceRaw: o.Confidence, CweIds: o.CWE, EvidenceJson: o.Evidence, Description: o.Description, Remediation: o.Remediation, CreatedAt: observationTime})
		if e != nil {
			return Import{}, e
		}
		observationTime = observationTime.Add(time.Microsecond)
	}
	if e = s.audit(ctx, q, request, "report.import", "import", id); e != nil {
		return Import{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Import{}, e
	}
	return Import{
		ID: id, TargetID: targetID, Hash: r.Hash, ParserVersion: zap.Version,
		FindingCount: int32(len(unique)), ObservationCount: int32(len(r.Observations)),
		UnlocatedCount: len(r.Unlocated), Warnings: r.Warnings, Provenance: prov, ReportDeclaredVersion: declaredVersion, ParserDetectedFormat: zap.ReportFormat,
		CreatedAt: now.Truncate(time.Microsecond),
	}, nil
}
func importDTO(v db.GetImportRow) (Import, error) {
	r := Import{ID: v.ID, TargetID: v.TargetID, Hash: v.ReportSha256, ParserVersion: v.ParserVersion, FindingCount: v.FindingCount, ObservationCount: v.ObservationCount, CreatedAt: v.CreatedAt.UTC()}
	r.ParserDetectedFormat = zap.ReportFormat
	if v.ReportDeclaredVersion.Valid {
		r.ReportDeclaredVersion = &v.ReportDeclaredVersion.String
	}
	var u []json.RawMessage
	if e := json.Unmarshal(v.WarningsJson, &r.Warnings); e != nil {
		return r, e
	}
	if e := json.Unmarshal(v.UnlocatedAlertsJson, &u); e != nil {
		return r, e
	}
	r.UnlocatedCount = len(u)
	if e := json.Unmarshal(v.SourceMetadataJson, &r.Provenance); e != nil {
		return r, e
	}
	return r, nil
}
func (s *Service) GetImport(ctx context.Context, id string) (Import, error) {
	v, e := s.Store.Q.GetImport(ctx, id)
	if e != nil {
		return Import{}, e
	}
	return importDTO(v)
}
func (s *Service) Raw(ctx context.Context, id, request string) ([]byte, error) {
	tx, q, e := s.Store.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	raw, e := q.GetRaw(ctx, id)
	if e != nil {
		return nil, e
	}
	if e = s.audit(ctx, q, request, "evidence.raw.read", "import", id); e != nil {
		return nil, e
	}
	if e = tx.Commit(ctx); e != nil {
		return nil, e
	}
	return raw, nil
}

type FindingFilter struct{ TargetID, ImportID, Category, Verification, Severity string }

func finding(v db.ListFindingsRow) Finding {
	title, _ := Redact(v.Title)
	path, _ := Redact(v.Path)
	param, _ := Redact(v.ParameterName)
	return Finding{ID: v.ID, TargetID: v.TargetID, Title: title, Method: v.Method, Origin: v.CanonicalOrigin, Path: path, Parameter: param, QueryNames: v.QueryParameterNames, ObservationCount: v.ObservationCount, Severity: v.ScannerSeverity, SeverityObservationID: v.SeverityObservationID, Verification: "unreviewed", Classification: Classification{Status: "unmapped", Candidates: []string{}}, PluginID: v.PluginID, FingerprintVersion: v.FingerprintVersion, Fingerprint: v.Fingerprint, CreatedAt: v.CreatedAt.UTC()}
}
func (s *Service) Findings(ctx context.Context, f FindingFilter, limit int, raw string) (Page[Finding], error) {
	p := Page[Finding]{Items: []Finding{}}
	if _, e := s.Target(ctx, f.TargetID); e != nil {
		return p, e
	}
	if f.ImportID != "" {
		im, e := s.GetImport(ctx, f.ImportID)
		if e != nil {
			return p, e
		}
		if im.TargetID != f.TargetID {
			return p, domain.Fail(422, "validation_failed", "/import_id", "Import belongs to a different target")
		}
	}
	filter := Filter("findings", f.TargetID, f.ImportID, f.Category, f.Verification, f.Severity)
	c, e := s.Cursor(raw, filter)
	if e != nil {
		return p, e
	}
	vs, e := s.Store.Q.ListFindings(ctx, db.ListFindingsParams{TargetID: f.TargetID, ImportID: f.ImportID, Category: f.Category, Verification: f.Verification, Severity: f.Severity, AfterID: c.ID, PageLimit: int32(limit + 1)})
	if e != nil {
		return p, e
	}
	for i, v := range vs {
		if i == limit {
			p.Next = s.Next(Cursor{ID: vs[i-1].ID, Filter: filter})
			break
		}
		p.Items = append(p.Items, finding(v))
	}
	return p, nil
}
func (s *Service) Finding(ctx context.Context, id string) (Finding, error) {
	// Dedicated identity lookup keeps the target-scoped list query and DTO identical.
	targetID, e := s.Store.FindingTarget(ctx, id)
	if e != nil {
		return Finding{}, e
	}
	vs, e := s.Store.Q.ListFindings(ctx, db.ListFindingsParams{TargetID: targetID, FindingID: id, PageLimit: 1})
	if e != nil {
		return Finding{}, e
	}
	if len(vs) == 0 {
		return Finding{}, pgx.ErrNoRows
	}
	return finding(vs[0]), nil
}
func (s *Service) Observations(ctx context.Context, id string, limit int, raw string) (Page[Observation], error) {
	p := Page[Observation]{Items: []Observation{}}
	f, e := s.Finding(ctx, id)
	if e != nil {
		return p, e
	}
	filter := Filter("observations", id)
	c, e := s.Cursor(raw, filter)
	if e != nil {
		return p, e
	}
	after := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	afterID := uuid.Nil.String()
	if c.ID != "" {
		afterID = c.ID
		after, e = time.Parse(time.RFC3339Nano, c.Time)
		if e != nil {
			return p, domain.Fail(400, "invalid_request", "/cursor", "Invalid timestamp")
		}
	}
	vs, e := s.Store.Q.ListObservations(ctx, db.ListObservationsParams{FindingID: id, AfterID: afterID, AfterTime: after, PageLimit: int32(limit + 1)})
	if e != nil {
		return p, e
	}
	for i, v := range vs {
		if i == limit {
			prev := vs[i-1]
			p.Next = s.Next(Cursor{ID: prev.ID, Time: prev.CreatedAt.UTC().Format(time.RFC3339Nano), Filter: filter})
			break
		}
		desc, d := Redact(v.Description)
		rem, r := Redact(v.Remediation)
		// Scanner evidence/attack/otherinfo are unstructured and may hold arbitrary secrets.
		// Withhold their contents by default rather than claiming heuristic redaction is complete.
		redactions := []string{"attack, evidence and other_info withheld; use audited raw access for private review"}
		if d || r {
			redactions = append(redactions, "description/remediation redacted or truncated")
		}
		p.Items = append(p.Items, Observation{ID: v.ID, ImportID: v.ReportImportID, Pointer: v.SourcePointer, Method: f.Method, Path: f.Path, Parameter: f.Parameter, Severity: v.ScannerSeverity, SeverityRaw: v.ScannerSeverityRaw, Confidence: v.ConfidenceRaw, CWE: v.CweIds, Description: desc, Remediation: rem, Excerpt: "[Private scanner payload withheld]", Redactions: redactions, CreatedAt: v.CreatedAt.UTC()})
	}
	return p, nil
}
