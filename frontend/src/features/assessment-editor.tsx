import { useEffect, useRef, type FormEvent, type ReactNode } from "react";
import type { Finding, Observation } from "../api/client";
import { contextFactorLabels, FactorsView } from "./context-panels";
import { ErrorState, Pager } from "./evidence";
import { EvidenceSelection } from "./verification-panel";
import {
  baseMetrics,
  emptyChoice,
  environmentalMetrics,
  linkedMetrics,
  type Choice,
  type Draft,
  type Metric,
  type ResolvedContext,
  type Snapshot,
} from "./assessment-draft";

export type EvidencePage = {
  observations: Observation[];
  pending: boolean;
  error: unknown;
  next: string | null;
  active: boolean;
  onNext: (cursor: string) => void;
  onFirst: () => void;
};

function MetricField({
  metric,
  choice,
  observations,
  onChange,
  defaultLabel = "Unknown / missing",
  environmental = false,
}: {
  metric: Metric;
  choice: Choice;
  observations: Observation[];
  onChange: (choice: Choice) => void;
  defaultLabel?: string;
  environmental?: boolean;
}) {
  const offPage = choice.evidence_refs.filter(
    (ref) =>
      !observations.some(
        (observation) => observation.id === ref.observation_id,
      ),
  );
  return (
    <fieldset className="stack">
      <legend>
        {metric.label} ({metric.name})
      </legend>
      <label>
        Value
        <select
          value={choice.value}
          onChange={(event) =>
            onChange({ ...choice, value: event.target.value })
          }
        >
          <option value="">{defaultLabel}</option>
          {choice.value === "X" && (
            <option value="X">Recorded Not Defined (X)</option>
          )}
          {Object.entries(metric.values).map(([code, label]) => (
            <option key={code} value={code}>
              {label} ({code})
            </option>
          ))}
        </select>
      </label>
      {choice.value && (
        <>
          <label>
            Rationale
            <textarea
              rows={2}
              maxLength={2048}
              value={choice.reason}
              onChange={(event) =>
                onChange({ ...choice, reason: event.target.value })
              }
            />
          </label>
          <details>
            <summary>
              Observation evidence · {choice.evidence_refs.length} selected
            </summary>
            <EvidenceSelection
              observations={observations}
              refs={choice.evidence_refs}
              onChange={(evidence_refs) =>
                onChange({ ...choice, evidence_refs })
              }
            />
            {offPage.map((ref) => (
              <label className="checkbox-label" key={ref.observation_id}>
                <input
                  type="checkbox"
                  checked
                  onChange={() =>
                    onChange({
                      ...choice,
                      evidence_refs: choice.evidence_refs.filter(
                        (item) => item.observation_id !== ref.observation_id,
                      ),
                    })
                  }
                />
                <span>
                  <code>{ref.source_pointer}</code>
                  <small className="location">
                    Observation <code>{ref.observation_id}</code> · another page
                  </small>
                </span>
              </label>
            ))}
          </details>
          {environmental && choice.value !== "X" && (
            <fieldset className="stack">
              <legend>Context factors used</legend>
              {contextFactorLabels.map(({ key, label }) => (
                <label key={key} className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={choice.context_factor_refs.includes(key)}
                    onChange={(event) =>
                      onChange({
                        ...choice,
                        context_factor_refs: event.target.checked
                          ? [...choice.context_factor_refs, key]
                          : choice.context_factor_refs.filter(
                              (factor) => factor !== key,
                            ),
                      })
                    }
                  />
                  {label}
                </label>
              ))}
            </fieldset>
          )}
        </>
      )}
    </fieldset>
  );
}

