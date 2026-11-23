# Implementation coordination

Role: main implementation agent. Scope: M0 and M1.

Current milestone: M0/M1. Hosted CI verified the Python test wiring on `94f431f`. Genuine-report acceptance awaits an authorized collection and its provenance. Lab preparation is separate from the frozen documentation checkpoint.

Current checkout baseline: `94f431f479b7b707c84ee9d79ebf5f6710aa9ced`. No application files are under edit. See checkpoint 19 for the frozen documentation and CI evidence, and `review.md` for independent results. Earlier checkpoints preserve the implementation history.

## Checkpoint 1 — M0 foundation — READY FOR REVIEW

Saved files: `backend/go.mod`, `backend/go.sum`, `backend/cmd/api/main.go`, `backend/internal/config/config.go`, `backend/db/migrations/00001_imports.sql`, `00002_runtime_privileges.sql`, `00003_target_lock_privilege.sql`, `backend/db/sqlc.yaml`, `backend/internal/store/sqlc/`, `frontend/package.json`, `frontend/package-lock.json`, `frontend/tsconfig.json`, `frontend/vite.config.ts`, `contracts/openapi.yaml`, `frontend/src/api/schema.d.ts`, `tools.versions.json`, `Makefile`, `.env.example`, `.gitignore`, `compose.yaml`, and `scripts/test-integration.sh`.

Checks: `make check` exited 0, including byte-identical generation, Go vet, React lint and both UI unit tests. The disposable PostgreSQL script exited 0 under the race detector with migrations from empty to revision 1, from revision 1 to 3, and a repeat upgrade. The prior two-phase Chromium run exited 0 and observed health 200, readiness 503, and unauthenticated access 401 during DB outage. The latest Chromium run hit an ambiguous status selector in the test, so the browser test is being revised and will receive a fresh checkpoint. Compose config validation exited 0. No running Docker daemon was available for container startup verification.

Current work: M1 final verification and operator documentation. Files being edited: `frontend/tests/browser/import.spec.ts`, `README.md`, `docs/status.md`, `docs/fixture-provenance.md`, and local/Compose setup scripts.

## Checkpoint 2 — M0 host startup and M1 offline import — READY FOR REVIEW

This checkpoint supersedes checkpoint 1 for the files listed below. Reviewer approval has not been received.

- Foundation and startup: `Makefile`, `compose.yaml`, all three Dockerfiles, `.dockerignore` files, `scripts/setup-local.py`, `scripts/dev-local.sh`, `scripts/runtime-login.sql`, `scripts/db-init.sh`, `backend/docker-entrypoint.sh`, `backend/cmd/api/main.go`, `backend/internal/config/` and `tools.versions.json`.
- Import and persistence: `backend/internal/domain/`, `backend/internal/scope/`, `backend/internal/zap/`, `backend/internal/service/`, `backend/internal/store/`, `backend/internal/httpapi/`, `backend/db/`, `backend/testdata/zap/`, `configs/` and `contracts/openapi.yaml`.
- Browser and verification: `frontend/src/`, `frontend/tests/`, `frontend/playwright.config.ts`, `frontend/vite.config.ts`, `scripts/test-integration.sh`, `scripts/test-e2e.sh` and `scripts/generate-check.py`.
- Operator documentation: `README.md`, `docs/status.md` and `docs/fixture-provenance.md`.

Checks after saving the implementation: `make check`, `npm run build` in frontend, `make test-integration` and `make test-e2e` all exited 0. The PostgreSQL suite ran with the race detector. Chromium passed the import/reload and post-restart suites, plus the outage check. `make dev-local` reached a healthy, ready, empty stack and returned a protected target list through the frontend proxy. The stack was stopped after the smoke checks. `docker compose config --quiet` exited 0. Container startup is unverified because the daemon socket is unavailable.

Current milestone: M1. Offline implementation and synthetic verification are ready for review. Acceptance still requires a genuine authorized Juice Shop report plus provenance and sanitized derivative. Files currently being edited: none. No reviewer findings were present when this checkpoint was published.

## Checkpoint 3 — M1 evidence access and pagination checks — READY FOR REVIEW

Saved files: `frontend/tests/browser/import.spec.ts`, `backend/internal/httpapi/api_test.go` and `docs/status.md`. The application code in checkpoint 2 is unchanged. Use checkpoint 3 for these test files and updated verification notes.

`make check`, `make test-integration` and `make test-e2e` exited 0. The PostgreSQL test paged through both observations of a grouped finding without repeating a pointer. Chromium verified that the private download control starts disabled and that the downloaded file matches the uploaded bytes after explicit acknowledgement. Both bounded Go fuzz runs passed. The counts and commands are in `docs/status.md`.

Review findings: none posted when this checkpoint was saved. Review status: pending.

## Checkpoint 4 — M1 decoded query-name validation — READY FOR REVIEW

Saved files: `backend/internal/scope/url.go`, `backend/internal/scope/url_test.go`, `backend/internal/httpapi/api_test.go` and `docs/status.md`. This checkpoint supersedes the earlier checkpoints for these files. The change came from implementation verification. No reviewer finding ID has been posted.

The regression test reproduced acceptance of decoded NUL and invalid UTF-8 in a query name. Both inputs now receive a documented 422 with the instance URI pointer. PostgreSQL checks confirm no partial rows after either invalid import. A valid Unicode query name still normalizes independently of query values.

`make check` passed with byte-identical generated files and two Vitest tests. `make test-integration` passed under the race detector, including all six rejection cases and the prior import and persistence checks. Both Chromium tests and the database outage checks passed with `make test-e2e`. The three-second URL parser fuzz run passed with 288,782 executions.

Reviewed-file SHA-256 values:

- `backend/internal/scope/url.go` — `cb5dd01dddeefe95d9425b01496e8eda0695b7050633d6d32fce01bffb46fc5b`
- `backend/internal/scope/url_test.go` — `017ab16245ee1b9d6178b1b74fbb1ee08547fa05b6467a0622ca3771cffcc0db`
- `backend/internal/httpapi/api_test.go` — `d47af2c12929a8b544fbaa493b2b7e2dce42b776c193370afc7ecd2d64d026a7`

Review status: pending.

## Checkpoint 5 — M1 parser review fixes — READY FOR REVIEW

Saved files: `backend/internal/scope/url.go`, `backend/internal/scope/url_test.go`, `backend/internal/zap/parser.go`, `backend/internal/zap/parser_test.go`, `backend/internal/httpapi/api_test.go`, `README.md`, and `configs/lab-targets.example.yaml`.

- R001 addressed. Decode only query names, splitting pairs on ampersands. Malformed name escapes still get a pointer and 422. Raw semicolons and percent characters in values remain private and do not affect identity.
- R005 addressed. The 64 KiB scalar limit covers numeric strings and numeric JSON tokens.
- R007 addressed through documentation. The offline registry continues to require a DNS-style service name without underscores.
- R008 addressed. Regression cases cover site-level origin mismatch, missing instances, empty fingerprint fields and rejection without partial rows when a field or nesting depth exceeds the limit.
- R009 addressed. Fingerprints normalize nil, empty, unsorted and repeated query names without changing the caller's slice.

`make check` and `make test-integration` passed. PostgreSQL verified nine invalid reports against the documented status/code and checked for zero partial rows. The URL parser fuzz run passed with 290,454 executions. Ready for reviewer recheck. The reviewer has only rendered a verdict on checkpoints 1–3.

