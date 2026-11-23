# M0 and M1 implementation status

Updated 2026-11-23, Asia/Almaty. The workspace initially held only the root contract and had no Git metadata, application source or report. Work is limited to the foundation and offline import.

| Milestone | Implementation | Verification and remaining gates |
| --- | --- | --- |
| M0 | Go, React and PostgreSQL startup, four goose migrations, least-privilege database login, sqlc, OpenAPI types, local secrets and reproducible setup | The host setup and M0 behavioral gates passed. Compose config validated. Container boot remains unverified because the Docker daemon socket is unavailable. |
| M1 | Offline Traditional JSON import, immutable original bytes, SHA-256, typed provenance, duplicate handling, normalized findings and observations, audited private downloads, filtered and paged UI | Executed using the labeled synthetic fixture. Acceptance is blocked by the absence of a genuine, provenance-backed Juice Shop report and its sanitized derivative. |
| M2–M6 | Planned | No implementation started. |

## Recorded checks

Executed on macOS arm64 with Go 1.27.1, PostgreSQL 15.19, goose 3.30.1, sqlc 1.31.1, Node 26.13.0, npm 12.2.1, Playwright 1.65.1 and Chromium 159.0.8154.6. Docker CLI 29.7.2 and Compose 5.5.1 were available. The engine socket connection failed.

| Command | Result |
| --- | --- |
| `make dev-local` followed by loopback HTTP smoke checks | Installed the locked frontend, created the local cluster, applied migrations 1–3 and started Vite and the API. Health, readiness, frontend and authenticated empty target list returned 200. Missing token returned 401 through the Vite proxy. Stopped the development services after the check and retained the empty database files. |
| `make check` | Exit 0. Go format and vet, frontend format, strict TypeScript and ESLint passed. sqlc and OpenAPI generation produced identical bytes. Go tests passed and all three Vitest tests passed. The database suite runs separately. |
| `cd frontend && npm run build` | Exit 0. TypeScript check and Vite production build passed. |
| `make test-integration` | Exit 0 under the Go race detector. Migrated empty PostgreSQL to revision 1, upgraded to revision 3 and then revision 4, and repeated with no migration pending. Four concurrent uploads returned one creation and three duplicates of the same import. Stored bytes matched the upload. Verified finding and observation pagination and filters, invalid input without partial inserts, runtime write restrictions, target consistency and audit on private downloads. |
| `make test-e2e` | Exit 0. Two Chromium tests passed against disposable PostgreSQL and the same-origin Vite proxy. Import and duplicate reuse, filtering, token cleared on reload, escaped scanner text, acknowledgement-gated download with bytes matching the original, and persisted rows after API and database restart passed. With PostgreSQL stopped, health was 200, readiness was 503, and missing bearer credentials returned 401. |
| `docker compose config --quiet` | Exit 0. `docker info --format '{{.ServerVersion}}'` failed to connect to Docker's local socket, so no Compose startup was observed. |

Both bounded Go fuzz runs exited 0 on 2026-11-23. From `backend/`, run `GOCACHE=/private/tmp/diploma-go-cache go test ./internal/zap -run '^$' -fuzz '^FuzzDecode$' -fuzztime=3s -parallel=2` for the report decoder (23,459 executions) and `GOCACHE=/private/tmp/diploma-go-cache go test ./internal/scope -run '^$' -fuzz '^FuzzParse$' -fuzztime=3s -parallel=2` for URL parsing (276,635 executions). The first attempts could not write into the default user cache and were rerun with the temporary cache.

An earlier integration run found that PostgreSQL requires column-level UPDATE permission for the target row lock. Migration 3 grants only UPDATE on the target display name. The runtime still cannot update or delete evidence. An earlier Chromium run failed because the import confirmation and findings loading messages shared the status role. The test now selects the import message and passed the rerun.

## Review follow-up

Review pass 1 withheld approval on checkpoints 1–3. Checkpoints 5 and 6 record the implementation responses to R001–R009 and await reviewer recheck.

