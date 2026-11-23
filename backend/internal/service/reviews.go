package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"diploma/internal/domain"
	db "diploma/internal/store/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EvidenceRef struct {
	ObservationID string `json:"observation_id"`
	SourcePointer string `json:"source_pointer"`
	Note          string `json:"note"`
}

func (ref *EvidenceRef) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if len(fields) != 3 {
		return errors.New("expected the three evidence reference fields")
	}
	for _, key := range []string{"observation_id", "source_pointer", "note"} {
		if value := fields[key]; len(value) == 0 || value[0] != '"' {
			return errors.New("invalid evidence reference field")
		}
	}
	type wire EvidenceRef
	return json.Unmarshal(raw, (*wire)(ref))
}

type VerificationInput struct {
	State              string          `json:"state"`
	Procedure          string          `json:"procedure"`
	Rationale          string          `json:"rationale"`
	EvidenceRefs       []EvidenceRef   `json:"evidence_refs"`
	ExpectedPreviousID json.RawMessage `json:"expected_previous_id"`
}

type Verification struct {
	ID           string        `json:"id"`
	FindingID    string        `json:"finding_id"`
	State        string        `json:"state"`
	Procedure    string        `json:"procedure"`
	Rationale    string        `json:"rationale"`
	EvidenceRefs []EvidenceRef `json:"evidence_refs"`
	Actor        string        `json:"actor"`
	CreatedAt    time.Time     `json:"created_at"`
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func reviewState(value string) bool {
	switch value {
	case "unreviewed", "confirmed", "false_positive", "inconclusive":
		return true
	default:
		return false
	}
}

func validUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id.String() == value
}

func expectedRevision(raw json.RawMessage) (string, error) {
	var id *string
	if len(raw) == 0 {
		return "", revisionConflict()
	}
	if json.Unmarshal(raw, &id) != nil || id != nil && !validUUID(*id) {
		return "", domain.Fail(400, "invalid_request", "/expected_previous_id", "Expected a canonical UUID or null")
	}
	if id == nil {
		return "", nil
	}
	return *id, nil
}

func revisionConflict() error {
	return domain.Fail(409, "conflict", "/expected_previous_id", "The current revision changed. Load the latest revision before saving.")
}

func boundedText(value, pointer string, minimum, maximum int) error {
	n := utf8.RuneCountInString(value)
	if n < minimum || n > maximum || minimum > 0 && strings.TrimSpace(value) == "" {
		return domain.Fail(422, "validation_failed", pointer, fmt.Sprintf("Enter text with %d to %d characters", minimum, maximum))
	}
	return domain.ValidateStorageText(value, pointer)
}

func validateObservationRef(ctx context.Context, q *db.Queries, ref EvidenceRef, findingID, targetID, pointer string) error {
	if !validUUID(ref.ObservationID) {
		return domain.Fail(422, "validation_failed", pointer+"/observation_id", "Expected an observation UUID")
	}
	if err := boundedText(ref.SourcePointer, pointer+"/source_pointer", 1, 256); err != nil {
		return err
	}
	if err := boundedText(ref.Note, pointer+"/note", 0, 512); err != nil {
		return err
	}
	observation, err := q.EvidenceObservation(ctx, ref.ObservationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Fail(422, "validation_failed", pointer+"/observation_id", "Referenced observation is unavailable")
	}
	if err != nil {
		return err
	}
	if findingID != "" && observation.FindingID != findingID || targetID != "" && observation.TargetID != targetID {
		return domain.Fail(422, "validation_failed", pointer+"/observation_id", "Observation does not belong to this resource")
	}
	if observation.SourcePointer != ref.SourcePointer {
		return domain.Fail(422, "validation_failed", pointer+"/source_pointer", "Source pointer does not match the observation")
	}
	return nil
}

func validateEvidenceRefs(ctx context.Context, q *db.Queries, refs []EvidenceRef, findingID string) error {
	if len(refs) < 1 || len(refs) > 20 {
		return domain.Fail(422, "validation_failed", "/evidence_refs", "Select between 1 and 20 source observations")
	}
	seen := make(map[string]bool, len(refs))
	for i, ref := range refs {
		pointer := fmt.Sprintf("/evidence_refs/%d", i)
		if seen[ref.ObservationID] {
			return domain.Fail(422, "validation_failed", pointer+"/observation_id", "Observation is already referenced")
		}
		if err := validateObservationRef(ctx, q, ref, findingID, "", pointer); err != nil {
			return err
		}
		seen[ref.ObservationID] = true
	}
	return nil
}

