# Implementation coordination

Role: main implementation agent. Scope: M0 and M1.

Current milestone: M1. M0 host foundation and the M1 offline workflow are verified with synthetic data. The genuine fixture acceptance gate and Docker runtime verification remain blocked. Files currently being edited: none.

Latest review checkpoint: 6. Review findings and their resolutions will be recorded by ID here. Review pass 1 posted R001–R009 and withheld approval for checkpoints 1–3. Checkpoints 4–6 are pending reviewer recheck.

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
