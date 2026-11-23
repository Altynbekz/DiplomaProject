# M0 and M1 implementation status

Updated 2026-11-23, Asia/Almaty. This status covers the foundation and offline import. This continuation started from `94f431f479b7b707c84ee9d79ebf5f6710aa9ced`, which matched GitHub `main`.

| Milestone | Implementation | Verification and remaining gates |
| --- | --- | --- |
| M0 | Go, React and PostgreSQL startup, four goose migrations, least-privilege database login, sqlc, OpenAPI types, local secrets, pinned CI and reproducible setup | The host setup and M0 behavioral gates passed, including disposable Compose startup and restart on ARM64. Hosted CI passed for `94f431f`, including the Python tests in `make check`. [GitHub Actions run 45170314104](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104) |
| M1 | Offline Traditional JSON import, immutable evidence, provenance, duplicate handling, normalized observations, audited private downloads, filtered and paged UI, reusable genuine fixture | The synthetic and genuine automated suites passed locally and in an independent isolated copy. Hosted CI for the published fixture revision remains pending. [Independent verification](coordination/review.md#independent-verification--m1-20261123-22) |
| M2–M6 | Planned | No implementation started. |

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
