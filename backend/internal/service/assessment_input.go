package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"diploma/internal/cvss"
	"diploma/internal/domain"
	db "diploma/internal/store/sqlc"
)

type MetricRationale struct {
	Reason            string        `json:"reason"`
	EvidenceRefs      []EvidenceRef `json:"evidence_refs"`
	ContextFactorRefs []string      `json:"context_factor_refs"`
}

type FactorDisposition struct {
	Reason  string   `json:"reason"`
	Metrics []string `json:"metrics"`
}

type FactorDispositions map[string]FactorDisposition

func (d *FactorDispositions) UnmarshalJSON(raw []byte) error {
	if len(raw) == 0 || raw[0] != '{' {
		return errors.New("expected a factor disposition object")
	}
	type wire FactorDispositions
	return json.Unmarshal(raw, (*wire)(d))
}

func requireJSONFields(raw []byte, fields map[string]byte) error {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return err
	}
	if len(values) != len(fields) {
		return errors.New("unexpected object fields")
	}
	for key, prefix := range fields {
		if value := values[key]; len(value) == 0 || value[0] != prefix {
			return errors.New("missing or invalid object field")
		}
	}
	return nil
}

func (r *MetricRationale) UnmarshalJSON(raw []byte) error {
	if err := requireJSONFields(raw, map[string]byte{"reason": '"', "evidence_refs": '[', "context_factor_refs": '['}); err != nil {
		return err
	}
	type wire MetricRationale
	return json.Unmarshal(raw, (*wire)(r))
}

func (d *FactorDisposition) UnmarshalJSON(raw []byte) error {
	if err := requireJSONFields(raw, map[string]byte{"reason": '"', "metrics": '['}); err != nil {
		return err
	}
	type wire FactorDisposition
	return json.Unmarshal(raw, (*wire)(d))
}

type AssessmentInput struct {
	Kind                     string                     `json:"kind"`
	Finalize                 *bool                      `json:"finalize"`
	BaseAssessmentID         json.RawMessage            `json:"base_assessment_id"`
	ExpectedVerificationID   json.RawMessage            `json:"expected_verification_review_id"`
	ExpectedClassificationID json.RawMessage            `json:"expected_classification_revision_id"`
	ExpectedContextIDs       []string                   `json:"expected_context_revision_ids"`
	Metrics                  map[string]string          `json:"metrics"`
	MetricRationales         map[string]MetricRationale `json:"metric_rationales"`
	FactorDispositions       FactorDispositions         `json:"factor_dispositions"`
	Notes                    *string                    `json:"notes"`
}

func parseAssessmentID(raw json.RawMessage, pointer string) (string, error) {
	var value *string
	if json.Unmarshal(raw, &value) != nil || (value != nil && !validUUID(*value)) {
		return "", domain.Fail(400, "invalid_request", pointer, "Expected a canonical UUID or explicit null")
	}
	if value == nil {
		return "", nil
	}
	return *value, nil
}

func validation(pointer, message string) error {
	return domain.Fail(422, "validation_failed", pointer, message)
}

func incomplete(pointer, message string) error {
	return domain.Fail(422, "incomplete_assessment", pointer, message)
}

func assessmentConflict(pointer, message string) error {
	return domain.Fail(409, "conflict", pointer, message)
}

func contextFactor(name string) bool {
	for _, factor := range contextDimensions {
		if factor.name == name {
			return true
		}
	}
	return false
}