On 2026-11-23, `make check`, `cd frontend && npm run build`, `make test-integration`, `make test-e2e` and `docker compose config --quiet` exited 0 after the review fixes. The PostgreSQL suite verified migration 4, all ten malformed report cases without partial writes, a successful import with unusual query values, ordered observation history across two reports with the clock fixed, import-scoped severity, and unchanged provenance after a rejected duplicate. Chromium verified upload/reload and persistence after API and database restart. A three-second URL fuzz run after the query parsing fix passed with 290,454 executions.

## Checkout continuation

Baseline `21a096f899ed43b340f26fbf6cceabf288f46530` on `main` matched GitHub's branch head through read-only inspection. The two existing untracked frontend and backend guides were left intact.

The redesigned frontend keeps the same evidence workflow and protected download. Chromium rendered unlock and findings at 1280×900 and findings and evidence at 390×844. The mobile page stayed within the viewport, with horizontal scrolling confined to the findings table. Both browser cases passed using `E2E_SCREENSHOTS=/private/tmp/diploma-ui-21a096f make test-e2e`. The suite exercised keyboard unlock and rejection of an invalid token, import and idempotency, immutable provenance conflict, filtering and return navigation, safe scanner text, audited download, outage behavior and persistence across restart. The three Vitest tests and the frontend build passed. Screenshots show the synthetic fixture only.

`make test-integration` reproduced six escaped-NUL HTTP 500s before the fix. The last run passed under the race detector, with precise 422 pointers and no writes for the invalid fields, preserved raw bytes, and passing concurrency and authorization checks. The verifier independently confirmed this at checkpoint 07. The verifier reproduced V010 and V011. The implementation suite reproduced them again and then passed with both fixes on disposable PostgreSQL. Large numbers in unlocated alert extensions now round-trip through a versioned JSON-text envelope. Provenance over the database-rendered 64 KiB limit returns 413. The final `make check` passed with identical generated files and three Vitest tests.

Independent verification of checkpoint 09 passed with matching file hashes and `GOFLAGS=-mod=readonly make test-integration`. V010 and V011 are resolved. The verifier also found V012, where an invalid cursor could leave the findings page without an in-app recovery action. Checkpoint 10 restores the First page control after errors. The final Chromium test reproduced the invalid-cursor response and verified that the control returns to the original persisted rows. Source, formatting, build and browser checks passed on the final combined implementation. The verifier independently ran `GOFLAGS=-mod=readonly make test-e2e` against checkpoint 10 at matching hashes. Both browser phases passed and V012 is resolved. No scoped review findings remain open.

The scoped changes are in the coordination checkpoints. Review conclusions are recorded separately in `docs/coordination/review.md`. Genuine lab acceptance and Compose startup remain separate gates.

## Synthetic fixture and artifact hashes

The URL validation follow-up passed `make check`, `make test-integration` and `make test-e2e`. The regression test first reproduced acceptance of percent-decoded NUL and invalid UTF-8 query names. The fixed parser returned `422 validation_failed` at the instance URI pointer for both cases. The integration test verified zero import, finding and observation rows after each rejected request. The same bounded URL fuzz command above passed again with 288,782 executions. 

The browser and PostgreSQL tests used authored data and observed **5 observations**, **4 findings**, **1 unlocated alert**, and an unknown severity warning. The fixture's run involved no scanner. Scan authorization, times, versions and sanitized hashes for genuine data are unavailable. See [fixture provenance](fixture-provenance.md).

| Artifact | SHA-256 |
| --- | --- |
| `backend/testdata/zap/synthetic-traditional.json` | `c91f5451043fcac49e8246e04dedb90eb362ebe983d064600a6914330eaa304a` |
| `contracts/openapi.yaml` | `5bfed9a5e7383cdd80e0805762a39d141fcd1c81b6277dc17a8886ec57e47c4a` |
| `backend/internal/store/sqlc/imports.sql.go` | `ff1c65eee2854d201d7fe4a2ad62a72fec9c256aca454f60b191f40234e28714` |
| `frontend/src/api/schema.d.ts` | `2d9aa5630be8a3c631cbea8e286a8189a33e92893bd426d1de5f5d69f9c900e3` |

For the open M1 gate, a genuine existing report from an authorized local Juice Shop lab must be paired with its provenance and sanitized derivative. Compare fixture-derived counts and fingerprints with rows persisted by the API, then inspect those rows in the browser after restart. The root contract requires explicit named lab authorization and verified network isolation before any scan. The current scope is offline import.
