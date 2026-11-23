# Review log

Role: reviewing agent (read-only except this file). Scope: checkpoints marked READY FOR REVIEW in `docs/coordination/status.md`.

## Review pass 1: checkpoints 1–3 (M0 foundation, M1 offline import, evidence/pagination tests)

- Started 2026-11-23T15:10+05:00, finished at about 15:15+05:00.
- `docs/coordination/status.md` at review time: sha256 `9a5988546f326e214feb1712f6a525d915e47639a7a0f7783ff5c66f25bcc685` (states "Latest review checkpoint: 3").
- Snapshot method: SHA-256 of every non-ignored workspace file, taken at the start and again at the end. 68 files, **no changes during inspection**. I excluded `node_modules`, `artifacts`, `.secrets`, `.env` and `docs/coordination`.
- **Caveat:** status.md says `backend/internal/scope/url.go`, `url_test.go` and `httpapi/api_test.go` are *currently being edited* (decoded query-name validation slice). `docs/status.md` already reports that slice as passing. The findings below for those three files apply only to the hashes listed. When that slice is finished, publish a new checkpoint.

### Reviewed file hashes (sha256)

| File | sha256 |
| --- | --- |
| backend/cmd/api/main.go | 4cc1b3575718747c553e36c2d3cb28197abdb2ae729826058f7f57854fde770d |
| backend/internal/config/config.go | b1c9cfab5784acba1bec32329d2168d05a2b79f5e8411cd18f39a516a54b67cd |
| backend/internal/domain/errors.go | b319d2c5d1ae21701c8dde432577a5eff1248588713166662e3dd2bb326938f5 |
| backend/internal/httpapi/api.go | b5385848a176ed008efe1f0b1886b5ddcb1422f51ef402c2cb672cc455065ec1 |
| backend/internal/httpapi/api_test.go | d47af2c12929a8b544fbaa493b2b7e2dce42b776c193370afc7ecd2d64d026a7 |
| backend/internal/scope/url.go | cb5dd01dddeefe95d9425b01496e8eda0695b7050633d6d32fce01bffb46fc5b |
| backend/internal/scope/url_test.go | 017ab16245ee1b9d6178b1b74fbb1ee08547fa05b6467a0622ca3771cffcc0db |
| backend/internal/service/service.go | eafbdef0933fdfed94e409596ba23d9622b82b31f62b44d99654741662916b41 |
| backend/internal/service/types.go | fd9e025fafb5cddf203f65763dfb132024ae62413541f6a95f99fec3873d0cc9 |
| backend/internal/service/paging.go | abbb2d8c556a91f3622271a211044dae29468a35ea8e2a9e03c3af7fe61e32ea |
| backend/internal/service/redact.go | 02d45682e610969ad06ad28bc17dd2d93e8c03729d2157637c7ed7b3eedc5e79 |
| backend/internal/store/store.go | ab5581f420b9eddf7447aba17a9df0203491945a1288c52476668aa9f18ef221 |
| backend/internal/zap/parser.go | b66e9ac5764f91f4990e5ec924f2d3cefffed44cc8df7e2dc1539f756594f4ec |
| backend/internal/zap/parser_test.go | f0f32b83ba043508ffa5accf348e96bd9cc3e88e235a4fc4ca06ef7daf7123f0 |
| backend/db/migrations/00001_imports.sql | 74a3de9f13544b4c3e59fa6715a9bd4cdee31a65fdec95fbbd0031fe8efda2d2 |
| backend/db/migrations/00002_runtime_privileges.sql | 69556e133389488e885e2a2f7e14392414417f7276f7d02becdd031a886faed7 |
| backend/db/migrations/00003_target_lock_privilege.sql | f420da8bd8d862e66bc006e032ec51b7ef0f29d3e7b8a7e29ac69783689a0599 |
| backend/db/queries/imports.sql | a153a57e84957fccec4fd3c44d03a58dadc70fb86d28d7c023d588c585f72f39 |
| backend/internal/store/sqlc/imports.sql.go | c29d644e0072cc7d28db21ed9a8c38cea73934d527dea4d4749449d705d9cb77 (matches docs/status.md) |
| backend/testdata/zap/synthetic-traditional.json | c91f5451043fcac49e8246e04dedb90eb362ebe983d064600a6914330eaa304a (matches docs/status.md) |
| contracts/openapi.yaml | 3bd038fe661fd985e12cbf41f4729e34d12e988e656af1561e65bb8ca3650ba6 (matches docs/status.md) |
| frontend/src/api/schema.d.ts | 95f1d52a0a657f8d161cc7762c53e9d52a96135ad26e08afc32d18beb67c209f (matches docs/status.md) |
| frontend/src/api/client.ts | d12f97fe6f984e196362eddee15996bce2f8fae5326183647340c51f5474b929 |
| frontend/src/app/main.tsx | e010406af3e8ff5528dee2b4b249ae93841fb273c0ddcb8fc25916225d0ed501 |
| frontend/src/features/evidence.tsx | 1f8da6aaa8e639f7003ee2f9f22503a1f3c77b2bf61a54cd6d340127fd3797e8 |
| frontend/tests/browser/import.spec.ts | d948a16fd66e67f3d15dbeb71633bd84a5cf4be0ce28aafe1009ca1f96f9b8e4 |
| frontend/tests/workspace.test.tsx | 57c21a2d0658d5775f79b82df1c568cee5a0f78f2cab54cc37fe9f91288e7207 |
| docs/status.md | b84648f5cf66a880a911ec1e59675f1204e4a457b38dd1e9c54d1259d8144fac |
| README.md | dea59ab84034ffe9c1b8e052cc7b8b47b45104cdf9dc53a159299b983801d6b4 |

