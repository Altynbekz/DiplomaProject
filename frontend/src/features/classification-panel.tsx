import { type FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  APIError,
  request,
  type Finding,
  type Observation,
} from "../api/client";
import type { components } from "../api/schema";
import { ErrorState, Pager } from "./evidence";
import { EvidenceSelection } from "./verification-panel";

type Input = components["schemas"]["ClassificationInput"];
type Revision = components["schemas"]["ClassificationRevision"];
type History = components["schemas"]["ClassificationPage"];
type Category = components["schemas"]["StudyCategory"];
type Status = components["schemas"]["ClassificationStatus"];

const categories: Category[] = ["A01:2025", "A02:2025", "A05:2025"];
const statuses: { id: Status; label: string }[] = [
  { id: "unmapped", label: "Unmapped" },
  { id: "mapped", label: "Mapped" },
  { id: "ambiguous", label: "Ambiguous" },
  { id: "out_of_scope", label: "Out of scope" },
];

export function ClassificationPanel({
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
  const [status, setStatus] = useState<Status>(finding.classification.status);
  const [category, setCategory] = useState<Category | null>(
    finding.classification.category,
  );
  const [candidates, setCandidates] = useState<Category[]>(
    finding.classification.candidate_categories,
  );
  const [rationale, setRationale] = useState("");
  const [refs, setRefs] = useState<Input["evidence_refs"]>([]);
  const [predecessor, setPredecessor] = useState(
    finding.classification.revision_id,
  );
  const [notice, setNotice] = useState("");
  const [latestError, setLatestError] = useState<unknown>(null);
  const path = `/findings/${finding.id}/classifications`;
  const key = ["classifications", finding.id];
  const history = useQuery({
    queryKey: [...key, cursor],
    queryFn: () =>
      request<History>(
        token,
        `${path}?limit=10&cursor=${encodeURIComponent(cursor)}`,
      ),
    enabled: open,
  });
  const save = useMutation({
    mutationFn: (body: Input) =>
      request<Revision>(token, path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    onSuccess: (revision) => {
      setPredecessor(revision.id);
      setCursor("");
      setNotice("Classification saved.");
      void client.invalidateQueries({ queryKey: key });
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
    const chosen = status === "mapped" ? category : null;
    save.mutate({
      status,
      category: chosen,
      candidate_categories: chosen
        ? [chosen]
        : status === "ambiguous"
          ? candidates
          : [],
      rationale,
      evidence_refs: refs,
      expected_previous_id: predecessor,
    });
  };
  async function loadLatest() {
    try {
      const page = await request<History>(token, `${path}?limit=10`);
      setPredecessor(page.items[0]?.id ?? null);
      setCursor("");
      client.setQueryData([...key, ""], page);
      setNotice(
        "Latest classification loaded. Review your draft before saving.",
      );
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
      <summary>OWASP classification</summary>
      {open && (
        <div className="stack">
          <form className="stack" onSubmit={submit}>
            <label>
              Classification status
              <select
                value={status}
                onChange={(event) => {
                  const value = statuses.find(
                    (state) => state.id === event.target.value,
                  );
                  if (value) setStatus(value.id);
                }}
              >
                {statuses.map(({ id, label }) => (
                  <option key={id} value={id}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
            {status === "mapped" && (
              <label>
                Study category
                <select
                  value={category ?? ""}
                  required
                  onChange={(event) =>
                    setCategory(
                      categories.find(
                        (value) => event.target.value === value,
                      ) ?? null,
                    )
                  }
                >
                  <option value="">Select a category</option>
                  {categories.map((value) => (
                    <option key={value}>{value}</option>
                  ))}
                </select>
              </label>
            )}
            {status === "ambiguous" && (
              <fieldset className="stack">
                <legend>Competing categories</legend>
                {categories.map((value) => (
                  <label key={value} className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={candidates.includes(value)}
                      onChange={(event) =>
                        setCandidates(
                          event.target.checked
                            ? [...candidates, value]
                            : candidates.filter((current) => current !== value),
                        )
                      }
                    />
                    {value}
                  </label>
                ))}
              </fieldset>
            )}
            <label>
              Classification rationale
              <textarea
                value={rationale}
                required
                maxLength={4096}
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
                disabled={
                  save.isPending ||
                  !refs.length ||
                  (status === "ambiguous" && candidates.length < 2)
                }
              >
                {save.isPending
                  ? "Saving classification…"
                  : "Save classification"}
              </button>
            </div>
          </form>
          {notice && <p role="status">{notice}</p>}
          {save.error && <ErrorState error={save.error} />}
          {save.error instanceof APIError &&
            save.error.problem.status === 409 && (
              <button type="button" onClick={() => void loadLatest()}>
                Load latest classification
              </button>
            )}
          {!!latestError && <ErrorState error={latestError} />}
          <section className="stack" aria-label="Classification history">
            <h3>Classification history</h3>
            {history.isPending && (
              <p role="status">Loading classification history…</p>
            )}
            {history.error && <ErrorState error={history.error} />}
            {history.data?.items.length === 0 && (
              <p>No classification revisions yet.</p>
            )}
            {history.data?.items.map((item) => (
              <article key={item.id} className="panel stack">
                <strong>
                  {item.status}{" "}
                  {item.category ?? item.candidate_categories.join(", ")}
                </strong>
                <small>
                  {item.mapping_version} · {item.actor} ·{" "}
                  <time dateTime={item.created_at}>{item.created_at}</time>
                </small>
                <p className="scanner-text">{item.rationale}</p>
                {item.rule_refs.map((rule) => (
                  <div key={rule.id}>
                    <p>{rule.id}</p>
                    <p className="location">
                      Policy review: {rule.review_record}
                    </p>
                    <p className="location">
                      Policy SHA-256: <code>{rule.policy_sha256}</code>
                    </p>
                    {rule.sources.map((source) => (
                      <p className="location" key={source}>
                        {source}
                      </p>
                    ))}
                  </div>
                ))}
                {item.evidence_refs.map((ref) => (
                  <div key={ref.observation_id}>
                    <code>{ref.source_pointer}</code>
                    <small className="location">
                      Observation <code>{ref.observation_id}</code>
                    </small>
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
