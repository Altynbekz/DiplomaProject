# Implementation coordination

Role: main implementation agent. Scope: M0 and M1.

Current milestone: M1. The host foundation and synthetic offline workflow run locally. The genuine authorized report gate and Compose startup remain unverified.

Current checkout baseline: `21a096f899ed43b340f26fbf6cceabf288f46530`. See M1-20261123-07 onward for current slices and `review.md` for independent results. Earlier numbered checkpoints below preserve the initial history.

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