The full 68-file manifest is in the reviewer's temp dir (`/tmp/review-99f0/hashes-t1.txt`), which is not shared.

### Checks I ran myself

- `GOCACHE=/tmp/review-99f0/gocache GOFLAGS=-mod=readonly go test -count=1 ./internal/zap ./internal/scope ./internal/httpapi ./internal/service` exited 0. **`TestIntegrationImport` skipped**: no `TEST_DATABASE_URL`, and I did not touch any database.
- A standalone Go program in `/tmp/review-99f0/q` confirmed R001: `url.ParseQuery("q=1;DROP")` returns `invalid semicolon separator in query`, and `url.ParseQuery("q=100%")` returns `invalid URL escape "%"`.
- I did **not** run `make check`, `make test-integration`, `make test-e2e`, `npm run build`, the fuzzers, migrations, or Compose. For those I rely on the main agent's results in `docs/status.md` and did not reproduce them.

### Findings

| ID | Priority | Location | Status |
| --- | --- | --- | --- |
| R001 | P1 | backend/internal/scope/url.go:104-107 | OPEN |
| R002 | P2 | backend/internal/service/service.go:148,168; backend/db/queries/imports.sql:33 | OPEN |
| R003 | P2 | backend/internal/service/service.go:97-103 | OPEN |
| R004 | P2 | backend/internal/service/service.go:138-143 | OPEN |
| R005 | P3 | backend/internal/zap/parser.go:61 | OPEN |
| R006 | P3 | frontend/src/features/evidence.tsx:428,487; backend/internal/service/service.go:233 | OPEN |
| R007 | P3 | backend/internal/scope/url.go:32-36 | OPEN |
| R008 | P3 | backend/internal/zap/parser_test.go, backend/internal/httpapi/api_test.go | OPEN |
| R009 | P3 | backend/internal/zap/parser.go:170-176 | OPEN |

**R001 (P1): valid ZAP instance URIs reject the whole import.** `url.ParseQuery` fails on a raw `;` in the query and on a bad `%` escape inside a query *value*. Then `Parse` returns `bad()`, and the importer rejects the whole report atomically with 422. RFC 3986 allows `;` in a query. Real ZAP instance URIs carry attack payloads in query values, such as SQL injection or traversal strings, so a genuine Juice Shop report can be rejected outright. This blocks the M1 genuine-report gate. The comment on lines 110-111 says value bytes are "immaterial", but `ParseQuery` still parses them.
*Fix:* parse names yourself. Split `RawQuery` on `&` only, take the text before the first `=`, then `QueryUnescape` the name alone. Keep the NUL/UTF-8 rejection for names. Never decode values. Decide how a name with a bad escape is handled (reject with a pointer, or keep it raw plus a warning) and document the choice. Add tests for `?q=1;2`, `?q=100%`, `?q=%27%3B--`, and an empty name (`?=x` currently gives the name `""`).

**R002 (P2): "latest scanner observation" is arbitrary within one import.** Every observation in an import shares one `created_at` (`now`). The tie on `(created_at DESC, id DESC)` therefore falls to a random UUIDv4. Fingerprint v1 omits `alertRef` and `riskcode`, so one grouped finding can hold observations with different severities. One example is the same plugin and endpoint/parameter with different alert variants. In that case `scanner_severity` and `severity_observation_id` depend on random IDs, not on the evidence. Observation order also does not follow source order.
*Fix:* give the tie a deterministic, documented order, for example the source order of observations within the import (an ordinal column) after `created_at`. Optionally add a `severity_conflict` warning when one finding gets differing severities in one import. Add a test with two severities in one group. If the order changes from `(created_at,id)`, update OpenAPI.

**R003 (P2): provenance mixes self-reported report content with operator-supplied metadata.** When the operator leaves `zap_version` null, the server copies it from the report's `@version`. It sets `report_template` to `traditional-json` without being told. For the synthetic fixture the stored provenance then reads `zap_version: "SYNTHETIC-TEST-ONLY"`. §5 requires unknown metadata to stay null, and provenance is the evidence for the genuine-report gate. Copying a field from the uploaded file presents it as independent metadata.
*Fix:* keep operator provenance exactly as submitted (null stays null). Put report-derived values in separate, labeled fields such as `report_declared_version` and `parser_detected_format`, and update OpenAPI and the UI.

**R004 (P2): duplicate uploads silently drop new provenance.** On an exact duplicate, the stored import comes back with `duplicate=true`, and any different `provenance` in the new request is discarded with no warning. An operator who re-uploads to add an authorization reference or scan time will believe it was recorded.
*Fix:* compare the submitted provenance with the stored value. If they differ, return 409 or a warning in the response that the provenance was not applied. Document that immutable imports cannot gain provenance later, at least until an append-only provenance revision exists. Add a test.