function EnvironmentalInputs({
  draft,
  context,
  parent,
  onChange,
}: {
  draft: Draft;
  context: ResolvedContext | undefined;
  parent: Snapshot | undefined;
  onChange: (draft: Draft) => void;
}) {
  const contextID = context?.source_revision_ids[0] ?? null;
  const draftContextID = draft.contextIDs?.[0] ?? null;
  return (
    <>
      <div className="stack">
        <strong>Selected Base parent</strong>
        <code>{draft.parentID}</code>
        {parent?.vector && <code>{parent.vector}</code>}
        {parent?.score != null && (
          <span>CVSS-B score {parent.score.toFixed(1)}</span>
        )}
        {parent && (
          <details>
            <summary>Inherited Base values</summary>
            <dl>
              {baseMetrics.map(({ name, label, values }) => (
                <div className="stack" key={name}>
                  <dt>
                    {label} ({name})
                  </dt>
                  <dd>
                    {values[parent.metrics[name]]} ({parent.metrics[name]})
                  </dd>
                </div>
              ))}
            </dl>
          </details>
        )}
      </div>
      {context && (
        <div className="stack">
          <strong>Resolved context</strong>
          <p className="location">
            Current revision <code>{contextID ?? "None"}</code>
          </p>
          <details>
            <summary>Context values and provenance</summary>
            <FactorsView
              factors={context.factors}
              provenance={context.factor_provenance}
            />
          </details>
          {(draft.contextIDs === null || contextID !== draftContextID) && (
            <p className="location">
              Draft context{" "}
              <code>
                {draft.contextIDs === null
                  ? "Awaiting review"
                  : (draftContextID ?? "None")}
              </code>
            </p>
          )}
          {draft.contextIDs === null && (
            <div>
              <button
                type="button"
                onClick={() =>
                  onChange({
                    ...draft,
                    contextIDs: context.source_revision_ids,
                  })
                }
              >
                Use displayed context
              </button>
            </div>
          )}
        </div>
      )}
      <fieldset className="stack">
        <legend>Context factor dispositions</legend>
        <p>
          Explain each factor. When a factor justifies a metric choice, select
          it under that metric.
        </p>
        {contextFactorLabels.map(({ key, label }) => {
          const affected = linkedMetrics(draft, key);
          return (
            <label key={key}>
              {label} disposition
              <small>
                {affected.length
                  ? "Linked metrics · " + affected.join(", ")
                  : "No metric change linked"}
              </small>
              <textarea
                rows={2}
                maxLength={2048}
                value={draft.dispositions[key] ?? ""}
                onChange={(event) =>
                  onChange({
                    ...draft,
                    dispositions: {
                      ...draft.dispositions,
                      [key]: event.target.value,
                    },
                  })
                }
              />
            </label>
          );
        })}
      </fieldset>
    </>
  );
}

