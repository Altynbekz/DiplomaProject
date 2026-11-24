package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"diploma/internal/cvss"
	"diploma/internal/domain"
	db "diploma/internal/store/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const assessmentSchemaVersion = "assessment-snapshot/v1"
const assessmentPolicyVersion = "context-policy/v1"

type AssessmentSnapshot struct {
	Assessment
	FindingID                string                       `json:"finding_id"`
	TargetID                 string                       `json:"target_id"`
	BaseAssessmentID         *string                      `json:"base_assessment_id"`
	VerificationReviewID     *string                      `json:"verification_review_id"`
	ClassificationRevisionID *string                      `json:"classification_revision_id"`
	Context                  ResolvedContext              `json:"context_snapshot"`
	Metrics                  map[string]string            `json:"metrics"`
	MetricRationales         map[string]MetricRationale   `json:"metric_rationales"`
	FactorDispositions       map[string]FactorDisposition `json:"factor_dispositions"`
	Notes                    string                       `json:"notes"`
	Calculator               *cvss.Provenance             `json:"calculator"`
	SchemaVersion            string                       `json:"schema_version"`
	PolicyVersion            string                       `json:"policy_version"`
	Actor                    string                       `json:"actor"`
}

func nullableID(v pgtype.UUID) string {
	if !v.Valid {
		return ""
	}
	return uuid.UUID(v.Bytes).String()
}

func snapshotDTO(row db.AssessmentSnapshot, stale bool) (AssessmentSnapshot, error) {
	result := AssessmentSnapshot{
		Assessment: Assessment{
			ID: row.ID, Kind: row.Kind, Status: row.Status, Vector: optionalString(row.Vector.String),
			Stale: stale, CreatedAt: row.CreatedAt.UTC(),
		},
		FindingID: row.FindingID, TargetID: row.TargetID,
		BaseAssessmentID:         optionalString(nullableID(row.BaseAssessmentID)),
		VerificationReviewID:     optionalString(nullableID(row.VerificationReviewID)),
		ClassificationRevisionID: optionalString(nullableID(row.ClassificationRevisionID)),
		Notes:                    row.Notes, SchemaVersion: row.SchemaVersion, PolicyVersion: row.PolicyVersion, Actor: row.Actor,
	}
	if row.Score.Valid {
		value, err := row.Score.Float64Value()
		if err != nil {
			return AssessmentSnapshot{}, err
		}
		result.Score = &value.Float64
	}
	fields := []struct {
		raw []byte
		dst any
	}{
		{row.ContextSnapshotJson, &result.Context},
		{row.MetricsJson, &result.Metrics},
		{row.MetricRationalesJson, &result.MetricRationales},
		{row.FactorDispositionsJson, &result.FactorDispositions},
	}
	if row.CalculatorJson != nil {
		fields = append(fields, struct {
			raw []byte
			dst any
		}{row.CalculatorJson, &result.Calculator})
	}
	for _, field := range fields {
		if err := json.Unmarshal(field.raw, field.dst); err != nil {
			return AssessmentSnapshot{}, err
		}
	}
	return result, nil
}

func verifyBaseParent(ctx context.Context, q *db.Queries, findingID, baseID, reviewID, classificationID string) (AssessmentSnapshot, error) {
	row, err := q.GetAssessment(ctx, baseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return AssessmentSnapshot{}, validation("/base_assessment_id", "Base assessment is unavailable")
	}
	if err != nil {
		return AssessmentSnapshot{}, err
	}
	parent := row.AssessmentSnapshot
	if parent.FindingID != findingID || parent.Kind != "B" || parent.Status != "complete" {
		return AssessmentSnapshot{}, validation("/base_assessment_id", "Select a complete Base assessment of this finding")
	}
	latest, err := q.LatestCompleteBaseID(ctx, findingID)
	if err != nil {
		return AssessmentSnapshot{}, err
	}
	if latest != baseID || nullableID(parent.VerificationReviewID) != reviewID || nullableID(parent.ClassificationRevisionID) != classificationID {
		return AssessmentSnapshot{}, assessmentConflict("/base_assessment_id", "The complete Base assessment or its review and classification changed")
	}
	return snapshotDTO(parent, false)
}

