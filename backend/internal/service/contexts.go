package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"diploma/internal/domain"
	"diploma/internal/scope"
	db "diploma/internal/store/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var contextDimensions = []struct {
	name   string
	values []string
}{
	{"exposure", []string{"public", "internal", "restricted", "unknown"}},
	{"required_privileges", []string{"anonymous", "user", "admin", "unknown"}},
	{"data_sensitivity", []string{"public", "internal", "personal", "highly_sensitive", "unknown"}},
	{"business_criticality", []string{"low", "medium", "high", "critical", "unknown"}},
}

var contextMethod = regexp.MustCompile("^[!#$%&'*+.^_\x60|~0-9A-Za-z-]+$")
var contextDigest = regexp.MustCompile("^[0-9a-f]{64}$")

const factorProvenanceVersion = "factor-provenance/v1"
const contextResolverVersion = "exact-endpoint-v1"

type ContextSource struct {
	ObservationID *string `json:"observation_id,omitempty"`
	SourcePointer *string `json:"source_pointer,omitempty"`
	DocumentRef   *string `json:"document_ref,omitempty"`
	SHA256        *string `json:"sha256,omitempty"`
	Note          string  `json:"note"`
}

func (source *ContextSource) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if len(fields) != 3 {
		return errors.New("expected three fields for a context source")
	}
	fieldNames := []string{"document_ref", "sha256", "note"}
	if _, found := fields["observation_id"]; found {
		fieldNames = []string{"observation_id", "source_pointer", "note"}
	}
	for _, field := range fieldNames {
		value := fields[field]
		if len(value) == 0 || value[0] != '"' {
			return errors.New("invalid context source field")
		}
	}
	type wire ContextSource
	return json.Unmarshal(raw, (*wire)(source))
}

type ContextFactorNote struct {
	Rationale string          `json:"rationale"`
	Sources   []ContextSource `json:"sources"`
}

func (note *ContextFactorNote) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	rationale, sources := fields["rationale"], fields["sources"]
	if len(fields) != 2 || len(rationale) == 0 || rationale[0] != '"' || len(sources) == 0 || sources[0] != '[' {
		return errors.New("expected a rationale and source array")
	}
	type wire ContextFactorNote
	return json.Unmarshal(raw, (*wire)(note))
}

type FactorProvenance map[string]ContextFactorNote
type ContextFactors map[string]string

type ContextInput struct {
	Method             json.RawMessage  `json:"method"`
	Path               json.RawMessage  `json:"path"`
	Factors            ContextFactors   `json:"factors"`
	FactorProvenance   FactorProvenance `json:"factor_provenance"`
	ExpectedPreviousID json.RawMessage  `json:"expected_previous_id"`
}

type ContextRevision struct {
	ID               string           `json:"id"`
	TargetID         string           `json:"target_id"`
	Method           *string          `json:"method"`
	Path             *string          `json:"path"`
	Revision         int32            `json:"revision"`
	Factors          ContextFactors   `json:"factors"`
	FactorProvenance FactorProvenance `json:"factor_provenance"`
	Actor            string           `json:"actor"`
	CreatedAt        time.Time        `json:"created_at"`
}

type ResolvedContext struct {
	Factors           ContextFactors   `json:"factors"`
	FactorProvenance  FactorProvenance `json:"factor_provenance"`
	SourceRevisionIDs []string         `json:"source_revision_ids"`
	ResolverVersion   string           `json:"resolver_version"`
	ResolvedAt        time.Time        `json:"resolved_at"`
}

type storedFactorProvenance struct {
	SchemaVersion string           `json:"schema_version"`
	Factors       FactorProvenance `json:"factors"`
}

func parseContextScope(in ContextInput, origin string) (pgtype.Text, pgtype.Text, error) {
	var method, path *string
	if json.Unmarshal(in.Method, &method) != nil || json.Unmarshal(in.Path, &path) != nil ||
		(method == nil) != (path == nil) {
		return pgtype.Text{}, pgtype.Text{}, domain.Fail(422, "validation_failed", "/method", "Supply both method and path or set both to null")
	}
	if method == nil {
		return pgtype.Text{}, pgtype.Text{}, nil
	}
	if len(*method) > 64 || !contextMethod.MatchString(*method) {
		return pgtype.Text{}, pgtype.Text{}, domain.Fail(422, "validation_failed", "/method", "Expected an HTTP method")
	}
	if !strings.HasPrefix(*path, "/") || len(*path) > 4096 || strings.ContainsAny(*path, "?#") {
		return pgtype.Text{}, pgtype.Text{}, domain.Fail(422, "validation_failed", "/path", "Expected an exact endpoint path without a query or fragment")
	}
	location, err := scope.Parse(origin + *path)
	if err != nil || location.Origin != origin || len(location.Path) > 4096 {
		return pgtype.Text{}, pgtype.Text{}, domain.Fail(422, "validation_failed", "/path", "Endpoint path is invalid")
	}
	return pgtype.Text{String: strings.ToUpper(*method), Valid: true}, pgtype.Text{String: location.Path, Valid: true}, nil
}