**R005 (P3): the 64 KiB field limit skips numbers.** The check only covers `string` tokens. A `json.Number` literal of any length up to the 20 MiB body passes, and its text is stored, for example in `scanner_severity_raw` or `confidence_raw`.
*Fix:* apply the same length limit to `json.Number`, and add a test.

**R006 (P3): the UI and service hardcode assessment and review states.** The UI always prints "Not assessed" instead of reading `latest_base` and `latest_environmental`. The service always returns `verification: "unreviewed"` and `classification: "unmapped"`. Both are correct for M1. Once M2 or M3 add data, though, nothing ties the UI or service to it, and the display could misreport assessment state.
*Fix:* render from the API fields (`latest_base == null` → "Not assessed"), and add a test that a non-null fixture renders its score and vector.

**R007 (P3): hosts with underscores are rejected.** The host character set leaves out `_`. Docker or Compose hostnames containing underscores, which older Compose container names use, make `JUICE_SHOP_ORIGIN` fail at startup or make an import fail.
*Fix:* decide explicitly. Either allow `_` in registered lab hostnames, or document the restriction in README and `lab-targets`.

**R008 (P3): some acceptance cases are untested.**
- No test where the site-level `@name` origin differs (only the instance-level `uri` is tested).
- No test for an alert with no `instances` key at all.
- No golden case with empty `query_parameter_names` or an empty parameter name.
- No test that limit-exceeded and depth errors at the API level leave no partial rows. Today this holds by construction, because parsing finishes before the transaction starts.

*Fix:* add these as unit or integration cases.

**R009 (P3): `Fingerprint` is exported but does not enforce its own input invariants.** A nil `QueryNames` serializes as `null` and an empty slice as `[]`, so the two give different hashes. `Parse` always passes a non-nil slice, but any future caller, such as a reviewed-split migration, could produce non-canonical fingerprints.
*Fix:* convert nil to an empty slice inside `Fingerprint`, sort and deduplicate there, and add golden tests.

### Verified as acceptable at this checkpoint (no finding)

- Bearer auth uses a constant-time comparison, and only `/healthz` and `/readyz` skip it. 401s use the problem shape and include `X-Request-ID`. Every response sends `Cache-Control: no-store` and `nosniff`.
- The raw download requires the acknowledgement header, is audited in the same transaction, and returns the stored bytes as an attachment. List and detail responses never expose `raw_uri` or `raw_bytes`. Attack, evidence and other_info are withheld rather than heuristically redacted.
- Import order: parse (no I/O), lock the target row, check for a duplicate, then insert everything in one transaction. On any error nothing is persisted. Uniqueness on `(target_id, sha256, parser_version)` and on `(import, pointer)` is enforced in the schema. Composite FKs keep findings and imports on the same target.
- The runtime role has SELECT and INSERT only on evidence tables, plus UPDATE on `targets.name` alone so row locks work. Readiness checks this. No score, vector, `risk_score` or invented classification exists anywhere. CVSS is absent from the schema.
- Fingerprint v1 inputs match §5. Method is uppercased and validated as a token. Missing `pluginid` on a located alert is a 422 with a pointer. Unlocated alerts are kept, with a warning. Unknown `riskcode` becomes `unknown`, with a warning.
- Mutation JSON rejects duplicate keys, unknown fields, keys matched only case-insensitively, and trailing data. Cursors are HMAC-signed and bound to the filter. Limits are 1–200.
- Scanner text renders as React text. I found no `dangerouslySetInnerHTML`. The token is kept in memory only.
- OpenAPI covers the 11 implemented routes, and the Problem code enum matches the server. Generated artifacts match the hashes in docs/status.md.

### Acceptance gates still unverified

1. **M1 genuine-report gate: BLOCKED.** No genuine, provenance-backed, sanitized Juice Shop report exists, and every recorded result uses authored synthetic data. docs/status.md and README state this correctly; I found no misleading completion claim. R001 may also stop a real report from importing.
2. Integration (PostgreSQL), e2e (Chromium), fuzz, `make check` and build were **not independently run** by me. I rely on the main agent's results.
3. Compose container startup is unverified because no Docker daemon was available, as the main agent also records.
4. The "frontend renders persisted rows after restart" check exists only for synthetic data.

### Verdict

**Not approved** for checkpoints 1–3 at the hashes above, because R001 is open. Please fix it, then publish a new checkpoint listing changed files. I will update statuses to RESOLVED or NEEDS RECHECK on the next pass.

---

Reviewer note: I cannot watch for changes automatically. The last checkpoint reviewed is **3**, at the hashes above. A new invocation is needed to review later checkpoints.

## Independent verification — M1-20261123-07

Verified on 2026-11-23 at baseline commit `21a096f899ed43b340f26fbf6cceabf288f46530`. Scope: the five frozen backend files listed in checkpoint 07. All five SHA-256 values matched before and after verification. Frontend files were outside the frozen slice. PostgreSQL 15.19 and Go 1.27.1 on macOS arm64.

### Independent checks

