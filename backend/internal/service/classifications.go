package service

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"time"

	"diploma/internal/classify"
	"diploma/internal/domain"
	db "diploma/internal/store/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const manualMappingVersion = "manual-v1"

type ClassificationInput struct {
	Status             string          `json:"status"`
	Category           json.RawMessage `json:"category"`
	Candidates         []string        `json:"candidate_categories"`
	Rationale          string          `json:"rationale"`
	EvidenceRefs       []EvidenceRef   `json:"evidence_refs"`
	ExpectedPreviousID json.RawMessage `json:"expected_previous_id"`
}

type ClassificationRevision struct {
	ID           string          `json:"id"`
	FindingID    string          `json:"finding_id"`
	Status       string          `json:"status"`
	Category     *string         `json:"category"`
	Candidates   []string        `json:"candidate_categories"`
	Version      string          `json:"mapping_version"`
	Rules        []classify.Rule `json:"rule_refs"`
	Rationale    string          `json:"rationale"`
	EvidenceRefs []EvidenceRef   `json:"evidence_refs"`
	Actor        string          `json:"actor"`
	CreatedAt    time.Time       `json:"created_at"`
}

func validateClassification(in ClassificationInput, selected *string) error {
	bad := func(pointer, message string) error {
		return domain.Fail(422, "validation_failed", pointer, message)
	}
	if in.Candidates == nil || len(in.Candidates) > 3 {
		return bad("/candidate_categories", "Supply zero to three distinct study categories")
	}
	seen := map[string]bool{}
	for _, category := range in.Candidates {
		if !classify.ValidCategory(category) || seen[category] {
			return bad("/candidate_categories", "Use distinct study categories")
		}
		seen[category] = true
	}
	switch in.Status {
	case "mapped":
		if selected == nil || len(in.Candidates) != 1 || in.Candidates[0] != *selected {
			return bad("/category", "Mapped status needs one matching study category")
		}
	case "ambiguous":
		if selected != nil || len(in.Candidates) < 2 {
			return bad("/candidate_categories", "An ambiguous manual revision requires competing categories")
		}
	case "unmapped", "out_of_scope":
		if selected != nil || len(in.Candidates) != 0 {
			return bad("/category", "This status has no categories")
		}
	default:
		return bad("/status", "Select a classification status")
	}
	return boundedText(in.Rationale, "/rationale", 1, 4096)
}

func (s *Service) appendClassification(ctx context.Context, q *db.Queries, revision ClassificationRevision, requestID string) error {
	refs, err := json.Marshal(revision.EvidenceRefs)
	if err != nil {
		return err
	}
	rules, err := json.Marshal(revision.Rules)
	if err != nil {
		return err
	}
	var category pgtype.Text
	if revision.Category != nil {
		category = pgtype.Text{Valid: true, String: *revision.Category}
	}
	if err := q.CreateClassification(ctx, db.CreateClassificationParams{
		ID: revision.ID, FindingID: revision.FindingID, Status: revision.Status,
		Category: category, CandidateCategories: revision.Candidates, MappingVersion: revision.Version,
		Rationale: revision.Rationale, EvidenceRefsJson: refs, RuleRefsJson: rules,
		Actor: revision.Actor, CreatedAt: revision.CreatedAt,
	}); err != nil {
		return err
	}
	return s.audit(ctx, q, requestID, "classification.append", "classification_revision", revision.ID)
}