func revisionTime(now, previous time.Time) time.Time {
	now = now.UTC().Truncate(time.Microsecond)
	if !now.After(previous) {
		return previous.UTC().Add(time.Microsecond)
	}
	return now
}

func (s *Service) CreateVerification(ctx context.Context, findingID string, in VerificationInput, requestID string) (Verification, error) {
	expected, err := expectedRevision(in.ExpectedPreviousID)
	if err != nil {
		return Verification{}, err
	}
	if !reviewState(in.State) {
		return Verification{}, domain.Fail(422, "validation_failed", "/state", "Select a valid verification state")
	}
	if err := boundedText(in.Procedure, "/procedure", 1, 4096); err != nil {
		return Verification{}, err
	}
	if err := boundedText(in.Rationale, "/rationale", 1, 4096); err != nil {
		return Verification{}, err
	}
	tx, q, err := s.Store.Begin(ctx)
	if err != nil {
		return Verification{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := q.LockFindingRevision(ctx, findingID); err != nil {
		return Verification{}, err
	}
	head, err := q.LatestVerification(ctx, findingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Verification{}, err
	}
	if head.ID != expected {
		return Verification{}, revisionConflict()
	}
	if err := validateEvidenceRefs(ctx, q, in.EvidenceRefs, findingID); err != nil {
		return Verification{}, err
	}
	refs, err := json.Marshal(in.EvidenceRefs)
	if err != nil {
		return Verification{}, err
	}
	now := revisionTime(s.Now(), head.CreatedAt)
	review := Verification{
		ID: uuid.NewString(), FindingID: findingID, State: in.State, Procedure: in.Procedure,
		Rationale: in.Rationale, EvidenceRefs: in.EvidenceRefs, Actor: s.Actor, CreatedAt: now,
	}
	if err := q.CreateVerification(ctx, db.CreateVerificationParams{
		ID: review.ID, FindingID: findingID, State: review.State, Procedure: review.Procedure,
		Rationale: review.Rationale, EvidenceRefsJson: refs, Actor: review.Actor, CreatedAt: now,
	}); err != nil {
		return Verification{}, err
	}
	if err := s.audit(ctx, q, requestID, "verification.append", "verification_review", review.ID); err != nil {
		return Verification{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Verification{}, err
	}
	return review, nil
}

func (s *Service) historyCursor(raw, filter string) (Cursor, time.Time, error) {
	cursor, err := s.Cursor(raw, filter)
	if err != nil {
		return Cursor{}, time.Time{}, err
	}
	if cursor.ID == "" {
		return Cursor{ID: uuid.Nil.String()}, time.Time{}, nil
	}
	when, err := time.Parse(time.RFC3339Nano, cursor.Time)
	if err != nil || !validUUID(cursor.ID) {
		return Cursor{}, time.Time{}, domain.Fail(400, "invalid_request", "/cursor", "Invalid history cursor")
	}
	return cursor, when, nil
}

func (s *Service) Verifications(ctx context.Context, findingID string, limit int, raw string) (Page[Verification], error) {
	page := Page[Verification]{Items: []Verification{}}
	if _, err := s.Store.FindingTarget(ctx, findingID); err != nil {
		return page, err
	}
	filter := Filter("verifications", findingID)
	cursor, when, err := s.historyCursor(raw, filter)
	if err != nil {
		return page, err
	}
	rows, err := s.Store.Q.ListVerifications(ctx, db.ListVerificationsParams{
		FindingID: findingID, FirstPage: raw == "", AfterTime: when, AfterID: cursor.ID, PageLimit: int32(limit + 1),
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
		var refs []EvidenceRef
		if err := json.Unmarshal(row.EvidenceRefsJson, &refs); err != nil {
			return page, err
		}
		page.Items = append(page.Items, Verification{
			ID: row.ID, FindingID: row.FindingID, State: row.State, Procedure: row.Procedure,
			Rationale: row.Rationale, EvidenceRefs: refs, Actor: row.Actor, CreatedAt: row.CreatedAt.UTC(),
		})
	}
	return page, nil
}
