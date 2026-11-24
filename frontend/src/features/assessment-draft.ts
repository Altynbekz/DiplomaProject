import type { Finding } from "../api/client";
import type { components } from "../api/schema";

export type Snapshot = components["schemas"]["AssessmentSnapshot"];
export type MetricRationale = components["schemas"]["MetricRationale"];
export type AssessmentInput = components["schemas"]["AssessmentInput"];
export type ResolvedContext = components["schemas"]["ResolvedContext"];
type Factor = keyof components["schemas"]["ContextFactors"];
export type Kind = Snapshot["kind"];

type Values = Readonly<Record<string, string>>;
export type Metric = {
  name: string;
  label: string;
  values: Values;
  inherits?: string;
};

const impact = { H: "High", L: "Low", N: "None" };

// Display names and choice codes from the FIRST CVSS v4.0 specification.
// The server validates the choices and calculates every vector and score.
export const baseMetrics: readonly Metric[] = [
  {
    name: "AV",
    label: "Attack vector",
    values: { N: "Network", A: "Adjacent", L: "Local", P: "Physical" },
  },
  { name: "AC", label: "Attack complexity", values: { L: "Low", H: "High" } },
  {
    name: "AT",
    label: "Attack requirements",
    values: { N: "None", P: "Present" },
  },
  {
    name: "PR",
    label: "Privileges required",
    values: { N: "None", L: "Low", H: "High" },
  },
  {
    name: "UI",
    label: "User interaction",
    values: { N: "None", P: "Passive", A: "Active" },
  },
  { name: "VC", label: "Vulnerable system confidentiality", values: impact },
  { name: "VI", label: "Vulnerable system integrity", values: impact },
  { name: "VA", label: "Vulnerable system availability", values: impact },
  { name: "SC", label: "Subsequent system confidentiality", values: impact },
  { name: "SI", label: "Subsequent system integrity", values: impact },
  { name: "SA", label: "Subsequent system availability", values: impact },
];

const modifiedSubsequentImpact = { ...impact, N: "Negligible" };
const modifiedSafetyImpact = { ...modifiedSubsequentImpact, S: "Safety" };
const requirements = { H: "High", M: "Medium", L: "Low" };
export const environmentalMetrics: readonly Metric[] = [
  { name: "CR", label: "Confidentiality requirement", values: requirements },
  { name: "IR", label: "Integrity requirement", values: requirements },
  { name: "AR", label: "Availability requirement", values: requirements },
  ...baseMetrics.map(({ name, label, values }) => ({
    name: "M" + name,
    label: "Modified " + label.toLowerCase(),
    inherits: name,
    values:
      name === "SC"
        ? modifiedSubsequentImpact
        : name === "SI" || name === "SA"
          ? modifiedSafetyImpact
          : values,
  })),
];

export type Choice = MetricRationale & { value: string };
export const emptyChoice: Choice = {
  value: "",
  reason: "",
  evidence_refs: [],
  context_factor_refs: [],
};

export type Draft = {
  kind: Kind;
  sourceID: string | null;
  parentID: string | null;
  reviewID: string | null;
  classificationID: string | null;
  contextIDs: string[] | null;
  choices: Record<string, Choice>;
  dispositions: Partial<Record<Factor, string>>;
  notes: string;
};

export function assessmentDraft(
  finding: Pick<Finding, "verification_review_id" | "classification">,
  { snapshot, parent }: { snapshot?: Snapshot; parent?: Snapshot } = {},
): Draft {
  const kind = snapshot?.kind ?? (parent ? "BE" : "B");
  const reference = snapshot ?? parent;
  const metrics = kind === "B" ? baseMetrics : environmentalMetrics;
  return {
    kind,
    sourceID: snapshot?.id ?? null,
    parentID: snapshot?.base_assessment_id ?? parent?.id ?? null,
    reviewID: reference
      ? reference.verification_review_id
      : finding.verification_review_id,
    classificationID: reference
      ? reference.classification_revision_id
      : finding.classification.revision_id,
    contextIDs:
      kind === "B"
        ? []
        : (snapshot?.context_snapshot.source_revision_ids ?? null),
    choices: Object.fromEntries(
      metrics.flatMap(({ name }) => {
        const value = snapshot?.metrics[name];
        const rationale = snapshot?.metric_rationales[name];
        // Stored Environmental defaults have no entry in the submitted choices.
        return value && (value !== "X" || rationale)
          ? [[name, { ...emptyChoice, ...rationale, value }]]
          : [];
      }),
    ),
    dispositions: Object.fromEntries(
      Object.entries(snapshot?.factor_dispositions ?? {}).map(([key, item]) => [
        key,
        item.reason,
      ]),
    ),
    notes: snapshot?.notes ?? "",
  };
}

export function linkedMetrics(draft: Draft, factor: Factor) {
  return environmentalMetrics
    .filter(({ name }) => {
      const choice = draft.choices[name];
      return (
        choice?.value &&
        choice.value !== "X" &&
        choice.context_factor_refs.includes(factor)
      );
    })
    .map(({ name }) => name);
}

export function assessmentInput(
  draft: Draft,
  finalize: boolean,
): AssessmentInput {
  const metrics: AssessmentInput["metrics"] = {};
  const metric_rationales: AssessmentInput["metric_rationales"] = {};
  for (const [name, { value, ...rationale }] of Object.entries(draft.choices)) {
    if (value) {
      metrics[name] = value;
      metric_rationales[name] =
        value === "X" ? { ...rationale, context_factor_refs: [] } : rationale;
    }
  }
  return {
    kind: draft.kind,
    finalize,
    base_assessment_id: draft.parentID,
    expected_verification_review_id: draft.reviewID,
    expected_classification_revision_id: draft.classificationID,
    expected_context_revision_ids: draft.contextIDs ?? [],
    metrics,
    metric_rationales,
    ...(draft.kind === "BE" && {
      factor_dispositions: Object.fromEntries(
        Object.entries(draft.dispositions).map(([key, reason]) => [
          key,
          { reason, metrics: linkedMetrics(draft, key as Factor) },
        ]),
      ),
    }),
    notes: draft.notes,
  };
}
