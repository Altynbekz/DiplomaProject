# Diploma vulnerability evidence workspace

A local research prototype using React and TypeScript, Go `net/http`, sqlc and PostgreSQL. The app imports an existing [ZAP Traditional JSON report](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/), records its original bytes and SHA-256, and displays normalized findings and their source observations. Scanner severity, confidence, OWASP classification, verification and CVSS are separate. Imported findings start unreviewed, with both assessment fields null.

The automated offline suites exercise both an authored test fixture and a reviewed derivative of a genuine report from an authorized local Juice Shop session. The M2 software checks passed. M2 acceptance remains open pending a separately authorized role/object test record with observed access decisions. [Acceptance checklist](docs/status.md#m2-acceptance-audit)

## Start locally

The host path was verified on macOS arm64. Install Go 1.27.1, Node 26.13.0, npm 12.2.1, PostgreSQL 15.19 with its server binaries on PATH, goose 3.30.1, Python 3.11 or newer, Make and ripgrep. Use sqlc 1.31.1 for generation. Frontend dependencies are locked in `frontend/package-lock.json`. Exact direct dependencies and tool versions are in `tools.versions.json`.

```sh
make dev-local
```

Open [the local workspace](http://127.0.0.1:5173) and enter the token from `.secrets/api_token`. The browser keeps it in memory until reload or logout. The API binds to loopback port 8080 and the private host database uses loopback port 55440 by default. Ctrl-C stops the API and frontend. To stop the database while retaining its files, run `pg_ctl -D artifacts/local-db/data -m fast stop`.

The Compose stack targets a native Linux ARM64 Docker engine, including Docker Desktop on Apple Silicon. Its images are pinned by SHA-256. AMD64 and emulated engines are outside the supported Compose configuration. Start with `make dev` and stop with `docker compose down` to retain the volume. The API and frontend use a project bridge with loopback-only host ports. PostgreSQL uses a separate internal network with no published port. For observed runtime results, see [implementation status](docs/status.md).

Setup generates ignored secret files. `.env` contains the local operator ID and server-owned `JUICE_SHOP_ORIGIN`. Set the latter to the exact origin recorded in the report **before creating the target** and restart the API if it was running. The default is `http://juice-shop:3000`. Use a DNS-style service hostname without underscores. Legacy Compose container names with underscores are not supported as registered origins. For endpoint identity, the importer decodes only query names. Malformed escapes in a name reject the report with the source URI pointer. Values stay private and are ignored for identity, including raw semicolons and percent characters. This setting registers an offline identity. Imports never resolve hosts or contact report URLs. Keep `.env`, `.secrets`, database files and raw reports private. Host setup, Compose setup and tests start no scanner.

## Import and inspect a report

The workspace uses a light interface with findings and evidence kept on separate pages. Optional provenance and advanced filters open inline. On mobile, scroll the findings table sideways to reach the assessment columns. Return to findings from an observation to restore the filter and page.

The importer rejects decoded NUL characters in database-bound fields with a field pointer. Private data kept only in the original report is preserved byte for byte. Unlocated alerts keep their full decoded JSON as reversible text in a versioned envelope, so large numeric literals in extensions survive storage. Oversized stored provenance receives a size error and leaves the existing data intact.

1. Unlock the workspace and create a Juice Shop target. The server supplies the registered origin.
2. Open the target and select [the genuine Juice Shop fixture](backend/testdata/zap/juice-shop-traditional.json). Paste the contents of [its provenance file](backend/testdata/zap/juice-shop-provenance.json) into **Optional report provenance**. See [fixture provenance](docs/fixture-provenance.md) for the transformation and report hashes.
3. Import the report. The UI shows its hash, parser version, finding and observation counts, unlocated-alert count, warnings and provenance. The report-declared ZAP version and parser-detected format appear separately from operator provenance. An identical upload reuses the existing import ID when provenance is omitted or matches the stored record. Different provenance receives a 409 conflict. Import provenance is immutable. For imports created before migration 4, the separately stored report version is unavailable and legacy provenance may include report-derived defaults. See [the provenance decision](docs/decisions/0001-immutable-import-provenance.md) for upgrade behavior.
4. Filter findings by report ID, category, severity or verification and page through the results. New imports keep observations in report source order by assigning strictly increasing receipt timestamps. The recorded scan time is separate. Older observations keep their original timestamps. Open a finding to read source pointers, scanner confidence, descriptions and remediation. Private payloads stay behind explicit, audited download. CVSS displays “Not assessed.”
5. Open **Human verification** on a finding. Select its source observations and record what you reviewed and what remains uncertain. For the passive report demo, use **Inconclusive** when the scanner evidence is insufficient to establish the result. For a stale draft, load the latest verification, review the preserved draft, and save when ready. The history shows each saved review with its source pointers, operator and time.
6. Open **OWASP classification** to see the current mapping, matched rule sources, and history. A manual revision records its categories, rationale and selected evidence independently of the human review.
7. On the target page, open **Application context** and save a target default or an exact method and path. Give each factor a rationale and attach a source for each known value. Open **Resolved context** on a finding to see the selected revision and unknown values. Open context history to inspect or revise an earlier scope. [Access-control evidence protocol](lab/scenarios/a01-role-object.md)

To import the same fixture from the repository root, set `TARGET_ID` to the ID of the registered target. This example passes the bearer header through standard input.

```sh
REPORT=backend/testdata/zap/juice-shop-traditional.json
PROVENANCE=backend/testdata/zap/juice-shop-provenance.json
printf 'header = "Authorization: Bearer %s"\n' "$(cat .secrets/api_token)" |
  curl --fail-with-body --config - \
    --form "report=@${REPORT};type=application/json" \
    --form "provenance=<${PROVENANCE}" \
    "http://127.0.0.1:8080/api/v1/targets/${TARGET_ID}/imports/zap"
```

The genuine fixture contains 21 observations, grouped into 17 findings with zero unlocated alerts. Both fixtures have separate Chromium reload and API/database restart checks.

## Verify and maintain

```sh
cd frontend && npm ci && npx --no-install playwright install chromium && cd ..
make check
make test-integration
make test-e2e
```

`make test-scripts` runs the fixture preflight unit tests with Python's standard-library `unittest`. Both `make test` and `make check` include these synthetic tests. CI pins Python 3.14.3.

The root Go workspace points to the module in `backend/`. From the repository root, run `go test ./backend/...` for Go unit tests or `go build ./backend/...` to compile the Go packages. Inside `backend/`, use `go test ./...`.

Database and browser suites create and remove their own disposable clusters. They require loopback sockets and PostgreSQL shared memory. `make migrate-up` uses an explicit `MIGRATION_DATABASE_URL` for the migration role. The API connects as a separate role with append-only permissions for evidence and audit records. For local evidence retention, back up the database privately before any operator-managed reset. A stack stop retains evidence.

To exercise Compose on an ARM64 engine with Docker Compose 2.24.4 or newer, run `python3 scripts/test-compose.py`. It verifies startup, migrations, readiness, authentication, synthetic import, duplicate handling, and persistence across shutdown and restart. It creates a random project, disposable credentials and volume, and ephemeral loopback ports. Cleanup is scoped to that project. Use `--config-only` to validate the effective isolation settings without starting containers. The smoke test uses Compose’s port replacement syntax. [Compose merge reference](https://docs.docker.com/reference/compose-file/merge/#replace-value)

The CI workflow runs formatting, lint, type, generation, unit and frontend build checks on Ubuntu 24.04. It builds the checksum-pinned PostgreSQL 15.19 release for separate integration and Chromium runs under the non-root runner account. Hosted run `45170314104` passed for commit `94f431f`, including the Python preflight tests. [Workflow](.github/workflows/ci.yml) · [Hosted run](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104) · [Implementation status](docs/status.md)

[OpenAPI](contracts/openapi.yaml) lists implemented routes for imports and evidence, human verification, OWASP classification, context revisions and resolution. The [OWASP Top 10:2025](https://top10.owasp.org/) study categories are available as filter values. CVSS scoring, ranking, scanning and research evaluation are planned.