func (s *Service) CreateClassification(ctx context.Context, findingID string, input ClassificationInput, requestID string) (ClassificationRevision, error) {
	predecessor, err := expectedRevision(input.ExpectedPreviousID)
	if err != nil {
		return ClassificationRevision{}, err
	}
	var category *string
	if json.Unmarshal(input.Category, &category) != nil {
		return ClassificationRevision{}, domain.Fail(400, "invalid_request", "/category", "Supply a study category or explicit null")
	}
	if err := validateClassification(input, category); err != nil {
		return ClassificationRevision{}, err
	}
	tx, q, err := s.Store.Begin(ctx)
	if err != nil {
		return ClassificationRevision{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := q.LockFindingRevision(ctx, findingID); err != nil {
		return ClassificationRevision{}, err
	}
	head, err := q.LatestClassification(ctx, findingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ClassificationRevision{}, err
	}
	if head.ID != predecessor {
		return ClassificationRevision{}, revisionConflict()
	}
	if err := validateEvidenceRefs(ctx, q, input.EvidenceRefs, findingID); err != nil {
		return ClassificationRevision{}, err
	}
	slices.Sort(input.Candidates)
	revision := ClassificationRevision{
		ID: uuid.NewString(), FindingID: findingID, Status: input.Status, Category: category,
		Candidates: input.Candidates, Version: manualMappingVersion, Rules: []classify.Rule{},
		Rationale: input.Rationale, EvidenceRefs: input.EvidenceRefs,
		Actor: s.Actor, CreatedAt: revisionTime(s.Now(), head.CreatedAt),
	}
	if err := s.appendClassification(ctx, q, revision, requestID); err != nil {
		return ClassificationRevision{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ClassificationRevision{}, err
	}
	return revision, nil
}

func (s *Service) Classifications(ctx context.Context, findingID string, limit int, raw string) (Page[ClassificationRevision], error) {
	page := Page[ClassificationRevision]{Items: []ClassificationRevision{}}
	if _, err := s.Store.FindingTarget(ctx, findingID); err != nil {
		return page, err
	}
	filter := Filter("classifications", findingID)
	cursor, when, err := s.historyCursor(raw, filter)
	if err != nil {
		return page, err
	}
	rows, err := s.Store.Q.ListClassifications(ctx, db.ListClassificationsParams{
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
		var rules []classify.Rule
		if err := json.Unmarshal(row.RuleRefsJson, &rules); err != nil {
			return page, err
		}
		page.Items = append(page.Items, ClassificationRevision{
			ID: row.ID, FindingID: row.FindingID, Status: row.Status, Category: optionalString(row.Category.String),
			Candidates: row.CandidateCategories, Version: row.MappingVersion, Rules: rules,
			Rationale: row.Rationale, EvidenceRefs: refs, Actor: row.Actor, CreatedAt: row.CreatedAt.UTC(),
		})
	}
	return page, nil
}

// Import already holds the target lock. Each finding lock serializes the rule
// append with operator decisions and prevents an import superseding a manual head.
func (s *Service) classifyImported(ctx context.Context, q *db.Queries, eligible map[string]string, requestID string) error {
	for _, findingID := range slices.Sorted(maps.Keys(eligible)) {
		if _, err := q.LockFindingRevision(ctx, findingID); err != nil {
			return err
		}
		head, err := q.LatestClassification(ctx, findingID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if head.MappingVersion == manualMappingVersion {
			continue
		}
		observations, err := q.ClassificationEvidence(ctx, findingID)
		if err != nil {
			return err
		}
		cwes := []int32{}
		for _, observation := range observations {
			cwes = append(cwes, observation.CweIds...)
		}
		plugin := eligible[findingID]
		decision := s.Mapping.Match(plugin, cwes)
		if decision.Status == "unmapped" {
			continue
		}
		if head.MappingVersion == s.Mapping.Version {
			var previousRules []classify.Rule
			if err := json.Unmarshal(head.RuleRefsJson, &previousRules); err != nil {
				return err
			}
			if reflect.DeepEqual(previousRules, decision.Rules) {
				continue
			}
		}
		refs := []EvidenceRef{}
		selected := map[string]bool{}
		for _, rule := range decision.Rules {
			for _, observation := range observations {
				if rule.Matches(plugin, observation.CweIds) {
					if !selected[observation.ID] {
						refs = append(refs, EvidenceRef{ObservationID: observation.ID, SourcePointer: observation.SourcePointer, Note: "Source observation for the recorded mapping rule"})
						selected[observation.ID] = true
					}
					break
				}
			}
		}
		revision := ClassificationRevision{
			ID: uuid.NewString(), FindingID: findingID, Status: decision.Status, Category: decision.Category,
			Candidates: decision.Candidates, Version: s.Mapping.Version, Rules: decision.Rules,
			Rationale: decision.Rationale, EvidenceRefs: refs, Actor: "mapping:" + s.Mapping.Version,
			CreatedAt: revisionTime(s.Now(), head.CreatedAt),
		}
		if err := s.appendClassification(ctx, q, revision, requestID); err != nil {
			return err
		}
	}
	return nil
}
