# Implementation evidence

This is engineering evidence for the paper, using the [artifact IDs](artifact-register.md) and [claim map](claim-evidence-map.md). M1 is complete for the genuine fixture. M2 software is independently verified at the current baseline. The live A01 role/object gate is open. [Milestone contract](../../AGENTS.md#11-testing-and-concrete-acceptance-gates) · [Implementation status](../status.md)

| Slice | Implementation and concrete tests | Independent evidence and scope |
| --- | --- | --- |
| M0 local stack, migrations and permissions | [Compose](../../compose.yaml), [local startup](../../scripts/dev-local.sh), [Go API entry](../../backend/cmd/api/main.go), [initial migration](../../backend/db/migrations/00001_imports.sql), [runtime grants](../../backend/db/migrations/00002_runtime_privileges.sql) and [Compose smoke script](../../scripts/test-compose.py). Existing tests cover fresh and staged upgrade, authentication and readiness with PostgreSQL unavailable. | [Checkpoint 14](../coordination/review.md#independent-verification--m0-20261123-14) verified disposable ARM64 Compose behavior. [Checkpoint 17](../coordination/review.md#independent-verification--m0-20261123-17) verified test automation, including failure propagation. These are platform-scoped checks. |
| M1 authorized genuine capture and provenance | ART-01, ART-03, ART-06, ART-07 and ART-15 identify the historical passive session. ART-02 is the reviewed derivative. [Sanitizer and oracle](../../scripts/zap-fixture.py) · [Offline preflight](../../scripts/check-zap-fixture.py) | [Checkpoint 20](../coordination/review.md#independent-verification--m1-20261123-20) verified authorization, recorded isolation, stopped lab, source bytes and derivative. A genuine scanner report contains scanner assertions. |
| M1 import, immutable provenance and evidence | [Parser](../../backend/internal/zap/parser.go), [import service](../../backend/internal/service/service.go), [provenance migration](../../backend/db/migrations/00004_import_source_metadata.sql), [API integration](../../backend/internal/httpapi/api_test.go), [genuine integration](../../backend/internal/httpapi/genuine_test.go) and [genuine browser tests](../../frontend/tests/browser/genuine.spec.ts). ART-04 supplied the independently derived 17 groups and 21 pointers. Concurrent imports returned one creation and three duplicate responses sharing that import ID. Provenance conflict, exact submitted bytes, audit-gated downloads and persistence after restart were checked. | [Checkpoints 21–22](../coordination/review.md#independent-verification--m1-20261123-21) verified genuine API/DB results, Chromium persistence and disclosure. FIG-01–FIG-02 illustrate stored data. |
| M2 human verification | [Migration 5](../../backend/db/migrations/00005_verification.sql), [service](../../backend/internal/service/reviews.go), [integration](../../backend/internal/httpapi/review_integration_test.go), [panel](../../frontend/src/features/verification-panel.tsx). Tests exercise all four states, exact evidence references, cross-finding rejection, append-only history, audit, concurrent and stale revisions and browser draft recovery. | [Checkpoints 24–25](../coordination/review.md#independent-verification--m2-20261123-24) covered history and precise evidence selection. FIG-03 is illustrative. |
| M2 OWASP classification | [Migration 6](../../backend/db/migrations/00006_classification.sql), [ART-10 rules](../../backend/internal/classify/owasp-mapping.v1.yaml), [service](../../backend/internal/service/classifications.go), [integration](../../backend/internal/httpapi/classification_integration_test.go). Rules address selected 2025 categories. A deliberately conflicting input tests ambiguity. The genuine fixture initially maps five A01 and one A02 findings, with one out of scope and ten unmapped. Tests cover manual history, stale writes, concurrency and runtime mutation restrictions. | [Checkpoints 26–27](../coordination/review.md#independent-verification--m2-20261123-26) verified source-backed classification and exact policy attribution. [Checkpoint 31](../coordination/review.md#independent-verification--m2-20261124-31) verified one 201 plus one 409, keyboard recovery and persistence. FIG-04 shows one manual demo revision. |
| M2 application context | [Migration 7](../../backend/db/migrations/00007_context.sql), [context SQL](../../backend/db/queries/contexts.sql), [service](../../backend/internal/service/contexts.go), [integration](../../backend/internal/httpapi/context_integration_test.go), [editor and resolver view](../../frontend/src/features/context-panels.tsx). Tests cover four factors, provenance and foreign references, input limits, precedence, history, audit, immutability and conflicts. | [Checkpoints 28–29](../coordination/review.md#independent-verification--m2-20261123-28) checked populated upgrade, permissions, precedence and fixes. [Checkpoint 31](../coordination/review.md#independent-verification--m2-20261124-31) verified scope-safe draft recovery and restart. FIG-05–FIG-06 show assumed lab exposure. |
| M2 live A01 acceptance | ART-05 and ART-08 document a blocked attempt. [Scenario and evidence request](../../lab/scenarios/a01-role-object.md) | [Checkpoint 31](../coordination/review.md#independent-verification--m2-20261124-31) found no live decision record. CORS observations, hypothetical expected decisions and in-memory handler tests remain separate. |

## Hosted CI

Public GitHub run metadata was checked for each exact revision. Dates are run starts in UTC, distinct from the capture and review dates. The workflow uses disposable PostgreSQL and Chromium and starts no scanner. [Workflow](../../.github/workflows/ci.yml)

| Slice | Exact tested SHA | Run and observed conclusion | Started (UTC) |
| --- | --- | --- | --- |
| Foundation and test automation | `94f431f479b7b707c84ee9d79ebf5f6710aa9ced` | [45170314104](https://github.com/Altynbekz/DiplomaProject/actions/runs/45170314104), success | 2026-11-23T12:00:55Z |
| Reusable genuine fixture and acceptance | `f5712a49b5676badb05486b30c73e0d4b916146b` | [45187645541](https://github.com/Altynbekz/DiplomaProject/actions/runs/45187645541), success | 2026-11-23T13:48:53Z |
| Human verification | `948351c45008048c601a827594d3ddf84d1a1cd8` | [45193568143](https://github.com/Altynbekz/DiplomaProject/actions/runs/45193568143), success | 2026-11-23T14:22:48Z |
| Reviewed classification | `20c6a76536c3b2d921ef3a27a36b96b119c0af13` | [45195527459](https://github.com/Altynbekz/DiplomaProject/actions/runs/45195527459), success | 2026-11-23T14:33:45Z |
| Immutable context resolution | `da9844eabb14c6c14646b407c652864afc0012e9` | [45197471901](https://github.com/Altynbekz/DiplomaProject/actions/runs/45197471901), success | 2026-11-23T14:44:42Z |
| Current M2 software baseline | `f1613e6fd5d64625dc239d2673cb019ecb964efd` | [45365896739](https://github.com/Altynbekz/DiplomaProject/actions/runs/45365896739), success | 2026-11-24T09:30:47Z |

## Reproducing software checks

With the pinned tools and installed frontend and browser dependencies, run the existing verification commands on the recorded baseline. These commands exercise the application using disposable resources. Current paper artifacts use prior test evidence and the existing demo rows.

```sh
GOFLAGS=-mod=readonly make check
npm run build --prefix frontend
env -u TEST_DATABASE_URL -u TEST_GENUINE_DATABASE_URL GOFLAGS=-mod=readonly make test-integration
GOFLAGS=-mod=readonly make test-e2e
```

Commands, passes and platform limits are recorded in the status and review logs. Tool versions come from ART-14. The M3 foundation evidence is recorded below. M4–M6 and comparative results remain planned. [Local verification](../status.md#draft-recovery-follow-up) · [Protocol proposal](../research-protocol.md)


## M3 foundation

| Slice | Implementation and evidence |
| --- | --- |
| Offline CVSS-B and CVSS-BE | [Adapter](../../backend/internal/cvss/cvss.go) and ART-16. Unmodified BSD-2-Clause FIRST reference at `c5b0d409ae9f57c44264c6ce5f27d89298e1d32a`; pinned Goja executes only that source with validated values. Separate Node fixture generation and byte-for-byte regeneration. Thirteen curated parity cases and independent 4,096-vector comparison passed. [ADR](../decisions/0002-cvss-v4-calculator.md) · [Independent review](../coordination/review.md#independent-verification--m3-20261124-34) |
| Immutable evidence and context snapshots | [Migration 8](../../backend/db/migrations/00008_assessments.sql), [queries](../../backend/db/queries/assessments.sql), [service](../../backend/internal/service/assessments.go) and ART-17. M2 rows matched before and after a populated upgrade. Tests exercise current verified/mapped evidence, exact parent inheritance, context/factor requirements, stale conflict, immutability and atomic audit rollback. |
| API and read-only UI | [Routes](../../backend/internal/httpapi/assessments.go), [contract](../../contracts/openapi.yaml), [application history](../../frontend/src/features/assessment-panel.tsx) and [synthetic Chromium flow](../../frontend/tests/browser/assessment-flow.ts). Drafts preserve null scores, and completed B/BE test snapshots show calculator provenance and historical status after restart. Genuine imports remain unassessed until separately reviewed. |

These results describe the assessment foundation for RQ-01. The analyst-facing metric and rationale editor is a later M3 slice. Independent verification and exact hosted run status are recorded separately. [Local checks](../status.md#m3-assessment-foundation) · [Review log](../coordination/review.md)