- From `backend/`, `GOCACHE=/private/tmp/diploma-verifier-go-cache GOFLAGS=-mod=readonly go test -count=1 ./internal/zap ./internal/scope ./internal/httpapi ./internal/service ./internal/domain` exited 0. Service/domain are compilation checks. Database cases were run separately.
- `GOFLAGS=-mod=readonly make test-integration` exited 0 against a disposable PostgreSQL cluster under the Go race detector. Applied migrations 1 through 4 in stages and confirmed a repeat upgrade was a no-op. All six escaped-NUL tests returned the expected 422 and pointer for finding title, parameter identity, private evidence, unlocated JSONB, provenance value and provenance key, with no partial rows. Existing import, exact bytes and audit, concurrent duplicate, provenance conflict, import scope, deterministic history and runtime privilege checks also passed.
- `git diff --check` on the checkpoint files exited 0.
- A second disposable integration run used a temporary Go test overlay to probe two adjacent PostgreSQL representability limits. The overlay asserted the observed generic 500 responses for the two inputs below, so its exit 0 reproduces the failures. It is not a fix or an acceptance pass.

### Earlier findings recheck

| ID | Current disposition | Evidence |
| --- | --- | --- |
| R001 | Verified resolved | Query names alone are decoded. Unit cases and the PostgreSQL variant import accepted raw semicolons and malformed percent text confined to private values. |
| R002 | Verified resolved for new imports | The fixed-clock PostgreSQL test proved deterministic source and receipt order, distinct severities in one finding, and import-scoped severity. Legacy observation timestamps remain as stored. |
| R003 | Verified resolved | The test proved operator version/template remain null when omitted and the report-declared version remains separate. |
| R004 | Verified resolved | Conflicting provenance returns 409 with the existing record unchanged. Concurrent duplicates and matching provenance reuse the import. |
| R005 | Verified resolved | Both the unit token limit test and the API test cover a 65,537-byte numeric token. |
| R006 | Baseline fix present; current UI pending verification | Git baseline 21a096f reads finding verification, classification and both assessment fields from the API; current frontend edits were outside this checkpoint. Backend M1 default states are expected. |
| R007 | Verified documented decision | README and the target registry example specify DNS-style service names without underscores. |
| R008 | Verified resolved | Unit and PostgreSQL coverage now includes site-origin mismatch, missing instances, empty fingerprint entries, and atomic field/count/depth failures. |
| R009 | Verified resolved | The golden fingerprint tests pass for nil, empty, unsorted and duplicate query names and verify caller data remains unchanged. |

### New findings

- **V010 (P2), open.** `backend/internal/zap/parser.go` preserves an unlocated alert’s entire decoded map in JSONB. A synthetic extra field containing the short numeric token `1e999999` passed scalar limits and then returned HTTP 500 during report creation. PostgreSQL JSONB represents numbers with its bounded numeric type. Validate JSONB representability for this path and return a typed field error, or use a reversible representation that keeps the full alert available. Add the reproducer to disposable integration coverage.
- **V011 (P2), open.** `backend/internal/service/service.go` checks the byte length of compact Go JSON, while `00001_imports.sql` bounds `octet_length(source_metadata_json::text)`. PostgreSQL adds spaces in that rendering. A synthetic supplied authorization reference sized so the normalized Go JSON was exactly 65,536 bytes passed Go validation and then returned HTTP 500. Align the check with the stored representation or map this specific database constraint to the documented 413. Retain the database constraint and add a boundary test.

### Verdict

The escaped-NUL slice is independently verified at the recorded hashes. V010 and V011 remain open for import reliability. Genuine-report acceptance, the new frontend and Compose startup were outside this verification pass. No scanner was started and no existing database was accessed.

## Independent verification — M1-20261123-09

Verified on 2026-11-23 at baseline commit `21a096f899ed43b340f26fbf6cceabf288f46530`. The five backend file hashes matched the frozen checkpoint. `git diff --check -- backend` reported no whitespace errors.

- `GOFLAGS=-mod=readonly make test-integration` exited 0 on macOS arm64 with Go 1.27.1, PostgreSQL 15.19 and the Go race detector. Applied migrations 1 through 4 in the disposable cluster and repeated the migration command successfully.
- **V010 resolved.** The previously failing large numeric extension imported successfully. The stored unlocated alert retained the original numeric literal in its versioned JSON text field, and the independently executed test confirmed identical original bytes.
- **V011 resolved.** PostgreSQL's provenance size constraint remains in place. The independently executed exact-size case returned the documented `413 limit_exceeded` and left no import row.
- The six escaped-NUL cases still returned pointer-specific 422 with no partial import, finding or observation rows. The existing concurrency, provenance conflict, scope, ordering, raw download and least-privileged role checks passed.

**Verdict:** independently verified for checkpoint 09. The genuine authorized report and Compose startup remain separate, unverified gates.

## Independent verification — M1-20261123-08

Verified on 2026-11-23 at baseline commit `21a096f899ed43b340f26fbf6cceabf288f46530`. All five frontend hashes matched before and after verification.

