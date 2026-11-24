import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import type { Finding } from "../src/api/client";
import { AssessmentEditor } from "../src/features/assessment-editor";
import {
  assessmentDraft,
  assessmentInput,
  emptyChoice,
  baseMetrics,
  environmentalMetrics,
  type Snapshot,
} from "../src/features/assessment-draft";

afterEach(cleanup);

const finding = {
  verification: "confirmed",
  verification_review_id: "current-synthetic-review",
  classification: {
    revision_id: "current-synthetic-classification",
    status: "mapped",
    category: "A02:2025",
    candidate_categories: ["A02:2025"],
    mapping_version: "manual-v1",
  },
} satisfies Pick<
  Finding,
  "verification" | "verification_review_id" | "classification"
>;
const ref = {
  observation_id: "synthetic-observation",
  source_pointer: "/site/0/alerts/0/instances/0",
  note: "Synthetic test reference",
};

test("a Base draft sends only selected values, matching rationale and pinned revisions", () => {
  const draft = assessmentDraft(finding);
  draft.choices.AV = {
    ...emptyChoice,
    value: "N",
    reason: "Synthetic rationale",
    evidence_refs: [ref],
  };
  draft.choices.AC = {
    ...emptyChoice,
    reason: "Unsaved notes for a missing choice",
  };
  draft.notes = "Complexity still unknown";
  expect(assessmentInput(draft, false)).toEqual({
    kind: "B",
    finalize: false,
    base_assessment_id: null,
    expected_verification_review_id: finding.verification_review_id,
    expected_classification_revision_id: finding.classification.revision_id,
    expected_context_revision_ids: [],
    metrics: { AV: "N" },
    metric_rationales: {
      AV: {
        reason: "Synthetic rationale",
        evidence_refs: [ref],
        context_factor_refs: [],
      },
    },
    notes: "Complexity still unknown",
  });
});

test("revising BE omits inherited Base and implicit defaults and keeps factor links reciprocal", () => {
  const note = { rationale: "Synthetic context note", sources: [] };
  const snapshot: Snapshot = {
    id: "prior-synthetic-be",
    finding_id: "synthetic-finding",
    target_id: "synthetic-target",
    kind: "BE",
    status: "incomplete",
    score: null,
    vector: null,
    stale: true,
    base_assessment_id: "exact-synthetic-base",
    verification_review_id: null,
    classification_revision_id: "prior-synthetic-classification",
    context_snapshot: {
      factors: {
        exposure: "unknown",
        required_privileges: "unknown",
        data_sensitivity: "unknown",
        business_criticality: "unknown",
      },
      factor_provenance: {
        exposure: note,
        required_privileges: note,
        data_sensitivity: note,
        business_criticality: note,
      },
      source_revision_ids: ["prior-synthetic-context"],
      resolver_version: "context-resolver/v1",
      resolved_at: "2026-11-24T00:00:00Z",
    },
    metrics: { AV: "N", CR: "X", MAV: "X", MAC: "H" },
    metric_rationales: {
      AV: {
        reason: "Inherited Base",
        evidence_refs: [ref],
        context_factor_refs: [],
      },
      CR: {
        reason: "Explicitly recorded default",
        evidence_refs: [],
        context_factor_refs: [],
      },
      MAC: {
        reason: "Synthetic Environmental review",
        evidence_refs: [],
        context_factor_refs: ["exposure"],
      },
    },
    factor_dispositions: {
      exposure: { reason: "Synthetic choice", metrics: ["MAC"] },
      data_sensitivity: { reason: "No change justified", metrics: [] },
    },
    notes: "Stored note",
    calculator: null,
    schema_version: "assessment-snapshot/v1",
    policy_version: "context-policy/v1",
    actor: "synthetic-operator",
    created_at: "2026-11-24T00:00:00Z",
  };
  const draft = assessmentDraft(finding, { snapshot });
  expect(assessmentInput(draft, false)).toEqual({
    kind: "BE",
    finalize: false,
    base_assessment_id: snapshot.base_assessment_id,
    expected_verification_review_id: snapshot.verification_review_id,
    expected_classification_revision_id: snapshot.classification_revision_id,
    expected_context_revision_ids:
      snapshot.context_snapshot.source_revision_ids,
    metrics: { CR: "X", MAC: "H" },
    metric_rationales: {
      CR: snapshot.metric_rationales.CR,
      MAC: snapshot.metric_rationales.MAC,
    },
    factor_dispositions: snapshot.factor_dispositions,
    notes: "Stored note",
  });
  draft.choices.MAC = { ...draft.choices.MAC, value: "" };
  expect(
    assessmentInput(draft, false).factor_dispositions?.exposure.metrics,
  ).toEqual([]);
  expect(snapshot.metrics.MAC).toBe("H");
  expect(snapshot.factor_dispositions.exposure.metrics).toEqual(["MAC"]);
});

test("off-page selections with the same pointer identify separate observations", () => {
  const repeated = { ...ref, observation_id: "second-synthetic-observation" };
  const draft = assessmentDraft(finding);
  draft.choices.AV = {
    ...emptyChoice,
    value: "N",
    evidence_refs: [ref, repeated],
  };
  const onChange = vi.fn();
  render(
    <AssessmentEditor
      finding={finding}
      draft={draft}
      evidence={{
        observations: [],
        pending: false,
        error: null,
        next: null,
        active: true,
        onNext: vi.fn(),
        onFirst: vi.fn(),
      }}
      pending={false}
      onChange={onChange}
      onSubmit={vi.fn()}
      onDiscard={vi.fn()}
      feedback={null}
    />,
  );
  fireEvent.click(screen.getByText("Observation evidence · 2 selected"));
  fireEvent.click(screen.getByLabelText(/second-synthetic-observation/));
  expect(onChange).toHaveBeenCalledWith({
    ...draft,
    choices: { AV: { ...draft.choices.AV, evidence_refs: [ref] } },
  });
});

test("modified subsequent choices use FIRST labels and keep Safety scoped", () => {
  for (const name of ["SC", "SI", "SA"]) {
    expect(baseMetrics.find((metric) => metric.name === name)?.values.N).toBe(
      "None",
    );
    const modified = environmentalMetrics.find(
      (metric) => metric.name === "M" + name,
    );
    expect(modified?.values.N).toBe("Negligible");
  }
  expect(
    environmentalMetrics
      .filter((metric) => "S" in metric.values)
      .map(({ name }) => name),
  ).toEqual(["MSI", "MSA"]);
});