## Checkpoint 6 — M1 persistence, provenance, and display review fixes — READY FOR REVIEW

Saved files: `backend/db/migrations/00004_import_source_metadata.sql`, `backend/db/queries/imports.sql`, generated `backend/internal/store/sqlc/imports.sql.go` and `models.go`, `backend/internal/service/service.go`, `backend/internal/service/types.go`, `backend/internal/httpapi/api.go`, `backend/internal/httpapi/api_test.go`, `backend/internal/zap/parser.go`, `contracts/openapi.yaml`, `frontend/src/api/schema.d.ts`, `frontend/src/features/evidence.tsx`, `frontend/tests/workspace.test.tsx`, `scripts/test-integration.sh`, `README.md`, `docs/status.md`, and `docs/decisions/0001-immutable-import-provenance.md`. This checkpoint supersedes earlier snapshots for the listed files. Checkpoint 5 covers the other parser/test files.

- R002 addressed. Under the target lock, new observations receive strictly increasing receipt timestamps in source order. The database retains the contract's `(created_at,id)` order and existing observation rows. A fixed-clock test verifies distinct severities, cross-report order and the import-filtered summary.
- R003 addressed. New records keep operator ZAP version and template null when omitted. Report-declared version and parser-detected format are separate fields. Legacy records remain immutable and are labeled in the UI.
- R004 addressed. Omitted or matching provenance returns the existing import. Differing supplied provenance receives 409, and the existing record is unchanged.
- R006 addressed. The UI reads both assessment fields and displays their supplied values. A synthetic UI test covers a non-null vector and a zero score. The M1 API still returns unreviewed, unmapped and null assessments because review and assessment persistence are outside this milestone.
- R001 and R008 follow-up. PostgreSQL accepts a synthetic report with private semicolon and percent query values, using the same finding identity across observations. The oversized site-count request is also verified to create zero partial rows.

`make check` passed, including byte-identical generation and three Vitest tests. `cd frontend && npm run build` passed. `make test-integration` passed under the race detector, including the revision 3 to 4 upgrade and a repeat upgrade with nothing pending. `make test-e2e` passed both Chromium tests and the database outage checks. `docker compose config --quiet` passed. Container startup and the genuine fixture gate remain unverified.

Reviewer verdict is still limited to the earlier checkpoints in review pass 1. Acceptance and reviewer approval have not been claimed.

## Continuation at 21a096f

Baseline commit: `21a096f899ed43b340f26fbf6cceabf288f46530` on `main`. Read-only remote inspection returned the same commit. Existing untracked `frontend/FRONTEND.md` and `backend/BACKEND.md` are preserved.

In progress: reproduce escaped-NUL import failures through PostgreSQL before adding validation. Backend slice files under edit: `backend/internal/httpapi/api_test.go`, `backend/internal/domain/`, `backend/internal/zap/parser.go`, and `backend/internal/service/service.go`. No verification checkpoint is active for this slice yet.

## M1-20261123-07 — Escaped NUL validation — READY FOR VERIFICATION

Baseline commit: `21a096f899ed43b340f26fbf6cceabf288f46530`. This checkpoint covers only the backend files below.

Behavior: PostgreSQL-bound report fields and operator provenance reject decoded NUL with a safe field pointer and HTTP 422. Raw-only extensions remain accepted and successful uploads still retain the identical report bytes. No database schema, fingerprint, or API shape change.

Checks: `make test-integration` reproduced HTTP 500 in all six escaped-NUL cases before the fix. After the fix it exited 0 under the race detector. All six cases returned 422 and created zero partial rows. The existing import, concurrent duplicate, byte preservation, target scope and privilege assertions passed. Backend `GOCACHE=/private/tmp/diploma-go-cache gofmt -l backend/internal` produced no paths. From `backend/`, `GOCACHE=/private/tmp/diploma-go-cache go vet ./...` and `GOCACHE=/private/tmp/diploma-go-cache go test ./...` exited 0. The bounded decoder fuzz run `GOCACHE=/private/tmp/diploma-go-cache go test ./internal/zap -run '^$' -fuzz '^FuzzDecode$' -fuzztime=3s -parallel=2` passed with 13,879 executions.

Remaining gates: genuine authorized lab report with provenance and sanitized derivative, host Compose startup if its daemon becomes available, and independent review. Frontend is a separate slice under active edit.

```text
c96db8e7650dd55b7cc70d6438973be37ef67256baa1316139318b833a52206a  backend/internal/domain/storage_text.go
ed401ef0d2a22b0166a360ffb4ebd1928e6d338d3817931b6e3bbf9af15c3dde  backend/internal/httpapi/api_test.go
110e0823bdce6ecec41678e1b79448e69e61bea50cf3ecfdc92907674bb9534c  backend/internal/zap/parser.go
a131e4bdedf260efe97ca1a1e300d2cb27712e1b831b37fd78cb6db5d612b82c  backend/internal/zap/parser_test.go
af51178bb22e6e8fef1ebf292f158bc27ca032a5642f581c749e76a615f00579  backend/internal/service/service.go
```

Backend checkpoint M1-20261123-07 was independently verified at its recorded hashes and is now superseded for further edits. Backend files are being reopened to address V010 and V011 from the verifier. A new backend checkpoint will replace it. The frontend slice remains separate.

## M1-20261123-08 — Light evidence workspace — READY FOR VERIFICATION

Baseline commit: `21a096f899ed43b340f26fbf6cceabf288f46530`. This checkpoint covers only the five frontend files below.

Behavior: compact light unlock, targets, import, findings and evidence pages. Unlock checks the protected API before showing the workspace. The one-target registry hides duplicate creation once a target exists. Optional provenance and less common filters use inline disclosures. Finding links restore the originating filter and cursor. The mobile findings table scrolls inside a keyboard-focusable region. Scanner severity, verification, OWASP category and CVSS remain separate. Original report download still requires acknowledgement and uses the audited endpoint.

Checks: from `frontend/`, `npm run fmt`, `npm run lint`, `npm test` (three tests) and `npm run build` exited 0. `E2E_SCREENSHOTS=/private/tmp/diploma-ui-21a096f make test-e2e` exited 0. Chromium verified keyboard navigation and token rejection, upload, duplicate reuse, conflict with immutable provenance, filtering and return navigation, escaped evidence, acknowledged download, outage responses and persistence after restart. The viewport checks verified no page overflow and local table scrolling at 390×844. Rendered unlock and findings at 1280×900, plus findings and evidence at 390×844, were visually inspected. Screenshots contain synthetic data only.

Remaining gates: independent frontend verification and a genuine authorized lab report with provenance and a sanitized derivative. Backend findings V010 and V011 are handled separately.

```text
3f7bf882d9c8da64879d491d04f2b5869ded91ea7c2d425a6a83cc832ac3185a  frontend/src/app/main.tsx
c68b7eb039fb6a2d2830aa5e8361b1a7509d1b90f1d6ae93be75707bff79f1b6  frontend/src/app/style.css
dacca59b51b3d7e2d7d31b06a519174fa09e3471dbc7d11051211d0ad5b0139b  frontend/src/features/evidence.tsx
bbf9786d71f148974dcdc7afeb794e230c169e136e41d555f0da0a2ca4ca0d94  frontend/tests/workspace.test.tsx
30604946c08cb2d9cb4fdcb44ed18c42c4a2d94d2029bdfab4a8780824f4c5bb  frontend/tests/browser/import.spec.ts
```