- `npm --prefix frontend run fmt:check`, `npm --prefix frontend run lint` (TypeScript and ESLint), `npm --prefix frontend test` (three tests), and `git diff --check -- frontend` passed. The production Vite build to a disposable output directory passed.
- `E2E_SCREENSHOTS=/private/tmp/diploma-verifier-ui-08 GOFLAGS=-mod=readonly make test-e2e` exited 0. Chromium and the separately created PostgreSQL cluster verified the protected unlock, target creation, synthetic import, exact duplicates and provenance conflict, safe evidence rendering, preserved severity filter on return, acknowledgement-gated byte-identical download, reload and restart persistence, and readiness during database outage.
- Reviewed the captured 1280×900 unlock/findings and 390×844 findings/evidence layouts. The screens have white panels, gray page backgrounds and restrained controls. The narrow findings table has its own scroll region, and the evidence detail independently shows scanner severity, verification, OWASP category and CVSS. **R006 is resolved** on the verified UI.
- **V012 (P3), open.** A cursor-only findings URL with an invalid or outdated cursor has no visible recovery action. A separate temporary Chromium probe against a fresh local API and PostgreSQL target opened the target URL with only `?cursor=stale-cursor`. The API returned `400 invalid_request` and the page displayed “Invalid or mismatched cursor.” Neither Clear filters nor First page was offered. `filtersActive` excludes cursor and the pager renders only after a successful response. Cursor links can become stale after token rotation, since the cursor signing key comes from the API token. Offer a First page action on a cursor error that clears cursor alone and preserves any filters, and cover this recovery in the browser test.

**Verdict:** verified the light workspace and existing happy path at checkpoint 08. A superseding checkpoint should address V012. This pass used synthetic data only.

## Independent verification — M1-20261123-10

Verified on 2026-11-23 at baseline commit `21a096f899ed43b340f26fbf6cceabf288f46530`. The five frozen frontend SHA-256 values matched before and after the checks. `git diff --check -- frontend` passed.

- `GOFLAGS=-mod=readonly make test-e2e` exited 0. Both Chromium phases passed against a separately created PostgreSQL database. The first exercised the real API 400 for a cursor-only findings URL, clicked **First page**, verified that the URL no longer included the cursor, and recovered all four original synthetic rows. The database outage check returned health 200, readiness 503, and authentication 401 where expected. The second Chromium phase verified persisted rows after PostgreSQL and API restart.
- The findings recovery handler clears only cursor and retains other query filters. For targets and observations, the pagers are likewise outside the success-only rendering branches, and a First page action remains available whenever those views have an active cursor.

**Verdict:** V012 resolved and frontend checkpoint 10 independently verified at its recorded hashes. Together with the independently verified backend checkpoint 09, no review findings remain open for the scoped changes. Genuine authorized report acceptance and Compose startup have not been verified.

## Independent verification — M0-20261123-11

Verified on 2026-11-23. Both `HEAD` and the local `refs/remotes/origin/main` resolved to checkpoint baseline `a1b523d8ad25b170466a09e4d758e8afacc55b39`. The three SHA-256 values matched the frozen checkpoint.

- An independent Python check found 13 relative Markdown links across `AGENTS.md`, `frontend/FRONTEND.md`, and `backend/BACKEND.md`. All resolved to files from the linking document's directory. No relative link used an anchor.
- `git diff --check -- AGENTS.md frontend/FRONTEND.md backend/BACKEND.md` exited 0. The scoped diff contains only guide links and status wording. The root contract keeps its milestone evidence rule and directs readers to the implementation and independent verification logs. Both layer guides reference the current checkout and its API contract.
- The documentation does not claim hosted CI, a running Compose stack, or a genuine report acceptance pass. I checked the local tracking ref and did not query GitHub again. No application check was rerun for this docs-only checkpoint.

**Verdict:** checkpoint 11 independently verified at its recorded hashes. No findings.

## Independent verification — M0-20261123-12

Verified on 2026-11-23 at baseline commit `a1b523d8ad25b170466a09e4d758e8afacc55b39`. The SHA-256 values for the CI workflow and PostgreSQL bootstrap matched the frozen checkpoint before and after verification.

- `GOCACHE=/private/tmp/diploma-verifier-go-cache go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 .github/workflows/ci.yml` exited 0 when the Go module proxy was available. `sh -n` passed separately for the installer and both disposable test scripts. The installer was executable and rejected a relative prefix and an existing directory before downloading anything.
- Read-only upstream tag lookups matched all three full GitHub Action pins. The PostgreSQL project's published SHA-256 for its 15.19 source archive matched the installer. Go 1.27.1 in `backend/go.mod` and the locally reported Node 26.13.0, npm 12.2.1, sqlc 1.31.1, goose 3.30.1, and lockfile Playwright 1.65.1 align with the workflow and tool manifest.
- Workflow permissions are `contents: read`, checkout does not persist credentials, and the workflow references no repository secrets. Both suites provision separate temporary, loopback-only clusters. The installer refuses UID 0 and the workflow separately checks for a non-root runner. No source archive build or container run was repeated in this verification.
- **V013 (P2), open.** The workflow's `make check` indirectly invokes `rg -q` in `Makefile:21`, yet the workflow never provisions ripgrep. The [official Ubuntu 24.04 runner software manifest](https://raw.githubusercontent.com/actions/runner-images/main/images/ubuntu/Ubuntu2404-Readme.md) does not list it. Independently reproduced `make tools-check` with pinned sqlc and goose on a constrained PATH containing only those tool links and system utilities. It exited 2 with `rg: command not found` and Make error 127. Remove the dependency from the version check with an exact shell comparison, or explicitly provision ripgrep in the workflow.

**Verdict:** the syntax, pins and test isolation checks passed at checkpoint 12. CI verification is pending V013 and an observed hosted run. No workflow was dispatched and no application data was accessed.

## Independent verification — M0-20261123-13

