# Backend guide

> Architecture and implementation guidance for the diploma project. This is a development contract supplement, not a claim that all endpoints, tables, or later milestones exist.

## 1. Read this first

Read [AGENTS.md](../AGENTS.md) before changing code. It contains the authoritative domain model, exact planned API interfaces, database requirements, safety rules, research method, and milestone acceptance gates. [FRONTEND.md](../frontend/FRONTEND.md) defines the intended user experience.

Inspect the backend, migrations, [OpenAPI contract](../contracts/openapi.yaml), and tests in the current checkout. Read the [implementation status](../docs/status.md) for observed results and remaining gates. Preserve working code and applied migrations. Add features only for the active milestone.

**Backend goal: a small, reliable service that preserves evidence and explains every assessment.** Keep uncertainty visible. Never manufacture a score, verification result, classification, or successful scan.

## 2. Architecture at a glance

```text
React client
    ↓ typed HTTP requests
Go net/http handlers
    ↓ validated request DTOs
Application services
    ↓ explicit transactions and sqlc queries
PostgreSQL

Later: authorized worker → fixed ZAP service → same report importer
Later: frozen research export → Python analysis
```

| Area | Owns | Does not own |
| --- | --- | --- |
| `cmd/api` | Startup, dependencies, HTTP server, graceful shutdown | Business rules |
| `internal/config` | Validated server configuration and secret loading | Request-controlled destinations |
| `internal/httpapi` | Auth, request limits, decoding, DTOs, problem responses | SQL and scoring formulas |
| `internal/domain` | Domain values, enums, invariants, typed errors | HTTP handler dependencies |
| `internal/service` | Use cases, transactions, audit, revision checks | Scanner transport details in API responses |
| `internal/store` | Database access and transaction adapters | Presentation logic |
| `internal/store/sqlc` | Generated query code | Handwritten edits |
| `internal/zap` | Traditional JSON parsing; later a fixed scanner adapter | Evidence verification |
| `internal/scope` | Offline canonicalization; later enforced scan scope policy | Fetching imported report URLs |
| `internal/cvss` | Later: validated, pinned calculator adapter | Invented metrics or scores |
| `internal/jobs` | Later: durable job leases and lifecycle | Unrestricted execution |

Use interfaces at a consumer boundary when they make a real adapter or behavior test easier. Avoid a generic repository, global mutable service state, microservices, and extra queues.

## 3. Stack and conventions

- Go standard `net/http`, `context.Context`, `log/slog`, pgx/v5, PostgreSQL, sqlc, and goose.
- Use versions pinned in the repository. Do not upgrade dependencies as part of unrelated work.
- UUID v4 identifiers; UTC timestamps; JSON and SQL `snake_case`.
- Keep generated sqlc models separate from public API DTOs.
- Parameterize SQL. Use explicit columns and deterministic ordering.
- Keep transactions short enough to bound resource use. Parse and validate report structure before acquiring write locks; recheck identity and idempotency inside the transaction.
- Propagate cancellation and errors. Close streams, release connections, and roll back failed transactions.
- Use typed validation/conflict errors. Unexpected failures receive safe HTTP messages and redacted diagnostic logging.
- Applied migrations are immutable. Add numbered migrations and regenerate sqlc after query/schema changes.

Do not create the worker, scoring engine, or complete future schema simply to fill the planned directory tree.

## 4. Domain concepts that must stay separate

| Concept | Meaning | Default or rule |
| --- | --- | --- |
| Report | Original uploaded bytes and provenance | Immutable bytes and SHA-256 |
| Observation | A scanner assertion at a source location | Retain source pointer and scanner values |
| Finding | Stable grouping of observations | Fingerprint is a grouping key, not proof of root cause |
| Verification | Human review of the finding | Starts `unreviewed`; append revisions |
| OWASP classification | Reviewed 2025 category mapping | May remain unmapped or ambiguous |
| Scanner severity | Scanner-provided technical label | Preserve original and normalized values |
| Context | Scenario exposure, privileges, sensitivity, criticality | Unknown is valid and explicit |
| CVSS-B | Reviewed Base assessment | Null until complete and calculated |
| CVSS-BE | Base plus justified Environmental choices | References the exact completed Base snapshot |
| Ranking | Frozen eligible cohort and assessment choices | Never recomputed silently in place |

Research categories remain `A01:2025`, `A02:2025`, and `A05:2025`. Import other alerts as well. Do not classify by title substring, convert scanner severity into CVSS, treat scanner confidence as human confirmation, or call a generic alert a CVE.

## 5. First priority: offline report import

The first product milestone is a real ZAP Traditional JSON report persisted and visible in the frontend. Scanning is a separate later capability.

### Import sequence