func (s *Service) CreateAssessment(ctx context.Context, findingID string, in AssessmentInput, requestID string) (AssessmentSnapshot, error) {
	var empty AssessmentSnapshot
	if err := validateAssessmentInput(in); err != nil {
		return empty, err
	}
	baseID, err := parseAssessmentID(in.BaseAssessmentID, "/base_assessment_id")
	if err != nil {
		return empty, err
	}
	if (in.Kind == "BE") != (baseID != "") {
		return empty, validation("/base_assessment_id", "BE requires a complete Base parent and B requires null")
	}
	expectedReview, err := parseAssessmentID(in.ExpectedVerificationID, "/expected_verification_review_id")
	if err != nil {
		return empty, err
	}
	expectedClassification, err := parseAssessmentID(in.ExpectedClassificationID, "/expected_classification_revision_id")
	if err != nil {
		return empty, err
	}
	tx, q, err := s.Store.Begin(ctx)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(ctx)

	endpoint, err := q.FindingEndpoint(ctx, findingID)
	if err != nil {
		return empty, err
	}
	// Match import's target-before-finding lock order. Context edits take
	// the target lock and verification/classification take the finding lock.
	if _, err = q.LockTarget(ctx, endpoint.TargetID); err != nil {
		return empty, err
	}
	if _, err = q.LockFindingRevision(ctx, findingID); err != nil {
		return empty, err
	}
	review, err := q.LatestVerification(ctx, findingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return empty, err
	}
	if review.ID != expectedReview {
		return empty, assessmentConflict("/expected_verification_review_id", "The current verification review changed")
	}
	classification, err := q.LatestClassification(ctx, findingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return empty, err
	}
	if classification.ID != expectedClassification {
		return empty, assessmentConflict("/expected_classification_revision_id", "The current classification revision changed")
	}
	resolved, err := s.resolveContext(ctx, q, endpoint)
	if err != nil {
		return empty, err
	}
	var contextID string
	if len(resolved.SourceRevisionIDs) == 1 {
		contextID = resolved.SourceRevisionIDs[0]
	}
	baseMetrics := in.Metrics
	rationales := maps.Clone(in.MetricRationales)
	var environment map[string]string
	if in.Kind == "BE" {
		if !slices.Equal(in.ExpectedContextIDs, resolved.SourceRevisionIDs) {
			return empty, assessmentConflict("/expected_context_revision_ids", "The resolved context changed")
		}
		parent, err := verifyBaseParent(ctx, q, findingID, baseID, expectedReview, expectedClassification)
		if err != nil {
			return empty, err
		}
		baseMetrics, environment = parent.Metrics, in.Metrics
		maps.Copy(rationales, parent.MetricRationales)
	}
	if err := validateAssessmentEvidence(ctx, q, in, findingID); err != nil {
		return empty, err
	}
	previous, err := q.LatestAssessmentTime(ctx, findingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return empty, err
	}
	snapshot := AssessmentSnapshot{
		Assessment: Assessment{ID: uuid.NewString(), Kind: in.Kind, Status: "incomplete", CreatedAt: revisionTime(s.Now(), previous)},
		FindingID:  findingID, TargetID: endpoint.TargetID,
		BaseAssessmentID: optionalString(baseID), VerificationReviewID: optionalString(review.ID),
		ClassificationRevisionID: optionalString(classification.ID),
		Context:                  resolved, Metrics: cvss.SnapshotMetrics(baseMetrics, environment), MetricRationales: rationales,
		FactorDispositions: in.FactorDispositions, Notes: *in.Notes,
		SchemaVersion: assessmentSchemaVersion, PolicyVersion: assessmentPolicyVersion, Actor: s.Actor,
	}
	if snapshot.FactorDispositions == nil {
		snapshot.FactorDispositions = map[string]FactorDisposition{}
	}
	if *in.Finalize {
		if review.State != "confirmed" {
			return empty, incomplete("/expected_verification_review_id", "Finalization requires current confirmed verification")
		}
		if classification.Status != "mapped" || classification.Category.String == "" {
			return empty, incomplete("/expected_classification_revision_id", "Finalization requires a current mapped study classification")
		}
		if in.Kind == "BE" {
			for _, factor := range contextDimensions {
				if resolved.Factors[factor.name] == "unknown" {
					return empty, incomplete("/context_snapshot/factors/"+factor.name, "Review the unknown context factor before finalizing BE")
				}
			}
		}
		result, err := cvss.Evaluate(baseMetrics, environment)
		if err != nil {
			return empty, err
		}
		if result.Score == nil {
			return empty, incomplete("/metrics", "Complete all Base metrics and explicitly define an Environmental metric for BE")
		}
		calculator := cvss.Calculator()
		snapshot.Status, snapshot.Score, snapshot.Vector, snapshot.Calculator = "complete", result.Score, result.Vector, &calculator
	}
	if err := insertAssessment(ctx, q, snapshot, baseID, review.ID, classification.ID, contextID); err != nil {
		return empty, err
	}
	if err := s.audit(ctx, q, requestID, "assessment.append", "assessment_snapshot", snapshot.ID); err != nil {
		return empty, err
	}
	if err := tx.Commit(ctx); err != nil {
		return empty, err
	}
	return snapshot, nil
}