## M1-20261123-09 — JSONB storage boundaries — READY FOR VERIFICATION

Baseline commit: `21a096f899ed43b340f26fbf6cceabf288f46530`. This checkpoint supersedes M1-20261123-07 for backend files.

- V010 addressed. The exact reproducer returned HTTP 500 before the fix. New unlocated alerts retain their decoded JSON as text inside a versioned JSONB envelope. The PostgreSQL regression proves the large exponent survives and the original report bytes are unchanged. Existing stored alerts and API response shapes stay unchanged.
- V011 addressed. The exact compact 65,536-byte normalized provenance reproducer returned HTTP 500. The existing database constraint still applies. Its specific PostgreSQL violation now returns HTTP 413 at `/provenance`, with no partial rows.
- The escaped-NUL rejection from checkpoint 07 remains covered in the same suite.

Checks: `make test-integration` exited 0 after both fixes under the race detector, covering migrations 1 through 4, repeat upgrade, six NUL cases, both boundary cases, exact bytes, observation order, import filters and concurrent duplicates. `make check` exited 0, including Go formatting, vet, unit tests, TypeScript, ESLint, generation comparison and three Vitest tests. `docker compose config --quiet` exited 0 and `docker info` failed because the Docker daemon socket is absent.

Remaining gates: genuine authorized lab report with provenance and sanitized derivative, Compose runtime, and independent recheck of this checkpoint. Checkpoint 08 separately covers the frontend.

```text
c96db8e7650dd55b7cc70d6438973be37ef67256baa1316139318b833a52206a  backend/internal/domain/storage_text.go
56ba98eed0258907c37a603ad65c4c50dcdb473215edfeac98e4f8071f7b80ee  backend/internal/httpapi/api_test.go
7236663f6b6169e36a411aa51f5a0e77046ce82833842b72da81491465a67637  backend/internal/zap/parser.go
a131e4bdedf260efe97ca1a1e300d2cb27712e1b831b37fd78cb6db5d612b82c  backend/internal/zap/parser_test.go
b18f44a377e7ef376b4e499b19d7b2ba9fedddb3243b4829b23812c3d88954c8  backend/internal/service/service.go
```

Frontend checkpoint M1-20261123-08 is superseded for further edits after the verifier finished inspecting its frozen files. V012 reproduced an invalid cursor with no recovery action. Reopening `frontend/src/features/evidence.tsx` and `frontend/tests/browser/import.spec.ts` for the recovery slice. The verified backend checkpoint 09 remains frozen.

## M1-20261123-10 — Recover from stale pagination — READY FOR VERIFICATION

Baseline commit: `21a096f899ed43b340f26fbf6cceabf288f46530`. This checkpoint supersedes frontend checkpoint 08. Backend checkpoint 09 was independently verified at its recorded hashes.

V012 addressed. A first-page action remains visible while a paged targets, findings or observations request is loading or has failed. No pager is displayed when there is no return path or next page. The browser regression opens a target with an invalid cursor as its only query parameter, receives the real API 400, uses First page and sees all four original synthetic rows.

Checks: from `frontend/`, `npm run fmt`, `npm run lint` and `npm run build` exited 0. `make check` passed Go formatting, vet, unit tests, TypeScript, lint, byte-identical generation and three Vitest tests. `E2E_SCREENSHOTS=/private/tmp/diploma-ui-21a096f make test-e2e` passed both Chromium tests on the final combined stack, including invalid cursor recovery, desktop/mobile viewport checks, outage behavior and persistence across restart. `git diff --check` exited 0.

Remaining gates: independent recheck of V012, genuine authorized Juice Shop report with provenance and sanitized derivative, and Compose startup on a host with an available daemon. No application files are under edit while this checkpoint is verified.

```text
3f7bf882d9c8da64879d491d04f2b5869ded91ea7c2d425a6a83cc832ac3185a  frontend/src/app/main.tsx
c68b7eb039fb6a2d2830aa5e8361b1a7509d1b90f1d6ae93be75707bff79f1b6  frontend/src/app/style.css
5d07dd0f74f35c65f3a69f308c5cb6ec2e73f797f4372d65f2988f4ed01db855  frontend/src/features/evidence.tsx
bbf9786d71f148974dcdc7afeb794e230c169e136e41d555f0da0a2ca4ca0d94  frontend/tests/workspace.test.tsx
10514aa7bace3324b4e95711409ed6c4d42e7d7f6fb375597957ac5a3f19a17c  frontend/tests/browser/import.spec.ts
```

Independent verification completed for checkpoints 09 and 10. The verifier matched both sets of file hashes and recorded passing disposable PostgreSQL and Chromium checks in `review.md`. V010, V011 and V012 are resolved. No application files are under edit. Genuine-report acceptance and Compose startup still require the external report and a running Docker daemon, respectively.

## M0-20261123-11 — Repository guidance — READY FOR VERIFICATION

Baseline commit: `a1b523d8ad25b170466a09e4d758e8afacc55b39` on `main`. GitHub main matched through read-only inspection.

Behavior: the root contract links to both layer guides and the status/verification logs. Guide links resolve from their own directories, and both guides describe guidance for the current checkout. Implemented and unverified states remain in docs/status.md.

Checks: the Python relative-link check for the three Markdown files passed, and git diff --check exited 0. No application code changed.

Remaining gates: GitHub Actions setup and hosted run, disposable Compose startup, and genuine report acceptance. The three files below are frozen for review.

```text
6f6def7d685dfb81dc408f615abe13c2acaeb0f335cebebd5b80cd97b0a1d193  AGENTS.md
27cf216aa6557c8d98b6f647b0b7b6110bf2c9d5ca595fcb519805b4d9f3d60e  frontend/FRONTEND.md
af84ab8c056d4c12162658d0549d4b9e929ba002716e7cbb69569b9299b8f3a6  backend/BACKEND.md
```

## M0-20261123-12 — GitHub Actions — READY FOR VERIFICATION

Baseline commit: `a1b523d8ad25b170466a09e4d758e8afacc55b39`. The workflow and PostgreSQL bootstrap are frozen below.

Behavior: GitHub Actions is configured for main pushes, pull requests and manual runs on Ubuntu 24.04. Actions use full release hashes, checkout does not persist credentials, and permissions are read-only. Go follows go.mod and Node, npm, sqlc, goose, and the npm lockfile use repository pins. PostgreSQL 15.19 is built from a checksum-pinned release and both PostgreSQL and Chromium suites use separate disposable clusters under the non-root runner. No scan jobs or real credentials are part of the workflow.

Local checks: `GOCACHE=/private/tmp/diploma-go-cache go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 .github/workflows/ci.yml`, `sh -n scripts/install-test-postgres.sh`, `GOFLAGS=-mod=readonly make check`, `npm run build --prefix frontend`, `GOFLAGS=-mod=readonly make test-integration`, and `GOFLAGS=-mod=readonly make test-e2e` exited 0. Go and frontend checks included three Vitest tests and byte-identical generation. Disposable PostgreSQL applied revisions 1 through 4 and passed the import suite with the race detector. Both Chromium phases passed, including outage and restart. Local pinned Homebrew PostgreSQL 15.19 powered these runs. The additional archive-build exercise is underway. Go 1.27.1, Node 26.13.0, npm 12.2.1, sqlc 1.31.1, goose 3.30.1 and lockfile Playwright 1.65.1 were confirmed locally.

