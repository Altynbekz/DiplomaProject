import { useState, type FormEvent } from "react";
import {
  Link,
  useLocation,
  useParams,
  useSearchParams,
} from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  APIError,
  request,
  type Target,
  type Finding,
  type Observation,
  type ImportResult,
} from "../api/client";
import type { components } from "../api/schema";
import { useToken } from "../app/main";
import { VerificationPanel } from "./verification-panel";
import { ClassificationPanel } from "./classification-panel";
type Page<T> = { items: T[]; next_cursor: string | null };
export function ErrorState({ error }: { error: unknown }) {
  return (
    <div className="error" role="alert">
      <strong>Could not complete the request.</strong>
      <p>{error instanceof Error ? error.message : "Unexpected error"}</p>
      {error instanceof APIError && (
        <>
          <small>
            {error.problem.code} · Request {error.problem.request_id}
          </small>
          {error.problem.errors.map((v, i) => (
            <p key={i}>
              {v.pointer && <code>{v.pointer} · </code>}
              {v.message}
            </p>
          ))}
        </>
      )}
    </div>
  );
}
export function AssessmentValue({
  assessment,
}: {
  assessment: Finding["latest_base"];
}) {
  if (assessment === null) return <>Not assessed</>;
  if (assessment.score === null)
    return <>Incomplete{assessment.stale && " (historical)"}</>;
  return (
    <>
      {assessment.score}
      {assessment.stale && " (historical)"}
      {assessment.vector && (
        <>
          {" "}
          · <code>{assessment.vector}</code>
        </>
      )}
    </>
  );
}
export function Pager({
  next,
  onNext,
  onFirst,
  active,
}: {
  next: string | null;
  onNext: (s: string) => void;
  onFirst: () => void;
  active: boolean;
}) {
  if (!active && !next) return null;
  return (
    <nav aria-label="Pagination">
      {active && (
        <button type="button" onClick={onFirst}>
          First page
        </button>
      )}
      {next && (
        <button type="button" onClick={() => onNext(next)}>
          Next page
        </button>
      )}
    </nav>
  );
}
export function Targets() {
  const token = useToken(),
    client = useQueryClient();
  const [cursor, setCursor] = useState("");
  const query = useQuery({
    queryKey: ["targets", cursor],
    staleTime: 5_000,
    queryFn: () =>
      request<components["schemas"]["TargetPage"]>(
        token,
        `/targets?limit=50&cursor=${encodeURIComponent(cursor)}`,
      ),
  });
  const create = useMutation({
    mutationFn: (name: string) =>
      request<Target>(token, "/targets", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, lab_target_key: "juice-shop" }),
      }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["targets"] }),
  });
  return (
    <div className="page-stack">
      <h1>Lab targets</h1>
      {query.isPending && <p role="status">Loading targets…</p>}
      {query.error && <ErrorState error={query.error} />}
      {query.data && (
        <>
          {query.data.items.length === 0 && !cursor ? (
            <section className="panel stack">
              <h2>Create Juice Shop target</h2>
              <p className="muted">
                No reports imported. Create a target using the registered lab
                origin.
              </p>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  const name = new FormData(e.currentTarget).get("name");
                  if (typeof name === "string") create.mutate(name);
                }}
              >
                <label>
                  Target name
                  <input
                    name="name"
                    required
                    maxLength={200}
                    placeholder="Juice Shop lab"
                  />
                </label>
                <button className="primary" disabled={create.isPending}>
                  {create.isPending ? "Creating…" : "Create Juice Shop target"}
                </button>
              </form>
              {create.error && <ErrorState error={create.error} />}
            </section>
          ) : (
            <ul className="targets stack">
              {query.data.items.map((t) => (
                <li className="panel target-row" key={t.id}>
                  <Link to={`/targets/${t.id}`}>{t.name}</Link>
                  <span className="muted">{t.canonical_origin}</span>
                </li>
              ))}
            </ul>
          )}
        </>
      )}
      <Pager
        next={query.data?.next_cursor ?? null}
        onNext={setCursor}
        onFirst={() => setCursor("")}
        active={!!cursor}
      />
    </div>
  );
}
function PrivateEvidence({ id }: { id: string }) {
  const token = useToken();
  const [ack, setAck] = useState(false);
  const download = useMutation({
    mutationFn: async () => {
      const response = await fetch(`/api/v1/imports/${id}/raw`, {
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Evidence-Acknowledgement": "private-lab-data",
        },
        cache: "no-store",
      });
      if (!response.ok)
        throw new Error(
          `Private evidence download failed (${response.status}).`,
        );
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = "zap-report.json";
      link.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    },
  });
  return (
    <details
      className="evidence-info"
      onToggle={(event) => {
        if (!event.currentTarget.open) setAck(false);
      }}
    >
      <summary>Explicit private evidence access</summary>
      <div className="stack">
        <p>
          The original report can contain credentials and personal data.
          Download is audited. Review locally and retain it privately.
        </p>
        <label className="checkbox-label">
          <input
            type="checkbox"
            checked={ack}
            onChange={(e) => setAck(e.target.checked)}
          />
          I acknowledge this contains private lab data
        </label>
        <button
          disabled={!ack || download.isPending}
          onClick={() => download.mutate()}
        >
          Download original report
        </button>
        {download.error && <ErrorState error={download.error} />}
      </div>
    </details>
  );
}
export function ImportCard({ id }: { id: string }) {
  const token = useToken();
  const q = useQuery({
    queryKey: ["import", id],
    queryFn: () => request<ImportResult>(token, `/imports/${id}`),
  });
  return (
    <div className="stack">
      {q.isPending && <p role="status">Loading report details…</p>}
      {q.error && <ErrorState error={q.error} />}
      {q.data && (
        <>
          {q.data.finding_count === 0 && (
            <p>This report contains no located findings.</p>
          )}
          <p className="notice">
            {q.data.observation_count} observations · {q.data.finding_count}{" "}
            findings
            {q.data.unlocated_alert_count > 0 &&
              ` · ${q.data.unlocated_alert_count} unlocated`}
            {q.data.warnings.some((w) => w.code === "incomplete_provenance") &&
              " · Provenance incomplete"}
          </p>
          <details className="panel report-details">
            <summary>Source report details</summary>
            <div className="stack">
              <dl>
                <dt>SHA-256 of uploaded bytes</dt>
                <dd>
                  <code>{q.data.report_sha256}</code>
                </dd>
                <dt>Parser</dt>
                <dd>{q.data.parser_version}</dd>
                <dt>Parser-detected report format</dt>
                <dd>{q.data.parser_detected_format}</dd>
                <dt>ZAP version declared by the report</dt>
                <dd>{q.data.report_declared_version ?? "Unknown"}</dd>
                <dt>Import ID</dt>
                <dd>
                  <code>{q.data.id}</code>
                </dd>
              </dl>
              <details className="evidence-info">
                <summary>Recorded provenance (null means unknown)</summary>
                {q.data.provenance.schema_version ===
                  "report-provenance/v1" && (
                  <p>
                    Legacy provenance may contain values copied from the report.
                  </p>
                )}
                <pre>{JSON.stringify(q.data.provenance, null, 2)}</pre>
              </details>
              <PrivateEvidence id={id} />
              {q.data.warnings.map((w, i) => (
                <p className="notice" key={i}>
                  {w.code} · {w.pointer} · {w.message}
                </p>
              ))}
            </div>
          </details>
        </>
      )}
    </div>
  );
}
export function TargetPage() {
  const { id } = useParams();
  const token = useToken(),
    client = useQueryClient();
  const [params, setParams] = useSearchParams();
  const [localError, setLocalError] = useState<unknown>(null);
  const t = useQuery({
    queryKey: ["target", id],
    queryFn: () => request<Target>(token, `/targets/${id}`),
  });
  const { search } = useLocation();
  const importID = params.get("import_id");
  const filtersActive = [
    "import_id",
    "category",
    "verification",
    "scanner_severity",
  ].some((key) => Boolean(params.get(key)));
  const filter = new URLSearchParams(params);
  filter.set("limit", "50");
  const findings = useQuery({
    queryKey: ["findings", id, filter.toString()],
    queryFn: () =>
      request<Page<Finding>>(token, `/targets/${id}/findings?${filter}`),
  });
  const upload = useMutation({
    mutationFn: (form: FormData) =>
      request<ImportResult>(token, `/targets/${id}/imports/zap`, {
        method: "POST",
        body: form,
      }),
    onSuccess: (im) => {
      setParams({ import_id: im.id });
      void client.invalidateQueries({ queryKey: ["findings", id] });
      void client.invalidateQueries({ queryKey: ["import", im.id] });
    },
  });
  const reportID = importID || upload.data?.id;
  function change(key: string, value: string) {
    const next = new URLSearchParams(params);
    next.delete("cursor");
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next);
  }
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (upload.isPending) return;
    upload.reset();
    setLocalError(null);
    const data = new FormData(e.currentTarget);
    const provenance = data.get("provenance");
    if (typeof provenance === "string" && provenance.trim()) {
      try {
        JSON.parse(provenance);
      } catch {
        setLocalError(new Error("Provenance must be valid JSON."));
        return;
      }
    } else data.delete("provenance");
    upload.mutate(data);
  }
  return (
    <div className="page-stack">
      <div className="stack">
        <Link to="/">← Targets</Link>
        <h1>{t.data?.name ?? "Lab target"}</h1>
        {t.data && <p className="muted">{t.data.canonical_origin}</p>}
      </div>
      {t.isPending && <p role="status">Loading target…</p>}
      {t.error && <ErrorState error={t.error} />}
      {t.data && (
        <>
          <section className="panel stack">
            <h2>Import report</h2>
            <p className="muted">
              ZAP Traditional JSON from an authorized lab. Maximum upload: 20
              MiB.
            </p>
            <form onSubmit={submit}>
              <label>
                Report file
                <input
                  name="report"
                  type="file"
                  accept=".json,application/json"
                  required
                />
              </label>
              <button className="primary" disabled={upload.isPending}>
                {upload.isPending ? "Importing…" : "Import report"}
              </button>
              <details className="subsection">
                <summary>Optional report provenance</summary>
                <label>
                  Provenance JSON
                  <textarea
                    name="provenance"
                    rows={4}
                    placeholder={'{"authorization_ref":null,"scan_time":null}'}
                  />
                </label>
              </details>
            </form>
            {localError != null && <ErrorState error={localError} />}
            {upload.error &&
              (upload.error instanceof APIError ? (
                <ErrorState error={upload.error} />
              ) : (
                <p role="alert" className="error">
                  Upload outcome is unknown. You can select the same report and
                  submit again after checking the connection.
                </p>
              ))}
            {upload.data && (
              <p role="status">
                {upload.data.duplicate
                  ? "This report was already imported."
                  : "Report imported."}
              </p>
            )}
          </section>
          {reportID && <ImportCard key={reportID} id={reportID} />}
          <section className="findings stack">
            <div className="toolbar">
              <h2 id="findings-title">Findings</h2>
              <div className="filters">
                <label>
                  Scanner severity
                  <select
                    aria-label="Scanner severity"
                    value={params.get("scanner_severity") ?? ""}
                    onChange={(e) => change("scanner_severity", e.target.value)}
                  >
                    <option value="">All severities</option>
                    {["informational", "low", "medium", "high", "unknown"].map(
                      (v) => (
                        <option key={v}>{v}</option>
                      ),
                    )}
                  </select>
                </label>
                {filtersActive && (
                  <button onClick={() => setParams({})}>Clear filters</button>
                )}
              </div>
            </div>
            <details className="filter-panel">
              <summary>More filters</summary>
              <div className="filters">
                <label>
                  Verification
                  <select
                    aria-label="Verification"
                    value={params.get("verification") ?? ""}
                    onChange={(e) => change("verification", e.target.value)}
                  >
                    <option value="">All verification states</option>
                    {[
                      "unreviewed",
                      "confirmed",
                      "false_positive",
                      "inconclusive",
                    ].map((v) => (
                      <option key={v}>{v}</option>
                    ))}
                  </select>
                </label>
                <label>
                  OWASP category
                  <select
                    aria-label="OWASP category"
                    value={params.get("category") ?? ""}
                    onChange={(e) => change("category", e.target.value)}
                  >
                    <option value="">All categories (includes unmapped)</option>
                    {["A01:2025", "A02:2025", "A05:2025"].map((v) => (
                      <option key={v}>{v}</option>
                    ))}
                  </select>
                </label>
                <form
                  key={importID ?? ""}
                  onSubmit={(event) => {
                    event.preventDefault();
                    const id = String(
                      new FormData(event.currentTarget).get("import_id") ?? "",
                    ).trim();
                    change("import_id", id);
                  }}
                >
                  <label>
                    Import ID filter
                    <input name="import_id" defaultValue={importID ?? ""} />
                  </label>
                  <button type="submit">Apply import ID</button>
                </form>
                <p className="notice">
                  Unmapped findings have no reviewed OWASP category.
                </p>
              </div>
            </details>
            {findings.isPending && (
              <p role="status">Loading persisted findings…</p>
            )}
            {findings.error && <ErrorState error={findings.error} />}
            {findings.data && (
              <>
                {findings.data.items.length === 0 ? (
                  <p>
                    {filtersActive
                      ? "No findings match these filters."
                      : "No located findings yet. Import a report to get started."}
                  </p>
                ) : (
                  <>
                    <p className="notice scroll-hint">
                      Scroll across the table for all columns.
                    </p>
                    <div
                      className="table-wrap"
                      role="region"
                      aria-labelledby="findings-title"
                      tabIndex={0}
                    >
                      <table>
                        <caption>
                          Latest scanner observation in this view.
                        </caption>
                        <thead>
                          <tr>
                            <th>Finding / location</th>
                            <th>Latest scanner severity</th>
                            <th>Verification</th>
                            <th>OWASP category</th>
                            <th>CVSS</th>
                          </tr>
                        </thead>
                        <tbody>
                          {findings.data.items.map((f) => (
                            <tr key={f.id}>
                              <td>
                                <Link
                                  to={`/findings/${f.id}`}
                                  state={{ targetSearch: search }}
                                >
                                  {f.title}
                                </Link>
                                <span className="location">
                                  {f.method} {f.path}
                                </span>
                                <small>
                                  Parameter: {f.parameter_name || "(none)"} ·{" "}
                                  {f.observation_count}{" "}
                                  {f.observation_count === 1
                                    ? "observation"
                                    : "observations"}
                                </small>
                              </td>
                              <td>{f.scanner_severity}</td>
                              <td>{f.verification}</td>
                              <td>
                                {f.classification.category ??
                                  f.classification.status}
                              </td>
                              <td>
                                {f.latest_base === null &&
                                f.latest_environmental === null ? (
                                  "Not assessed"
                                ) : (
                                  <>
                                    CVSS-B{" "}
                                    <AssessmentValue
                                      assessment={f.latest_base}
                                    />{" "}
                                    / CVSS-BE{" "}
                                    <AssessmentValue
                                      assessment={f.latest_environmental}
                                    />
                                  </>
                                )}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </>
                )}
              </>
            )}
            <Pager
              next={findings.data?.next_cursor ?? null}
              onNext={(v) => {
                const p = new URLSearchParams(params);
                p.set("cursor", v);
                setParams(p);
              }}
              onFirst={() => change("cursor", "")}
              active={!!params.get("cursor")}
            />
          </section>
        </>
      )}
    </div>
  );
}
export function FindingPage() {
  const { id } = useParams();
  return <FindingDetail key={id} />;
}
function FindingDetail() {
  const { id } = useParams();
  const token = useToken();
  const [cursor, setCursor] = useState("");
  const f = useQuery({
    queryKey: ["finding", id],
    queryFn: () => request<Finding>(token, `/findings/${id}`),
  });
  const o = useQuery({
    queryKey: ["observations", id, cursor],
    queryFn: () =>
      request<Page<Observation>>(
        token,
        `/findings/${id}/observations?limit=50&cursor=${encodeURIComponent(cursor)}`,
      ),
  });
  const location = useLocation();
  const savedSearch: unknown = location.state?.targetSearch;
  const returnParams = new URLSearchParams();
  if (typeof savedSearch === "string" && savedSearch.startsWith("?")) {
    const prior = new URLSearchParams(savedSearch);
    for (const name of [
      "import_id",
      "category",
      "verification",
      "scanner_severity",
      "cursor",
    ]) {
      const value = prior.get(name);
      if (value) returnParams.set(name, value);
    }
  }
  const returnSearch = returnParams.toString();
  return (
    <div className="page-stack">
      {f.data && (
        <>
          <div className="stack">
            <Link
              to={`/targets/${f.data.target_id}${returnSearch ? `?${returnSearch}` : ""}`}
            >
              ← Findings
            </Link>
            <h1>{f.data.title}</h1>
            <p className="muted">
              {f.data.method} {f.data.origin}
              {f.data.path}
            </p>
          </div>
          <dl className="panel">
            <dt>Latest scanner observation</dt>
            <dd>
              {f.data.scanner_severity} · {f.data.severity_observation_id}
            </dd>
            <dt>Verification</dt>
            <dd>{f.data.verification}</dd>
            <dt>Category mapping</dt>
            <dd>
              {f.data.classification.category ?? f.data.classification.status}
              {f.data.classification.status === "ambiguous" && (
                <small className="location">
                  {f.data.classification.candidate_categories.join(", ")}
                </small>
              )}
            </dd>
            <dt>CVSS-B / CVSS-BE</dt>
            <dd>
              <AssessmentValue assessment={f.data.latest_base} /> /{" "}
              <AssessmentValue assessment={f.data.latest_environmental} />
            </dd>
          </dl>
          <VerificationPanel
            finding={f.data}
            observations={o.data?.items ?? []}
            token={token}
          />
          <ClassificationPanel
            finding={f.data}
            observations={o.data?.items ?? []}
            token={token}
          />
          <details>
            <summary>
              Finding fingerprint ({f.data.fingerprint_version})
            </summary>
            <code>{f.data.fingerprint}</code>
          </details>
        </>
      )}
      {f.isPending && <p>Loading finding…</p>}
      {f.error && <ErrorState error={f.error} />}
      <h2>Source observations</h2>
      {o.isPending && <p role="status">Loading observations…</p>}
      {o.error && <ErrorState error={o.error} />}{" "}
      {o.data && (
        <>
          {o.data.items.map((v) => (
            <article className="panel stack" key={v.id}>
              <h3>
                {v.method} {v.path} · {v.scanner_severity}
              </h3>
              <dl>
                <dt>Observation ID</dt>
                <dd>
                  <code>{v.id}</code>
                </dd>
                <dt>Source JSON pointer</dt>
                <dd>
                  <code>{v.source_pointer}</code>
                </dd>
                <dt>Scanner confidence (raw ZAP code)</dt>
                <dd>{v.confidence_raw || "Unknown"}</dd>
                <dt>CWE identifiers (scanner assertions)</dt>
                <dd>{v.cwe_ids.join(", ") || "Unknown"}</dd>
                <dt>Description</dt>
                <dd className="scanner-text">{v.description || "(none)"}</dd>
                <dt>Remediation (scanner assertion)</dt>
                <dd className="scanner-text">{v.remediation || "(none)"}</dd>
                <dt>Private evidence</dt>
                <dd>{v.evidence_excerpt}</dd>
              </dl>
              {v.redactions.map((r, i) => (
                <p className="notice" key={i}>
                  {r}
                </p>
              ))}
              <details>
                <summary>Report linked to this observation</summary>
                <ImportCard id={v.import_id} />
              </details>
            </article>
          ))}
          {o.data.items.length === 0 && <p>No observations.</p>}
        </>
      )}
      <Pager
        next={o.data?.next_cursor ?? null}
        onNext={setCursor}
        onFirst={() => setCursor("")}
        active={!!cursor}
      />
    </div>
  );
}
