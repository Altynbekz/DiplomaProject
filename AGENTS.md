# AGENTS.md — Diploma implementation instructions

Project: **Development of a Context-Aware Web Application Vulnerability Assessment System with OWASP Top 10 Risk Prioritization**

## 0. Status, authority, and operating rules

This document defines the implementation plan and contract. Check the current checkout and [implementation status](docs/status.md) for working features, pending work, and verification results. Preserve unrelated files, including any QR images. Mark a milestone complete only when its acceptance gates have executable evidence.

Use the [frontend guide](frontend/FRONTEND.md) and [backend guide](backend/BACKEND.md) for layer-specific guidance. Review checkpoints and independent results live in [coordination status](docs/coordination/status.md) and the [verification log](docs/coordination/review.md).

These instructions apply to coding agents working anywhere in this repository. Follow the user's current requirements and applicable nested instructions. Read the actual files and working tree before each task; preserve unrelated changes. Never infer implementation status from this document alone.

Work in small, runnable vertical slices. Implement only the current milestone and its prerequisites, then report what runs, what was tested, and what remains. If a required tool, real report, reviewer judgment, or authorization is missing, identify that specific gap; do not generate evidence to conceal it. A mock, synthetic fixture, illustrative screenshot, or predicted result is never a real lab measurement.

**First user-visible deliverable:** import a genuine OWASP ZAP Traditional JSON report from an authorized local Juice Shop lab, persist its normalized findings in PostgreSQL, and display them in React. Do not build a scoring engine or scan launcher before this path works. A user-supplied genuine report can satisfy this milestone without running a new scan. An agent may scan only after the explicit lab authorization and network isolation gates in §10 are satisfied.

## 1. Purpose, scope, and non-goals

Build a local academic prototype that preserves scanner evidence, supports human verification and contextual assessment, and compares technical-only with contextual vulnerability rankings. Research question: does contextual assessment agree more closely with independent reference prioritization for the studied scenarios? Improvement is a hypothesis, not a promised outcome.

Only these OWASP Top 10 **2025** categories are in the research scope:

| Stable identifier | Name |
| --- | --- |
| `A01:2025` | Broken Access Control |
| `A02:2025` | Security Misconfiguration |
| `A05:2025` | Injection |