Verified on 2026-11-23 at baseline commit `a1b523d8ad25b170466a09e4d758e8afacc55b39`. The three frozen SHA-256 values matched the checkpoint. The workflow and installer match their independently reviewed checkpoint 12 hashes. `git diff --check -- Makefile .github/workflows/ci.yml scripts/install-test-postgres.sh` exited 0.

- **V013 resolved.** Repeated the restricted-PATH probe with a temporary directory containing symlinks to pinned goose and sqlc, followed by only `/usr/bin:/bin`. Verified that `command -v rg` failed. Goose emitted `goose version: v3.30.1`, sqlc emitted `v1.31.1`, and `PATH="$probe_path" make tools-check` exited 0. The only Makefile change replaces the ripgrep pipeline with an exact shell builtin comparison.

**Verdict:** checkpoint 13 independently verified for the CI compatibility fix. No hosted workflow, source archive build, application test or Compose stack was run for this narrow recheck. Hosted CI and the separate runtime and acceptance gates remain unverified by this pass.

## Independent verification — M0-20261123-14

Verified on 2026-11-23 at baseline commit `a1b523d8ad25b170466a09e4d758e8afacc55b39`. All five frozen file hashes matched. `python3 scripts/test-compose.py --config-only`, `docker compose config --quiet` and the scoped `git diff --check` exited 0. The README's Compose 2.24.4 minimum for the port replacement tag agrees with the [official Compose merge reference](https://docs.docker.com/reference/compose-file/merge/#replace-value).

- `python3 -u scripts/test-compose.py` exited 0 independently on Apple Silicon with Docker Desktop Engine 29.7.2 aarch64 and Compose 5.5.1. It provisioned project `diploma-smoke-4e62a840fdca1b400abf` with fresh credentials and a fresh volume.
- The effective configuration and smoke run verified ephemeral API/frontend bindings on loopback, no published PostgreSQL port, and separate internal storage networking for PostgreSQL and the migration service. The running stack reached Goose revision 4 and passed readiness, health, non-root frontend cache write access, frontend HTML and TypeScript-entry serving, and the missing-bearer rejection.
- One synthetic upload produced four findings, five observations and one unlocated alert. The digest matched the exact fixture bytes, the duplicate reused the original import ID, and PostgreSQL contained one report and five observations. Shutdown retained the test volume. After restarting, the same digest and all four findings were returned.
- Final cleanup removed only resources scoped to the independent test project, including its exact image tags. Separate read-only Docker queries found no matching containers, networks, volumes or image tags afterward. No existing application data was modified.

**Verdict:** checkpoint 14 independently verified at its recorded hashes. No findings. This closes the tested ARM64 Compose runtime gate. Hosted GitHub Actions and acceptance with a genuine authorized Juice Shop/ZAP report remain separate, unverified gates. No scanner was launched.

## Independent verification — M1-20261123-15

Verified on 2026-11-23 at baseline commit `5a1cd7427fbfca196d3dc59bea65555a256046e6`. All five frozen SHA-256 hashes matched.

- Public GitHub API metadata confirmed run `45166744371`, attempt 1, on `main` at that SHA, updated `2026-11-23T11:42:06Z`, with status `completed` and conclusion `success`. Job `137665526785` and all 15 reported steps succeeded, including PostgreSQL integration and Chromium. The API returned one run for the exact commit. [Hosted GitHub Actions run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45166744371)
- `python3 -B -m unittest discover -s scripts -p 'test_check_zap_fixture.py' -v` exited 0 with eight passing tests on Python 3.14.3. `python3 -B scripts/check-zap-fixture.py --help` and the scoped `git diff --check` exited 0. All 16 local links and anchors across the three checkpoint documents resolved.
- A separate synthetic CLI invocation returned four aggregate counts on stdout with empty stderr. A before-and-after snapshot confirmed that all three inputs were byte-identical and the temporary directory gained no files. Implementation reads the inputs locally in binary mode under byte limits. The checklist separately calls for human review of authorization and sanitization, verification of matching alert order and fingerprint groups, and PostgreSQL and browser checks using the accepted derivative.
- **V014 (P3), open.** `scripts/check-zap-fixture.py:143–147` uses the default `argparse` error handler. `python3 -B scripts/check-zap-fixture.py original.json derivative.json provenance.json /fictional/private-path-marker.json` exited 2 and echoed the synthetic path in stderr. This contradicts the checkpoint's fixed-diagnostics claim for invalid invocations. Use a fixed error diagnostic alongside usage and test that a malformed invocation does not disclose argument values.

**Verdict:** verified the hosted CI result, documentation links and normal preflight path at checkpoint 15. The diagnostic privacy claim remains open under V014. The genuine-report acceptance gate still requires an existing authorized Juice Shop ZAP Traditional JSON report and its provenance, followed by the checklist checks. This verification used synthetic inputs and public metadata only.

## Independent verification — M1-20261123-16

Verified on 2026-11-23 at baseline commit `5a1cd7427fbfca196d3dc59bea65555a256046e6`. All five frozen SHA-256 hashes matched before and after the checks. The scoped `git diff --check` exited 0.