func validateAssessmentInput(in AssessmentInput) error {
	if in.Kind != "B" && in.Kind != "BE" {
		return validation("/kind", "Select B or BE")
	}
	if in.Finalize == nil || in.Notes == nil || in.Metrics == nil || in.MetricRationales == nil || in.ExpectedContextIDs == nil {
		return domain.Fail(400, "invalid_request", "", "Supply finalize, notes, metrics, metric rationales and expected context revisions")
	}
	if err := boundedText(*in.Notes, "/notes", 0, 4096); err != nil {
		return err
	}
	if len(in.ExpectedContextIDs) > 1 || (len(in.ExpectedContextIDs) == 1 && !validUUID(in.ExpectedContextIDs[0])) {
		return validation("/expected_context_revision_ids", "Expected zero or one canonical revision UUID")
	}
	if in.Kind == "B" && (len(in.ExpectedContextIDs) != 0 || len(in.FactorDispositions) != 0) {
		return validation("/kind", "B accepts neither expected context revisions nor factor dispositions")
	}
	if err := cvss.Validate(in.Metrics, in.Kind == "BE"); err != nil {
		var metric *cvss.MetricError
		if errors.As(err, &metric) {
			pointer := "/metrics"
			if len(metric.Name) <= 4 {
				pointer += "/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(metric.Name)
			}
			return validation(pointer, metric.Error())
		}
		return err
	}
	if len(in.Metrics) != len(in.MetricRationales) {
		return validation("/metric_rationales", "Provide one rationale entry per supplied metric")
	}
	for _, name := range slices.Sorted(maps.Keys(in.Metrics)) {
		value := in.Metrics[name]
		rationale, found := in.MetricRationales[name]
		pointer := "/metric_rationales/" + name
		if !found || rationale.EvidenceRefs == nil || rationale.ContextFactorRefs == nil {
			return validation(pointer, "Supply a rationale and evidence and context reference arrays")
		}
		if err := boundedText(rationale.Reason, pointer+"/reason", 0, 2048); err != nil {
			return err
		}
		if *in.Finalize && strings.TrimSpace(rationale.Reason) == "" {
			return incomplete(pointer+"/reason", "Finalization requires a metric rationale")
		}
		if len(rationale.EvidenceRefs) > 20 || len(rationale.ContextFactorRefs) > len(contextDimensions) {
			return validation(pointer, "Too many evidence or context references")
		}
		if len(rationale.ContextFactorRefs) != 0 && (in.Kind == "B" || value == "X") {
			return validation(pointer+"/context_factor_refs", "Only defined Environmental choices can use context factors")
		}
		seen := map[string]bool{}
		for _, factor := range rationale.ContextFactorRefs {
			if !contextFactor(factor) || seen[factor] {
				return validation(pointer+"/context_factor_refs", "Use distinct recognized context factors")
			}
			seen[factor] = true
		}
		if *in.Finalize && value != "X" && len(rationale.EvidenceRefs)+len(rationale.ContextFactorRefs) == 0 {
			return incomplete(pointer, "A defined metric requires observation evidence or a reviewed context factor")
		}
	}
	for _, factor := range slices.Sorted(maps.Keys(in.FactorDispositions)) {
		disposition := in.FactorDispositions[factor]
		if !contextFactor(factor) {
			return validation("/factor_dispositions", "Unrecognized context factor")
		}
		pointer := "/factor_dispositions/" + factor
		if disposition.Metrics == nil || len(disposition.Metrics) > 14 {
			return validation(pointer+"/metrics", "Supply a bounded list of defined Environmental choices")
		}
		if err := boundedText(disposition.Reason, pointer+"/reason", 0, 2048); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, name := range disposition.Metrics {
			if value, found := in.Metrics[name]; !found || value == "X" || seen[name] ||
				!slices.Contains(in.MetricRationales[name].ContextFactorRefs, factor) {
				return validation(pointer+"/metrics", "Use distinct defined metrics that reference this factor")
			}
			seen[name] = true
		}
	}
	if *in.Finalize && in.Kind == "BE" {
		for _, factor := range contextDimensions {
			if strings.TrimSpace(in.FactorDispositions[factor.name].Reason) == "" {
				return incomplete("/factor_dispositions/"+factor.name, "Document this factor's disposition")
			}
		}
		for _, name := range slices.Sorted(maps.Keys(in.MetricRationales)) {
			rationale := in.MetricRationales[name]
			for _, factor := range rationale.ContextFactorRefs {
				if !slices.Contains(in.FactorDispositions[factor].Metrics, name) {
					return incomplete("/factor_dispositions/"+factor, "Record the same factor-to-metric reference in the disposition")
				}
			}
		}
	}
	return nil
}

func validateAssessmentEvidence(ctx context.Context, q *db.Queries, in AssessmentInput, findingID string) error {
	for _, name := range slices.Sorted(maps.Keys(in.MetricRationales)) {
		rationale := in.MetricRationales[name]
		seen := map[string]bool{}
		for i, ref := range rationale.EvidenceRefs {
			pointer := fmt.Sprintf("/metric_rationales/%s/evidence_refs/%d", name, i)
			if seen[ref.ObservationID] {
				return validation(pointer+"/observation_id", "Observation already referenced by this metric")
			}
			if err := validateObservationRef(ctx, q, ref, findingID, "", pointer); err != nil {
				return err
			}
			seen[ref.ObservationID] = true
		}
	}
	return nil
}
