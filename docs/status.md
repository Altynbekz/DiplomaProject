# M0 and M1 implementation status

Updated 2026-11-23, Asia/Almaty. This status covers the foundation and offline import. This continuation started from `5a1cd7427fbfca196d3dc59bea65555a256046e6`, which matched GitHub `main`.

| Milestone | Implementation | Verification and remaining gates |
| --- | --- | --- |
| M0 | Go, React and PostgreSQL startup, four goose migrations, least-privilege database login, sqlc, OpenAPI types, local secrets, pinned CI and reproducible setup | The host setup and M0 behavioral gates passed, including disposable Compose startup and restart on ARM64. Hosted CI passed for `5a1cd74`. [GitHub Actions run 45166744371](https://github.com/Altynbekz/DiplomaProject/actions/runs/45166744371) |
| M1 | Offline Traditional JSON import, immutable original bytes, SHA-256, typed provenance, duplicate handling, normalized findings and observations, audited private downloads, filtered and paged UI | Executed using the labeled synthetic fixture. Acceptance is blocked by the absence of a genuine, provenance-backed Juice Shop report and its sanitized derivative. |
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
- [GitHub Actions](../.github/workflows/ci.yml) configures Ubuntu 24.04 checks with pinned actions and tools. It runs formatting, lint, type checking, backend and frontend tests, generated-file consistency, the frontend build, and separate disposable PostgreSQL and Chromium suites. The workflow references no real secrets or scans and keeps PostgreSQL under a non-root user.
- `actionlint` 1.7.7, `sh -n scripts/install-test-postgres.sh`, `GOFLAGS=-mod=readonly make check`, and `npm run build --prefix frontend` exited 0. The PostgreSQL 15.19 release archive matched its pinned SHA-256 and built and installed to a temporary prefix. Both `GOFLAGS=-mod=readonly make test-integration` under the race detector and `GOFLAGS=-mod=readonly make test-e2e` passed using that installation. The suite checked migrations 1 through 4, the readiness outage response and both Chromium phases. The browser run used local execution with loopback binding enabled after the default sandbox rejected `socket.bind`.
- Independent checkpoint 12 found V013. Checkpoint 13 independently verified the fix, including the pinned sqlc and goose checks with ripgrep absent from PATH.
- GitHub marked run `45166744371`, attempt 1, `completed` with conclusion `success` for `5a1cd7427fbfca196d3dc59bea65555a256046e6` on `main`. The last run update was `2026-11-23T11:42:06Z`. The `check` job and all listed steps succeeded, including pinned tool installation, generated-file comparison, frontend build, isolated PostgreSQL migrations and integration tests, and the Chromium suite. GitHub listed one run for this commit and no failures. No CI fix or rerun was needed. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45166744371)

## Container startup fixes

- The original frontend image denied write access to `node_modules` as its non-root runtime user. Vite reproduced the cache write error in a disposable checkout. With the build and application files owned by `node`, the same image probe passed as that runtime user with networking disabled.
- With both Compose networks marked internal, Docker reported no effective host ports for the API or frontend even after both health checks passed. The app now uses a project-local bridge for loopback publishing. PostgreSQL and migrations remain on the separate internal storage network, with no database port exposed. The Compose smoke test checks the effective configuration and the actual host bindings.
- The passing smoke run used project `diploma-smoke-667b49d98d434dbab875`, disposable credentials and a fresh volume. Docker Desktop on Apple Silicon was the observed ARM64 environment. AMD64 and emulated Compose engines are outside the supported configuration.

## Synthetic fixture and artifact hashes

The URL validation follow-up passed `make check`, `make test-integration` and `make test-e2e`. The regression test first reproduced acceptance of percent-decoded NUL and invalid UTF-8 query names. The fixed parser returned `422 validation_failed` at the instance URI pointer for both cases. The integration test verified zero import, finding and observation rows after each rejected request. The same bounded URL fuzz command above passed again with 288,782 executions.

The browser and PostgreSQL tests used authored data and observed **5 observations**, **4 findings**, **1 unlocated alert**, and an unknown severity warning. The fixture's run involved no scanner. Scan authorization, times, versions and sanitized hashes for genuine data are unavailable. See [fixture provenance](fixture-provenance.md).

| Artifact | SHA-256 |
| --- | --- |
| `backend/testdata/zap/synthetic-traditional.json` | `c91f5451043fcac49e8246e04dedb90eb362ebe983d064600a6914330eaa304a` |
| `contracts/openapi.yaml` | `5bfed9a5e7383cdd80e0805762a39d141fcd1c81b6277dc17a8886ec57e47c4a` |
| `backend/internal/store/sqlc/imports.sql.go` | `ff1c65eee2854d201d7fe4a2ad62a72fec9c256aca454f60b191f40234e28714` |
| `frontend/src/api/schema.d.ts` | `2d9aa5630be8a3c631cbea8e286a8189a33e92893bd426d1de5f5d69f9c900e3` |

## Genuine-report acceptance readiness

The search covered the checkout, Git history and the documented `artifacts/` and `backend/testdata/zap/` locations. The only report found was the labeled synthetic fixture. M1 acceptance needs the local path to a genuine ZAP Traditional JSON report from an authorized Juice Shop lab and its provenance, including the authorization reference, scan time and any recorded scanner, target or scan plan details.

`python3 -B -m unittest discover -s scripts -p 'test_check_zap_fixture.py' -v` exited 0 with nine tests for the offline preflight. They use synthetic inputs to check digest agreement, structure and source positions, required provenance, rejection of the repository's synthetic fixture, and preservation of both input files. The verifier found V014, an argument error that exposed an unexpected path. The revised CLI uses a fixed diagnostic, and the regression test checks that the path stays private. Candidate-specific sanitization, expected fingerprints, integration coverage and browser restart await the genuine input. [Acceptance checklist](fixture-provenance.md)
