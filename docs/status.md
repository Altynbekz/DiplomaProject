# Implementation status

Updated 2026-11-24, Asia/Almaty. This status covers the foundation, offline import, M2 review workflow and the M3 assessment workflow.

| Milestone | Implementation | Verification and remaining gates |
| --- | --- | --- |
| M0 | Go, React and PostgreSQL startup, four goose migrations, least-privilege database login, sqlc, OpenAPI types, local secrets, pinned CI and reproducible setup | The host setup and M0 behavioral gates passed, including disposable Compose startup and restart on ARM64. Hosted CI passed for `94f431f`, including the Python tests in `make check`. [GitHub Actions run 45170314104](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104) |
| M1 | Offline Traditional JSON import, immutable evidence, provenance, duplicate handling, normalized observations, audited private downloads, filtered and paged UI, reusable genuine fixture | Acceptance passed, with synthetic and genuine automated suites passing locally, independently and in hosted CI for `f5712a49b5676badb05486b30c73e0d4b916146b`. [Independent verification](coordination/review.md#independent-verification--m1-20261123-22) · [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45187645541) |
| M2 | Human verification, OWASP mapping and immutable context history | Software passed at reviewed baseline `817c0449bbb9f6c040ba847a41c6f125bbda7d97`. The later draft recovery repair and checks are recorded below. M2 acceptance remains open pending authorized A01 role/object evidence. [Independent baseline review](coordination/review.md#independent-verification--m2-20261124-30) · [Hosted baseline run 45260100347](https://github.com/Altynbekz/DiplomaProject/actions/runs/45260100347) |
| M3 | Offline scoring, analyst authoring, immutable assessments and paired evidence history | Foundation accepted at `0cdf484759ed8f03253850ba213016311322f39f`. Analyst workflow passed local acceptance checks. Independent verification and hosted CI for this workflow are tracked separately. [M3 workflow](#m3-analyst-workflow-and-acceptance) |
| M4–M6 | Planned | No implementation started. |

## Recorded checks

Executed on macOS arm64 with Go 1.27.1, PostgreSQL 15.19, goose 3.30.1, sqlc 1.31.1, Node 26.13.0, npm 12.2.1, Playwright 1.65.1 and Chromium 159.0.8154.6. The current continuation also observed Docker Engine 29.7.2 on ARM64 and Compose 5.5.1 after starting Docker Desktop.

| Command | Result |
| --- | --- |
| `make dev-local` followed by loopback HTTP smoke checks | Installed the locked frontend, created the local cluster, applied migrations 1–3 and started Vite and the API. Health, readiness, frontend and authenticated empty target list returned 200. Missing token returned 401 through the Vite proxy. Stopped the development services after the check and retained the empty database files. |
| `make check` | Exit 0. Go format and vet, frontend format, strict TypeScript and ESLint passed. sqlc and OpenAPI generation produced identical bytes. Go tests passed and all three Vitest tests passed. The database suite runs separately. |
| `cd frontend && npm run build` | Exit 0. TypeScript check and Vite production build passed. |
| `make test-integration` | Exit 0 under the Go race detector. Migrated empty PostgreSQL to revision 1, upgraded to revision 3 and then revision 4, and repeated with no migration pending. Four concurrent uploads returned one creation and three duplicates of the same import. Stored bytes matched the upload. Verified finding and observation pagination and filters, invalid input without partial inserts, runtime write restrictions, target consistency and audit on private downloads. |
| `make test-e2e` | Exit 0. Two Chromium tests passed against disposable PostgreSQL and the same-origin Vite proxy. Import and duplicate reuse, filtering, token cleared on reload, escaped scanner text, acknowledgement-gated download with bytes matching the original, and persisted rows after API and database restart passed. With PostgreSQL stopped, health was 200, readiness was 503, and missing bearer credentials returned 401. |
| `python3 scripts/test-compose.py --config-only` | Exit 0. Validated a random project name, isolated networks and volume, disposable credentials, ephemeral loopback bindings and no published PostgreSQL port. |
| `python3 -u scripts/test-compose.py` | Exit 0 on Docker Desktop with ARM64 Docker Engine 29.7.2 and Compose 5.5.1. Applied migrations through revision 4, served the frontend and its TypeScript entry, returned health and readiness 200 and missing-token 401, and imported synthetic data through the frontend proxy. The exact duplicate returned the same import ID and the database held one report and five observations. Shutdown preserved the volume. Restart returned the same digest and four findings. Final cleanup removed the disposable project resources. |

Both bounded Go fuzz runs exited 0 on 2026-11-23. From `backend/`, run `GOCACHE=/private/tmp/diploma-go-cache go test ./internal/zap -run '^$' -fuzz '^FuzzDecode$' -fuzztime=3s -parallel=2` for the report decoder (23,459 executions) and `GOCACHE=/private/tmp/diploma-go-cache go test ./internal/scope -run '^$' -fuzz '^FuzzParse$' -fuzztime=3s -parallel=2` for URL parsing (276,635 executions). The first attempts could not write into the default user cache and were rerun with the temporary cache.

An earlier integration run found that PostgreSQL requires column-level UPDATE permission for the target row lock. Migration 3 grants only UPDATE on the target display name. The runtime still cannot update or delete evidence. An earlier Chromium run failed because the import confirmation and findings loading messages shared the status role. The test now selects the import message and passed the rerun.

## Review follow-up

Review pass 1 withheld approval on checkpoints 1–3. Checkpoints 5 and 6 record the implementation responses to R001–R009. Subsequent independent checks are recorded below.

On 2026-11-23, `make check`, `cd frontend && npm run build`, `make test-integration`, `make test-e2e` and `docker compose config --quiet` exited 0 after the review fixes. The PostgreSQL suite verified migration 4, all ten malformed report cases without partial writes, a successful import with unusual query values, ordered observation history across two reports with the clock fixed, import-scoped severity, and unchanged provenance after a rejected duplicate. Chromium verified upload/reload and persistence after API and database restart. A three-second URL fuzz run after the query parsing fix passed with 290,454 executions.

## Checkout continuation

Baseline `21a096f899ed43b340f26fbf6cceabf288f46530` on `main` matched GitHub's branch head through read-only inspection. The two existing untracked frontend and backend guides were left intact.

The redesigned frontend keeps the same evidence workflow and protected download. Chromium rendered unlock and findings at 1280×900 and findings and evidence at 390×844. The mobile page stayed within the viewport, with horizontal scrolling confined to the findings table. Both browser cases passed using `E2E_SCREENSHOTS=/private/tmp/diploma-ui-21a096f make test-e2e`. The suite exercised keyboard unlock and rejection of an invalid token, import and idempotency, immutable provenance conflict, filtering and return navigation, safe scanner text, audited download, outage behavior and persistence across restart. The three Vitest tests and the frontend build passed. Screenshots show the synthetic fixture only.

`make test-integration` reproduced six escaped-NUL HTTP 500s before the fix. The last run passed under the race detector, with precise 422 pointers and no writes for the invalid fields, preserved raw bytes, and passing concurrency and authorization checks. The verifier independently confirmed this at checkpoint 07. The verifier reproduced V010 and V011. The implementation suite reproduced them again and then passed with both fixes on disposable PostgreSQL. Large numbers in unlocated alert extensions now round-trip through a versioned JSON-text envelope. Provenance over the database-rendered 64 KiB limit returns 413. The final `make check` passed with identical generated files and three Vitest tests.

Independent verification of checkpoint 09 passed with matching file hashes and `GOFLAGS=-mod=readonly make test-integration`. V010 and V011 are resolved. The verifier also found V012, where an invalid cursor could leave the findings page without an in-app recovery action. Checkpoint 10 restores the First page control after errors. The final Chromium test reproduced the invalid-cursor response and verified that the control returns to the original persisted rows. Source, formatting, build and browser checks passed on the final combined implementation. The verifier independently ran `GOFLAGS=-mod=readonly make test-e2e` against checkpoint 10 at matching hashes. Both browser phases passed and V012 is resolved. No scoped review findings remain open.

The scoped changes and review conclusions are recorded in the [coordination log](coordination/status.md) and [independent verification log](coordination/review.md).

## Current continuation and CI

- The root and layer guides now link to the actual guides, contract and status files. Checkpoint 11 independently verified all 13 relative links.
- [GitHub Actions](../.github/workflows/ci.yml) configures Ubuntu 24.04 checks with pinned actions and tools. It runs formatting, lint, type checking, unit tests, generated-file consistency, the frontend build, and separate disposable PostgreSQL and Chromium suites. The workflow references no real secrets or scans and keeps PostgreSQL under a non-root user.
- `actionlint` 1.7.7, `sh -n scripts/install-test-postgres.sh`, `GOFLAGS=-mod=readonly make check`, and `npm run build --prefix frontend` exited 0. The PostgreSQL 15.19 release archive matched its pinned SHA-256 and built and installed to a temporary prefix. Both `GOFLAGS=-mod=readonly make test-integration` under the race detector and `GOFLAGS=-mod=readonly make test-e2e` passed using that installation. The suite checked migrations 1 through 4, the readiness outage response and both Chromium phases. The browser run used local execution with loopback binding enabled after the default sandbox rejected `socket.bind`.
- Independent checkpoint 12 found V013. Checkpoint 13 independently verified the fix, including the pinned sqlc and goose checks with ripgrep absent from PATH.
- GitHub marked run `45166744371`, attempt 1, `completed` with conclusion `success` for `5a1cd7427fbfca196d3dc59bea65555a256046e6` on `main`. The last run update was `2026-11-23T11:42:06Z`. The `check` job and all listed steps succeeded, including pinned tool installation, generated-file comparison, frontend build, isolated PostgreSQL migrations and integration tests, and the Chromium suite. GitHub listed one run for this commit and no failures. No CI fix or rerun was needed. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45166744371)
- GitHub marked run `45168962574`, attempt 1, `completed` with conclusion `success` for `4d0240c4f50559b0e21feeeead7348c03cd74a7f` on `main`. The last run update was `2026-11-23T11:54:53Z`. Job `137673174362` and all 15 listed steps passed, including isolated PostgreSQL integration and Chromium. GitHub listed one run for this exact commit. The baseline's workflow did not run the Python unit tests. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45168962574)
- GitHub marked run `45170314104`, attempt 1, `completed` with conclusion `success` for `94f431f479b7b707c84ee9d79ebf5f6710aa9ced` on `main`. The last run update was `2026-11-23T12:07:24Z`. Job `137677821087` and all 17 listed steps passed, including Python setup, nine `unittest` tests in `make check`, PostgreSQL integration and Chromium. This was the only run listed for the exact commit. The independent verifier checked the job log. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104)

## Python script tests in normal verification

`make test-scripts` runs the nine synthetic fixture preflight tests with standard-library `unittest`. `make test` depends on it and `make check` includes it through `make test`. Host verification uses Python 3.11 or newer. The updated workflow pins CPython 3.14.3 with a commit-pinned setup action, and uses the existing `make check` step to run the suite.

Locally on Python 3.14.3, `make test-scripts`, `GOFLAGS=-mod=readonly make test` and `GOFLAGS=-mod=readonly make check` exited 0. Generation produced identical files, Go tests passed, and all three Vitest tests passed. In a temporary copy of the checkout, one intentionally failing Python assertion made both `make test` and `make check` exit 2. `actionlint` 1.7.7, the tool manifest JSON check and `git diff --check` exited 0.

`npm run build --prefix frontend`, `GOFLAGS=-mod=readonly make test-integration` and `GOFLAGS=-mod=readonly make test-e2e` exited 0. Disposable PostgreSQL applied migrations through revision 4 and passed the synthetic integration suite under the Go race detector. Both Chromium phases and the database outage check passed. Hosted run `45170314104` later verified the Python test wiring on `94f431f`.

## Go workspace

The root `go.work` selects the module in `backend/`. From the repository root, `go test ./backend/...`, `go build ./backend/...`, and `go vet ./backend/...` exited 0 with Go 1.27.1. `make check` also exited 0 with byte-identical generated files. The local guard and canary build as separate packages in their own ignored workspace, where `go test ./...` and `go vet ./...` passed.

## Container startup fixes

- The original frontend image denied write access to `node_modules` as its non-root runtime user. Vite reproduced the cache write error in a disposable checkout. With the build and application files owned by `node`, the same image probe passed as that runtime user with networking disabled.
- With both Compose networks marked internal, Docker reported no effective host ports for the API or frontend even after both health checks passed. The app now uses a project-local bridge for loopback publishing. PostgreSQL and migrations remain on the separate internal storage network, with no database port exposed. The Compose smoke test checks the effective configuration and the actual host bindings.
- The passing smoke run used project `diploma-smoke-667b49d98d434dbab875`, disposable credentials and a fresh volume. Docker Desktop on Apple Silicon was the observed ARM64 environment. AMD64 and emulated Compose engines are outside the supported configuration.

## Synthetic fixture and artifact hashes

The URL validation follow-up passed `make check`, `make test-integration` and `make test-e2e`. The regression test first reproduced acceptance of percent-decoded NUL and invalid UTF-8 query names. The fixed parser returned `422 validation_failed` at the instance URI pointer for both cases. The integration test verified zero import, finding and observation rows after each rejected request. The same bounded URL fuzz command above passed again with 288,782 executions.

The synthetic browser and PostgreSQL tests used authored data and observed **5 observations**, **4 findings**, **1 unlocated alert**, and an unknown severity warning. The fixture's run involved no scanner. See [fixture provenance](fixture-provenance.md).

| Artifact | SHA-256 |
| --- | --- |
| `backend/testdata/zap/synthetic-traditional.json` | `c91f5451043fcac49e8246e04dedb90eb362ebe983d064600a6914330eaa304a` |
| `contracts/openapi.yaml` | `5bfed9a5e7383cdd80e0805762a39d141fcd1c81b6277dc17a8886ec57e47c4a` |
| `backend/internal/store/sqlc/imports.sql.go` | `ff1c65eee2854d201d7fe4a2ad62a72fec9c256aca454f60b191f40234e28714` |
| `frontend/src/api/schema.d.ts` | `2d9aa5630be8a3c631cbea8e286a8189a33e92893bd426d1de5f5d69f9c900e3` |

## Genuine-report acceptance readiness

One passive session observed ordinary browser traffic for `http://juice-shop:3000` in the isolated local lab. The named authorization `M1-20261123T132231Z-cef3083eaa` was valid from `2026-11-23T13:22:31.466198Z` through `2026-11-23T13:52:31.466198Z`. The isolation recheck passed before collection. The allowed Chromium → ZAP → guard path worked, and the denied canaries received zero requests. Three concurrent allowed canary requests took 1.084 seconds. The guard unit tests covered the request limit, rate and deadline.

The capture ran from `2026-11-23T13:25:43.520670Z` through `2026-11-23T13:26:59.738124Z`. ZAP 2.17.0 completed its passive queue wait and Traditional JSON export. Docker confirmed that all lab containers stopped. The report contains one site, eight scanner alerts, 21 observations and zero unlocated alerts. The site and all observation URI origins match `http://juice-shop:3000`. Three observations describe browser POST attempts to `/socket.io/`. The origin guard forwards only GET and HEAD. The session did not record a total forwarded request count.

The ignored lab directory `artifacts/zap/diploma-m1-passive-cef3083eaa/` holds `original.json`, `sanitized-traditional.json` and the import `provenance.json`. The private original matched the stopped ZAP container export byte for byte. The sanitizer reproduced the derivative and provenance, removed query values and private scanner text, and preserved alert and instance order and grouping inputs. The offline `scripts/check-zap-fixture.py` check exited 0 for the three files. A separate derivation from the report found 17 expected finding groups across the 21 source pointers.

| Artifact | SHA-256 |
| --- | --- |
| `original.json` | `121e82b36df0b343f43015b9faed851e3bc8fc78c6dc91e3caed867ea8406bad` |
| `sanitized-traditional.json` | `036a4eb798d55fa68fa56a9389933b52b2bc3fefe86fd35eff741bb15efebdf7` |

`python3 -B -m unittest discover -s scripts -p 'test_check_zap_fixture.py' -v` previously exited 0 with nine synthetic preflight tests. The CLI checks digests, shape, and source positions. Review finding V014 led to a fixed diagnostic that keeps unexpected argument values private, with a regression test. Independent checkpoint 20 matched the original to the stopped container, verified authorization and chronology, and confirmed the sanitized report structure and grouping. [Independent verification](coordination/review.md#independent-verification--m1-20261123-20)

`env -u TEST_DATABASE_URL GOFLAGS=-mod=readonly sh artifacts/zap/diploma-m1-passive-cef3083eaa/verify-offline.sh` exited 0 on a disposable PostgreSQL cluster. Migrations reached revision 4. Four concurrent genuine uploads returned one creation and three duplicates with one import ID. A later duplicate reused that ID and conflicting provenance returned 409. The stored provenance and audited download matched the derivative submission exactly. The API and PostgreSQL matched all 17 independently derived fingerprints and their 21 observation pointers. Chromium selected the derivative and submitted its provenance, saw the duplicate result, and displayed 17 findings and source evidence after both reload and API/database restart. Browser page requests stayed on loopback. The database outage check returned health 200, readiness 503 and missing-bearer 401.

Independent checkpoint 21 reran the disposable acceptance command successfully and matched all 17 fingerprints and 21 pointers, provenance, duplicate handling, and browser persistence across restart. [Independent replay](coordination/review.md#independent-verification--m1-20261123-21)

The reviewed derivative, shareable provenance and independent fingerprint oracle are available under `backend/testdata/zap/`. Reproduction from the private original matched the fixture byte for byte. The normal integration command uses separate disposable databases for the synthetic and genuine imports. Its genuine test returned one creation and three duplicates under concurrency, reused the ID on a subsequent upload, rejected provenance changes, preserved the exact submitted bytes, and matched all 17 fingerprints and 21 pointers in the API and PostgreSQL. Both Chromium suites passed import, reload and independent API/database restart checks. `GOFLAGS=-mod=readonly make check`, `npm run build --prefix frontend`, `make test-integration` and `make test-e2e` exited 0 locally. [Fixture provenance](fixture-provenance.md)

## Human verification

Migration 5 adds append-only verification reviews. The runtime role can create and read reviews and can lock a finding through a narrow database function. The role cannot edit, delete or truncate review records or update evidence rows. Every accepted review validates its observation UUID and exact source pointer against the finding and writes an audit row in the same transaction. The finding's latest human review can be filtered separately from scanner severity.

GOFLAGS=-mod=readonly `make check`, `make test-integration`, `make test-e2e` and `npm run build --prefix frontend` exited 0 locally. Goose upgraded through migration 4 to 5 and also migrated a fresh genuine-test database. The API suite checked all four review states, evidence-reference failures, newest-first history, stale predecessors, two concurrent requests sharing a predecessor, and runtime write restrictions. Chromium covered two tabs editing the same finding, preservation and recovery of a stale draft, and persisted history following restart. Tests used isolated databases and no scanner traffic.


## Reviewed OWASP classification

Migration 6 adds immutable classification revisions. The versioned data file is embedded in the backend for host and container builds. The existing path under configs links to that canonical file. The rules match only plugin and CWE identifiers. A matched rule set records its sources, rationale, supporting observations, and the independent policy review bound to the exact policy digest. Conflicts preserve all candidate study categories and an ambiguous state. An unmatched historical finding stays unmapped until it receives an operator revision or a new distinct report supplies usable evidence. Exact report duplicates retain their existing identity. A manual head takes precedence over subsequent automatic rules.

The genuine fixture maps five findings to A01:2025 and one to A02:2025, places one diagnostic outside study scope, and leaves ten findings unmapped. These are classifications of scanner assertions. The disposable integration test uses a separately authored contradictory CWE to verify that both categories survive in ambiguity and neither enters category-filtered results. It also checks audit history, same-finding references, stale predecessors, runtime immutability, and manual precedence. No live role/object validation was performed. [A01 test protocol](../lab/scenarios/a01-role-object.md) · [Mapping rules](../configs/owasp-mapping.v1.yaml)

`GOFLAGS=-mod=readonly make check`, `npm run build --prefix frontend`, `make test-integration` and `make test-e2e` exited 0 locally for this slice. Goose applied migration 6 on clean and upgraded schemas. Both Chromium suites passed, including the genuine classification form and retained history after restart. Hosted CI passed for human verification at `948351c45008048c601a827594d3ddf84d1a1cd8` and for classification at `20c6a76536c3b2d921ef3a27a36b96b119c0af13`. [Verification CI](https://github.com/Altynbekz/DiplomaProject/actions/runs/45193568143) · [Classification CI](https://github.com/Altynbekz/DiplomaProject/actions/runs/45195527459)


## Application context

Migration 7 adds immutable target-default and exact-method/path context revisions. The API requires the four context factors and per-factor rationales. Known factors require same-target observation evidence or a lab document reference with a SHA-256 digest. The resolver selects the latest matching endpoint revision, then the latest target default, then four unknown factors. It reads the finding's canonical endpoint and does not write history. Revisions use per-scope predecessor checks under a target lock and each append writes an audit event.

`GOFLAGS=-mod=readonly make check`, `make test-integration`, `make test-e2e` and `npm run build --prefix frontend` passed locally. Goose upgraded through revision 6 to 7 and created a clean dedicated database. The race-mode API test checked missing provenance, invalid enums and scopes, foreign-target observations, exact-over-default precedence, unknowns, concurrent stale predecessors, pagination, immutable runtime permissions and audit rows. Chromium created target and endpoint contexts, preserved a stale draft through conflict recovery, displayed the exact source and unknown factors, and verified context, verification and classification after restart. The editor fit a 390-pixel viewport. Tests used disposable stacks. No new scans or live role/object checks were performed.

The final context checks also verified HTTP 413 when PostgreSQL's rendered provenance exceeds 64 KiB, rejection of non-ASCII methods, and exact handling of an incompressible 3,841-byte path. A bounded SHA-256 index key supports the full path limit while the resolver compares the actual path. Chromium held scope controls stable during a delayed save. No production data or credentials were used. [Independent review](coordination/review.md)

## M2 acceptance audit

Audited against `AGENTS.md` §11 at `da9844eabb14c6c14646b407c652864afc0012e9`. GitHub `main` matched the checkout and run `45197471901` completed successfully for that SHA. The acceptance audit changed documentation only. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45197471901)

| Gate | Observed result |
| --- | --- |
| Genuine report import remains reproducible | Pass. PostgreSQL and the API matched 17 fingerprints and 21 observation pointers. Concurrent duplicates reused the import ID. Chromium showed the findings after reload and after restarting PostgreSQL and the API. The checked fixture SHA-256 was `036a4eb798d55fa68fa56a9389933b52b2bc3fefe86fd35eff741bb15efebdf7`. |
| Confirmed, false-positive and inconclusive reviews remain in history | Pass. The disposable workflow appended each state, retained each prior revision and its evidence reference, and read them in order. The browser found the saved review and classification histories after restart. Test states are workflow assertions on disposable data. |
| Wrong-finding evidence references fail | Pass. HTTP 422 for a foreign-finding reference and a mismatched source pointer in verification. Classification rejected evidence from a different finding, and context rejected an observation from another target. |
| Stale predecessors return a recoverable 409 | Pass. API tests rejected stale verification, classification and context revisions. Concurrent verification and context requests sharing a predecessor produced one 201 and one 409 per pair. Chromium recovered the review and context drafts. |
| Default and exact endpoint context resolve deterministically, including unknown values | Pass. A newer target default applied to unmatched endpoints. The exact method and path still used their own revision and retained unknown factors and their provenance. Chromium showed resolved context after restart. |
| Histories are append-only and writes are audited | Pass. The runtime grants allow selection and insertion on all three revision tables. Tests exercised denied mutations, oldest and newest records retained their values, and accepted appends recorded audit events. Fresh and staged migrations reached revision 7. |
| OWASP conflicts remain ambiguous and severity stays separate | Pass. Competing rules left both candidate categories with no mapped category. A manual classification survived a subsequent import, with scanner severity and human verification unchanged by the classification edit. |
| A01 role/object claims have manual evidence | **Open.** The scoped artifact inventory found the passive capture and a private record of the authorized standalone fixture attempt. Automatic approval review blocked that attempt before execution. The record contains zero live requests and no observed decisions. An in-memory test passed on fictional cases. Independent review determined that the evidence gate remains open. The genuine report's A01 CORS assertion remains separate. [Operator evidence checklist](../lab/scenarios/a01-role-object.md#evidence-to-supply) |

Local commands and exit status on macOS arm64 with Go 1.27.1, PostgreSQL 15.19, goose 3.30.1, sqlc 1.31.1, Node 26.13.0, npm 12.2.1 and Python 3.14.3.

| Command | Result |
| --- | --- |
| `GOFLAGS=-mod=readonly make check` | Exit 0. Go formatting and vet, frontend formatting, TypeScript and ESLint, byte-identical generation, Go unit tests, nine Python tests and four Vitest tests passed. |
| `npm run build --prefix frontend` | Exit 0. TypeScript and the Vite production build passed. |
| `env -u TEST_DATABASE_URL -u TEST_GENUINE_DATABASE_URL GOFLAGS=-mod=readonly make test-integration` | Exit 0. Disposable PostgreSQL migrated from empty and through revision 6 to 7. Both import suites and all three M2 workflow subtests passed under the race detector. |
| `GOFLAGS=-mod=readonly make test-e2e` | Exit 0. Four Chromium tests passed, including both restart cases. Both outage checks returned health 200, readiness 503 and missing bearer 401. |

### Draft recovery follow-up

The working-tree audit started from `817c0449bbb9f6c040ba847a41c6f125bbda7d97`, which matched GitHub `main` at inspection. The genuine Chromium regression exited 1 before the fix. After an endpoint revision received HTTP 409, selecting a revision for the target default changed the draft to that scope. Recovery now permits only a revision matching the draft scope, and preserves its factors and provenance. Endpoint recovery uses the method and path shown in history. [Hosted baseline run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45260100347)

`GOFLAGS=-mod=readonly make check` and `npm run build --prefix frontend` exited 0. Go and frontend formatting, vet, strict TypeScript, ESLint, byte-identical generation, Go unit tests, nine Python tests and four Vitest tests passed. `env -u TEST_DATABASE_URL -u TEST_GENUINE_DATABASE_URL GOFLAGS=-mod=readonly make test-integration` exited 0 under the race detector. Fresh and staged migrations reached revision 7. Concurrent classification writes with a shared predecessor produced one 201 and one 409. The runtime role received SQLSTATE 42501 for `UPDATE`, `DELETE` and `TRUNCATE` on classification history.

`GOFLAGS=-mod=readonly make test-e2e` exited 0. All four Chromium cases and both outage checks passed. The genuine workflow kept a conflicting draft on its endpoint, saved it as the next revision, and recovered a separate classification draft by keyboard. Both classification revisions were present after restart. Tests used disposable application stacks and the checked-in sanitized report. No live A01 checks ran.


## M3 assessment foundation

The initial baseline is `3dedfa8d45554844c421f263e6db3ce7efe0dbeb`. The adapter uses unmodified FIRST scoring files at `c5b0d409ae9f57c44264c6ce5f27d89298e1d32a` and the pinned pure-Go Goja engine `v0.0.0-20261122185237-35fdc4b88e39`. Thirteen Node-generated reference cases passed offline, including defined and inherited Environmental metrics, Safety, zero and rounding regressions. The verifier separately reproduced 4,096 vectors and scores and ran concurrent evaluations with the race detector. [Calculator decision](decisions/0002-cvss-v4-calculator.md) · [Independent calculator review](coordination/review.md#independent-verification--m3-20261124-34)

Migration 8 adds append-only snapshots and a current staleness view. The API captures review, classification, exact Base parent and context links. It validates expected revisions, evidence, justification and all four BE dispositions before calculating. Failed checks persist no score. An incomplete save persists a new snapshot with null vector, score and calculator. The latest finding summary retains an incomplete newest snapshot. The frontend can inspect histories, metric rationales, vectors and provenance. [API contract](../contracts/openapi.yaml) · [Context policy](assessment-context-policy.md)

| Command | Local result |
| --- | --- |
| `GOFLAGS=-mod=readonly make check` | Exit 0. Go/Prettier formatting, vet, TypeScript, ESLint, byte-identical generation, Go tests, nine Python and four Vitest cases passed. |
| `npm run build --prefix frontend` | Exit 0. TypeScript and production build passed. |
| `env -u TEST_DATABASE_URL -u TEST_GENUINE_DATABASE_URL GOFLAGS=-mod=readonly make test-integration` | Exit 0. Fresh and staged migrations reached 8. All seven populated M2 table snapshots matched across v7 to v8 and the new assessment history was empty. The normal and genuine report suites and the synthetic assessment suite passed under the race detector. Auditing failure rolled back the append. |
| `GOFLAGS=-mod=readonly make test-e2e` | Exit 0. Four Chromium cases passed with both outage checks. The synthetic UI test used Node/FIRST parity values, opened history by keyboard, preserved incomplete and historical state and verified snapshots after restart. The genuine workflow remained unchanged. |

These runs used disposable macOS arm64 resources with Go 1.27.1, Node 26.13.0, PostgreSQL 15.19, goose 3.30.1, Playwright 1.65.1 and Chromium 159.0.8154.6. All complete test assessments used explicitly authored synthetic data. The live A01 gate remains open. The foundation preceded analyst-facing authoring. No ranking or research experiment was run.


### Assessment history refresh follow-up

Checkpoint 36 independently verified the calculator, migration and storage/API behavior, with V027 open for browser freshness. The API correctly reported historical status after a context or human review change. The open page still held the old query result. After a verification or classification save, the finding's open assessment history refetches. A new context revision refreshes the target's findings table. The synthetic Chromium regression confirmed that those views update without reloading them. `GOFLAGS=-mod=readonly make test-e2e` passed all four cases and both outage checks. `npm run --prefix frontend fmt:check`, `npm run --prefix frontend lint` and `npm run --prefix frontend build` exited 0. [Independent findings](coordination/review.md#independent-verification--m3-20261124-36)

## M3 analyst workflow and acceptance

The authoring slice starts from `0cdf484759ed8f03253850ba213016311322f39f`, which passed [hosted run 45378616340](https://github.com/Altynbekz/DiplomaProject/actions/runs/45378616340). It reuses the accepted offline calculator, HTTP contract, migrations and transaction logic. The browser sends choices, evidence and expected revisions. The server verifies eligibility, resolves context and creates a new snapshot.

| M3 gate | Evidence on the authoring slice |
| --- | --- |
| FIRST parity and provenance | All 13 recorded parity cases carry forward, including boundaries, zero, modified-value inheritance, Environmental requirements, legal Safety values and ordering. The pinned source and adapter are unchanged. Checkpoints 34–37 established byte-identical regeneration and 4,096 independent comparisons. [Reference fixture](../backend/testdata/cvss/first-v4.json) · [Calculator review](coordination/review.md#independent-verification--m3-20261124-34) |
| Explicit missing values and legal choices | Chromium selected supported values, saved unselected Base metrics as omissions and read null score and vector on incomplete B and BE. Unit tests verify omitted implicit Environmental defaults and retained explicit Not Defined rationale. Duplicate and unsupported inputs still fail at the server. |
| Review, rationale, evidence and context | The browser rendered server rejection for missing rationale or factor dispositions and unknown BE context. Database tests rejected an unmapped current classification, a mismatched source pointer and a correctly addressed observation from another finding. |
| Exact parent and factor dispositions | Chromium selected the Base parent by snapshot, checked inherited and default labels, supplied contextual and observation support and explained all four factors. The completed BE matched the pinned fixture and retained the selected parent's Base values and rationales. |
| Stale inputs and immutable history | HTTP 409 recovery retained edits across verification, classification, context and Base-parent changes. Loading current revisions and selecting them are distinct actions. The browser kept the BE draft while revising Base. Completed and draft snapshots remained distinct, with unchanged vectors, scores and provenance after later changes. Runtime mutation remained denied and a failed audit append rolled back. |
| Readable and accessible pair | Chromium used keyboard controls at narrow width, checked the open editor for page overflow, read historical labels and inspected the exact paired Base after later revisions. The read API and UI retained a score of zero as distinct from absent. The display uses the server result. |

| Command | Local result |
| --- | --- |
| `GOFLAGS=-mod=readonly make check` | Exit 0. Go and Prettier formatting, vet, strict TypeScript and ESLint passed. sqlc and OpenAPI generation reproduced identical bytes. Go tests, nine Python tests and eight Vitest cases passed. |
| `npm run build --prefix frontend` | Exit 0, production Vite build passed. |
| `env -u TEST_DATABASE_URL -u TEST_GENUINE_DATABASE_URL GOFLAGS=-mod=readonly make test-integration` | Exit 0. Fresh, staged and populated PostgreSQL upgrades reached schema 8. Both import suites, M2 workflow subtests and M3 assessment regression passed under the race detector. |
| `GOFLAGS=-mod=readonly make test-e2e` | Exit 0. Four Chromium cases and both outage checks passed. The synthetic authoring and genuine-report flows persisted across API/PostgreSQL restart. |

The checks ran on macOS arm64 with Go 1.27.1, PostgreSQL 15.19, Node 26.13.0, Playwright 1.65.1 and Chromium 159.0.8154.6. The pinned FIRST fixture SHA-256 is `c292463487db168c2d121067b8b9148d8a772c522a817a52fde95778ee853dd8`. Completed assessments in these checks are synthetic. The M2 live A01 role/object evidence gate remains open. M4–M6 remain planned. [Demo walkthrough](../README.md#author-and-inspect-cvss-assessments)

### Independent-review follow-up

Checkpoint 38 isolated the accepted authoring flow and found three editor issues. V028 now requires the loaded revision IDs to match the visible current inputs before they can be adopted. Chromium reproduced the original race in a separate copy, then verified a fresh reload and adoption after a same-page verification change. The resulting Base pinned the newly displayed review ID. V029 now shows the stored observation ID on off-page selections. A DOM test removed the intended reference when two selections had the same source pointer. V030 labels `N` as Negligible for MSC, MSI and MSA, with Base labels and actual vector codes unchanged. A focused test checks those names and the scope of Safety. [Independent findings](coordination/review.md#independent-verification--m3-20261124-38) · [FIRST CVSS v4.0 specification, Table 15](https://www.first.org/cvss/v4.0/specification-document#Modified-Base-Metrics-and-Safety)

The repaired snapshot passed `GOFLAGS=-mod=readonly make check`, `npm run build --prefix frontend` and `GOFLAGS=-mod=readonly make test-e2e`. Manual inspection of the synthetic editor's 390px and 1280px renders confirmed the neutral layout, legible labels and visible focus. Checkpoint 39 is the independent retest.