1. Authenticate and validate target identity and multipart structure.
2. Apply the total request limit before reading an unbounded body.
3. Compute SHA-256 over the exact report bytes.
4. Decode with bounded depth, field sizes, and counts; reject malformed or unsupported structure with source pointers.
5. Validate every site and instance origin against the registered target, entirely offline.
6. Normalize observations and calculate deterministic finding fingerprints using AGENTS.md.
7. Validate submitted provenance without inventing missing values.
8. Begin a transaction, serialize concurrent imports for that target, and check the unique import identity.
9. Return the existing immutable result for an allowed duplicate, or insert the report, findings, observations, and audit event atomically.
10. Commit before returning success. Any failure leaves no partial import.

### Limits

| Limit | Value |
| --- | --- |
| Total multipart request | 20 MiB |
| Ordinary JSON mutation body | 1 MiB |
| Sites per report | 10 |
| Alert objects per report | 5,000 |
| Instances per report | 20,000 |
| Text/scalar field | 64 KiB |
| JSON nesting depth | 64 |
| List page | Default 50, maximum 200 |

Enforce counts across the whole report, not separately for each site. Multipart overhead counts toward the request limit. Enforce the scalar limit on numeric literals as well as strings.

### Important edge cases

- Preserve report bytes even when a normalized representation differs; document any reversible transformation. Never silently remove payload characters.
- Account for database representability before writes. Escaped NUL in decoded text or JSONB-bound values must not become an unexplained database error. Choose explicit pointer-specific rejection or a documented reversible representation, and test the full API/database behavior. Apply this to provenance and unlocated-alert data as well as ordinary observations.
- Do not parse private query values to derive endpoint identity. Decode and validate query names only, following the canonicalization contract. Scanner attack strings can contain unusual value bytes.
- Missing instances become recorded unlocated alerts, not fabricated endpoints or silently discarded entries.
- Unknown severity remains `unknown` with a warning. Numeric strings and numeric scanner fields use the documented supported representation.
- Keep observation ordering deterministic. Preserve the existing documented receipt-order approach unless a deliberate versioned migration changes it. Do not depend on random UUID ordering to choose between observations received together.
- Preserve operator provenance separately from metadata merely declared by the report. Unknown operator fields stay null.
- Re-uploading identical bytes cannot silently replace provenance. Preserve the existing documented duplicate/conflict semantics and test them.
- Do not change the fingerprint algorithm in place. A change needs an explicit version and compatibility/migration decision.

## 6. HTTP contract

The implemented `contracts/openapi.yaml` is the transport source of truth. Update it with code and regenerate frontend types. The root AGENTS.md describes the full planned interface; do not expose fake-success stubs for future routes.

### Initial endpoint groups

| Group | Routes | Purpose |
| --- | --- | --- |
| Process | `GET /healthz`, `GET /readyz` | Liveness and database/schema readiness |
| Targets | `POST /api/v1/targets`, `GET /api/v1/targets`, `GET /api/v1/targets/{id}` | Server-registered lab identities |
| Import | `POST /api/v1/targets/{id}/imports/zap`, `GET /api/v1/imports/{id}` | Atomic normalization and provenance |
| Private evidence | `GET /api/v1/imports/{id}/raw` | Explicit, acknowledged, audited access |
| Findings | `GET /api/v1/targets/{id}/findings`, `GET /api/v1/findings/{id}` | Filtered lists and finding detail |
| Observations | `GET /api/v1/findings/{id}/observations` | Paged source observations |

Review, context, assessment, ranking, export, and scan routes belong to their later milestones. Check actual implementation status before describing any route as available.

### Request and response rules

- Authenticate every route except health and readiness using the configured bearer token.
- Reject unknown/duplicate mutation fields, invalid canonical IDs, malformed query encoding, repeated query parameters, and trailing JSON.
- Bind opaque cursors to the active filter set and deterministic ordering.
- Return empty lists as `items: []` and no next page as `next_cursor: null`.
- Return missing assessments as null, not zero or a default vector.
- Never return raw report bytes, raw private URIs, credentials, or unstructured payloads implicitly in list/detail responses.
- Use no-store for sensitive responses. Raw reports are attachments with explicit acknowledgement and an audit event.
- Keep health independent of database availability; readiness checks the required schema and runtime access.

### Errors

Use `application/problem+json` with `type`, `title`, `status`, `code`, `detail`, `request_id`, and an `errors` array. Field errors include a pointer, code, and safe message.

| Status | Meaning |
| --- | --- |
| 400 | Malformed input, query, cursor, or ID |
| 401 | Missing or invalid authentication |
| 404 | Resource or route not found |
| 409 | Duplicate semantic conflict, stale revision, or invalid lifecycle transition |
| 413 | Request, field, count, or depth limit exceeded |
| 415 | Unsupported media type |
| 422 | Structurally valid request with invalid domain/report semantics |
| 429 | Enforced resource/concurrency budget; include Retry-After |
| 502 | Later: scanner communication failure where applicable |
| 503 | Readiness unavailable |
| 500 | Unexpected internal failure |

Do not expose SQL, internal topology, payloads, tokens, or stack traces in error messages. Log enough safe context to diagnose failures using request ID, operation, and stable error classification. Never log raw database errors without reviewing their potentially sensitive details.