func insertAssessment(ctx context.Context, q *db.Queries, snapshot AssessmentSnapshot, baseID, reviewID, classificationID, contextID string) error {
	params := db.CreateAssessmentParams{
		ID: snapshot.ID, FindingID: snapshot.FindingID, TargetID: snapshot.TargetID,
		Kind: snapshot.Kind, Status: snapshot.Status, BaseID: baseID, VerificationID: reviewID,
		ClassificationID: classificationID, ContextID: contextID, Notes: snapshot.Notes,
		SchemaVersion: snapshot.SchemaVersion, PolicyVersion: snapshot.PolicyVersion, Actor: snapshot.Actor,
		CreatedAt: snapshot.CreatedAt,
	}
	objects := []struct {
		value any
		dst   *[]byte
	}{
		{snapshot.Context, &params.ContextSnapshotJson},
		{snapshot.Metrics, &params.MetricsJson},
		{snapshot.MetricRationales, &params.MetricRationalesJson},
		{snapshot.FactorDispositions, &params.FactorDispositionsJson},
	}
	for _, object := range objects {
		bytes, err := json.Marshal(object.value)
		if err != nil {
			return err
		}
		*object.dst = bytes
	}
	if snapshot.Calculator != nil {
		var err error
		params.CalculatorJson, err = json.Marshal(snapshot.Calculator)
		if err != nil {
			return err
		}
		params.Vector = pgtype.Text{String: *snapshot.Vector, Valid: true}
		if err := params.Score.Scan(fmt.Sprintf("%.1f", *snapshot.Score)); err != nil {
			return err
		}
	}
	err := q.CreateAssessment(ctx, params)
	var violation *pgconn.PgError
	if errors.As(err, &violation) && violation.Code == "23514" && violation.ConstraintName == "assessment_snapshot_size" {
		return domain.Fail(413, "limit_exceeded", "/metric_rationales", "Stored assessment fields exceed size limits")
	}
	return err
}

func (s *Service) GetAssessment(ctx context.Context, id string) (AssessmentSnapshot, error) {
	row, err := s.Store.Q.GetAssessment(ctx, id)
	if err != nil {
		return AssessmentSnapshot{}, err
	}
	return snapshotDTO(row.AssessmentSnapshot, row.Stale.Bool)
}

func (s *Service) Assessments(ctx context.Context, findingID string, limit int, raw string) (Page[AssessmentSnapshot], error) {
	page := Page[AssessmentSnapshot]{Items: []AssessmentSnapshot{}}
	if _, err := s.Store.FindingTarget(ctx, findingID); err != nil {
		return page, err
	}
	filter := Filter("assessments", findingID)
	c, after, err := s.historyCursor(raw, filter)
	if err != nil {
		return page, err
	}
	rows, err := s.Store.Q.ListAssessments(ctx, db.ListAssessmentsParams{
		FindingID: findingID, FirstPage: raw == "", AfterTime: after, AfterID: c.ID, PageLimit: int32(limit + 1),
	})
	if err != nil {
		return page, err
	}
	for i, row := range rows {
		if i == limit {
			last := page.Items[i-1]
			page.Next = s.Next(Cursor{ID: last.ID, Time: last.CreatedAt.Format(time.RFC3339Nano), Filter: filter})
			break
		}
		item, err := snapshotDTO(row.AssessmentSnapshot, row.Stale.Bool)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, item)
	}
	return page, nil
}
