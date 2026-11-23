# Diploma vulnerability evidence workspace

A local research prototype using React and TypeScript, Go `net/http`, sqlc and PostgreSQL. The app imports an existing [ZAP Traditional JSON report](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/), records its original bytes and SHA-256, and displays normalized findings and their source observations. Scanner severity, confidence, OWASP classification, verification and CVSS are separate. M1 findings start unmapped and unreviewed, with both assessment fields null.

The offline workflow has been exercised against PostgreSQL using an explicitly synthetic fixture. M1 acceptance still requires a genuine report from an authorized local Juice Shop lab and a sanitized derivative with recorded provenance. See [status and verification](docs/status.md).

## Start locally

The host path was verified on macOS arm64. Install Go 1.27.1, Node 26.13.0, npm 12.2.1, PostgreSQL 15.19 with its server binaries on PATH, goose 3.30.1, Python 3, Make and ripgrep. Use sqlc 1.31.1 for generation. Frontend dependencies are locked in `frontend/package-lock.json`. Exact direct dependencies and tool versions are in `tools.versions.json`.

```sh
make dev-local
```

Open [the local workspace](http://127.0.0.1:5173) and enter the token from `.secrets/api_token`. The browser keeps it in memory until reload or logout. The API binds to loopback port 8080 and the private host database uses loopback port 55440 by default. Ctrl-C stops the API and frontend. To stop the database while retaining its files, run `pg_ctl -D artifacts/local-db/data -m fast stop`.

The Compose path uses Docker Compose with a running daemon and targets Linux arm64 images pinned by SHA-256. Its services include PostgreSQL, one-shot migrations, the API and Vite. Start with `make dev` and stop with `docker compose down` to retain the volume. Compose configuration was validated locally. Container startup requires separately verifying the stack on a running daemon. The database has no published port on this path.

Setup generates ignored secret files. `.env` contains the local operator ID and server-owned `JUICE_SHOP_ORIGIN`. Set the latter to the exact origin recorded in the report **before creating the target** and restart the API if it was running. The default is `http://juice-shop:3000`. Use a DNS-style service hostname without underscores. Legacy Compose container names with underscores are not supported as registered origins. For endpoint identity, the importer decodes only query names. Malformed escapes in a name reject the report with the source URI pointer. Values stay private and are ignored for identity, including raw semicolons and percent characters. This setting registers an offline identity. Imports never resolve hosts or contact report URLs. Keep `.env`, `.secrets`, database files and raw reports private. Host setup, Compose setup and tests start no scanner.

## Import and inspect a report

1. Unlock the workspace and create a Juice Shop target. The server supplies the registered origin.
2. Open the target and select your existing genuine ZAP Traditional JSON report. Supply available metadata in the provenance JSON field. See [fixture provenance](docs/fixture-provenance.md) for the fields and the distinction between original and sanitized hashes.
3. Import the report. The UI shows its hash, parser version, finding and observation counts, unlocated-alert count, warnings and provenance. The report-declared ZAP version and parser-detected format appear separately from operator provenance. An identical upload reuses the existing import ID when provenance is omitted or matches the stored record. Different provenance receives a 409 conflict. Import provenance is immutable. For imports created before migration 4, the separately stored report version is unavailable and legacy provenance may include report-derived defaults. See [the provenance decision](docs/decisions/0001-immutable-import-provenance.md) for upgrade behavior.
4. Filter findings by report ID, category, severity or verification and page through the results. New imports keep observations in report source order by assigning strictly increasing receipt timestamps. The recorded scan time is separate. Older observations keep their original timestamps. Open a finding to read source pointers, scanner confidence, descriptions and remediation. Private payloads stay behind explicit, audited download. CVSS displays “Not assessed.”

The same import can be exercised with curl after setting `TARGET_ID` to the ID of the registered target and setting `REPORT` and `PROVENANCE` to private local files. This example passes the header through curl config on standard input so the token stays out of curl’s argument list.

```sh
printf 'header = "Authorization: Bearer %s"\n' "$(cat .secrets/api_token)" |
  curl --fail-with-body --config - \
    --form "report=@${REPORT};type=application/json" \
    --form "provenance=<${PROVENANCE}" \
    "http://127.0.0.1:8080/api/v1/targets/${TARGET_ID}/imports/zap"
```

The automated synthetic browser walkthrough observed five observations grouped into four findings, one unlocated alert and an unknown severity warning. Reload and API/database restart checks used persisted rows. These are test results for authored data. Genuine scan measurements are unavailable.

## Verify and maintain

```sh
cd frontend && npm ci && npx --no-install playwright install chromium && cd ..
make check
make test-integration
make test-e2e
```

Database and browser suites create and remove their own disposable clusters. They require loopback sockets and PostgreSQL shared memory. `make migrate-up` uses an explicit `MIGRATION_DATABASE_URL` for the migration role. The API connects as a separate role with append-only permissions for evidence and audit records. For local evidence retention, back up the database privately before any operator-managed reset. A stack stop retains evidence.

[OpenAPI](contracts/openapi.yaml) lists implemented routes for health, readiness, targets, ZAP imports, findings and observations. Review edits, context revisions, CVSS scoring, rankings, scanner orchestration and research evaluation are later milestones. The [OWASP Top 10:2025](https://top10.owasp.org/) study categories are available as filter values. Until rules have been reviewed, findings remain unmapped.
# DiplomaProject
