import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { request } from "../api/client";
import type { components } from "../api/schema";
import { FactorsView } from "./context-panels";
import { ErrorState, Pager } from "./evidence";

type Snapshot = components["schemas"]["AssessmentSnapshot"];
type History = components["schemas"]["AssessmentPage"];

function SnapshotDetails({ item }: { item: Snapshot }) {
  return (
    <details>
      <summary>Assessment evidence and provenance</summary>
      <dl>
        <dt>Snapshot ID</dt>
        <dd>
          <code>{item.id}</code>
        </dd>
        {item.base_assessment_id && (
          <>
            <dt>Base parent</dt>
            <dd>
              <code>{item.base_assessment_id}</code>
            </dd>
          </>
        )}
        <dt>Verification revision</dt>
        <dd>
          <code>{item.verification_review_id ?? "None"}</code>
        </dd>
        <dt>Classification revision</dt>
        <dd>
          <code>{item.classification_revision_id ?? "None"}</code>
        </dd>
        <dt>Policy and schema</dt>
        <dd>
          {item.policy_version} · {item.schema_version}
        </dd>
        {item.calculator && (
          <>
            <dt>Calculator</dt>
            <dd>
              {item.calculator.name} · {item.calculator.version}
            </dd>
            <dt>FIRST source revision</dt>
            <dd>
              <code>{item.calculator.source_commit}</code>
            </dd>
            <dt>Specification and parity fixture</dt>
            <dd>
              CVSS {item.calculator.specification_version}, document{" "}
              {item.calculator.specification_document_version} ·{" "}
              {item.calculator.fixture_version}
            </dd>
            <dt>Execution engine</dt>
            <dd>
              <code>{item.calculator.engine_version}</code>
            </dd>
          </>
        )}
      </dl>
      {item.notes && <p className="scanner-text">{item.notes}</p>}
      <h4>Stored metric choices</h4>
      {item.kind === "BE" && (
        <p>
          Base choices come from the recorded parent. Not Defined leaves a
          modified choice inherited or an Environmental requirement at its CVSS
          default.
        </p>
      )}
      {Object.keys(item.metrics).length === 0 && (
        <p>No metric choices saved.</p>
      )}
      <dl>
        {Object.entries(item.metrics).map(([name, value]) => {
          const rationale = item.metric_rationales[name];
          return (
            <div className="stack" key={name}>
              <dt>{name}</dt>
              <dd>
                {value === "X" ? "Not Defined" : value}
                {rationale?.reason && (
                  <p className="scanner-text">{rationale.reason}</p>
                )}
                {rationale?.evidence_refs.map((ref) => (
                  <small className="location" key={ref.observation_id}>
                    Observation <code>{ref.observation_id}</code> ·{" "}
                    <code>{ref.source_pointer}</code>
                  </small>
                ))}
                {rationale && rationale.context_factor_refs.length > 0 && (
                  <small className="location">
                    {rationale.context_factor_refs.join(", ")}
                  </small>
                )}
              </dd>
            </div>
          );
        })}
      </dl>
      <h4>Context at assessment time</h4>
      <p>
        {item.context_snapshot.resolver_version} ·{" "}
        {item.context_snapshot.source_revision_ids.length ? (
          <code>{item.context_snapshot.source_revision_ids.join(", ")}</code>
        ) : (
          "No context revision"
        )}
      </p>
      <FactorsView
        factors={item.context_snapshot.factors}
        provenance={item.context_snapshot.factor_provenance}
      />
      {Object.keys(item.factor_dispositions).length > 0 && (
        <>
          <h4>Environmental factor dispositions</h4>
          <dl>
            {Object.entries(item.factor_dispositions).map(
              ([name, disposition]) => (
                <div className="stack" key={name}>
                  <dt>{name.replaceAll("_", " ")}</dt>
                  <dd>{disposition.reason}</dd>
                  <dd>
                    {disposition.metrics.length
                      ? disposition.metrics.join(", ")
                      : "No metric change recorded"}
                  </dd>
                </div>
              ),
            )}
          </dl>
        </>
      )}
    </details>
  );
}

export function AssessmentPanel({
  findingID,
  token,
}: {
  findingID: string;
  token: string;
}) {
  const [open, setOpen] = useState(false);
  const [cursor, setCursor] = useState("");
  const history = useQuery({
    queryKey: ["assessments", findingID, cursor],
    queryFn: () =>
      request<History>(
        token,
        `/findings/${findingID}/assessments?limit=10&cursor=${encodeURIComponent(cursor)}`,
      ),
    enabled: open,
  });
  return (
    <details
      className="panel"
      onToggle={(event) => setOpen(event.currentTarget.open)}
    >
      <summary>CVSS assessment history</summary>
      {open && (
        <section className="stack" aria-label="Stored CVSS assessments">
          {history.isPending && <p role="status">Loading assessments…</p>}
          {history.error && <ErrorState error={history.error} />}
          {history.data?.items.length === 0 && <p>No assessments saved.</p>}
          {history.data?.items.map((item) => (
            <article key={item.id} className="panel stack">
              <h3>{item.kind === "B" ? "CVSS-B" : "CVSS-BE"}</h3>
              <p>
                {item.score === null
                  ? "Incomplete · score unavailable"
                  : `Score ${item.score.toFixed(1)}`}
                {item.stale && " · Historical (stale)"}
              </p>
              {item.vector && (
                <p className="location">
                  <code>{item.vector}</code>
                </p>
              )}
              <small>
                {item.actor} ·{" "}
                <time dateTime={item.created_at}>{item.created_at}</time>
              </small>
              <SnapshotDetails item={item} />
            </article>
          ))}
          <Pager
            active={!!cursor}
            next={history.data?.next_cursor ?? null}
            onFirst={() => setCursor("")}
            onNext={setCursor}
          />
        </section>
      )}
    </details>
  );
}
