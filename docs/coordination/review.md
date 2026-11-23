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