func validateFactors(ctx context.Context, q *db.Queries, in ContextInput, targetID string) error {
	if len(in.Factors) != len(contextDimensions) || len(in.FactorProvenance) != len(contextDimensions) {
		return domain.Fail(422, "validation_failed", "/factors", "Supply all four context factors and their provenance")
	}
	for _, dimension := range contextDimensions {
		value := in.Factors[dimension.name]
		if !slices.Contains(dimension.values, value) {
			return domain.Fail(422, "validation_failed", "/factors/"+dimension.name, "Select a supported factor value")
		}
		pointer := "/factor_provenance/" + dimension.name
		note, found := in.FactorProvenance[dimension.name]
		if !found {
			return domain.Fail(422, "validation_failed", pointer, "Factor provenance is required")
		}
		if err := boundedText(note.Rationale, pointer+"/rationale", 1, 2048); err != nil {
			return err
		}
		if note.Sources == nil || len(note.Sources) > 4 || value != "unknown" && len(note.Sources) == 0 {
			return domain.Fail(422, "validation_failed", pointer+"/sources", "Known factors require one to four sources. Unknown factors may have an empty source list.")
		}
		for i, source := range note.Sources {
			at := fmt.Sprintf("%s/sources/%d", pointer, i)
			if err := boundedText(source.Note, at+"/note", 0, 256); err != nil {
				return err
			}
			if source.ObservationID != nil && source.SourcePointer != nil && source.DocumentRef == nil && source.SHA256 == nil {
				ref := EvidenceRef{ObservationID: *source.ObservationID, SourcePointer: *source.SourcePointer, Note: source.Note}
				if err := validateObservationRef(ctx, q, ref, "", targetID, at); err != nil {
					return err
				}
				continue
			}
			if source.DocumentRef == nil || source.SHA256 == nil || source.ObservationID != nil || source.SourcePointer != nil {
				return domain.Fail(422, "validation_failed", at, "Provide one observation or one lab document reference")
			}
			if err := boundedText(*source.DocumentRef, at+"/document_ref", 1, 128); err != nil {
				return err
			}
			if !contextDigest.MatchString(*source.SHA256) {
				return domain.Fail(422, "validation_failed", at+"/sha256", "Expected a lowercase SHA-256 digest")
			}
		}
	}
	return nil
}