- **V014 resolved.** Repeated the original invocation with the extra `/fictional/private-path-marker.json` positional argument. An unknown option with a synthetic value and an invocation missing two required arguments produced the same result. Each exited 2 with empty stdout and fixed stderr `Expected original, derivative and provenance file paths. Use --help.`. `--help` exited 0 with usage on stdout and empty stderr.
- `python3 -B -m unittest discover -s scripts -p 'test_check_zap_fixture.py' -v` exited 0 with all nine tests passing on Python 3.14.3, including the new diagnostic regression.
- The README and implementation status record the hosted CI success for the baseline commit and link to run `45166744371`, consistent with the independently observed metadata in checkpoint 15. The nine local Python tests are recorded separately from the hosted run. [Hosted GitHub Actions run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45166744371)

**Verdict:** checkpoint 16 independently verified at its recorded hashes. No new findings. Genuine-report acceptance still requires a ZAP Traditional JSON report from an authorized Juice Shop lab and its provenance, followed by the checklist checks.

## Independent verification — M0-20261123-17

Verified on 2026-11-23 at baseline `4d0240c4f50559b0e21feeeead7348c03cd74a7f`. All five frozen SHA-256 values matched before and after the checks. The scoped `git diff --check` and `actionlint` 1.7.7 exited 0.

- In a fresh disposable Git archive of the baseline, I copied in the five checkpoint files and a separate copy of the installed frontend dependencies. With Python 3.14.3 and `GOFLAGS=-mod=readonly`, `make test-scripts`, `make test` and `make check` each exited 0 and ran all nine Python tests. The check also ran formatting, Go vet, TypeScript, ESLint, byte-identical generation, Go unit tests and three Vitest tests.
- I introduced a deliberately failing Python assertion in that archive. Both `make test` and `make check` exited 2 and reported the exact assertion marker. The checkout's preflight and test sources were unchanged.
- Removing the added Python setup stanza reproduces the baseline workflow byte for byte. The normal run still includes `make check`, frontend build, disposable PostgreSQL integration and Chromium. The tool manifest and setup action select Python 3.14.3. The release tag `v7.1.0` resolves to its workflow pin `6d4fd22b20465ccc0f7354476588e8df7c9ca0e0`, and the official distribution manifest lists stable Python 3.14.3 for Ubuntu 24.04 x64. [Setup Python tag](https://api.github.com/repos/actions/setup-python/git/ref/tags/v7.1.0) · [Python versions manifest](https://raw.githubusercontent.com/actions/python-versions/main/versions-manifest.json)
- Public GitHub metadata independently confirmed one run for the exact baseline, `45168962574` attempt 1, `completed` with conclusion `success`, updated `2026-11-23T11:54:53Z`. Job `137673174362` and all 15 reported steps passed. The job's steps and baseline workflow predate the added Python setup and test target. The README and status correctly identify the baseline scope. [Hosted GitHub Actions run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45168962574)

**Verdict:** checkpoint 17 independently verified at its recorded hashes. No findings. A hosted run of the updated workflow and acceptance with a genuine authorized Juice Shop report remain separate gates. The independent rerun used the local check path and did not provision a database.

## Independent verification — M1-20261123-18

Verified on 2026-11-23 at baseline `94f431f479b7b707c84ee9d79ebf5f6710aa9ced`. All three frozen SHA-256 values matched before and after the checks. The scoped `git diff --check` exited 0, all 18 relative links and anchors resolved, and `git check-ignore -v` matched the example original and authorization paths to `.gitignore:3`. The frozen diff changes documentation only.

- Public GitHub API metadata confirmed run `45170314104`, attempt 1, on `main` for that exact SHA, updated `2026-11-23T12:07:24Z` with status `completed` and conclusion `success`. Job `137677821087` and all 17 steps passed, including Python setup, PostgreSQL integration and Chromium. GitHub returned one run for the exact commit. The read-only connector retrieved the job log, which contained the `python3 -B -m unittest -v test_check_zap_fixture` invocation, `Ran 9 tests`, and byte-identical generation. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104)
- The proposed collection limits passive observation to ordinary browser traffic against the named local target, with explicit authorization, verified isolation and budget enforcement before capture. It requires private provenance and a reviewed derivative followed by persistence tests. The stated Safe mode and passive observability agree with the ZAP documentation. The plan, preflight and hosted CI results are presented separately from real data and actual isolation evidence. [ZAP modes](https://www.zaproxy.org/docs/desktop/start/features/modes/) · [Passive Scanner](https://www.zaproxy.org/docs/desktop/addons/passive-scanner/) · [Traditional JSON format](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/)
- **V015 (P3), open.** `docs/fixture-provenance.md:24–27` requires the Juice Shop image digest and ZAP/add-on versions. Its preparation and run records omit digests for any other container images used, including a containerized ZAP. `AGENTS.md:107` requires version and digest pins for all selected images, and §9 includes scanner image provenance. Require pins for all chosen lab images before the isolation checks and retain the exact resolved image identities in the private run evidence.

**Verdict:** verified the hosted baseline result and documentation links at checkpoint 18. The plan's image-identification requirement remains open under V015. Actual lab isolation, explicit authorization and genuine-report acceptance remain pending. This review exercised no lab traffic and used no database.

## Independent verification — M1-20261123-19

Verified on 2026-11-23 at baseline `94f431f479b7b707c84ee9d79ebf5f6710aa9ced`. All three frozen SHA-256 values matched before and after verification. The scoped `git diff --check` exited 0, all 18 local links and anchors resolved, and `.gitignore` matched the example original and authorization paths.

- **V015 resolved.** The plan now requires version and SHA-256 digest pins for every lab container before isolation checks, plus exact ZAP and add-on versions. It requires keeping the resolved platform image identities with the private run evidence.
- The proposed test boundary keeps the browser on the private ZAP listener and ZAP on an allowlisting proxy restricted to the pinned Juice Shop target. The plan still calls for explicit authorization, controlled-sink and budget checks before any collection, ordinary passive browsing only, a private export, and shutdown. Actual isolation and authorization remain separate gates.
- The README is unchanged at the checkpoint 18 hash. The implementation status now records the nine hosted Python tests, consistent with the previously inspected step log and all 17 successful steps for the baseline. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104)

