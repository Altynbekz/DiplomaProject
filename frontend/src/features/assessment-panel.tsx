import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { APIError, request, type Finding } from "../api/client";
import type { components } from "../api/schema";
import {
  assessmentDraft,
  assessmentInput,
  type Draft,
  type Kind,
  type ResolvedContext,
  type Snapshot,
} from "./assessment-draft";
import { AssessmentEditor, type EvidencePage } from "./assessment-editor";
import { FactorsView } from "./context-panels";
import { ErrorState, Pager } from "./evidence";

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
                    {ref.note && (
                      <span className="scanner-text"> · {ref.note}</span>
                    )}
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

function useSnapshot(token: string, id: string | null) {
  return useQuery({
    queryKey: ["assessment", id],
    queryFn: () => request<Snapshot>(token, `/assessments/${id}`),
    enabled: id !== null,
  });
}

function PairedBase({ parentID, token }: { parentID: string; token: string }) {
  const [open, setOpen] = useState(false);
  const parent = useSnapshot(token, open ? parentID : null);
  return (
    <details onToggle={(event) => setOpen(event.currentTarget.open)}>
      <summary>Paired Base assessment</summary>
      {open && (
        <section className="stack" aria-label="Paired Base">
          {parent.isPending && <p role="status">Loading paired Base…</p>}
          {parent.error && <ErrorState error={parent.error} />}
          {parent.data && (
            <>
              <p>
                CVSS-B score {parent.data.score?.toFixed(1)}. This exact Base
                snapshot supplies the Base values and evidence for the
                Environmental assessment.
              </p>
              <p className="location">
                <code>{parent.data.vector}</code>
              </p>
              <SnapshotDetails item={parent.data} />
            </>
          )}
        </section>
      )}
    </details>
  );
}