## 7. Database and consistency

Implement only the tables required by the current milestone. The initial import model needs targets, report imports, findings, observations, and audit events. Use the full model in AGENTS.md as later capabilities are introduced.

Required invariants:

- Unique import identity: target + report SHA-256 + parser version.
- Unique finding identity: target + fingerprint version + fingerprint.
- Unique observation source: import + source pointer.
- A finding and its observations/imports belong to the same target, enforced by appropriate foreign keys.
- Original evidence and audit records are append-only for the runtime role.
- Database/schema ownership belongs to a separate migration role.
- Query indexes support the actual filter and ordering paths; inspect query plans before adding speculative indexes.
- Review/context revisions and assessment/ranking snapshots are immutable once inserted.
- Conflicting predecessor IDs fail with 409; they must not silently overwrite another revision.
- New context or reviews do not rewrite historical scores or rankings.

Never reset an existing developer or operator database as a test convenience. Use disposable database clusters or isolated test resources. Destructive lab reset is a separate explicit operator action.

## 8. Context, assessments, and rankings — later milestones

Follow the exact enums and payloads in AGENTS.md. Resolve exact endpoint context before target defaults; unresolved factors remain unknown. Snapshot factor values, provenance, source revisions, and resolver version.

Keep metric selection separate from numerical calculation. Analysts justify legal CVSS v4 metric choices; the server calculates using a pinned, validated adapter. The client cannot submit a trusted score. Complete BE assessments reference the same evidence/review state as their completed B parent.

Return null scores/vectors for incomplete drafts. Store calculator provenance and replayable vectors for completed assessments. Do not automatically map context enums to invented numeric weights.

Freeze each ranking's cohort, cutoff, evidence states, selected assessment IDs, exclusions, scores, and tie handling. Technical and contextual comparisons use the identical eligible cohort and paired Base assessments. Python reads frozen exports; it does not mutate the live database or recalculate CVSS.

## 9. Scanning, scope, and secrets

Offline import must never fetch or resolve report URLs. No preview, evidence rendering, context edit, or parser operation may create an outbound connection.

Later scanning remains disabled until the root contract's authorization and network isolation gates are met. Server-owned target registration is not scan authorization. Check an explicit valid authorization before scheduling, before network activity, and during the job lifecycle.

Scope checks must constrain ZAP's actual traffic, including redirects, DNS resolution, crawler/browser requests, and alternate destinations. A Go allowlist alone cannot prove that. Use the authorization, network policy, controlled sink tests, leases, cancellation, budgets, and stop-confirmation rules in AGENTS.md.

Keep tokens, credentials, raw reports, and database dumps outside Git. Use placeholders in examples and generated local secrets. Bind exposed local ports to loopback. Never pass scanner secrets to the frontend or mount the Docker socket into the API/worker.

## 10. Testing and delivery order

| Stage | Deliver | Required proof |
| --- | --- | --- |
| M0 | Reproducible service/database foundation | Startup, migration, auth, readiness, least privilege |
| M1 | Offline import and findings API/UI | Genuine provenance-backed report plus normalization/persistence tests |
| M2 | Verification, mapping, context revisions | History, precedence, evidence references, conflict behavior |
| M3 | CVSS-B and CVSS-BE | Reference-calculator parity and complete traceability |
| M4 | Frozen rankings and exports | Cohort consistency and deterministic replay |
| M5 | Authorized isolated scanning | Actual scope confinement and reliable job cancellation/recovery |
| M6 | Independent research evaluation | Frozen protocol, real reference judgments, reproducible analysis |

For substantive changes, run the relevant unit tests and real PostgreSQL integration tests. Parser/scope changes need malformed and boundary cases. Concurrency and append-only guarantees require database-level tests. Tests must not initiate scans.

At minimum, verify current import behavior for malformed reports, limits, cross-origin data, NUL/database representability, unusual query values, duplicate imports under concurrency, provenance conflict, transaction rollback, ordering, private download audit, and runtime role restrictions.

Run applicable formatting, lint, generation checks, frontend type/build checks, and browser tests when API behavior changes. Independently verify Compose startup rather than treating configuration parsing as a successful boot. Document tested host architectures and unverified environments.

CI uses isolated resources, locked dependencies, and no real secrets. Account for PostgreSQL's non-root execution requirement on runners. A workflow file is not evidence of a passing workflow; record the observed run when available.

Update README and `docs/status.md` with exact checks, results, artifact provenance, and blockers. Mark synthetic tests clearly. A genuine-report gate remains open until it has real evidence. Coordinate with the reviewing agent through the existing checkpoint files without editing the reviewer's notes.

## 11. Definition of done

A backend change is ready when its contract and migrations agree, domain invariants are enforced, meaningful success/failure tests pass, secrets and raw evidence stay protected, and actual verification results are recorded. Planned features, skipped checks, and external blockers remain explicitly labeled.