Remaining gates: hosted Ubuntu GitHub Actions run, completion of the local fresh source bootstrap check, runtime Compose validation, and the genuine authorized report. No workflow was pushed or dispatched.

```text
7e256e9a1150ce101c25622fafada3026d12b9040d78a138a9a3ac473666355f  .github/workflows/ci.yml
0d547dff5283fef8db5982b86795f77b70e1349933dc8ba1883068e318c00656  scripts/install-test-postgres.sh
```

Checkpoint 12 is superseded for CI compatibility changes. The verifier reproduced V013, an undeclared dependency on ripgrep in make tools-check. The workflow and source bootstrap are unchanged. Makefile is being edited for the replacement checkpoint.

## M0-20261123-13 — CI runner compatibility — READY FOR VERIFICATION

Baseline commit: `a1b523d8ad25b170466a09e4d758e8afacc55b39`. This checkpoint supersedes checkpoint 12. Only disjoint Compose and acceptance-documentation work continues while the three files below are frozen.

V013 addressed. The goose version check now compares the complete pinned version string with a shell builtin. `make tools-check` passed with a restricted PATH containing only system locations and links to the pinned sqlc and goose binaries. The same environment had reproduced the missing-ripgrep exit 127 under checkpoint 12.

Checks: `GOFLAGS=-mod=readonly make check` exited 0 after the fix, covering format, vet, strict TypeScript, ESLint, byte-identical regeneration, Go tests and three Vitest tests. `git diff --check` exited 0. The unchanged workflow and PostgreSQL bootstrap passed the actionlint and shell syntax checks recorded in checkpoint 12. The local build, isolated PostgreSQL integration and isolated Chromium runs there also passed. The fresh PostgreSQL source build has started after the archive matched its pinned checksum.

Remaining gates: independent V013 recheck, hosted Ubuntu GitHub Actions run, local fresh-source integration, disposable Compose runtime, and a genuine authorized report. No workflow has been pushed or dispatched.

```text
7e256e9a1150ce101c25622fafada3026d12b9040d78a138a9a3ac473666355f  .github/workflows/ci.yml
0d547dff5283fef8db5982b86795f77b70e1349933dc8ba1883068e318c00656  scripts/install-test-postgres.sh
4a0e60a3a0c9b2da1d00e50a18a1e3734825f76a57bed4200651c65aff265031  Makefile
```

Checkpoint 13 passed independent verification. The verifier matched all three hashes and repeated the restricted-PATH check without ripgrep. V013 is resolved. The PostgreSQL 15.19 archive matched its pinned checksum, built and installed in a temporary prefix, and both `GOFLAGS=-mod=readonly make test-integration` and `GOFLAGS=-mod=readonly make test-e2e` exited 0 on that build. Chromium passed both phases and the outage check. Hosted CI still awaits an actual GitHub run.

## M0-20261123-14 — Disposable Compose startup and persistence — READY FOR VERIFICATION

Baseline commit: `a1b523d8ad25b170466a09e4d758e8afacc55b39`. The five files below are frozen for verification.

Behavior: the frontend installs dependencies and runs Vite as `node`. The project app bridge publishes only loopback ports for the API and frontend. Migrations and PostgreSQL use the internal storage network and PostgreSQL publishes no port. The new smoke test uses a generated project name, disposable credentials, ephemeral ports and a separate volume. Following the persistence check, cleanup is limited to that project and its exact image tags.

Observed regressions: with networking disabled, a direct cache-write probe against the original frontend image returned EACCES. The same probe passed on the corrected image. With the corrected image and the previous Compose networks, both containers were healthy, but `docker inspect` reported empty effective port bindings and the smoke test exited 1 with `invalid IP:0`. The project app bridge fixed both loopback bindings.

Checks: `python3 scripts/test-compose.py --config-only`, `docker compose config --quiet`, and `git diff --check` exited 0. `python3 -u scripts/test-compose.py` exited 0 in project `diploma-smoke-667b49d98d434dbab875` on Apple Silicon with Docker Engine 29.7.2 and Compose 5.5.1. It verified clean setup through Goose revision 4, health/readiness 200, missing bearer 401, frontend HTML and TypeScript entry, one synthetic report with four findings, five observations and one unlocated alert, exact digest, HTTP 200 duplicate reuse, and the same digest and four findings after shutdown and restart. A separate Docker query confirmed no containers, networks, volume or tagged images remained for this test project.

Remaining gates: independent Compose recheck, the hosted CI run, and an existing genuine report with authorization and provenance. The repository, Git history and local report filename search yielded only the synthetic fixture.

```text
b670b08ba300cce25e74242806b0c59b145243d95a0333f4846cd4e633cec909  compose.yaml
3fe9859cf70e74f5b27ddf964fecbcba3748796fcfa60191a620fc7ed78c4ad2  frontend/Dockerfile
4becdbecff0a5b130072e8390f1d4cafeb61afa0dc3b56501e93b20d734845dd  scripts/test-compose.py
6c0cc248216f162ee5ec6d4b43d5576574ba87365dd6066a3a962e051750575c  README.md
b8177a4955a65d7e212e1460821688851dc97ca238de5b138bf42beba7acb034  docs/status.md
```

