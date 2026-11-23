import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  APIError,
  request,
  type Finding,
  type Observation,
} from "../api/client";
import type { components } from "../api/schema";
import { ErrorState, Pager } from "./evidence";

type EvidenceRef = components["schemas"]["EvidenceRef"];
type VerificationState = components["schemas"]["VerificationState"];
type Verification = components["schemas"]["Verification"];
type VerificationPage = components["schemas"]["VerificationPage"];
type VerificationInput = components["schemas"]["VerificationInput"];

const states: { value: VerificationState; label: string }[] = [
  { value: "unreviewed", label: "Unreviewed" },
  { value: "confirmed", label: "Confirmed" },
  { value: "false_positive", label: "False positive" },
  { value: "inconclusive", label: "Inconclusive" },
];

export function EvidenceSelection({
  observations,
  refs,
  onChange,
}: {
  observations: Observation[];
  refs: EvidenceRef[];
  onChange: (refs: EvidenceRef[]) => void;
}) {
  return (
    <fieldset className="stack">
      <legend>Source evidence</legend>
      {observations.map((observation) => {
        const selected = refs.some(
          (ref) => ref.observation_id === observation.id,
        );
        return (
          <label key={observation.id} className="checkbox-label">
            <input
              type="checkbox"
              checked={selected}
              disabled={!selected && refs.length >= 20}
              onChange={(event) =>
                onChange(
                  event.currentTarget.checked
                    ? [
                        ...refs,
                        {
                          observation_id: observation.id,
                          source_pointer: observation.source_pointer,
                          note: "",
                        },
                      ]
                    : refs.filter(
                        (ref) => ref.observation_id !== observation.id,
                      ),
                )
              }
            />
            <span>
              <code>{observation.source_pointer}</code>
              <small className="location">
                Import <code>{observation.import_id}</code>
              </small>
            </span>
          </label>
        );
      })}
      <small>
        {refs.length} selected · Use the observations page controls to see more.
      </small>
    </fieldset>
  );
}

export function VerificationPanel({
  finding,
  observations,
  token,
}: {
  finding: Finding;
  observations: Observation[];
  token: string;
}) {
  const client = useQueryClient();
  const [open, setOpen] = useState(false);
  const [cursor, setCursor] = useState("");
  const [state, setState] = useState<VerificationState>(finding.verification);
  const [procedure, setProcedure] = useState("");
  const [rationale, setRationale] = useState("");
  const [refs, setRefs] = useState<EvidenceRef[]>([]);
  const [predecessor, setPredecessor] = useState(
    finding.verification_review_id,
  );
  const [latestError, setLatestError] = useState<unknown>(null);
  const [notice, setNotice] = useState("");
  const path = `/findings/${finding.id}/verifications`;
  const historyKey = ["verifications", finding.id];
  const history = useQuery({
    queryKey: [...historyKey, cursor],
    queryFn: () =>
      request<VerificationPage>(
        token,
        `${path}?limit=10&cursor=${encodeURIComponent(cursor)}`,
      ),
    enabled: open,
  });
  const save = useMutation({
    mutationFn: (body: VerificationInput) =>
      request<Verification>(token, path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    onSuccess: (review) => {
      setPredecessor(review.id);
      setCursor("");
      setNotice("Verification saved.");
      void client.invalidateQueries({ queryKey: historyKey });
      void client.invalidateQueries({ queryKey: ["finding", finding.id] });
      void client.invalidateQueries({
        queryKey: ["findings", finding.target_id],
      });
    },
  });
  const submit = (event: FormEvent) => {
    event.preventDefault();
    setNotice("");
    setLatestError(null);
    save.mutate({
      state,
      procedure,
      rationale,
      evidence_refs: refs,
      expected_previous_id: predecessor,
    });
  };
  async function loadLatest() {
    try {
      const page = await request<VerificationPage>(token, `${path}?limit=10`);
      setPredecessor(page.items[0]?.id ?? null);
      setCursor("");
      client.setQueryData([...historyKey, ""], page);
      setNotice("Latest revision loaded. Review your draft before saving.");
      setLatestError(null);
      save.reset();
    } catch (error) {
      setLatestError(error);
    }
  }
  return (
    <details
      className="panel"
      onToggle={(event) => setOpen(event.currentTarget.open)}
    >
      <summary>Human verification</summary>
      {open && (
        <div className="stack">
          <form className="stack" onSubmit={submit}>
            <label>
              Review state
              <select
                value={state}
                onChange={(event) => {
                  const next = states.find(
                    ({ value }) => value === event.target.value,
                  );
                  if (next) setState(next.value);
                }}
              >
                {states.map((entry) => (
                  <option key={entry.value} value={entry.value}>
                    {entry.label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Verification procedure
              <textarea
                value={procedure}
                maxLength={4096}
                required
                onChange={(event) => setProcedure(event.target.value)}
              />
            </label>
            <label>
              Verification rationale
              <textarea
                value={rationale}
                maxLength={4096}
                required
                onChange={(event) => setRationale(event.target.value)}
              />
            </label>
            <EvidenceSelection
              observations={observations}
              refs={refs}
              onChange={setRefs}
            />
            <div>
              <button
                className="primary"
                type="submit"
                disabled={save.isPending || !refs.length}
              >
                {save.isPending ? "Saving verification…" : "Save verification"}
              </button>
            </div>
          </form>
          {notice && <p role="status">{notice}</p>}
          {save.error && (
            <>
              <ErrorState error={save.error} />
              {save.error instanceof APIError &&
                save.error.problem.status === 409 && (
                  <button type="button" onClick={() => void loadLatest()}>
                    Load latest verification
                  </button>
                )}
            </>
          )}
          {!!latestError && <ErrorState error={latestError} />}
          <section aria-label="Verification history" className="stack">
            <h3>Verification history</h3>
            {history.isPending && (
              <p role="status">Loading verification history…</p>
            )}
            {history.error && <ErrorState error={history.error} />}
            {history.data?.items.length === 0 && <p>No human reviews yet.</p>}
            {history.data?.items.map((item) => (
              <article key={item.id} className="panel stack">
                <strong>{item.state}</strong>
                <small>
                  {item.actor} ·{" "}
                  <time dateTime={item.created_at}>{item.created_at}</time>
                </small>
                <p className="scanner-text">{item.procedure}</p>
                <p className="scanner-text">{item.rationale}</p>
                {item.evidence_refs.map((ref) => (
                  <div key={ref.observation_id}>
                    <code>{ref.source_pointer}</code>
                    <small className="location">
                      Observation <code>{ref.observation_id}</code>
                    </small>
                    {ref.note && <p>{ref.note}</p>}
                  </div>
                ))}
              </article>
            ))}
            <Pager
              active={!!cursor}
              next={history.data?.next_cursor ?? null}
              onFirst={() => setCursor("")}
              onNext={setCursor}
            />
          </section>
        </div>
      )}
    </details>
  );
}