export function AssessmentEditor({
  finding,
  draft,
  evidence,
  pending,
  onChange,
  onSubmit,
  onDiscard,
  feedback,
  parent,
  context,
  contextFeedback,
}: {
  finding: Pick<
    Finding,
    "verification" | "verification_review_id" | "classification"
  >;
  draft: Draft;
  evidence: EvidencePage;
  pending: boolean;
  onChange: (draft: Draft) => void;
  onSubmit: (finalize: boolean) => void;
  onDiscard: () => void;
  feedback: ReactNode;
  parent?: Snapshot;
  context?: ResolvedContext;
  contextFeedback?: ReactNode;
}) {
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    heading.current?.focus();
  }, []);

  const environmental = draft.kind === "BE";
  const metrics = environmental ? environmentalMetrics : baseMetrics;
  const missingValues = metrics
    .filter(({ name }) => !draft.choices[name]?.value)
    .map(({ name }) => name);
  const missingSupport = metrics
    .filter(({ name }) => {
      const choice = draft.choices[name];
      return (
        choice?.value &&
        (!choice.reason.trim() ||
          (choice.value !== "X" &&
            choice.evidence_refs.length + choice.context_factor_refs.length ===
              0))
      );
    })
    .map(({ name }) => name);
  const unknown = contextFactorLabels.filter(
    ({ key }) => context?.factors[key] === "unknown",
  );
  const unresolved = contextFactorLabels.filter(
    ({ key }) => !draft.dispositions[key]?.trim(),
  );
  const canSubmit = !environmental || (draft.contextIDs !== null && !!parent);
  const reviewChanged = draft.reviewID !== finding.verification_review_id;
  const classificationChanged =
    draft.classificationID !== finding.classification.revision_id;
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSubmit || pending) return;
    const button = (event.nativeEvent as SubmitEvent).submitter;
    onSubmit(
      button instanceof HTMLButtonElement && button.value === "finalize",
    );
  }
  return (
    <section className="panel stack" aria-label="CVSS assessment editor">
      <h3 ref={heading} tabIndex={-1}>
        CVSS-{draft.kind} editor
      </h3>
      {draft.sourceID && (
        <p className="location">
          Revising <code>{draft.sourceID}</code> as a new snapshot.
        </p>
      )}
      <p>
        {environmental
          ? "The server takes Base values and evidence from the selected parent. Define Environmental choices only when you can justify them."
          : "Drafts save selected metrics and notes. Unknown Base metrics stay missing. Finalization requires rationale and observation evidence for every Base metric."}
      </p>
      <div className="notice stack">
        <strong>Before finalizing</strong>
        <ul>
          {finding.verification !== "confirmed" && (
            <li>
              Verification must be confirmed. Current state is{" "}
              {finding.verification}.
            </li>
          )}
          {finding.classification.status !== "mapped" && (
            <li>
              A mapped study classification is required. Current state is{" "}
              {finding.classification.status}.
            </li>
          )}
          {!environmental && missingValues.length > 0 && (
            <li>Missing Base values · {missingValues.join(", ")}</li>
          )}
          {environmental &&
            !metrics.some(({ name }) => {
              const value = draft.choices[name]?.value;
              return value && value !== "X";
            }) && <li>Define at least one Environmental metric.</li>}
          {missingSupport.length > 0 && (
            <li>Review rationale and support · {missingSupport.join(", ")}</li>
          )}
          {environmental && unknown.length > 0 && (
            <li>
              Unknown context factors ·{" "}
              {unknown.map(({ label }) => label).join(", ")}
            </li>
          )}
          {environmental && unresolved.length > 0 && (
            <li>
              Disposition explanations needed ·{" "}
              {unresolved.map(({ label }) => label).join(", ")}
            </li>
          )}
          {environmental && draft.contextIDs === null && (
            <li>Review the displayed context before saving.</li>
          )}
          {(reviewChanged || classificationChanged) && (
            <li>
              Verification or classification has changed since you started this
              draft.
            </li>
          )}
          {environmental &&
            draft.contextIDs !== null &&
            context &&
            (draft.contextIDs[0] ?? null) !==
              (context.source_revision_ids[0] ?? null) && (
              <li>Resolved context has changed since you selected it.</li>
            )}
          {environmental &&
            parent &&
            (parent.verification_review_id !== finding.verification_review_id ||
              parent.classification_revision_id !==
                finding.classification.revision_id) && (
              <li>
                The parent uses earlier review or classification revisions.
                Select a completed Base for the current revisions.
              </li>
            )}
          <li>
            The server checks current revisions and evidence when you finalize.
          </li>
        </ul>
        {(reviewChanged || classificationChanged) && (
          <details>
            <summary>Review and classification revisions</summary>
            <dl>
              <dt>Draft verification</dt>
              <dd>
                <code>{draft.reviewID ?? "None"}</code>
              </dd>
              <dt>Current verification</dt>
              <dd>
                {finding.verification} ·{" "}
                <code>{finding.verification_review_id ?? "None"}</code>
              </dd>
              <dt>Draft classification</dt>
              <dd>
                <code>{draft.classificationID ?? "None"}</code>
              </dd>
              <dt>Current classification</dt>
              <dd>
                {finding.classification.category ??
                  finding.classification.status}{" "}
                · <code>{finding.classification.revision_id ?? "None"}</code>
              </dd>
            </dl>
          </details>
        )}
      </div>
      <form className="stack" onSubmit={submit}>
        <fieldset
          className="stack"
          disabled={pending}
          aria-label="Assessment fields"
        >
          {environmental && (
            <>
              <EnvironmentalInputs
                draft={draft}
                context={context}
                parent={parent}
                onChange={onChange}
              />
              {contextFeedback}
            </>
          )}
          <details>
            <summary>Observation pages for metric evidence</summary>
            {evidence.pending && <p role="status">Loading observations…</p>}
            {!!evidence.error && <ErrorState error={evidence.error} />}
            <p>
              Evidence choices use the current observation page. Selections are
              retained when you change pages.
            </p>
            <Pager
              next={evidence.next}
              active={evidence.active}
              onNext={evidence.onNext}
              onFirst={evidence.onFirst}
            />
          </details>
          {metrics.map((metric) => (
            <MetricField
              key={metric.name}
              metric={metric}
              choice={draft.choices[metric.name] ?? emptyChoice}
              observations={evidence.observations}
              environmental={environmental}
              defaultLabel={
                environmental
                  ? metric.inherits
                    ? "Not Defined (X) · inherits " +
                      (parent?.metrics[metric.inherits]
                        ? metric.inherits +
                          ":" +
                          parent.metrics[metric.inherits]
                        : metric.inherits)
                    : "Not Defined (X) · CVSS default High"
                  : undefined
              }
              onChange={(choice) =>
                onChange({
                  ...draft,
                  choices: { ...draft.choices, [metric.name]: choice },
                })
              }
            />
          ))}
          <label>
            Assessment notes
            <textarea
              rows={3}
              maxLength={4096}
              value={draft.notes}
              onChange={(event) =>
                onChange({ ...draft, notes: event.target.value })
              }
            />
          </label>
          <div className="actions">
            <button type="submit" value="draft" disabled={!canSubmit}>
              Save incomplete assessment
            </button>
            <button
              type="submit"
              className="primary"
              value="finalize"
              disabled={!canSubmit}
            >
              Finalize CVSS-{draft.kind}
            </button>
            <button type="button" onClick={onDiscard}>
              Discard editor changes
            </button>
          </div>
        </fieldset>
      </form>
      {pending && <p role="status">Updating assessment inputs…</p>}
      {feedback}
    </section>
  );
}