Checkpoint 14 passed independent verification at all five frozen hashes. The verifier repeated startup, revision 4 migration, protected frontend/API checks, synthetic import, duplicate reuse, restart persistence and scoped cleanup in their own project. No findings were reported. See the [independent verification entry](review.md#independent-verification--m0-20261123-14).

Final local checks: `GOFLAGS=-mod=readonly make check` and `npm run build --prefix frontend` exited 0, including three Vitest tests and byte-identical generated files. No scoped review findings remain open. Hosted CI has no observed run. Genuine-report acceptance needs an existing authorized Juice Shop ZAP Traditional JSON report and its provenance record.

## M1-20261123-15 — Hosted CI result and genuine fixture preparation — READY FOR VERIFICATION

Baseline commit: `5a1cd7427fbfca196d3dc59bea65555a256046e6`. Read-only GitHub comparison found `main` identical to this baseline, and the local checkout started clean.

Observed CI: GitHub's run list showed one run for the commit. Run `45166744371`, attempt 1, initially reported `in_progress` and then `completed` with conclusion `success`. The `check` job (ID `137665526785`) and all its listed steps succeeded, including PostgreSQL integration and Chromium. No CI changes were needed. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45166744371)

Behavior: the checklist and read-only CLI check private original and sanitized derivative digests against supplied provenance, require authorization reference, scan time and a sanitization record, and compare site/alert/instance counts by position. Output is limited to aggregate counts and fixed diagnostics. The CLI rejects the repository's labeled synthetic report. Authorization, sanitization review, independent fingerprints, integration and browser persistence are separate checklist items for the genuine input.

Checks: on Python 3.14.3, `python3 -B -m unittest discover -s scripts -p 'test_check_zap_fixture.py' -v` exited 0 with eight tests. `python3 scripts/check-zap-fixture.py --help` and `git diff --check` exited 0. A local link check resolved all 16 relative links in the three changed Markdown files. Existing app and Compose sources are unchanged. The search checked only the repository, its history and documented artifact locations and found only the synthetic fixture.

Remaining gates: independent verification of these files and the genuine report gate. The missing input is the local path to an existing genuine ZAP Traditional JSON report from an authorized Juice Shop lab, plus its provenance record with the authorization reference, scan time and available tool/target/plan metadata.

```text
b018f4e806965bcea6c1872d8ccd47a887b72f3f1af23cf9c74affd1ed9298be  README.md
d66cf4195dedfadecbb4126d95f9ae1f86aa2108e0535283850e5596658e4712  docs/status.md
68c260cf6df4ea43e94d6cdd749d63e1d90fa7d22c7b0a699bcaaef387a01333  docs/fixture-provenance.md
0bedc1dada3be3ed3939d46446e285e0e770cb8781d39e7aad5c28cde8300107  scripts/check-zap-fixture.py
dc48c5edeb93fac571e84bcbbf030a45bc72d136b5933629751a394b1bdd7a0d  scripts/test_check_zap_fixture.py
```

Checkpoint 15 is superseded for the preflight CLI, its tests and the check record in `docs/status.md`. The verifier confirmed the hosted result and reproduced V014, which exposed an unexpected path in an argument error. The README and acceptance checklist hashes are unchanged.

## M1-20261123-16 — Private preflight diagnostics — READY FOR VERIFICATION

Baseline commit: `5a1cd7427fbfca196d3dc59bea65555a256046e6`. This checkpoint supersedes 15 for preflight verification. The verifier already confirmed the hosted result in the checkpoint 15 review. The README and checklist hashes match that review.

V014 addressed. Passing an extra fictional path originally exited 2 and printed the path verbatim. The same invocation now exits 2, leaves stdout empty and prints a fixed diagnostic with a `--help` hint. The regression test verifies that the path stays private. Success counts and preflight validation remain covered by the eight prior tests.

Checks: `python3 -B -m unittest discover -s scripts -p 'test_check_zap_fixture.py' -v` exited 0 with nine tests. `python3 scripts/check-zap-fixture.py --help`, the standalone privacy probe and `git diff --check` exited 0. The hosted CI success was verified separately for run `45166744371` on the baseline commit.

Remaining gates: independent V014 recheck and a genuine authorized Juice Shop report plus provenance. A candidate-specific sanitization review, expected fingerprints and real-data integration and browser persistence depend on that report.

```text
b018f4e806965bcea6c1872d8ccd47a887b72f3f1af23cf9c74affd1ed9298be  README.md
8ec839ac3649d11a36bc5e64e6e668b5769c8b5a31cd8c978a98d1b1f489871c  docs/status.md
68c260cf6df4ea43e94d6cdd749d63e1d90fa7d22c7b0a699bcaaef387a01333  docs/fixture-provenance.md
4bd930682ddf0e5b33a57d85c1aa38a92ec7e66e4dca2d73bc08308016dd7ad8  scripts/check-zap-fixture.py
6fee1d4fd7a50d7b516c9005d301092a04af2db95eda4f736c82bab0d51a5af2  scripts/test_check_zap_fixture.py
```

Checkpoint 16 passed independent verification at all five frozen hashes. The verifier repeated the original V014 command, checked unknown-option and missing-argument diagnostics, and ran all nine tests. V014 is resolved and no new findings were reported. The only remaining acceptance gate requires the genuine authorized report and provenance. [Independent verification](review.md#independent-verification--m1-20261123-16)

## M0-20261123-17 — Python tests in the default verification path — READY FOR VERIFICATION

Baseline commit: `4d0240c4f50559b0e21feeeead7348c03cd74a7f`. GitHub `main` matched this baseline and no local changes were present at the start. The five files below are frozen.

Behavior: `make test-scripts` runs the existing standard-library `unittest` suite. Both `make test` and `make check` require it. The existing hosted `make check` step picks up the tests after setting up CPython 3.14.3 with a full action commit pin. The Go, frontend, disposable PostgreSQL and Chromium commands remain in the workflow. Host verification requires Python 3.11 or newer.

Hosted evidence: run `45168962574`, attempt 1, went from `in_progress` to `completed` with conclusion `success` for the exact baseline. Job `137673174362` and all 15 listed steps succeeded. GitHub listed one run for that commit, last updated `2026-11-23T11:54:53Z`. The run tested the committed baseline and predates this Python wiring. [GitHub Actions run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45168962574)

Local checks on Python 3.14.3: `make test-scripts`, `GOFLAGS=-mod=readonly make test` and `GOFLAGS=-mod=readonly make check` exited 0. Nine Python tests, Go tests, three Vitest tests, lint, formatting, strict TypeScript and byte-identical generation passed. In a temporary Git archive with one deliberately failing Python assertion, `make test` and `make check` each exited 2 with the expected failure. The working-tree test was untouched.

`npm run build --prefix frontend`, `GOFLAGS=-mod=readonly make test-integration`, and `GOFLAGS=-mod=readonly make test-e2e` exited 0. Disposable migrations reached revision 4, PostgreSQL integration passed under the race detector, and both Chromium phases and outage checks passed. `actionlint` 1.7.7, `git diff --check`, the tool manifest JSON check and a 12-link Markdown check exited 0. The upstream tag resolves to the pinned action hash and the official tool versions manifest lists Python 3.14.3 for Ubuntu 24.04 x64.

Remaining gates: independent review of this checkpoint, a hosted run of the Python wiring after publication, and genuine-report acceptance. The repository, history and documented artifact locations contained only the labeled synthetic fixture. The missing input is the local path to the genuine authorized Juice Shop ZAP Traditional JSON report and its provenance record.

```text
407b914f10f6d1eb3edf63f66d347860d1db2ba6d1baf590f695cdea3c8f2697  .github/workflows/ci.yml
c129bd59fd47488da92d7158f95e61d69b1e8eab074e0a8b84a16422a04bc7a6  Makefile
072f65f621e5ce9a38f796d978c744f50edc8d4fcd1502ab78213cbfc4551809  README.md
80f8843031065e4ed4713fc6dc76fc31bbf01bdea46c9a81820fca2ed6089b10  docs/status.md
d8ac3e8710c4b12586cc73fd35edc5351ca113aeaee4debe72f6f10f93411fb2  tools.versions.json
```

Checkpoint 17 passed independent verification at all five frozen hashes. The verifier ran all three Make targets in a separate archive, confirmed the deliberate failure exits from `make test` and `make check`, and checked the Python and action pins. The observed hosted success applies to the baseline only. No findings remain open. [Independent verification](review.md#independent-verification--m0-20261123-17)

## M1-20261123-18 — Hosted CI evidence and passive collection plan — READY FOR VERIFICATION

Baseline commit: `94f431f479b7b707c84ee9d79ebf5f6710aa9ced`. The local checkout was clean and GitHub `main` matched the baseline. The three files below are frozen for review.

Hosted evidence: run `45170314104`, attempt 1, went from `in_progress` to `completed` with conclusion `success` on the exact baseline. GitHub reported one run for that commit, last updated `2026-11-23T12:07:24Z`. Job `137677821087` and all 17 listed steps passed. Its log includes `python3 -B -m unittest -v test_check_zap_fixture` in `make check`. PostgreSQL integration and Chromium passed in the same workflow. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104)

Changes: recorded the hosted result, updated the README, and documented a scoped passive collection plan for the disposable `http://juice-shop:3000` lab. The plan covers prior authorization, verified isolation and request limits, ordinary browser traffic in ZAP Safe mode, a private Traditional JSON export and provenance, and shutdown. It does not treat the plan as executed or use the preflight as proof of a genuine report. No application, test, or workflow files changed.

Local checks: `git diff --check` exited 0. A Python relative-link and anchor check resolved all 18 links in `README.md`, `docs/status.md` and `docs/fixture-provenance.md`. `git check-ignore artifacts/zap/example/original.json` exited 0 and confirmed the proposed private report path is ignored. The repository and documented artifact location search found only the labeled synthetic fixture, and the operator confirmed a new report is needed.

Remaining gates: independent review of this documentation checkpoint, a hosted run of the eventual pushed commit, completion of lab isolation verification, explicit authorization for the collection, and genuine-report acceptance after collection. No scan was launched.

```text
75e1110ca71464a000e7b4c5e498502acd0dc70b02c5ab1195aeb3fa2714be2c  README.md
1429e22eaaa725541aadae5d862d3d31f0368d9855eaae0a5ab3ec08b79a8107  docs/status.md
c6d85d1f08e1ec3cab27f8a2ce5352710c6fdde57195a0dbe3cd9258c071cc00  docs/fixture-provenance.md
```

Checkpoint 18 is superseded for the collection plan and CI check detail. The verifier confirmed the exact hosted run and all nine Python tests. V015 requests version and digest pins for every selected lab image. The plan is being amended before retesting.

## M1-20261123-19 — Pinned lab image provenance — READY FOR VERIFICATION

Baseline commit: `94f431f479b7b707c84ee9d79ebf5f6710aa9ced`. This checkpoint supersedes 18 for the three frozen documentation files.

V015 addressed. Before isolation tests, the plan requires lab container versions and SHA-256 digests plus exact ZAP and add-on versions. The resolved platform image identities stay with private run evidence. The plan spells out the browser-to-ZAP-to-allowlisting-proxy boundary. `docs/status.md` records nine passing Python tests from the separately reviewed hosted job log.

Checks: `git diff --check` exited 0, all 18 local links and anchors resolved, and `git check-ignore` confirmed that the private lab work and proposed original report path are excluded. No workflow or application code was edited. The independent verifier confirmed run `45170314104` on the exact baseline, all 17 successful steps, and the nine Python tests. [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104)

Remaining gates: independent recheck of V015, hosted CI of the commit after push, completed isolation verification on the local lab, a named authorization, and the genuine report and acceptance evidence. Private lab work continues outside the frozen files. No scan was launched.

```text
75e1110ca71464a000e7b4c5e498502acd0dc70b02c5ab1195aeb3fa2714be2c  README.md
3cd52ac68edf27dec0f69a1e6472398d2f5ec602f67ecb1fd907e2316630a0b3  docs/status.md
533a982b58ac1b07922335cbce5fb2a29c2738e9a34618c035baf5611d74232e  docs/fixture-provenance.md
```

Checkpoint 19 passed independent verification at all three frozen hashes. V015 is resolved. The verifier found no new issues in the CI record, scoped passive plan, private path exclusion or documentation links. Actual lab isolation and collection are separate gates. [Independent verification](review.md#independent-verification--m1-20261123-19)


## M1-20261123-20 — Genuine passive collection and derivative — READY FOR VERIFICATION

Baseline commit `0c193d51c989866e8d49932454803853cb16df11`. The three documentation files and the ignored local artifacts listed below are frozen. The private original and authorization stay outside Git. Review report contents locally and record aggregate results.

The authorized passive session for `http://juice-shop:3000` ran from `2026-11-23T13:25:43.520670Z` through `2026-11-23T13:26:59.738124Z`. Actual Chromium and ZAP passed the allowed canary, with eleven forbidden-request checks and zero requests on denied sinks. ZAP reported passive queue and export completion. Docker reported zero running lab containers. The Traditional JSON contains one site and eight alerts with 21 observations. Offline identity grouping yields 17 expected findings. The session did not record a total forwarded request count.

Checks executed:

- In the private lab module, `go test ./...` exited 0, covering the guard request limit, rate and duration.
- `python3 -B artifacts/zap/diploma-m1-passive-cef3083eaa/collect-once.py` exited 0 after two preflight-only attempts stopped before target traffic. The final isolation recheck passed before the single capture began.
- `cmp` of the private original and a new copy of the stopped ZAP export exited 0. Deployed guard and pinned helper SHA-256 hashes matched.
- `python3 -B artifacts/zap/diploma-m1-passive-cef3083eaa/sanitize-report.py --check` exited 0 and reproduced both the derivative and provenance byte for byte.
- `python3 -B scripts/check-zap-fixture.py "$ORIGINAL_REPORT" "$SANITIZED_REPORT" "$PROVENANCE"` exited 0, with one site, eight alerts, 21 observations and zero unlocated alerts. A separate offline review checked registered origin, unchanged grouping inputs and source positions, and disclosure risks in the derivative. `git diff --check` exited 0 and `git check-ignore` matched all three local report/provenance artifacts.

Remaining gates are independent review of collection evidence and the derivative, genuine API and database import and duplicate checks, and browser persistence after restart. No genuine fixture has been added to Git.

```text
fa919f70d1c9da55e54cb8731278400ae3d4f044eaa647f4d6bb89cb5f0492c6  README.md
bcbc17523c09518e71625d1143fbe0d11dbc4fe6bb4da37e576707f0e9582a4e  docs/status.md
bdffc77f7cf7c49913d29d3d13ecfae29062ded8294be05c9a2307e5cd434733  docs/fixture-provenance.md
121e82b36df0b343f43015b9faed851e3bc8fc78c6dc91e3caed867ea8406bad  artifacts/zap/diploma-m1-passive-cef3083eaa/original.json
036a4eb798d55fa68fa56a9389933b52b2bc3fefe86fd35eff741bb15efebdf7  artifacts/zap/diploma-m1-passive-cef3083eaa/sanitized-traditional.json
fa18726990acdb2857247bb3a46a2495fb2df59b470faa41a18d8dde81fd0cbe  artifacts/zap/diploma-m1-passive-cef3083eaa/provenance.json
882c37ed199b743bb385a3eca97ad15a586b20bf0044d024ad02229c4a4329fd  artifacts/zap/diploma-m1-passive-cef3083eaa/authorization.json
40f8e410d9976736b6ef02db543569f76f5c41af153b184505f30983d6283a22  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-isolation.json
e1d0e8f2cdbcf9e9622853d09f1cc6d97f22635d23f3492a6a9cf72fb5252833  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-start.json
a60df522dedc9ea17a766fa831ed2b35a12f45d44ae94b86b42106dbedb1d29d  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-end.json
b0c2058b40f4655a885ac1ba96799a2c1efc74b3ca739da4b7618474bcd2e0ee  artifacts/zap/diploma-m1-passive-cef3083eaa/passive-capture.yaml
b562ab54b9e50e000cd2ac2a7aabc33c8df37d3e19e9358248d5096f2167242b  artifacts/zap/diploma-m1-passive-cef3083eaa/sanitize-report.py
e581a400b7de2ab2648d39a84c10ef79ec2866075fdce8cd312dcf0017b77e27  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-zap.log
49fdfb4fa6aff32fba68f78ed5ff778c9333da4dfcf55d6be613d5411e1eef5e  artifacts/zap/diploma-m1-passive-cef3083eaa/collect-once.py
```


## M1-20261123-21 — Genuine offline application replay — READY FOR VERIFICATION

Baseline commit `0c193d51c989866e8d49932454803853cb16df11`. This checkpoint supersedes 20 for the three tracked documentation files, resolves V016 by removing the stale unavailable-provenance sentence, and retains the previously verified collection artifacts unchanged. The three disposable replay scripts are frozen in addition to the documentation and collection artifacts. No application code or workflow changed.

`env -u TEST_DATABASE_URL GOFLAGS=-mod=readonly sh artifacts/zap/diploma-m1-passive-cef3083eaa/verify-offline.sh` exited 0. Goose applied migrations 1–4. API and PostgreSQL matched 17 independently derived fingerprints and 21 observation pointers. Four first submissions returned one creation and three concurrent duplicates with one import ID. A sequential duplicate reused the ID, conflicting provenance returned 409, and the stored metadata and audited bytes matched the derivative submission. Chromium selected the derivative with its provenance, displayed the duplicate result, and displayed 17 persisted unreviewed findings and source evidence before and after API and PostgreSQL restart. The browser saw no external page requests. The outage check returned health 200, readiness 503 and missing-bearer 401. `git diff --check` exited 0 and a local check resolved all 18 relative links and anchors in the three documentation files.

Remaining gates are independent replay of the offline acceptance and adding the reviewed genuine fixture with repeatable repository coverage if approved for inclusion. Hosted CI covers the baseline without these local artifacts. The original and authorization remain out of Git.

```text
301e6a3c57b2b886b9ba4f219fa507115b2c60d85855b6105ee0d6c38e0b8e25  README.md
329ced9a03dd0c7c20aede6b17a14be39e2daf8095f8c335fa59fce3b92a2467  docs/status.md
830785925282a00af1cee8801a1aa3c631a410020902059819209094ad2ddf3d  docs/fixture-provenance.md
121e82b36df0b343f43015b9faed851e3bc8fc78c6dc91e3caed867ea8406bad  artifacts/zap/diploma-m1-passive-cef3083eaa/original.json
036a4eb798d55fa68fa56a9389933b52b2bc3fefe86fd35eff741bb15efebdf7  artifacts/zap/diploma-m1-passive-cef3083eaa/sanitized-traditional.json
fa18726990acdb2857247bb3a46a2495fb2df59b470faa41a18d8dde81fd0cbe  artifacts/zap/diploma-m1-passive-cef3083eaa/provenance.json
882c37ed199b743bb385a3eca97ad15a586b20bf0044d024ad02229c4a4329fd  artifacts/zap/diploma-m1-passive-cef3083eaa/authorization.json
40f8e410d9976736b6ef02db543569f76f5c41af153b184505f30983d6283a22  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-isolation.json
e1d0e8f2cdbcf9e9622853d09f1cc6d97f22635d23f3492a6a9cf72fb5252833  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-start.json
a60df522dedc9ea17a766fa831ed2b35a12f45d44ae94b86b42106dbedb1d29d  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-end.json
b0c2058b40f4655a885ac1ba96799a2c1efc74b3ca739da4b7618474bcd2e0ee  artifacts/zap/diploma-m1-passive-cef3083eaa/passive-capture.yaml
b562ab54b9e50e000cd2ac2a7aabc33c8df37d3e19e9358248d5096f2167242b  artifacts/zap/diploma-m1-passive-cef3083eaa/sanitize-report.py
e581a400b7de2ab2648d39a84c10ef79ec2866075fdce8cd312dcf0017b77e27  artifacts/zap/diploma-m1-passive-cef3083eaa/collection-zap.log
49fdfb4fa6aff32fba68f78ed5ff778c9333da4dfcf55d6be613d5411e1eef5e  artifacts/zap/diploma-m1-passive-cef3083eaa/collect-once.py
6c056db68f87bca6bc6d1a1f34b0dfff6e36e42933d80663c3879c20b84128b4  artifacts/zap/diploma-m1-passive-cef3083eaa/verify-offline.sh
4f3c08a988787e79679d0280c53d3547b9952f264a537f4cff0ec7e2d3e63907  artifacts/zap/diploma-m1-passive-cef3083eaa/verify-import.py
8c290f79d00e94d1866240a87034e561d12da2aab6603f33fe696cd218fb7e5b  artifacts/zap/diploma-m1-passive-cef3083eaa/verify-browser.cjs
```

Checkpoint 21 passed independent verification at all 17 frozen hashes. V016 is resolved. The verifier reran the disposable API, PostgreSQL and Chromium replay successfully. The reusable genuine repository fixture and repeatable coverage remain open. [Independent replay](review.md#independent-verification--m1-20261123-21)

After the checkpoint closed, `docs/status.md` was updated to cite the verifier's result. Its refreshed SHA-256 is `77ecf347b33d8a0e7ee17e9c3f9b74440d1c72e912dc7c1a72eb9851bd07abcc`. The application, replay scripts and lab artifacts retain their checkpoint hashes.


## M1-20261123-22 — Reusable genuine fixture — READY FOR VERIFICATION

Baseline 5e90808ae4bc1fde6e3241df77313229f7e7ff4f matches GitHub main. Hosted baseline [run 45185756939](https://github.com/Altynbekz/DiplomaProject/actions/runs/45185756939) succeeded. It predates the reusable fixture.

The public derivative is byte-identical to the reviewed private derivative. The provenance's reproduction command now points to the public offline sanitizer. Tests use disposable databases and no private originals, live targets or scanner. The independent Python oracle supplies 17 exact fingerprints and 21 ordered observation pointers for a fixed test target.

Observed on macOS ARM64: the sanitizer reproduced the same bytes with cmp exit 0. python3 -B scripts/check-zap-fixture.py exited 0 with one site, eight alerts, 21 observations, zero unlocated. GOFLAGS=-mod=readonly make check and npm run build --prefix frontend exited 0. GOFLAGS=-mod=readonly make test-integration exited 0 under the race detector for both fixtures. GOFLAGS=-mod=readonly make test-e2e exited 0 for four Chromium cases on separate synthetic and genuine stacks, including both restarts and readiness outage checks. git diff --check exited 0.

Remaining gates: independent review of this frozen slice and hosted CI for its published commit. Original and private authorization files remain ignored.

```text
6abc2912e49b7f2de72a02ae39a2f0675672657c30d423fa4d86a3b97dc2b864  README.md
5faf37d1dd35ba8d72cf03314778a27ca9092f5e88804a16c6a01b4b5aeffb53  docs/status.md
39a048d6204f02ec5c1dd647d2831bb9c2e53c4bc24dea9e93b79d15fbc8c80c  docs/fixture-provenance.md
b1c7977dce34b6da35f17871021a11b2a8f0bf28913d3680cfa2d593757266b3  backend/testdata/zap/README.md
036a4eb798d55fa68fa56a9389933b52b2bc3fefe86fd35eff741bb15efebdf7  backend/testdata/zap/juice-shop-traditional.json
aca865a2e0cf333e6cc7f21a7cd1b3beb5670b7c5fa488e830add9bc6d3c525e  backend/testdata/zap/juice-shop-provenance.json
fb48a29d966bba10aa5eb8cadd7007a5f7e8e8739b5225b6485abe461e6ebc22  backend/testdata/zap/juice-shop-expected.json
69eb0c6405be653a6c9623bacd54e22827bfafd20e87a6d2216c53e7730b5e3f  backend/internal/httpapi/genuine_test.go
38ef750f8197d1ba61a7d77f50d6899daed35e869bfacd104e82975e691345b4  frontend/tests/browser/genuine.spec.ts
10ee8b914d0692b60ae72163d2cb3141f7876c657d4f3bcbe77985e473a173d8  scripts/zap-fixture.py
edfc7decbd2ce76ef112642d04d84a75bd82137b8d2be165b64f902c3434a60b  scripts/test-integration.sh
0243f7b9a737a46cd3f4faa4121bc08c526e6ae12e348f0a0bfd2469291f22a9  scripts/test-e2e.sh
```

## M1-20261123-23 — Preflight documentation correction — READY FOR VERIFICATION

Baseline 5e90808ae4bc1fde6e3241df77313229f7e7ff4f. Checkpoint 22 passed independent archive format, lint, generation, build, race-mode integration and four Chromium cases. Only V017 remained. This documentation-only recheck records the two report digests checked by the CLI and points out the separate private plan digest check. All other frozen files in 22 retain their earlier hashes. git diff --check exited 0. Hosted verification still awaits publication.

```text
75a84aae59f15bd540ebeb3e99f07896d18952b10da02859ead1b038d326fb93  docs/status.md
091a19ae0337a09ba019a0d6ba6684de4e48b0c2e4a7b395408e6ac14c79d440  docs/fixture-provenance.md
```

Checkpoint 23 passed independent review. V017 is resolved. All ten unchanged files from checkpoint 22 matched their prior hashes. [Independent recheck](review.md#independent-verification--m1-20261123-23)

M1 acceptance passed on commit `f5712a49b5676badb05486b30c73e0d4b916146b`. The exact [GitHub Actions run 45187645541](https://github.com/Altynbekz/DiplomaProject/actions/runs/45187645541) completed successfully, including genuine and synthetic database and browser suites. Checkpoints 22 and 23 independently verified the fixture and documentation.

## M2 — Human verification — IN PROGRESS

Editing the migration, sqlc queries, service and HTTP contracts, finding detail panel, and disposable tests. Classification and context follow this slice.

## M2-20261123-24 — Human verification — READY FOR VERIFICATION

Baseline f5712a49b5676badb05486b30c73e0d4b916146b. Migration 5, the API, finding summary filter and keyboard-accessible verification panel deliver immutable review histories, exact observation references, server actor and timestamps, audit, and stale predecessor conflicts. Classification and context are separate slices.

GOFLAGS=-mod=readonly make check, make test-integration and make test-e2e, followed by npm run build --prefix frontend, exited 0 locally. Race-mode integration covered concurrent same-predecessor writes and runtime immutability. All four Chromium cases passed, including two-tab conflict recovery with the draft retained and evidence/review history after restart. Both fresh and v4-to-v5 migrations passed. git diff --check exited 0.

Remaining gates: independent review of the human verification slice and hosted CI for its published commit, then classification and context.

```text
2789a54478eb993fbd71c38e781b8231259ca3d8c7e0b9cc3cb9d40f6433e525  README.md
0ce5a9ab7fcd555594912aebbc3690233b7e4a17d55a8d43ba39b5e075b55adc  backend/db/migrations/00005_verification.sql
1131cdf237b68b91d4b532a757827223c13e25bb135abd6162bb96e2a298af8f  backend/db/queries/imports.sql
74019afb601a3ea9108b478aa121944f4a3ad706e834e32d999736b863b53d4e  backend/db/queries/reviews.sql
3a0f633dd33ca83d28d4107d7c66ed14cf53943d157123265ec6b7c9f671136f  backend/internal/httpapi/api.go
02d95eda385c232e2f3db96d2e98d3b7be2a3b69f01c0848e527eb18bd5c8aca  backend/internal/httpapi/genuine_test.go
373de8de0880b626f73fae3684e1efac7dcf3e39e51044b50d91227785672f9b  backend/internal/httpapi/review_integration_test.go
8a3d15d5f03eeac851916fcc0037778a1ea5cfcc82911cd203f35f29c67a6d37  backend/internal/httpapi/reviews.go
180c146392b3dccbb6cfc0696b592002f9b550742725cd0fd1892ab539d6c917  backend/internal/service/reviews.go
718b1224d60222a65615414e8ce7f0bd1b2d546c00d4e121fb5d440af6480c9e  backend/internal/service/service.go
8920b53db4cdaa94d259422265e275fcf938f4253ba8135196d491517b266845  backend/internal/store/sqlc/imports.sql.go
ee181f330b14534ef62eaf163fd4bebc76c7e91cbb3bdaaf29a8a586b8f371f3  backend/internal/store/sqlc/models.go
ea28d8e68ebe915903d923564375e5e9d6068b34faea83706bf2ed10d1042bbb  backend/internal/store/sqlc/reviews.sql.go
a208428d559fb4381c64ca64c266448888d43e4bf532ceca43d9f6c2abef5f85  contracts/openapi.yaml
586bae0262f8ccd6fdc6a9cdf14a515379d0860eac4e625dff903fb8c193df53  docs/status.md
9ea299eae89831494886922dd63b0f53b3bcdd0552fff0581d91be8a6f79f628  frontend/src/api/schema.d.ts
766e39dca1e0ccf1ce81ee6fb1466d838fa623b9c5c5ab295eaa5b63de9dec58  frontend/src/features/evidence.tsx
25a8dbdfbe3b5fee50eb7ecb72801133141cc8e25ce1cd1051f66d3ee433523b  frontend/src/features/verification-panel.tsx
15cbace647f48c3ac64356adc1da520e3f2e150cd72969f57f3b5355fc6b6b33  frontend/tests/browser/genuine.spec.ts
2fffde705ef0b6392d6ccccc899b51bf5906d9ff9f25a7679b9f665763fdc2cc  scripts/test-integration.sh
```

## M2-20261123-25 — Evidence reference labels — READY FOR VERIFICATION

Baseline f5712a49b5676badb05486b30c73e0d4b916146b. V018 addressed by displaying each observation's import identifier in the selection label and its observation identifier in the history and evidence card. A regression test selects one of two checkboxes with the same pointer and different imports. All other checkpoint 24 files retain their hashes. The unused classification rules package is independent work outside this checkpoint.

Frontend lint and four Vitest tests exited 0, including the two-import label regression. GOFLAGS=-mod=readonly ./scripts/test-e2e.sh genuine exited 0 with both Chromium cases, conflict recovery and restart. git diff --check exited 0. Hosted CI remains pending publication.

```text
a747946ec40d9264bf504b51bcac5049ddb24f5295785136cc133a8ca0d35d8c  frontend/src/features/verification-panel.tsx
9b913b0d3c028d7970d015316a630f84d23429c73c7b6a5bfc73f4540d3006e8  frontend/src/features/evidence.tsx
56da53598f42df9e1b903a4c6730355d99252f3b05bcabfe03d3dedd26bd43bb  frontend/tests/workspace.test.tsx
```

Checkpoint 25 passed independent review and resolves V018. The verifier used a second offline import at identical JSON positions, selected the intended source by keyboard, and traced the saved observation ID to its evidence card. [Independent recheck](review.md#independent-verification--m2-20261123-25)
