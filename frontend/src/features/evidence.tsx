import { useState, type FormEvent } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
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
              {v.pointer}: {v.message}
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
  if (assessment.score === null) return <>Incomplete</>;
  return (
    <>
      {assessment.score}
      {assessment.vector && (
        <>
          {" "}
          · <code>{assessment.vector}</code>
        </>
      )}
    </>
  );
}
function Pager({
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
  return (
    <nav aria-label="Pagination">
      <button disabled={!active} onClick={onFirst}>
        First page
      </button>
      <button disabled={!next} onClick={() => next && onNext(next)}>
        Next page
      </button>
    </nav>
  );
}
export function Targets() {
  const token = useToken(),
    client = useQueryClient();
  const [cursor, setCursor] = useState("");
  const query = useQuery({
    queryKey: ["targets", cursor],
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
    <>
      <h2>Lab targets</h2>
      <p>
        Origins are resolved from the server registry; this screen cannot
        schedule scans.
      </p>
      <section className="panel">
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
          <button disabled={create.isPending}>
            {create.isPending ? "Creating…" : "Create Juice Shop target"}
          </button>
        </form>
        {create.error && <ErrorState error={create.error} />}
      </section>
      {query.isPending && <p role="status">Loading targets…</p>}
      {query.error && <ErrorState error={query.error} />}
      {query.data && (
        <>
          <ul className="targets">
            {query.data.items.map((t) => (
              <li className="panel" key={t.id}>
                <Link to={`/targets/${t.id}`}>
                  <h3>{t.name}</h3>
                </Link>
                <code>{t.canonical_origin}</code>
              </li>
            ))}
          </ul>
          {query.data.items.length === 0 && (
            <p>No reports imported. Create a registered lab target to start.</p>
          )}
          <Pager
            next={query.data.next_cursor}
            onNext={setCursor}
            onFirst={() => setCursor("")}
            active={!!cursor}
          />
        </>
      )}
    </>
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
    <details>
      <summary>Explicit private evidence access</summary>
      <p>
        The original report can contain credentials and personal data. Download
        is audited. Review locally and retain it privately.
      </p>
      <label>
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
    <section className="panel">
      <h3>Import provenance</h3>
      {q.isPending && <p>Loading import…</p>}
      {q.error && <ErrorState error={q.error} />}
      {q.data && (
        <>
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
            <dt>Observations / findings / unlocated alerts</dt>
            <dd>
              {q.data.observation_count} / {q.data.finding_count} /{" "}
              {q.data.unlocated_alert_count}
            </dd>
          </dl>
          <details>
            <summary>Recorded provenance (null means unknown)</summary>
            {q.data.provenance.schema_version === "report-provenance/v1" && (
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
        </>
      )}
    </section>
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
  function change(key: string, value: string) {
    const next = new URLSearchParams(params);
    next.delete("cursor");
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next);
  }
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
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
    <>
      <Link to="/">← Targets</Link>
      <h2>{t.data?.name ?? "Lab target"}</h2>
      {t.error && <ErrorState error={t.error} />}
      <p>
        <code>{t.data?.canonical_origin}</code>
      </p>
      <section className="panel">
        <h3>Import ZAP Traditional JSON</h3>
        <p>
          Select a report from your authorized lab. Uploading does not contact
          report URLs. Maximum total upload: 20 MiB.
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
          <label>
            Provenance JSON (optional; unknown metadata stays null)
            <textarea
              name="provenance"
              rows={4}
              placeholder={'{"authorization_ref":null,"scan_time":null}'}
            />
          </label>
          <button disabled={upload.isPending}>
            {upload.isPending ? "Importing…" : "Import report"}
          </button>
        </form>
        {localError != null && <ErrorState error={localError} />}{" "}
        {upload.error && <ErrorState error={upload.error} />}{" "}
        {upload.data && (
          <p role="status">
            {upload.data.duplicate
              ? "Existing import reused."
              : "Report persisted."}{" "}
            {upload.data.observation_count} observations.
          </p>
        )}
      </section>
      {params.get("import_id") && <ImportCard id={params.get("import_id")!} />}
      <section>
        <h3>Persisted findings</h3>
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
          <label>
            Import ID filter
            <input
              value={params.get("import_id") ?? ""}
              onChange={(e) => change("import_id", e.target.value)}
            />
          </label>
          <button onClick={() => setParams({})}>Clear filters</button>
        </div>
        {findings.isPending && <p role="status">Loading persisted findings…</p>}
        {findings.error && <ErrorState error={findings.error} />}
        {findings.data && (
          <>
            {findings.data.items.length === 0 ? (
              <p>
                {params.size
                  ? "No findings match these filters."
                  : "No reports imported or no located findings. Upload a report to view observations."}
              </p>
            ) : (
              <div className="table-wrap">
                <table>
                  <caption>
                    Scanner findings · latest observation within the import
                    filter
                  </caption>
                  <thead>
                    <tr>
                      <th>Finding / location</th>
                      <th>Latest scanner severity</th>
                      <th>Verification</th>
                      <th>Classification</th>
                      <th>CVSS</th>
                    </tr>
                  </thead>
                  <tbody>
                    {findings.data.items.map((f) => (
                      <tr key={f.id}>
                        <td>
                          <Link to={`/findings/${f.id}`}>{f.title}</Link>
                          <br />
                          <code>
                            {f.method} {f.path}
                          </code>
                          <br />
                          <small>
                            Parameter: {f.parameter_name || "(none)"} ·{" "}
                            {f.observation_count} observations
                          </small>
                        </td>
                        <td>{f.scanner_severity}</td>
                        <td>{f.verification}</td>
                        <td>{f.classification.status}</td>
                        <td>
                          {f.latest_base === null &&
                          f.latest_environmental === null ? (
                            "Not assessed"
                          ) : (
                            <>
                              CVSS-B{" "}
                              <AssessmentValue assessment={f.latest_base} /> /
                              CVSS-BE{" "}
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
            )}
            <Pager
              next={findings.data.next_cursor}
              onNext={(v) => {
                const p = new URLSearchParams(params);
                p.set("cursor", v);
                setParams(p);
              }}
              onFirst={() => change("cursor", "")}
              active={!!params.get("cursor")}
            />
          </>
        )}
      </section>
    </>
  );
}
export function FindingPage() {
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
  return (
    <>
      {f.data && (
        <>
          <Link to={`/targets/${f.data.target_id}`}>← Findings</Link>
          <h2>{f.data.title}</h2>
          <code>
            {f.data.method} {f.data.origin}
            {f.data.path}
          </code>
          <dl className="panel">
            <dt>Latest scanner observation</dt>
            <dd>
              {f.data.scanner_severity} · {f.data.severity_observation_id}
            </dd>
            <dt>Verification</dt>
            <dd>{f.data.verification}</dd>
            <dt>Category mapping</dt>
            <dd>{f.data.classification.status}</dd>
            <dt>CVSS-B / CVSS-BE</dt>
            <dd>
              <AssessmentValue assessment={f.data.latest_base} /> /{" "}
              <AssessmentValue assessment={f.data.latest_environmental} />
            </dd>
            <dt>Fingerprint {f.data.fingerprint_version}</dt>
            <dd>
              <code>{f.data.fingerprint}</code>
            </dd>
          </dl>
        </>
      )}
      {f.isPending && <p>Loading finding…</p>}
      {f.error && <ErrorState error={f.error} />}
      <h3>Source observations</h3>
      {o.isPending && <p role="status">Loading observations…</p>}
      {o.error && <ErrorState error={o.error} />}{" "}
      {o.data && (
        <>
          {o.data.items.map((v) => (
            <article className="panel" key={v.id}>
              <h4>
                {v.method} {v.path} · {v.scanner_severity}
              </h4>
              <dl>
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
                <summary>Report hash and provenance</summary>
                <ImportCard id={v.import_id} />
              </details>
            </article>
          ))}
          {o.data.items.length === 0 && <p>No observations.</p>}
          <Pager
            next={o.data.next_cursor}
            onNext={setCursor}
            onFirst={() => setCursor("")}
            active={!!cursor}
          />
        </>
      )}
    </>
  );
}