These identifiers follow the [official OWASP 2025 list](https://top10.owasp.org/). A category is a weakness classification, not a numeric severity or proof of exploitation. Preserve other or unmapped scanner alerts; label them outside study scope rather than forcing them into these categories. Do not claim ZAP covers all weaknesses in a category. Access-control flaws often require manually verified role/object scenarios.

Non-goals: Internet or production scanning; arbitrary target discovery; multi-tenant SaaS; autonomous exploitation; automatic remediation; comprehensive OWASP coverage; a new CVSS formula; CVSS v3-to-v4 conversion; CVE detection claims from generic alerts; proof that a real business is vulnerable; enterprise identity management; Kubernetes; distributed queues; LLM-generated verification or scores. Manual checks remain lab-only and must record authorization and evidence.

## 2. Planned monorepo and boundaries

Create directories when needed; do not fill the tree with empty abstractions.

```text
.
├── AGENTS.md
├── README.md                       # actual setup, status, demo and limitations
├── Makefile
├── .env.example                    # placeholders only
├── .gitignore
├── compose.yaml
├── backend/
│   ├── go.mod / go.sum
│   ├── cmd/api/main.go
│   ├── cmd/worker/main.go           # introduced at scan milestone
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/                  # value types, enums, validation
│   │   ├── httpapi/                 # routes, DTOs, middleware, problem responses
│   │   ├── service/                 # imports, findings, context, assessments, ranks
│   │   ├── store/                   # repositories/transactions using sqlc
│   │   │   └── sqlc/                # generated; never edit by hand
│   │   ├── zap/                     # report parser and scanner adapter
│   │   ├── scope/                   # URL policy; no scanner dependency
│   │   ├── cvss/                    # validated scoring adapter
│   │   └── jobs/                    # PostgreSQL worker state/leases
│   ├── db/
│   │   ├── migrations/              # goose numbered SQL migrations
│   │   ├── queries/                 # named sqlc queries
│   │   └── sqlc.yaml
│   └── testdata/zap/                # sanitized real fixture + provenance
├── frontend/
│   ├── package.json / package-lock.json
│   ├── src/
│   │   ├── app/                    # router/providers
│   │   ├── api/                    # generated types + typed fetch client
│   │   ├── features/               # targets/imports/findings/assessments/rankings
│   │   └── components/             # reusable UI and accessibility primitives
│   └── tests/
├── contracts/openapi.yaml
├── configs/
│   ├── lab-targets.example.yaml
│   ├── owasp-mapping.v1.yaml
│   └── context-policy.v1.yaml
├── lab/
│   ├── zap/                        # versioned automation plans and scan policies
│   ├── scenarios/                  # lab reproductions; no real credentials
│   └── README.md                   # isolation, authorization, reset, stop procedure
├── analysis/
│   ├── pyproject.toml / uv.lock
│   ├── src/diploma_analysis/
│   ├── tests/
│   └── notebooks/                  # optional presentation, no hidden computation
├── tests/integration/
├── tests/e2e/
├── docs/
│   ├── status.md                   # implemented/planned/blocked with evidence
│   ├── decisions/                  # numbered architecture decision records
│   ├── research-protocol.md
│   ├── fixture-provenance.md
│   └── threat-model.md
└── artifacts/                      # ignored raw reports, exports, experiment runs
```

Architecture: React → Go `net/http` API → application services → sqlc repositories → PostgreSQL. A separate Go worker reads PostgreSQL scan jobs, talks to a fixed ZAP service, and imports output through the same normalizer. PostgreSQL stores metadata, bounded raw report bytes and immutable snapshots; no object store is needed for this prototype. Python consumes versioned exports, never the live production schema. No network calls from report parsing, categorization, context resolution, or scoring.

Handlers decode and validate transport input, call services, and serialize DTOs. Services own transactions, authorization checks, state changes and domain rules. Repositories own SQL. ZAP transport types and generated sqlc structs must not escape to the public API. Keep interfaces at consumers and only where tests or adapters need them. Avoid a generic repository, ORM, GraphQL, microservices or a dependency injection framework.

## 3. Exact development conventions

- Backend: Go modules, standard `net/http` method/path routing, `context.Context` first in I/O methods, `pgx/v5`, sqlc, goose, `log/slog`. No Gin/Echo/GORM. Pin a supported Go version in `go.mod`, tool versions in a checked-in tool manifest, and record them in README when scaffolding; do not guess version availability.
- Frontend: React, TypeScript `strict`, Vite, npm with committed lockfile, React Router, TanStack Query for server state. Use native forms and a small component set. Validate mutations on the server; browser validation is supplementary. Generate TypeScript request/response types from OpenAPI with a pinned tool. Do not duplicate scoring or scope logic in the browser.
- Python: supported Python version pinned in `pyproject.toml`, uv lockfile; numpy/pandas/scipy and matplotlib as needed. CLI modules produce analysis outputs; notebooks must call those modules.
- PostgreSQL and all container images: pin exact versions and deployment image digests once selected. No `latest` in reproducible runs. Use Compose v2 and health checks; document supported local platform.
- SQL `snake_case`; HTTP JSON `snake_case`; Go exported identifiers idiomatic PascalCase; TS variables camelCase and components PascalCase; file names kebab-case in frontend and snake_case in Python. UUIDs are application-generated v4; timestamps are UTC RFC3339 over HTTP and `timestamptz` in DB. Use the real clock in production and injectable clock in tests.
- SQL migrations `00001_description.sql` with goose Up/Down sections. Never edit an applied migration; add a new one. Use text + CHECK constraints for domain enums. Parameterize SQL, use explicit columns, include stable ORDER BY, and avoid `SELECT *`. Generated files carry generated headers and are committed. Regenerate and require a clean diff.
- Never serialize an absent score as zero. Numeric scores use `numeric(3,1)` in DB, validated finite decimal JSON values from 0 through 10. Ordering and equality use the stored official rounded value. Do not recompute in TS/Python.
- English code/API/docs. Accessible UI labels may be localized later. Keyboard access, visible focus, labeled forms, readable tables, and loading/empty/error states are required. Never use color alone for severity.
- No secrets, real personal information, hard-coded business judgments, arbitrary numeric risk multipliers, shell interpolation of user input, `dangerouslySetInnerHTML`, global mutable DB handles, `panic` for operational errors, or swallowed errors. Escape scanner HTML as plain text.
- Configuration is validated at startup. `.env.example` contains placeholders; secret values are generated locally. Ignore `.env`, credentials, DB dumps, raw scan output and experiment participant identifiers. Fixture sanitization must be reproducible and documented.

Create these Make targets alongside the feature that needs them, and distinguish host dependencies from container commands in README:

```text
make dev                 # db+migrations+API+frontend; never launches a scan
make migrate-up          # explicit schema upgrade
make generate            # sqlc + OpenAPI TS types
make fmt                 # Go + frontend formatting
make lint                # go vet + TS/ESLint + Python lint when introduced
make test                # deterministic unit tests; no network scanning
make test-integration    # disposable PostgreSQL; migrations + repository/API tests
make test-e2e            # browser tests against isolated test stack
make lab-up              # opt-in isolated Juice Shop + ZAP profile only
make lab-down            # stop lab; retain evidence by default
make analysis            # deterministic replay from an explicit frozen export
make check               # fmt-check, lint, generation diff and applicable tests
```

Do not let `make test`, `make dev`, migrations, or CI launch a security scan. Destructive database reset must be a separate clearly named command restricted to the test/lab database.

## 4. Domain semantics: keep these concepts separate

1. **Report/evidence:** original bytes, digest, provenance, parser version, extracted observations. Evidence can contain false positives. Retain original lab report bytes privately and show redacted excerpts by default. A sanitized committed derivative has a different digest and an explicit transformation record.
2. **Finding:** a stable grouping of related observations at one target endpoint/parameter. Finding identity is not proof that two reports describe the same root cause.
3. **Verification:** append-only human review `unreviewed | confirmed | false_positive | inconclusive`. New findings start unreviewed. Record procedure, actor, time and evidence references. Scanner confidence cannot change verification automatically. “Confirmed” requires a reproducible lab procedure.
4. **Classification:** versioned OWASP 2025 mapping with source CWE/plugin rule, rationale, reviewer and evidence. Separate from verification. Mapping can be `mapped | unmapped | ambiguous | out_of_scope`. Ambiguous findings stay excluded from category-specific claims until reviewed.
5. **Scanner severity/confidence:** original ZAP values and normalized enum on each observation. ZAP risk codes 0/1/2/3 map to informational/low/medium/high. Missing/unrecognized codes become `unknown`, with import warnings; never invent critical. Preserve differing severities across scans rather than overwriting history.
6. **CVSS assessment:** analyst-selected metrics with justification, independently calculated numeric output, exact vector and calculator provenance. Store Base (`CVSS-B`) and contextual Base+Environmental (`CVSS-BE`) separately as related assessments. An incomplete assessment has no numeric score.
7. **Ranking:** immutable ordered set of explicitly chosen assessment snapshots for an eligible cohort. No generic `risk_score` field that conflates these concepts.

Use the [FIRST CVSS v4 specification](https://www.first.org/cvss/v4.0/specification-document) for metric validation and score nomenclature. `CVSS-BE` includes Base with Environmental choices; it is not an independent environmental subscore. Threat is not assessed in this project; keep it Not Defined and do not label results BT/BTE. Supplemental information does not change the score. CVSS describes severity, not total business risk.

A ZAP risk level, CWE, OWASP category, finding title or context enum is insufficient to infer a complete CVSS vector. Do not equate a scanner alert with a CVE. Unknown values stay unknown even when a score would improve the demo.

## 5. Normalization and real-report import

Support **ZAP Traditional JSON** first, per its [documented report shape](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/). Distinguish its nested site/alert/instance report from other ZAP JSON formats; reject unsupported envelopes with a useful code. Accept documented numeric fields encoded as strings or numbers. Unknown extension fields in scanner JSON are retained in raw evidence, not rejected as API request fields would be.

- Read multipart upload with a hard 20 MiB total request limit, at most 10 sites, 5,000 alert objects and 20,000 instances; field text limit 64 KiB, parser nesting limit 64. Require UTF-8 JSON and a recognized report envelope. Enforce limits before large allocations; reject trailing JSON documents. Reject malformed required URI/method fields atomically, reporting JSON pointers. Never silently discard instances. Alerts with no instances persist as report-level `unlocated_alerts` and trigger a warning, not fabricated endpoints.
- Compute SHA-256 over original report bytes. Store upload bytes and normalization result in one transaction; roll back the entire import on structural or scope errors. If a limit prevents normalization, return a limit error with no persisted partial success. Import is synchronous in M1; it is not a scan job.
- M1 supports one canonical target origin per import. All site origins and instance URLs must belong to it. Reject cross-origin reports with 422 and site/instance pointers; do not silently select one site. Offline URL validation never resolves DNS or fetches the URL.
- Normalize method to uppercase. Canonical origin = lowercase ASCII host + explicit effective port + scheme. Normalize only RFC-unreserved path escapes, preserve case and significant path/trailing slash; reject invalid escapes, fragments and credentials. Preserve raw URI privately. Endpoint identity excludes query values, includes sorted unique query parameter names. Do not collapse numeric path segments or infer templates. Endpoint `path` rules use the actual normalized path.
- Finding fingerprint v1 = SHA-256 of a canonical JSON array `[target_id, scanner, plugin_id, method, canonical_origin, normalized_path, query_parameter_names, parameter_name]`. Use deterministic UTF-8 serialization and golden cases. Missing plugin ID is a structural error for an alert with instances. Empty parameter name is valid. Keep attack/evidence variations as observations, not additional findings. Document possible conflation and allow a future reviewed split migration.
- Each normalized instance retains source JSON pointer, instance index, plugin/alert reference, CWE if present, source severity/confidence, raw location, parameter, attack, evidence and other information. Descriptions and remediation are scanner assertions; render escaped plain text.
- Enforce uniqueness of `(target_id, report_sha256, parser_version)` for imports; concurrent duplicate uploads return the existing import and create no duplicate observations. Parser upgrades may create a new normalization run; never overwrite the old run. Finding upsert uniqueness is `(target_id, fingerprint_version, fingerprint)`.
- Require a genuine locally generated or user-provided report plus provenance: ZAP and add-on versions where available, report template, Juice Shop image/version, lab authorization, scan time, command/plan digest, raw and sanitized hashes, sanitization steps. Record unknown metadata as null. Unit fixtures can be synthetic if clearly labeled; at least one integration fixture must be genuinely generated and sanitized.

Classification mapping is a reviewed data file, not substring matching against titles. Rules use ZAP plugin IDs and/or CWE identifiers with cited OWASP 2025 source, mapping version and rationale. Do not copy 2021 IDs into 2025. Conflicting rules yield `ambiguous`; the API reports all candidate categories. Human overrides append a classification record. Import every alert even if classification is unknown.

## 6. Context and assessment policy

Four context dimensions are required as recorded input, even when unknown:

| Field | Allowed values | Meaning |
| --- | --- | --- |
| `exposure` | `public`, `internal`, `restricted`, `unknown` | documented intended reachability |
| `required_privileges` | `anonymous`, `user`, `admin`, `unknown` | privileges needed for the documented attack |
| `data_sensitivity` | `public`, `internal`, `personal`, `highly_sensitive`, `unknown` | affected data with justification |
| `business_criticality` | `low`, `medium`, `high`, `critical`, `unknown` | consequence to the documented lab scenario |

Context is a scenario assertion, not a fact inferred from the private lab IP. A simulated public banking endpoint can be hosted privately; record scenario context separately from actual scan scope. Each factor has rationale, source/evidence references, author and timestamp. If a factor is unknown, record why and mark contextual eligibility incomplete.

Context rules are immutable revisions scoped to a target and either target default (`method=null,path=null`) or exact endpoint (`method` and `path` both set). No regex, glob, query-value matching or wildcard host in v1. Resolve exact method+path first, otherwise target default, otherwise all unknown. Store the resolved factor values, rule revision IDs and resolver version in every assessment snapshot. New context revisions affect future assessments only; history is never recalculated in place.

Implement `context-policy.v1` as a documented **review aid**, not automatic numeric scoring:

- Exposure prompts review of environmental attack feasibility; do not translate `internal` mechanically into a CVSS network value.
- Required privileges prompts review of the attack's actual prerequisites, distinct from the assessor's login role.
- Sensitivity prompts justified confidentiality requirements; business criticality prompts integrity/availability and other relevant requirements.
- No one-to-one enum mapping or multiplicative weight is approved by default. The assessor selects appropriate Environmental metrics, records per-metric reasoning and references the context factors used. Every factor must have a documented disposition, including “no CVSS metric change justified.”
- Record non-CVSS business consequences as qualitative notes. Do not relabel them an official CVSS score. A future custom priority index needs a separate ADR, preregistered formula, validation and distinct name; it is outside the initial study.

Assessment draft input uses metric dictionaries with legal FIRST values. Base assessment requires every mandatory Base metric with evidence/rationale. For BE, copy the exact completed Base snapshot, then supply reviewed Environmental choices. Unknown Base metrics mean incomplete; Not Defined is not a replacement for a missing mandatory Base value. An unknown project context factor blocks finalization of BE until reviewed. Omitted optional Environmental metrics explicitly become Not Defined in the snapshot, with inheritance visible. BE finalization requires at least one explicitly defined Environmental metric; otherwise report it as B and reject a misleading BE label.

Scoring adapter: select and pin a licensed CVSS v4 implementation; compare it against the [FIRST reference calculator](https://github.com/FIRSTdotorg/cvss-v4-calculator) at a recorded commit. If no acceptable Go package is available, implement a narrowly scoped, licensed port behind the adapter and prove parity; do not substitute a homemade formula. No external scoring web service. Store calculator name/version/source commit, canonical input/output vector, metric justifications, score, spec version and fixture version. Clients submit metrics, never trusted scores or computed vector strings. The service generates and validates the canonical vector and calculates the score. Drafts have `score=null, vector=null`; complete snapshots are immutable.

## 7. Planned database contract

All tables use UUID primary keys unless a composite key is specified. All creation times are `timestamptz NOT NULL`. Use FK constraints, bounded text/JSON, CHECK constraints, transactions and indexes in migrations. JSONB snapshots are schema-versioned and validated in the service; do not use JSONB instead of relational identity. Required columns below are the minimum schema, not preexisting tables.

| Table | Required columns and invariants |
| --- | --- |
| `targets` | `id,name,canonical_origin,lab_target_key,created_at`; origin unique, no arbitrary scan host |
| `scope_authorizations` | `id,target_id,scope_snapshot_json,scope_sha256,authorized_by,authorization_ref,valid_from,valid_until,revoked_at,created_at`; expiry required; authorization reference is not a credential |
| `scan_runs` | `id,target_id,authorization_id,mode,status,policy_version,config_snapshot_json,zap_run_id,created_at,started_at,finished_at,lease_owner,lease_until,attempt_count,cancel_requested_at,error_code`; mode passive/active, status machine in §9 |
| `report_imports` | `id,target_id,scan_run_id nullable,report_sha256,parser_version,raw_bytes bytea,source_metadata_json,warnings_json,unlocated_alerts_json,created_at`; unique target/hash/parser; raw bytes access restricted; normalized import is immutable |
| `findings` | `id,target_id,fingerprint_version,fingerprint,plugin_id,method,canonical_origin,path,query_parameter_names text[],parameter_name,title,created_at`; unique target/version/fingerprint; title never establishes identity |
| `observations` | `id,finding_id,report_import_id,source_pointer,instance_index,raw_uri,scanner_severity_raw,scanner_severity,confidence_raw,cwe_ids int[],evidence_json,description,remediation,created_at`; unique import/pointer; finding and import must share target |
| `classification_revisions` | `id,finding_id,status,category nullable,candidate_categories text[],mapping_version,rule_refs_json,rationale,evidence_refs_json,actor,created_at`; category limited to study IDs when mapped, null otherwise; append-only |
| `verification_reviews` | `id,finding_id,state,procedure,rationale,evidence_refs_json,actor,created_at`; append-only; no entry means unreviewed |
| `context_revisions` | `id,target_id,method nullable,path nullable,revision,exposure,required_privileges,data_sensitivity,business_criticality,factor_provenance_json,actor,created_at`; default/endpoint CHECK; unique target/scope/revision with NULL-aware uniqueness |
| `assessment_snapshots` | `id,finding_id,kind,status,base_assessment_id nullable,verification_review_id,classification_revision_id,context_snapshot_json,metrics_json,metric_rationales_json,vector nullable,score nullable,calculator_json,policy_version,actor,created_at`; kind B/BE, status incomplete/complete; complete requires score+vector; BE references a complete B of same finding; immutable |
| `ranking_snapshots` | `id,target_id,mode,eligibility_policy_version,cohort_sha256,cutoff_at,exclusions_json,created_by,created_at`; mode technical/contextual; explicit candidate finding IDs and selected snapshot IDs frozen |
| `ranking_items` | composite PK `(ranking_id,finding_id)`; `assessment_id,score,tie_rank,display_order`; FK same finding; unique ranking/display_order; score must equal selected complete snapshot |
| `audit_events` | `id,request_id,actor,action,entity_type,entity_id,metadata_redacted_json,created_at`; append-only and no raw evidence/secrets |

Add target-consistency composite foreign keys or transactional service checks plus integration tests where composite constraints are impractical. Index report target/time, observations finding/import, reviews finding/time/id, context target/scope/revision and job status/lease. Use `(created_at,id)` for deterministic latest-review selection. Serialize context revision allocation with a target row lock. For v1, serialize new review/classification and assessment finalization on the finding row; reject stale review inputs with 409. Finalization verifies referenced reviews are current and confirmed/mapped. For BE, also require its B parent to reference those same verification and classification revisions; otherwise require a new Base assessment rather than pairing different evidence states. Later reviews do not mutate snapshots; display stale assessments as historical.

Use a least-privileged runtime DB role; a separate migration role owns DDL. Prevent UPDATE/DELETE of evidence/review/assessment/ranking tables for the runtime role. Do not expose generic deletes. Lab resets use an explicit maintenance role and procedure. Back up evidence before a requested reset; do not assume retention means permission to keep real secrets.

## 8. Exact HTTP API contract (implement incrementally)

`contracts/openapi.yaml` is the source of truth and must encode every implemented route, enum, nullable field, response and problem code. Update it before or with implementation. Planned endpoints stay marked planned in README until delivered. Breaking contract changes need an ADR and migration notes. No unimplemented endpoint should return fake success.

### Transport and security

- Base path `/api/v1`, JSON UTF-8; UUID path IDs; RFC3339 UTC timestamps. Mutation JSON rejects unknown fields, duplicate keys and trailing data. Multipart scanner bytes follow §5. JSON request cap 1 MiB unless specified. Lists default `limit=50`, max 200. Opaque cursor encodes stable sort key + active filter hash; bad/mismatched cursor → 400. No unbounded list endpoints.
- Local single-operator deployment only. Protected endpoints require `Authorization: Bearer <APP_API_TOKEN>` from a locally generated secret (minimum 32 random bytes). Compare safely; do not log. Browser asks for token, holds it in memory, clears on reload/logout, never embeds in bundle/localStorage/query string. Same-origin frontend through dev proxy/reverse proxy. No cookie auth, wildcard CORS or public registration. All mutations record the authenticated operator identifier from server configuration, never a client-supplied `actor`.
- Allow unauthenticated `/healthz` and `/readyz` only. Bind host ports to `127.0.0.1`; protect API even locally. No Internet deployment within this scope.
- Errors: `application/problem+json` `{type,title,status,code,detail,request_id,errors:[{pointer,code,message}]}`. `type` is `urn:diploma:problem:<code>`. `detail` is safe for display. Use a bounded generated/validated `X-Request-ID` and return it. Never include SQL errors, stack traces, credentials, payload bodies or internal network topology.
- 400 malformed JSON/query/UUID (`invalid_request`); 401 missing/bad token (`unauthorized`); 404 unknown ID (`not_found`); 409 stale review/context/base or lifecycle conflict (`conflict`); 413 byte/count/depth limits (`limit_exceeded`); 415 unsupported content type (`unsupported_media_type`); 422 valid structure but invalid semantics (`validation_failed`, `unsupported_report`, `out_of_scope`, `incomplete_assessment`); 429 resource budget/concurrent job cap (`rate_limited`, include Retry-After); 502 ZAP transport error (`scanner_unavailable`); 503 DB not ready (`not_ready`); 500 unexpected failure (`internal_error`). Log internal causes with request ID and redaction. Context cancellation rolls back work and is not presented as success.

### Resource shapes

Fields below are required unless marked `?`; `null` is distinct from omission. Responses never include `raw_bytes` or unredacted `raw_uri` implicitly.

```typescript
// Contract notation, to be encoded as OpenAPI schemas; not an existing TS file.
type Category = 'A01:2025' | 'A02:2025' | 'A05:2025';
type Target = { id:string; name:string; canonical_origin:string;
  lab_target_key:string; created_at:string };
type ImportResult = { id:string; target_id:string; report_sha256:string;
  parser_version:string; duplicate:boolean; finding_count:number;
  observation_count:number; unlocated_alert_count:number;
  warnings:{code:string; pointer:string; message:string}[]; created_at:string };
type FindingSummary = { id:string; target_id:string; title:string;
  method:string; origin:string; path:string; parameter_name:string;
  query_parameter_names:string[]; observation_count:number;
  scanner_severity:'informational'|'low'|'medium'|'high'|'unknown';
  severity_observation_id:string;
  verification:'unreviewed'|'confirmed'|'false_positive'|'inconclusive';
  verification_review_id:string|null;
  classification:{revision_id:string|null; status:'mapped'|'unmapped'|'ambiguous'|'out_of_scope';
    category:Category|null; candidate_categories:Category[]; mapping_version:string|null};
  latest_base:AssessmentSummary|null; latest_environmental:AssessmentSummary|null };
type AssessmentSummary = {id:string; kind:'B'|'BE'; status:'incomplete'|'complete';
  score:number|null; vector:string|null; stale:boolean; created_at:string};
type Page<T> = {items:T[]; next_cursor:string|null};
```

Summary scanner severity is taken from the most recent observation `(created_at,id)` within the active import filter, labeled “latest scanner observation,” with its observation ID. It is not the maximum or a replacement for observation history. `latest_base/environmental` means latest snapshot by kind, with incomplete state visible; do not silently select an older complete one. A rank request must supply explicit complete assessment IDs.

### Routes and payloads

| Route | Request | Success |
| --- | --- | --- |
| `GET /healthz` | none | 200 `{"status":"ok"}` without DB call |
| `GET /readyz` | none | 200 `{"status":"ready"}` after DB/schema checks; otherwise 503 |
| `POST /api/v1/targets` | `{name,lab_target_key}`; resolve origin from server lab registry, never caller URL | 201 `Target`, Location header |
| `GET /api/v1/targets` | `limit,cursor` | 200 `Page<Target>` |
| `GET /api/v1/targets/{id}` | none | 200 `Target` |
| `POST /api/v1/targets/{id}/imports/zap` | multipart field `report` exactly once; optional bounded `provenance` JSON with known scan metadata | 201 `ImportResult`; exact duplicate 200 with `duplicate=true` |
| `GET /api/v1/imports/{id}` | none | 200 `ImportResult` with `duplicate=false` as stored resource representation |
| `GET /api/v1/imports/{id}/raw` | explicit `X-Evidence-Acknowledgement: private-lab-data` | 200 original bytes as attachment; audited, `Cache-Control: no-store`; protected |
| `GET /api/v1/targets/{id}/findings` | `limit,cursor,import_id?,category?,verification?,scanner_severity?`; stable finding ID ascending | 200 `Page<FindingSummary>`; filters compose with AND; unmapped included without category filter |
| `GET /api/v1/findings/{id}` | none | 200 `FindingSummary` plus `plugin_id,fingerprint_version,fingerprint,created_at`; histories via separate routes |
| `GET /api/v1/findings/{id}/observations` | `limit,cursor`; creation time/ID ascending | 200 page of `{id,import_id,source_pointer,method,path,parameter_name,scanner_severity,scanner_severity_raw,confidence_raw,cwe_ids,description,remediation,evidence_excerpt,redactions,created_at}`; excerpt redacted and bounded to 4 KiB |
| `POST /api/v1/findings/{id}/verifications` | `{state,procedure,rationale,evidence_refs,expected_previous_id}`; previous ID nullable | 201 immutable review `{id,finding_id,state,procedure,rationale,evidence_refs,actor,created_at}` |
| `GET /api/v1/findings/{id}/verifications` | pagination | 200 review page, newest first |
| `POST /api/v1/findings/{id}/classifications` | `{status,category,candidate_categories,rationale,evidence_refs,expected_previous_id}` | 201 revision `{id,finding_id,status,category,candidate_categories,mapping_version,rule_refs,rationale,evidence_refs,actor,created_at}`; manual mapping_version=`manual-v1` |
| `GET /api/v1/findings/{id}/classifications` | pagination | 200 revision page, newest first |
| `POST /api/v1/targets/{id}/contexts` | `{method:null|string,path:null|string,factors,factor_provenance,expected_previous_id}` | 201 full immutable context revision with ID/revision/actor/time |
| `GET /api/v1/targets/{id}/contexts` | pagination | 200 context revision page, newest first |
| `GET /api/v1/findings/{id}/resolved-context` | none | 200 `{factors,factor_provenance,source_revision_ids,resolver_version,resolved_at}`; no DB writes |
| `POST /api/v1/findings/{id}/assessments` | assessment request below | 201 full assessment snapshot; incomplete permitted only with `finalize=false` |
| `GET /api/v1/findings/{id}/assessments` | pagination | 200 snapshot page, newest first |
| `GET /api/v1/assessments/{id}` | none | 200 full immutable snapshot |
| `POST /api/v1/targets/{id}/rankings` | `{mode,candidate_finding_ids,assessment_ids,cutoff_at}`; max 2,000 candidates | 201 frozen ranking response below |
| `GET /api/v1/rankings/{id}` | `limit,cursor` for items | 200 metadata plus item page, display order ascending |
| `POST /api/v1/targets/{id}/scans` | `{mode:'passive'|'active',authorization_id,policy_id}`; fixed server policies only | 202 `{id,target_id,status:'queued',created_at}` plus Location |
| `GET /api/v1/scans/{id}` | none | 200 scan state `{id,target_id,mode,status,created_at,started_at,finished_at,import_id,error_code,cancel_requested_at}` with nullable fields |
| `POST /api/v1/scans/{id}/cancel` | `{}` | 202 scan state after durable cancellation request; terminal scan 409 |
| `GET /api/v1/rankings/{id}/export` | none | 200 versioned reproducible sanitized JSON export; attachment |

Evidence references are typed `{observation_id,source_pointer,note}` arrays; require observation to belong to the same finding. Provenance sources for contextual scenario assertions can additionally use `{document_ref,sha256,note}` pointing to a lab document, never a URL fetched by the service. References are strings resolved by an operator; they are not filesystem paths to open.

Assessment request:

```json
{
  "kind": "BE",
  "finalize": false,
  "base_assessment_id": "<existing-complete-B-UUID>",
  "expected_verification_review_id": "<current-confirmed-review-UUID>",
  "expected_classification_revision_id": "<current-mapped-revision-UUID>",
  "expected_context_revision_ids": ["<resolved-revision-UUID>"],
  "metrics": {},
  "metric_rationales": {},
  "notes": "Draft; no score until all required evidence is entered"
}
```

For B, `base_assessment_id=null`, context revision IDs empty and `metrics` contains only Base choices. For BE, `metrics` contains only Environmental choices; inherit all Base choices from referenced B. Rationale map keys match supplied metric keys; each value `{reason,evidence_refs,context_factor_refs}`. Reject illegal metric names/values or attempts to set Threat. The response contains `id,finding_id,kind,status,base_assessment_id,verification_review_id,classification_revision_id,context_snapshot,metrics,metric_rationales,notes,vector,score,calculator,policy_version,actor,created_at`. `calculator=null` for uncalculated drafts. Finalization requires current confirmed verification, mapped in-scope classification, complete evidence and an eligible Base parent. Reassess by creating a new snapshot, never updating one.

Context `factors` is exactly the four fields in §6. `factor_provenance` has exactly those keys, each `{rationale,sources}`. Require both method/path null for default or both non-null for endpoint. Missing/stale expected revision IDs return 409 rather than overwriting another review.

Ranking rules: technical mode selects complete B; contextual mode selects complete BE paired to the same B used for comparison. Candidates must belong to the target, be confirmed and in scope at cutoff, and selected snapshots must match those reviews/context at cutoff. Reject foreign, duplicate or ambiguous assessment selections. Exclude ineligible candidates with explicit reason codes, not zero scores. Freeze metadata, candidate IDs, selected/rejected IDs and relevant histories at cutoff using a repeatable-read transaction. Current changes do not alter old rankings. Descending numeric score, ties share dense `tie_rank`; UUID ascending gives reproducible `display_order` only, not a claim of superior risk. Response: `{id,target_id,mode,cutoff_at,eligibility_policy_version,cohort_sha256,excluded:[{finding_id,reason}],items:[{finding_id,assessment_id,score,tie_rank,display_order}],next_cursor,created_at}`. Define `cohort_sha256` as the SHA-256 of canonical JSON containing sorted eligible finding IDs and their paired B assessment IDs; exclude mode-specific BE IDs and scores from this cohort hash. Keep selected snapshot IDs in the separately hashed export manifest. Paired-study cohort hashes must be identical.

All POSTs use DB transactions for related writes and audit events. Import idempotency is mandatory. Other POST retries may create additional immutable revisions; the client must not blindly retry mutations. Use optimistic predecessor checks for reviews/context. For scan scheduling, prevent more than one nonterminal scan per target with a partial unique index and return 409 on duplicate scheduling.

## 9. Scanner integration and lifecycle (later milestone)

ZAP is an adapter, not the domain model. Use the [ZAP Automation Framework](https://www.zaproxy.org/docs/automate/automation-framework/) with versioned fixed policies and its API only through a server-configured base URL. Never accept an API URL, raw YAML, arbitrary script, add-on installation or shell command from the frontend. No Docker socket in API or worker containers. Compose starts a dedicated scanner service; the worker controls jobs through a minimal fixed adapter. If isolated per-run containers require a separate local launcher, keep it outside API privileges and document the ADR.

State transitions:

```text
queued -> running -> importing -> completed
queued -> cancelled
running/importing -> cancelling -> cancelled
queued/running/importing/cancelling -> failed
```

Use PostgreSQL `FOR UPDATE SKIP LOCKED` to claim jobs, a 30-second lease and a 10-second heartbeat. At most one scan globally in v1. Default max scan runtime 10 minutes, crawler depth 3, total requests 1,000 and request rate 2/sec, with lower limits allowed. Verify that pinned ZAP policies and egress controls actually enforce each budget; a timeout alone is insufficient. Stop on budget exhaustion; record failure/reason and partial report provenance if retained, without treating it as completed coverage.

Persist the external run ID before polling. On lease loss/crash, reconcile with ZAP; do not replay active actions automatically. If ownership cannot be proved, stop the dedicated scanner run/service and mark failure for manual retry. Cancellation durably sets intent, stops ZAP work, waits for stop confirmation and prevents a completion race under a row lock; failure to stop is `failed` with an explicit error. Do not start a replacement while old scanner activity may continue. ZAP alert exit codes can signal findings, not necessarily infrastructure failure; distinguish policy outcome from transport/plan failure using documented pinned behavior.

Only a completed report uses the normal import service; same report hash remains idempotent. Passive scanning still sends crawler/proxy traffic and requires authorization. Store scanner/image/add-on/plan versions, limits, target, scope and authorization snapshots in the run. Manual role-based lab tests supplement ZAP for A01; lack of an alert is not absence of vulnerability.

## 10. Mandatory lab authorization, SSRF controls, and secrets

**Deny scans by default.** This file permits building the system; it does not authorize scanning arbitrary systems. Before any scan, the operator must explicitly authorize the named local lab target, modes and validity window in a local authorization record. A valid lab record may cover repeated runs; no repeated prompt is needed inside that scope. The scan API references this record and cannot create/extend it. Create records through an operator-only local setup command, validating the lab registry and recording an authorization reference, never through an unauthenticated web form. Recheck validity and revocation at enqueue, immediately before network activity, and on every worker heartbeat. Authorization expiry, cancellation or revocation stops affected jobs using the cancellation/stop-confirmation procedure.

Scope enforcement is two-layer: application URL policy **and** scanner network egress confinement. An API URL allowlist alone is not enough because ZAP itself resolves hosts, follows links, executes browser requests and sends injected requests.

1. Server-owned lab registry maps `lab_target_key` to one canonical HTTP(S) origin, exact effective port, allowed path prefixes (segment boundaries), dedicated lab network and allowed IP identities. v1 starts with only Juice Shop. No user-supplied destinations, wildcard domains, arbitrary CIDRs or scan-policy text.
2. Reject URL userinfo, unsupported schemes, fragments, malformed encoding/ports, ambiguous numeric IPv4 formats, IPv6 zone IDs and host aliases not registered. Normalize with the same tested Go parser for registration/import/scope. Check all A/AAAA answers and reject any forbidden answer, not only the first. Recheck at each connection and redirect; validate the actual peer IP. Disable implicit proxy environment variables and direct host escape paths.
3. Deny loopback, link-local, multicast, unspecified, metadata destinations (including `169.254.169.254`), host gateway, Docker management, API, DB and every destination outside the dedicated lab identity. Private address space is **not** automatically trusted. The explicit exception is the registered Juice Shop container IP/port on its dedicated isolated network. For IPv6, equivalent deny rules apply; disable it if not tested. Localhost host-port access from a browser does not imply scanner permission to scan loopback.
4. Never perform DNS lookups or HTTP requests on import, evidence rendering, report reference resolution, previews or context edits. URLs in reports are data. Redirects to unregistered origins/ports/paths are blocked. Block navigation, form actions, subresources, AJAX/WebSocket destinations and scanner-generated callback URLs outside scope, including browser-based crawling. Disable out-of-band callbacks and external telemetry for lab runs.
5. ZAP include/exclude context and “in scope only” controls are defense in depth. Build a verified network policy for the pinned scanner: a dedicated internal lab network plus IP/port egress enforcement at the scanner network namespace or equivalent allowlisting proxy that ZAP cannot bypass. `internal: true` alone does not prevent access to other containers on that network. Restrict scanner to the target port and necessary fixed control traffic; reject other origins before egress. If DNS pinning/redirect/browser enforcement cannot be demonstrated, keep launch disabled and use offline report import. Do not claim application-side Go URL checks govern ZAP's connections.
6. Compose profiles keep vulnerable Juice Shop and ZAP opt-in. No shared production network, host networking, privileged containers, host filesystem mounts or exposed scanner API. DB has no published port by default. Frontend/API loopback bindings only; Juice Shop may have an explicit loopback-only demo port. No Internet egress for scanner/lab. ZAP control secret available only to the worker/control channel; do not expose it in browser or logs.

SSRF tests must exercise both scope-unit cases and the actual ZAP path: redirects outside lab, disallowed port, mixed DNS answers/rebinding where supported, IPv4-mapped IPv6, metadata, localhost, alternate numeric IPs, cross-origin subresource and a link to another private container. Controlled sink servers must observe **zero unauthorized requests**. Run isolation tests before allowing active scans. Never test this by contacting a real metadata service or public victim.

Secrets: use local secret files or Compose secrets and a redacted `.env`; no committed tokens, passwords, auth headers or cookies. Authentication for Juice Shop uses disposable lab accounts; record role labels, never session values. Redact sensitive query values/headers/body excerpts in UI, logs and research exports. Keep raw reports in restricted DB storage with audited explicit access, no public static serving or automatic download. Use no-store on sensitive endpoints. Only sanitized synthetic-identity lab reports may become fixtures. Document manual retention/reset procedures and secret rotation after accidental disclosure. Do not “sanitize” the original silently: maintain original hash and derivative provenance.

## 11. Testing and concrete acceptance gates

Use Go `testing`, `httptest`, fuzz tests for parsers/scope, disposable PostgreSQL integration tests, Vitest/React Testing Library and Playwright. Use Python pytest for analysis. Unit tests isolate clock/IDs/calculator/scanner; real DB tests prove actual constraints/transactions. Avoid blanket coverage percentages as the sole gate.

### M0 — Reproducible foundation (prerequisite, not the first product demo)

Create minimal Go/React/Postgres stack, initial migration/import schema, OpenAPI, Make targets, `.env.example`, ignores, README and `docs/status.md`. Select and pin dependencies; record choices, not aspirational version numbers.

Acceptance: clean setup documented; migrations run on empty DB and on a previous revision; health/readiness distinguish DB outage; protected API rejects missing token; frontend loads; generation is repeatable. No scanner runs on setup. No `risk_score`, fake findings or fake database fallback. UI explicitly says no reports imported.

### M1 — First deliverable: real JSON → PostgreSQL → findings UI

Implement targets from lab registry, offline import, findings/observation list and detail. Category mapping may start unmapped except reviewed rules; verification is unreviewed, CVSS null. UI shows import digest/provenance, latest scanner severity, confidence, category mapping state and source evidence location, with visibly separate columns/states. Support category/severity/verification filters and pagination. Real report file is selected by the operator in the browser; no server path upload or remote URL fetch.

Acceptance tests:

- Import a provenance-backed genuine sanitized ZAP lab report. Derive expected finding fingerprints and instance counts from that fixture, not a made-up constant. API and DB counts match the expected normalization, including unlocated alerts and duplicate grouping; frontend renders actual persisted rows after reload/restart.
- Import same bytes twice and concurrently: one report normalization and one observation per source pointer; duplicate returns original ID. Invalid JSON, unsupported envelope, oversize/count/depth limits, mixed origins, malformed URL and trailing JSON return documented errors with no partial rows.
- Numeric-string fields parse; unknown severity becomes unknown with warning; out-of-study alerts remain visible. Two attacks on the same identity group correctly; distinct method or parameter creates separate finding.
- HTML/script and formula-like payloads display as text; no execution or unsolicited network requests. Pagination is stable. Token/raw evidence never appears in logs or list responses. Missing vectors/scores display “Not assessed,” not zero.
- README includes a genuine end-to-end import command/UI walkthrough, fixture provenance and observed results. If no real report is available, mark M1 blocked at that gate while completing parser/UI with labeled synthetic tests. Never present synthetic data as completing M1.

### M2 — Verification, mapping, and context history

Add manual review workflow and immutable context revisions, reviewed OWASP mapping file, resolution UI and audit events. Include reproducible lab access-control scenarios with two disposable roles/objects where applicable; do not assume Juice Shop guarantees every desired scenario without validation.

Acceptance: confirmed/false-positive/inconclusive histories survive edits as new revisions; stale predecessor gives 409; wrong-finding evidence refs fail; default/exact endpoint resolution deterministic; unknown factors visible; rules cannot silently overwrite history; mapping conflicts ambiguous; runtime role cannot alter immutable records. A01 claims have manual evidence. Category mapping changes do not change scanner severity.

### M3 — Traceable CVSS-B and CVSS-BE

Implement analyst metric entry, factor disposition, incomplete snapshots, calculator adapter, history and explanations. Store exact resolved context and parent Base relation. Allow reviewable drafts before scoring.

Acceptance: official/reference parity fixture suite covers boundaries, modified-metric inheritance, environmental requirements, safety-related Environmental values where legal, invalid/duplicate metrics and canonical ordering. Every parity expectation records source commit/vector/output; no numeric expected score is guessed. All completed B/BE vectors replay to identical official scores; missing required evidence or unknown context blocks finalization; invalid input rolls back; stale review/context gives 409. New context does not change an older vector/score. No score copied from ZAP severity; browser score matches API only because it displays the server output.

### M4 — Frozen ranking and research-ready export

Implement explicit cohort selection, technical/contextual comparison, dense ties, reasons for exclusion, snapshot detail and sanitized versioned JSON export. UI displays chosen mode as CVSS-B/CVSS-BE and links each rank to evidence, metrics, calculator and context.

Acceptance: same cohort and paired B parent across modes; ties stable with no implied extra precision; no false positives/out-of-scope/incomplete findings in study cohort; no fallback mixing scanner severity into CVSS ranking; concurrent context change cannot produce a mixed snapshot; exported artifact reproduces displayed order and hash after live context/review changes. Exports cannot leak secret fields. Unknown stays excluded with reason.

### M5 — Authorized ZAP orchestration and isolated lab

Only now implement worker, scanner adapter, authorization setup, network policy and cancellation. Add integration tests before enabling scan action.

Acceptance: queued→completed produces a real linked report; crash/lease/cancel/timeout tests prove no silent duplicate active scan; revoked/expired authorization and wrong mode refused; every forbidden network case in §10 is blocked with controlled sink evidence; restart reconciles external run state; active scan requires active-mode authorization. Capture actual versions/plan digests and limits. No scan from default test/dev/CI commands.

### M6 — Independent evaluation and final reproducibility

Freeze protocol before test-set assessment/ranking, obtain independent judgments, export cohorts, run analysis, and prepare demo and limitations. No accuracy/improvement claims until this stage has real data.

Acceptance: a fresh environment rebuilds selected artifacts from committed code + sanitized frozen inputs and locked dependencies; two runs with same seeds produce matching machine-readable results; hand-calculable tiny fixtures validate metric/tie handling; missing reference labels cause an explicit analysis error, never synthetic labels. Report both favorable and unfavorable results, exclusions, uncertainty, and threats to validity. Final README/status distinguishes successful local checks from unavailable external review/experiments.

For each milestone: record exact commands, versions, exit status, test output summary and artifact hashes in `docs/status.md`. Only check off the gate actually demonstrated. Stubs, screenshots and HTTP 200 alone do not satisfy behavior tests.

## 12. Research protocol and Python analysis

Unit of analysis: a verified normalized finding within a named lab context scenario. Predefine the scenario corpus and its known issues, intentional negatives and unobservable cases. Juice Shop is the primary lab; if extra controlled fixtures are necessary for missing categories, document and analyze them as separate sources, not as Juice Shop findings. Distinguish scanner detection coverage from prioritization performance.

Primary comparison: `technical-only = CVSS-B descending` versus `contextual = CVSS-BE descending`, on the **identical paired eligible finding cohort**. Freeze Base assessments before assigning contextual Environmental metrics. Do not change Base between methods. Scanner-severity ordering may be a separately labeled secondary descriptive baseline, never substituted for missing B assessments. Contextual ranking here means a scenario-informed CVSS-BE ranking, not a new universal business-risk metric.

Before looking at test results, commit a protocol specifying inclusion/exclusion, ranking cutoffs k, scenario groups, primary metric, ties, missing data, uncertainty method and any planned sensitivity analyses. Split calibration and evaluation by scenario/root weakness group to avoid duplicate observations leaking across sets. If no weights are fitted, still separate protocol design examples from held-out judgments. Never tune mappings or eligibility on test-reference outcomes.

Independent reference judgments:

- Aim for at least two security-competent independent reviewers who did not author the ranking policy. If unavailable, document the limitation and do not claim multi-reviewer validation.
- Give reviewers the same verified evidence and scenario context, randomized opaque finding IDs and a written prioritization rubric. Hide system scores, vectors, method names and produced ordering. Capture ties, rationale, reviewer confidence and missing judgments.
- Collect judgments before revealing system ranking. Preserve individual judgments and a prespecified adjudication/consensus procedure; report inter-reviewer agreement. Reference judgments are reasoned expert judgments, not objective ground truth.
- Separate the person assessing CVSS/context metrics from independent reference reviewers where possible. Record roles/conflicts and acknowledge any unavoidable overlap. Never use the formula, scanner severity or AI-generated labels as the independent reference.

Primary metric: Kendall tau-b with ties against prespecified consensus reference ordering, summarized by scenario. Secondary: Spearman correlation using average ranks; top-k agreement using tie-aware inclusion; nDCG@k only if reviewers independently provided graded relevance on a predefined scale. Define linear gain `g=r` for those grades and standard log2 position discount; do not derive relevance from either system score. All-zero relevance yields undefined nDCG and a reported count, not fabricated perfection. For a tie straddling k use fractional inclusion for overlap and expected gain across tied positions for nDCG; record implementation and validate with tiny fixtures. Fewer than two usable ranks or constant rankings can yield undefined correlation: emit null plus reason, never zero by default.

Report paired differences between methods, per-scenario results, cohort sizes, missingness and exclusions. If enough independent scenario groups exist, use seeded paired cluster bootstrap at scenario/root-weakness group level for confidence intervals; do not count repeated scans or correlated alert instances as independent samples. With too few groups, present descriptive results and state uncertainty limitations. Include disagreement analysis and no assumption that all contextual scores increase. Optional factor ablations must be predefined, use valid reassessments and remain secondary.

Export schema `diploma-research/v1` includes export ID/time, code commit (or `dirty=true` plus source digest), schema/parser/mapping/policy/resolver/calculator versions, scenario IDs, lab image/plan digests, report digests, candidate/included/excluded finding IDs and reasons, evidence provenance references, frozen reviews/classifications/context, paired assessment IDs/vectors/scores, ranking tie groups, and a separate reference-judgment file digest. Raw secrets and original reports are not embedded. Hash a canonical manifest and each artifact; keep private raw evidence available only through controlled operator access. A single ranking export can be merged into a study bundle only if cohort/parent/snapshot invariants pass validation.

Python reads frozen exports and independent references, validates IDs/cohorts/versions, and refuses mismatches. Output `results.json`, per-scenario CSV, exclusions CSV, reproducible figures, and a machine-readable run manifest with seeds and input/output hashes. Escape spreadsheet-formula prefixes in human-facing CSV text and record the transformation; canonical analysis JSON retains the underlying sanitized data. No database writes, rescanning or recalculation of CVSS during analysis.

Threats to validity to discuss: small artificial lab, judgment subjectivity, incomplete ZAP/access-control coverage, convenience sampling, ambiguous mappings, shared assessment author, contextual knowledge leakage, ties/missingness, root-cause grouping errors, lack of real operational impact and external generalization. Keep detection metrics separate: precision/recall require an independently verified scenario inventory and known expected negatives/positives; alert counts alone are not recall.

## 13. Staged AI implementation workflow

For every coding task:

1. Inspect real tree, applicable instructions, existing contracts, migration history, status and user changes. State the current milestone, actual starting point and bounded outcome.
2. Read the relevant contract and safety requirements. If this document leaves a routine choice open, choose the smallest option and record the decision. Ask only for missing decisions that prevent correctness, genuine data or authorization.
3. Define concrete acceptance cases for the slice. Update OpenAPI/schema/ADR together when needed. Implement the smallest working path through API, service, SQL and UI; do not generate all future milestones.
4. Add behavior tests for substantive logic and failure paths. Run appropriate formatting, type checks, generation checks, unit/integration/browser tests. Never call a skipped test passed. A missing Docker/tool/report is a recorded blocker; do not add fake fallbacks.
5. Review changed files for secrets, accidental raw evidence, unsafe target networking, invented vectors/scores, mutation of immutable history and unintended dependency upgrades. Confirm working-tree changes are within scope.
6. Update README and `docs/status.md` with actual commands/results, implemented vs planned endpoints and outstanding acceptance gates. Provide a short completion note: behavior delivered, evidence, limitations and next slice. Commit only if authorized by user/workflow; do not deploy or publish automatically.

Suggested first coding task after this document: “Implement M0 and M1 only. Build genuine ZAP Traditional JSON import into PostgreSQL and a React findings list/detail with source provenance and tests. Keep scanning disabled, verification unreviewed and all CVSS scores null. Complete the real-report acceptance gate or report the exact missing input.”

## 14. Sources and design provenance

Reviewed 2026-11-23. External facts follow primary sources; architecture, enums, limits, API, schema and milestones here are project design decisions. They are not prescribed by OWASP or FIRST. Pin actual upstream revisions used during implementation and revisit affected decisions if formats/specifications change.

- OWASP Top 10:2025 official release and categories: https://top10.owasp.org/
- FIRST CVSS v4.0 specification and nomenclature: https://www.first.org/cvss/v4.0/specification-document
- FIRST CVSS FAQ (interpretation and limitations): https://www.first.org/cvss/faq
- FIRST reference calculator (pin commit for parity): https://github.com/FIRSTdotorg/cvss-v4-calculator
- ZAP Traditional JSON report contract: https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/
- ZAP Automation Framework: https://www.zaproxy.org/docs/automate/automation-framework/
- ZAP API reference (verify capabilities against pinned service): https://www.zaproxy.org/docs/api/

Use source links and attribution in the implementation documentation. Preserve relevant upstream licenses for copied code/test data. The project must acknowledge FIRST's CVSS ownership/usage conditions when publishing CVSS assessments, and publish the associated vectors with scores.