export function AssessmentPanel({
  finding,
  token,
  evidence,
}: {
  finding: Finding;
  token: string;
  evidence: EvidencePage;
}) {
  const client = useQueryClient();
  const findingID = finding.id;
  const findingPath = `/findings/${findingID}`;
  const path = findingPath + "/assessments";
  const contextPath = findingPath + "/resolved-context";
  const historyKey = ["assessments", findingID];
  const contextKey = ["resolved-context", findingID];
  const [open, setOpen] = useState(false);
  const [cursor, setCursor] = useState("");
  const [activeKind, setActiveKind] = useState<Kind | null>(null);
  const [drafts, setDrafts] = useState<Partial<Record<Kind, Draft>>>({});
  const [notice, setNotice] = useState("");
  const draft = activeKind ? drafts[activeKind] : undefined;
  const history = useQuery({
    queryKey: [...historyKey, cursor],
    queryFn: () =>
      request<History>(
        token,
        `${path}?limit=10&cursor=${encodeURIComponent(cursor)}`,
      ),
    enabled: open,
  });
  const context = useQuery({
    queryKey: contextKey,
    queryFn: () => request<ResolvedContext>(token, contextPath),
    enabled: open && activeKind === "BE",
  });
  const parent = useSnapshot(token, open ? (draft?.parentID ?? null) : null);

  function setDraft(next: Draft) {
    setDrafts((previous) => ({ ...previous, [next.kind]: next }));
  }
  function reloadHistory() {
    setCursor("");
    void client.invalidateQueries({ queryKey: historyKey });
  }
  function closeDraft(kind: Kind) {
    setDrafts((previous) => ({ ...previous, [kind]: undefined }));
    setActiveKind(null);
  }
  const save = useMutation({
    mutationFn: ({ draft, finalize }: { draft: Draft; finalize: boolean }) =>
      request<Snapshot>(token, path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(assessmentInput(draft, finalize)),
      }),
    onSuccess: (snapshot) => {
      closeDraft(snapshot.kind);
      setNotice(
        snapshot.status === "complete"
          ? `CVSS-${snapshot.kind} finalized as a new snapshot.`
          : `Incomplete CVSS-${snapshot.kind} saved as a new snapshot.`,
      );
      reloadHistory();
      void client.invalidateQueries({ queryKey: ["finding", findingID] });
      void client.invalidateQueries({
        queryKey: ["findings", finding.target_id],
      });
    },
    onError: (error) => {
      if (!(error instanceof APIError) || error.problem.status >= 500) {
        setNotice(
          "Save outcome uncertain. Check the refreshed history before trying again.",
        );
        reloadHistory();
      }
    },
  });
  const refresh = useMutation({
    mutationFn: (kind: Kind) =>
      Promise.all([
        request<Finding>(token, findingPath),
        kind === "BE" ? request<ResolvedContext>(token, contextPath) : null,
      ]),
    onSuccess: ([current, resolved]) => {
      client.setQueryData(["finding", findingID], current);
      if (resolved) client.setQueryData(contextKey, resolved);
      const parentConflict =
        save.error instanceof APIError &&
        save.error.problem.errors.some(
          ({ pointer }) => pointer === "/base_assessment_id",
        );
      save.reset();
      setNotice(
        parentConflict
          ? "Your draft is retained. Select the newest completed Base in history before saving again."
          : "Current revisions loaded. Your draft is retained. Review the revisions before using them.",
      );
      reloadHistory();
    },
  });
  const pending = save.isPending || refresh.isPending;
  const loaded = refresh.data;
  const loadedMatchesDisplay =
    loaded &&
    loaded[0].verification_review_id === finding.verification_review_id &&
    loaded[0].classification.revision_id ===
      finding.classification.revision_id &&
    (!loaded[1] ||
      (context.data &&
        loaded[1].source_revision_ids[0] ===
          context.data.source_revision_ids[0]));
  const canAdopt =
    draft &&
    loaded &&
    loadedMatchesDisplay &&
    (draft.reviewID !== loaded[0].verification_review_id ||
      draft.classificationID !== loaded[0].classification.revision_id ||
      (loaded[1] &&
        draft.contextIDs !== null &&
        (draft.contextIDs[0] ?? null) !==
          (loaded[1].source_revision_ids[0] ?? null)));
  const needsRefresh =
    (save.error instanceof APIError && save.error.problem.status === 409) ||
    (draft &&
      (draft.reviewID !== finding.verification_review_id ||
        draft.classificationID !== finding.classification.revision_id ||
        (draft.kind === "BE" &&
          (context.error ||
            (draft.contextIDs !== null &&
              context.data &&
              (draft.contextIDs[0] ?? null) !==
                (context.data.source_revision_ids[0] ?? null))))));
  function edit(next: Draft) {
    setNotice("");
    save.reset();
    refresh.reset();
    setDraft(next);
    setActiveKind(next.kind);
  }
  return (
    <details
      className="panel"
      onToggle={(event) => setOpen(event.currentTarget.open)}
    >
      <summary>CVSS assessment history</summary>
      {open && (
        <div className="stack">
          {notice && <p role="status">{notice}</p>}
          <div className="actions">
            {activeKind !== "B" && (
              <button
                type="button"
                disabled={pending}
                onClick={() => edit(drafts.B ?? assessmentDraft(finding))}
              >
                {drafts.B
                  ? "Resume CVSS-B assessment"
                  : "New CVSS-B assessment"}
              </button>
            )}
            {drafts.BE && activeKind !== "BE" && (
              <button
                type="button"
                disabled={pending}
                onClick={() => edit(drafts.BE!)}
              >
                Resume CVSS-BE assessment
              </button>
            )}
          </div>
          {draft && (
            <AssessmentEditor
              key={draft.kind}
              finding={finding}
              draft={draft}
              evidence={evidence}
              pending={pending}
              onChange={setDraft}
              context={context.data}
              parent={parent.data}
              contextFeedback={
                <>
                  {context.isPending && (
                    <p role="status">Loading resolved context…</p>
                  )}
                  {context.error && <ErrorState error={context.error} />}
                  {draft.parentID && parent.isPending && (
                    <p role="status">Loading selected Base…</p>
                  )}
                  {parent.error && (
                    <>
                      <ErrorState error={parent.error} />
                      <div>
                        <button
                          type="button"
                          onClick={() => void parent.refetch()}
                        >
                          Reload selected Base
                        </button>
                      </div>
                    </>
                  )}
                </>
              }
              onSubmit={(finalize) => {
                setNotice("");
                save.mutate({ draft, finalize });
              }}
              onDiscard={() => {
                closeDraft(draft.kind);
                setNotice("");
                save.reset();
                refresh.reset();
              }}
              feedback={
                <>
                  {save.error && <ErrorState error={save.error} />}
                  {refresh.error && <ErrorState error={refresh.error} />}
                  {canAdopt ? (
                    <button
                      type="button"
                      disabled={pending}
                      onClick={() => {
                        setDraft({
                          ...draft,
                          reviewID: loaded[0].verification_review_id,
                          classificationID:
                            loaded[0].classification.revision_id,
                          contextIDs: loaded[1]?.source_revision_ids ?? [],
                        });
                        setNotice(
                          "Loaded revisions selected. Your draft is retained. Review its evidence before saving.",
                        );
                        refresh.reset();
                      }}
                    >
                      Use loaded revisions
                    </button>
                  ) : (
                    needsRefresh && (
                      <button
                        type="button"
                        disabled={pending}
                        onClick={() => refresh.mutate(draft.kind)}
                      >
                        {refresh.isPending
                          ? "Loading assessment inputs…"
                          : "Load current assessment inputs"}
                      </button>
                    )
                  )}
                </>
              }
            />
          )}
          <section className="stack" aria-label="Stored CVSS assessments">
            {history.isPending && <p role="status">Loading assessments…</p>}
            {history.error && <ErrorState error={history.error} />}
            {history.data?.items.length === 0 && <p>No assessments saved.</p>}
            {history.data?.items.map((item) => (
              <article key={item.id} className="panel stack">
                <h3>CVSS-{item.kind}</h3>
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
                {item.base_assessment_id && (
                  <PairedBase
                    parentID={item.base_assessment_id}
                    token={token}
                  />
                )}
                <div className="actions">
                  {!drafts[item.kind] && (
                    <button
                      type="button"
                      disabled={pending}
                      onClick={() =>
                        edit(assessmentDraft(finding, { snapshot: item }))
                      }
                    >
                      Revise as new snapshot
                    </button>
                  )}
                  {item.kind === "B" &&
                    item.status === "complete" &&
                    !item.stale && (
                      <button
                        type="button"
                        disabled={
                          pending ||
                          (activeKind === "BE" && draft?.parentID === item.id)
                        }
                        onClick={() =>
                          edit(
                            drafts.BE
                              ? { ...drafts.BE, parentID: item.id }
                              : assessmentDraft(finding, { parent: item }),
                          )
                        }
                      >
                        {drafts.BE
                          ? "Use this Base as parent"
                          : "Start CVSS-BE from this Base"}
                      </button>
                    )}
                </div>
              </article>
            ))}
            <Pager
              active={!!cursor}
              next={history.data?.next_cursor ?? null}
              onFirst={() => setCursor("")}
              onNext={setCursor}
            />
            {history.data?.items.some((item) => item.calculator) && (
              <small>CVSS is owned by FIRST and used by permission.</small>
            )}
          </section>
        </div>
      )}
    </details>
  );
}