func (s *Service) CreateContext(ctx context.Context, targetID string, in ContextInput, requestID string) (ContextRevision, error) {
	expected, err := expectedRevision(in.ExpectedPreviousID)
	if err != nil {
		return ContextRevision{}, err
	}
	tx, q, err := s.Store.Begin(ctx)
	if err != nil {
		return ContextRevision{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := q.LockTarget(ctx, targetID); err != nil {
		return ContextRevision{}, err
	}
	target, err := q.GetTarget(ctx, targetID)
	if err != nil {
		return ContextRevision{}, err
	}
	method, path, err := parseContextScope(in, target.CanonicalOrigin)
	if err != nil {
		return ContextRevision{}, err
	}
	head, err := q.LatestContext(ctx, db.LatestContextParams{TargetID: targetID, Method: method, Path: path})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ContextRevision{}, err
	}
	if head.ID != expected {
		return ContextRevision{}, revisionConflict()
	}
	if head.Revision == math.MaxInt32 {
		return ContextRevision{}, domain.Fail(409, "conflict", "/expected_previous_id", "Revision limit reached")
	}
	if err := validateFactors(ctx, q, in, targetID); err != nil {
		return ContextRevision{}, err
	}
	provenance, err := json.Marshal(storedFactorProvenance{SchemaVersion: factorProvenanceVersion, Factors: in.FactorProvenance})
	if err != nil {
		return ContextRevision{}, err
	}
	id, now := uuid.NewString(), revisionTime(s.Now(), head.CreatedAt)
	if err := q.CreateContext(ctx, db.CreateContextParams{
		ID: id, TargetID: targetID, Method: method, Path: path, Revision: head.Revision + 1,
		Exposure: in.Factors["exposure"], RequiredPrivileges: in.Factors["required_privileges"],
		DataSensitivity: in.Factors["data_sensitivity"], BusinessCriticality: in.Factors["business_criticality"],
		FactorProvenanceJson: provenance, Actor: s.Actor, CreatedAt: now,
	}); err != nil {
		var violation *pgconn.PgError
		if errors.As(err, &violation) && violation.Code == "23514" &&
			violation.TableName == "context_revisions" &&
			violation.ConstraintName == "context_revisions_factor_provenance_json_check" {
			return ContextRevision{}, domain.Fail(413, "limit_exceeded", "/factor_provenance", "Stored factor provenance exceeds 64 KiB")
		}
		return ContextRevision{}, err
	}
	if err := s.audit(ctx, q, requestID, "context.append", "context_revision", id); err != nil {
		return ContextRevision{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ContextRevision{}, err
	}
	return ContextRevision{
		ID: id, TargetID: targetID, Method: optionalString(method.String), Path: optionalString(path.String),
		Revision: head.Revision + 1, Factors: in.Factors, FactorProvenance: in.FactorProvenance,
		Actor: s.Actor, CreatedAt: now,
	}, nil
}

func contextDTO(row db.ContextRevision) (ContextRevision, error) {
	var provenance storedFactorProvenance
	if err := json.Unmarshal(row.FactorProvenanceJson, &provenance); err != nil {
		return ContextRevision{}, err
	}
	if provenance.SchemaVersion != factorProvenanceVersion {
		return ContextRevision{}, errors.New("unrecognized stored context provenance version")
	}
	return ContextRevision{
		ID: row.ID, TargetID: row.TargetID, Method: optionalString(row.Method.String), Path: optionalString(row.Path.String),
		Revision: row.Revision, Factors: ContextFactors{
			"exposure": row.Exposure, "required_privileges": row.RequiredPrivileges,
			"data_sensitivity": row.DataSensitivity, "business_criticality": row.BusinessCriticality,
		}, FactorProvenance: provenance.Factors, Actor: row.Actor, CreatedAt: row.CreatedAt.UTC(),
	}, nil
}

func (s *Service) Contexts(ctx context.Context, targetID string, limit int, raw string) (Page[ContextRevision], error) {
	page := Page[ContextRevision]{Items: []ContextRevision{}}
	if _, err := s.Store.Q.GetTarget(ctx, targetID); err != nil {
		return page, err
	}
	filter := Filter("contexts", targetID)
	cursor, when, err := s.historyCursor(raw, filter)
	if err != nil {
		return page, err
	}
	rows, err := s.Store.Q.ListContexts(ctx, db.ListContextsParams{
		TargetID: targetID, FirstPage: raw == "", AfterTime: when, AfterID: cursor.ID, PageLimit: int32(limit + 1),
	})
	if err != nil {
		return page, err
	}
	for i, row := range rows {
		if i == limit {
			prev := page.Items[i-1]
			page.Next = s.Next(Cursor{ID: prev.ID, Time: prev.CreatedAt.Format(time.RFC3339Nano), Filter: filter})
			break
		}
		revision, err := contextDTO(row)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, revision)
	}
	return page, nil
}

func (s *Service) ResolvedContext(ctx context.Context, findingID string) (ResolvedContext, error) {
	endpoint, err := s.Store.Q.FindingEndpoint(ctx, findingID)
	if err != nil {
		return ResolvedContext{}, err
	}
	row, err := s.Store.Q.ResolveContext(ctx, db.ResolveContextParams{TargetID: endpoint.TargetID, Method: endpoint.Method, Path: endpoint.Path})
	result := ResolvedContext{SourceRevisionIDs: []string{}, ResolverVersion: contextResolverVersion, ResolvedAt: s.Now().UTC()}
	if errors.Is(err, pgx.ErrNoRows) {
		result.Factors = ContextFactors{}
		result.FactorProvenance = FactorProvenance{}
		for _, dimension := range contextDimensions {
			result.Factors[dimension.name] = "unknown"
			result.FactorProvenance[dimension.name] = ContextFactorNote{Rationale: "No context revision applies to this endpoint.", Sources: []ContextSource{}}
		}
		return result, nil
	}
	if err != nil {
		return ResolvedContext{}, err
	}
	revision, err := contextDTO(row)
	if err != nil {
		return ResolvedContext{}, err
	}
	result.Factors, result.FactorProvenance = revision.Factors, revision.FactorProvenance
	result.SourceRevisionIDs = []string{revision.ID}
	return result, nil
}