**Verdict:** checkpoint 19 independently verified at its recorded hashes. No new findings. The lab, genuine-report acceptance and a hosted run of an eventual newer commit remain unverified by this documentation review.

## Independent verification — M1-20261123-20

Verified on 2026-11-23 at baseline `0c193d51c989866e8d49932454803853cb16df11`. All 14 frozen SHA-256 values matched before and after the checks. The scoped `git diff --check` exited 0, all 17 relative links and anchors resolved, and the original, derivative, provenance and authorization paths matched `.gitignore`. The private directory and reviewed JSON records were restricted to the local owner.

- The recorded authorization was unrevoked, scoped to one passive run at the registered origin, and carried explicit time, request and rate bounds. Its plan digest matched the saved plan and provenance. UTC chronology placed the successful isolation recheck after authorization began and before capture. Report generation and completion fell within the authorization window. The recorded 76.217-second elapsed duration agreed with the timestamps.
- The frozen isolation record shows the allowed Chromium and ZAP path, eleven forbidden-request probes, zero denied-sink requests, no target application requests during preflight, and a three-request pacing check of 1.084 seconds. Stored ZAP output records the passive queue, report job, and plan success and matches the add-on versions in provenance. Read-only Docker queries found zero running lab containers and confirmed all 12 retained lab containers stopped. Reading the stopped ZAP export into memory produced bytes identical to the private original. No live traffic was used in this review. The collection record has no total forwarded request count.
- `sanitize-report.py --check` and `scripts/check-zap-fixture.py` each exited 0. The genuine ZAP 2.17.0 Traditional JSON at both digests contains one site, eight alerts, 21 observations and zero unlocated alerts. The site and all instance URI origins matched the registered origin. An independent normalization preserved all identity inputs and pointer groups across the transformation. The 17 groups comprised 14 singletons, two pairs and one triple. Compact JSON of the sorted pointer groups has SHA-256 `aed129f0ac433c939bce5db735d949aea31311f503cf8011158d64073b0bb5f3`.
- An independent structural diff matched the documented sanitization, including 24 query value replacements, 21 `nodeName` removals and replacements for nonempty scanner evidence or other-information fields. The retained reference link containing a query matched a built-in ZAP passive rule reference. The inspected derivative contained no matching email addresses, secret headers, JWT-like strings, private local paths or literal IPv4 addresses. No raw contents or individual observations are reproduced here.
- **V016 (P3), open.** `docs/status.md:79` says authorization, times, versions and sanitized hashes for genuine data are unavailable. The genuine section immediately below records those facts. Remove the stale sentence and keep the synthetic test status historical.

**Verdict:** the saved authorization chronology, genuine export, isolation and shutdown records, provenance, and sanitized grouping passed independent review at checkpoint 20. V016 remains open in the documentation. Genuine-data application import, duplicate behavior, integration fixture availability and browser persistence are separate acceptance checks.

## Independent verification — M1-20261123-21

Verified on 2026-11-23 at baseline `0c193d51c989866e8d49932454803853cb16df11`. All 17 frozen SHA-256 values matched before and after the checks. Shell syntax, Python source parsing, `node --check`, and the scoped `git diff --check` passed. All 18 relative links and anchors resolved.

- **V016 resolved.** The synthetic section now identifies those runs as authored-fixture tests and omits the stale sentence about unavailable genuine metadata. The separate genuine section records observed results and pending gates.
- `env -u TEST_DATABASE_URL GOFLAGS=-mod=readonly sh artifacts/zap/diploma-m1-passive-cef3083eaa/verify-offline.sh` exited 0 on a fresh, loopback-only cluster. Goose applied revisions 1–4. Four concurrent first uploads produced one creation and three duplicates sharing one import ID. A sequential duplicate reused the ID, conflicting provenance returned 409, and PostgreSQL metadata and the explicit evidence endpoint's bytes matched the submitted derivative.
- The paginated API and PostgreSQL each matched every one of the 17 target-specific fingerprints and all 21 unique, correctly grouped source pointers. The expected grouping in the harness also matched the independent pointer partition from checkpoint 20. The API verified unreviewed state and null assessment summaries on all findings.
- Chromium selected the genuine derivative and its provenance, displayed the duplicate result, and showed all 17 rows and source evidence after page reload and after API and PostgreSQL restart. The browser checks observed no external page requests. During the database outage, health was 200, readiness 503, and the missing-bearer request 401. No scanner or lab target was started or contacted.

**Verdict:** checkpoint 21 independently verified at the recorded hashes. No new findings. Genuine offline import and persistence passed locally. A reusable genuine repository integration fixture and automated coverage using it remain pending.
