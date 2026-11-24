import { Fragment, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { APIError, request } from "../api/client";
import type { components } from "../api/schema";
import { ErrorState, Pager } from "./evidence";

type Input = components["schemas"]["ContextInput"];
type Revision = components["schemas"]["ContextRevision"];
type History = components["schemas"]["ContextPage"];
type Resolved = components["schemas"]["ResolvedContext"];
type Factors = components["schemas"]["ContextFactors"];
type Provenance = components["schemas"]["FactorProvenance"];
type Source = components["schemas"]["ContextSource"];
type Dimension = keyof Factors;

const dimensions: {
  key: Dimension;
  label: string;
  values: string[];
}[] = [
  {
    key: "exposure",
    label: "Exposure",
    values: ["unknown", "public", "internal", "restricted"],
  },
  {
    key: "required_privileges",
    label: "Required privileges",
    values: ["unknown", "anonymous", "user", "admin"],
  },
  {
    key: "data_sensitivity",
    label: "Data sensitivity",
    values: ["unknown", "public", "internal", "personal", "highly_sensitive"],
  },
  {
    key: "business_criticality",
    label: "Business criticality",
    values: ["unknown", "low", "medium", "high", "critical"],
  },
];
const unknown: Factors = {
  exposure: "unknown",
  required_privileges: "unknown",
  data_sensitivity: "unknown",
  business_criticality: "unknown",
};
const empty: Provenance = {
  exposure: { rationale: "", sources: [] },
  required_privileges: { rationale: "", sources: [] },
  data_sensitivity: { rationale: "", sources: [] },
  business_criticality: { rationale: "", sources: [] },
};

function FactorsView({
  factors,
  provenance,
}: {
  factors: Factors;
  provenance?: Provenance;
}) {
  return (
    <dl>
      {dimensions.map(({ key, label }) => (
        <Fragment key={key}>
          <dt>{label}</dt>
          <dd className="stack">
            <strong>{factors[key].replaceAll("_", " ")}</strong>
            {provenance && (
              <span className="scanner-text">{provenance[key].rationale}</span>
            )}
            {provenance?.[key].sources.map((source, index) => (
              <span key={index}>
                {"document_ref" in source ? (
                  <>
                    {source.document_ref} · SHA-256 <code>{source.sha256}</code>
                  </>
                ) : (
                  <>
                    Observation <code>{source.observation_id}</code> ·{" "}
                    <code>{source.source_pointer}</code>
                  </>
                )}
                {source.note && (
                  <small className="location">{source.note}</small>
                )}
              </span>
            ))}
          </dd>
        </Fragment>
      ))}
    </dl>
  );
}

function SourceFields({
  source,
  onChange,
  remove,
}: {
  source: Source;
  onChange: (source: Source) => void;
  remove: () => void;
}) {
  return (
    <div className="stack panel">
      {"document_ref" in source ? (
        <>
          <label>
            Lab document reference
            <input
              required
              value={source.document_ref}
              maxLength={128}
              onChange={(event) =>
                onChange({ ...source, document_ref: event.target.value })
              }
            />
          </label>
          <label>
            Document SHA-256
            <input
              required
              value={source.sha256}
              pattern="[0-9a-f]{64}"
              onChange={(event) =>
                onChange({ ...source, sha256: event.target.value })
              }
            />
          </label>
        </>
      ) : (
        <>
          <label>
            Observation ID
            <input
              required
              value={source.observation_id}
              onChange={(event) =>
                onChange({ ...source, observation_id: event.target.value })
              }
            />
          </label>
          <label>
            Observation source pointer
            <input
              required
              value={source.source_pointer}
              onChange={(event) =>
                onChange({ ...source, source_pointer: event.target.value })
              }
            />
          </label>
        </>
      )}
      <label>
        Source note
        <input
          value={source.note}
          maxLength={256}
          onChange={(event) =>
            onChange({ ...source, note: event.target.value })
          }
        />
      </label>
      <div>
        <button type="button" onClick={remove}>
          Remove source
        </button>
      </div>
    </div>
  );
}

export function ContextPanel({
  targetID,
  token,
}: {
  targetID: string;
  token: string;
}) {
  const client = useQueryClient();
  const [open, setOpen] = useState(false);
  const [scope, setScope] = useState<"default" | "endpoint">("default");
  const [method, setMethod] = useState("GET");
  const [path, setPath] = useState("");
  const [factors, setFactors] = useState<Factors>(unknown);
  const [provenance, setProvenance] = useState<Provenance>(empty);
  const [expected, setExpected] = useState<string | null>(null);
  const [cursor, setCursor] = useState("");
  const [notice, setNotice] = useState("");
  const scopeMethod = scope === "endpoint" ? method : null;
  const scopePath = scope === "endpoint" ? path : null;
  const endpoint = `/targets/${targetID}/contexts`;
  const key = ["contexts", targetID];
  const history = useQuery({
    queryKey: [...key, cursor],
    queryFn: () =>
      request<History>(
        token,
        `${endpoint}?limit=10&cursor=${encodeURIComponent(cursor)}`,
      ),
    enabled: open,
  });
  const save = useMutation({
    mutationFn: (body: Input) =>
      request<Revision>(token, endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    onSuccess: (revision) => {
      setExpected(revision.id);
      if (revision.method && revision.path) {
        setMethod(revision.method);
        setPath(revision.path);
      }
      setCursor("");
      setNotice("Context revision saved.");
      void client.invalidateQueries({ queryKey: key });
      void client.invalidateQueries({ queryKey: ["resolved-context"] });
    },
  });
  const recoveringConflict =
    save.error instanceof APIError && save.error.problem.status === 409;
  function submit(event: FormEvent) {
    event.preventDefault();
    setNotice("");
    save.mutate({
      method: scopeMethod,
      path: scopePath,
      factors,
      factor_provenance: provenance,
      expected_previous_id: expected,
    });
  }
  function sources(key: Dimension, next: Source[]) {
    setProvenance((previous) => ({
      ...previous,
      [key]: { ...previous[key], sources: next },
    }));
  }
  function matchesDraftScope(row: Revision) {
    return row.method === scopeMethod && row.path === scopePath;
  }
  function selectRevision(row: Revision) {
    if (recoveringConflict && !matchesDraftScope(row)) return;
    if (!recoveringConflict) {
      setScope(row.method === null ? "default" : "endpoint");
      setMethod(row.method ?? "GET");
      setPath(row.path ?? "");
      setFactors(row.factors);
      setProvenance(row.factor_provenance);
    }
    setExpected(row.id);
    setNotice(
      recoveringConflict
        ? "Predecessor updated. Review your draft before saving."
        : "Saved scope loaded for revision.",
    );
    save.reset();
  }
  return (
    <details
      className="panel"
      onToggle={(event) => setOpen(event.currentTarget.open)}
    >
      <summary>Application context</summary>
      {open && (
        <div className="stack">
          <form className="stack" onSubmit={submit}>
            <fieldset
              className="stack context-editor"
              disabled={save.isPending}
              aria-label="Context revision fields"
            >
              <label>
                Context scope
                <select
                  value={scope}
                  onChange={(event) => {
                    setScope(
                      event.target.value === "default" ? "default" : "endpoint",
                    );
                    setExpected(null);
                  }}
                >
                  <option value="default">Target default</option>
                  <option value="endpoint">Exact endpoint</option>
                </select>
              </label>
              {scope === "endpoint" && (
                <div className="stack">
                  <label>
                    Endpoint method
                    <input
                      required
                      value={method}
                      maxLength={64}
                      onChange={(event) => {
                        setMethod(event.target.value);
                        setExpected(null);
                      }}
                    />
                  </label>
                  <label>
                    Endpoint path
                    <input
                      required
                      value={path}
                      maxLength={4096}
                      onChange={(event) => {
                        setPath(event.target.value);
                        setExpected(null);
                      }}
                    />
                  </label>
                </div>
              )}
              <p className="muted">
                Explain each factor. Known values require an observation or a
                lab document reference.
              </p>
              {dimensions.map(({ key, label, values }) => (
                <fieldset key={key} className="stack">
                  <legend>{label}</legend>
                  <label>
                    Value
                    <select
                      value={factors[key]}
                      onChange={(event) =>
                        setFactors((previous) => ({
                          ...previous,
                          [key]: event.target.value,
                        }))
                      }
                    >
                      {values.map((value) => (
                        <option value={value} key={value}>
                          {value.replaceAll("_", " ")}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label>
                    Rationale
                    <textarea
                      required
                      value={provenance[key].rationale}
                      maxLength={2048}
                      onChange={(event) => {
                        const rationale = event.target.value;
                        setProvenance((previous) => ({
                          ...previous,
                          [key]: { ...previous[key], rationale },
                        }));
                      }}
                    />
                  </label>
                  {provenance[key].sources.map((source, index) => (
                    <SourceFields
                      key={index}
                      source={source}
                      onChange={(next) =>
                        sources(
                          key,
                          provenance[key].sources.map((existing, i) =>
                            i === index ? next : existing,
                          ),
                        )
                      }
                      remove={() =>
                        sources(
                          key,
                          provenance[key].sources.filter((_, i) => i !== index),
                        )
                      }
                    />
                  ))}
                  {provenance[key].sources.length < 4 && (
                    <div className="actions">
                      <button
                        type="button"
                        onClick={() =>
                          sources(key, [
                            ...provenance[key].sources,
                            { document_ref: "", sha256: "", note: "" },
                          ])
                        }
                      >
                        Add lab document
                      </button>
                      <button
                        type="button"
                        onClick={() =>
                          sources(key, [
                            ...provenance[key].sources,
                            {
                              observation_id: "",
                              source_pointer: "",
                              note: "",
                            },
                          ])
                        }
                      >
                        Add observation
                      </button>
                    </div>
                  )}
                </fieldset>
              ))}
              <div>
                <button
                  type="submit"
                  className="primary"
                  disabled={save.isPending}
                >
                  {save.isPending ? "Saving context…" : "Save context revision"}
                </button>
              </div>
            </fieldset>
          </form>
          {notice && <p role="status">{notice}</p>}
          {save.error && <ErrorState error={save.error} />}
          {recoveringConflict && (
            <div className="stack">
              <p>
                Refresh history and use the latest revision for this scope to
                continue with your draft. For an endpoint, match the method and
                path shown in history.
              </p>
              <div>
                <button
                  type="button"
                  onClick={() => {
                    setCursor("");
                    void client.invalidateQueries({ queryKey: key });
                  }}
                >
                  Refresh context history
                </button>
              </div>
            </div>
          )}
          <section className="stack" aria-label="Context history">
            <h3>Context history</h3>
            {history.isPending && <p role="status">Loading context history…</p>}
            {history.error && <ErrorState error={history.error} />}
            {history.data?.items.length === 0 && (
              <p>No context revisions yet.</p>
            )}
            {history.data?.items.map((row) => (
              <article className="panel stack" key={row.id}>
                <h4>
                  {row.method === null
                    ? "Target default"
                    : `${row.method} ${row.path}`}{" "}
                  · Revision {row.revision}
                </h4>
                <small>
                  {row.actor} ·{" "}
                  <time dateTime={row.created_at}>{row.created_at}</time>
                </small>
                <small>
                  Revision ID <code>{row.id}</code>
                </small>
                <details>
                  <summary>Factor provenance</summary>
                  <FactorsView
                    factors={row.factors}
                    provenance={row.factor_provenance}
                  />
                </details>
                <div>
                  <button
                    type="button"
                    disabled={
                      save.isPending ||
                      (recoveringConflict && !matchesDraftScope(row))
                    }
                    onClick={() => selectRevision(row)}
                  >
                    Use this revision
                  </button>
                </div>
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

export function ResolvedContextPanel({
  findingID,
  token,
}: {
  findingID: string;
  token: string;
}) {
  const [open, setOpen] = useState(false);
  const result = useQuery({
    queryKey: ["resolved-context", findingID],
    queryFn: () =>
      request<Resolved>(token, `/findings/${findingID}/resolved-context`),
    enabled: open,
  });
  return (
    <details
      className="panel"
      onToggle={(event) => setOpen(event.currentTarget.open)}
    >
      <summary>Resolved context</summary>
      {open && (
        <section className="stack" aria-label="Resolved context factors">
          {result.isPending && <p role="status">Resolving context…</p>}
          {result.error && <ErrorState error={result.error} />}
          {result.data && (
            <>
              <p>
                {result.data.source_revision_ids.length > 0 ? (
                  <>
                    Source revision{" "}
                    <code>{result.data.source_revision_ids.join(", ")}</code>
                  </>
                ) : (
                  "No context revision applies."
                )}
              </p>
              <FactorsView
                factors={result.data.factors}
                provenance={
                  result.data.source_revision_ids.length
                    ? result.data.factor_provenance
                    : undefined
                }
              />
            </>
          )}
        </section>
      )}
    </details>
  );
}
