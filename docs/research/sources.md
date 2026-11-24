# Primary source register

Accessed 2026-11-24 (Asia/Almaty). Versions below identify the cited edition or documentation, where supplied. [Repository tool pins](../../tools.versions.json) and [capture provenance](../../backend/testdata/zap/juice-shop-provenance.json) identify the versions used by the project. A current documentation page supplies definitions or API background. Its access date does not date the implementation or collection.

## Standards and lab sources

| ID | Title and authoring organization | Publication or version | Topic and link |
| --- | --- | --- | --- |
| SRC-01 | OWASP Top 10, OWASP Foundation | 2025 edition | The category names and identifiers for [A01 Broken Access Control](https://top10.owasp.org/2025/A01_2025-Broken_Access_Control/), [A02 Security Misconfiguration](https://top10.owasp.org/2025/A02_2025-Security_Misconfiguration/) and [A05 Injection](https://top10.owasp.org/2025/A05_2025-Injection/). [Release](https://top10.owasp.org/2025/) |
| SRC-02 | Common Vulnerability Scoring System version 4.0 Specification Document, FIRST | CVSS 4.0, document v1.2 | Required metrics, Base and Environmental semantics, and CVSS-B/CVSS-BE nomenclature. [Specification](https://www.first.org/cvss/v4.0/specification-document) |
| SRC-03 | CVSS v4.0 User Guide and CVSS v4.0 calculator, FIRST | CVSS 4.0 | Scoring guidance and planned reference parity. The M3 adapter uses the exact commit recorded in ART-16. [User guide](https://www.first.org/cvss/v4.0/user-guide) · [Reference calculator](https://github.com/FIRSTdotorg/cvss-v4-calculator) |
| SRC-04 | Traditional JSON Report, ZAP Dev Team | Traditional JSON template documentation | Site, alert and instance hierarchy and report-declared metadata. [Template and example](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/) |
| SRC-05 | Passive Scan, ZAP Dev Team | Desktop documentation | Passive analysis of proxied traffic and its distinction from active scanning. [Passive scan](https://www.zaproxy.org/docs/desktop/start/features/pscan/) |
| SRC-06 | OWASP Juice Shop, OWASP Foundation | Project page | Training application and suitability as a controlled security lab fixture. The exact capture build comes from ART-03. [Project page](https://owasp.org/projects/juice-shop) |
| SRC-07 | Cross-Domain Misconfiguration, ZAP Dev Team | Passive alert 10098 | CORS scanner assertion, CWE-264 field and OWASP 2025 A01 tag. [Alert](https://www.zaproxy.org/docs/alerts/10098/) |
| SRC-08 | Content Security Policy (CSP) Header Not Set, ZAP Dev Team | Passive rule 10038 and alert subtype 10038-1 | Distinguishes parent plugin from the subtype with CWE-693 and OWASP 2025 A02 tag. [Rule](https://www.zaproxy.org/docs/alerts/10038/) · [Alert subtype](https://www.zaproxy.org/docs/alerts/10038-1/) |
| SRC-09 | Modern Web Application, ZAP Dev Team | Passive alert 10109 | Informational exploration diagnostic. [Alert](https://www.zaproxy.org/docs/alerts/10109/) |
| SRC-10 | CWE-89, MITRE Common Weakness Enumeration | CWE 5.0 page at access | SQL injection weakness definition used with SRC-01 by the project mapping. [CWE entry](https://cwe.mitre.org/data/definitions/89.html) |

## Implementation sources

| ID | Title and authoring organization | Publication or version | Topic and link |
| --- | --- | --- | --- |
| SRC-11 | Package http, Go project | Go standard library documentation | HTTP server and request routing. [Go net/http](https://pkg.go.dev/net/http) |
| SRC-12 | PostgreSQL Documentation, PostgreSQL Global Development Group | PostgreSQL 15, section 5.7 | Object ownership and database privileges. [Privileges](https://www.postgresql.org/docs/15/ddl-priv.html) |
| SRC-13 | pgx and sqlc documentation, pgx and sqlc contributors | pgx v5 and sqlc documentation | PostgreSQL driver and typed query generation. [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5) · [sqlc](https://docs.sqlc.dev/en/latest/) |
| SRC-14 | pressly/goose overview, goose project | Project documentation | Versioned PostgreSQL schema migrations. [goose](https://pressly.github.io/goose/) |
| SRC-15 | Quick Start, React project | React 19.3 documentation | Component-state and conditional rendering model. [React](https://react.dev/learn) |
| SRC-16 | OpenAPI Specification, OpenAPI Initiative | 3.1.0, 2021-02-15 | Version used by the application's typed interface contract. [OAS 3.1.0](https://spec.openapis.org/oas/v3.1.0.html) |
| SRC-17 | Playwright documentation, Playwright project | Tool documentation | Browser automation and assertions. [Playwright](https://playwright.dev/docs/intro) |
| SRC-18 | Networking in Compose, Docker | Docker Compose documentation | Internal network and port-publishing terminology. [Compose networking](https://docs.docker.com/compose/how-tos/networking/) |

## Analysis sources

| ID | Reference | Verified topic and link |
| --- | --- | --- |
| SRC-19 | M. G. Kendall, “The Treatment of Ties in Ranking Problems,” *Biometrika* 33(3), 239–251, 1945. DOI 10.1093/biomet/33.3.239. | Bibliographic record for tie-aware ranking comparison. [Publisher record](https://academic.oup.com/biomet/article-abstract/33/3/239/216180) |
| SRC-20 | Kalervo Järvelin and Jaana Kekäläinen, “Cumulated Gain-Based Evaluation of IR Techniques,” *ACM Transactions on Information Systems* 20(4), 422–446, 2002. DOI 10.1145/582415.582418. | Graded gain and discounted gain metrics. [Authors’ university research record](https://researchportal.tuni.fi/en/publications/cumulated-gain-based-evaluation-of-ir-techniques/) |

## Project-specific decisions

The study category subset, alert-to-category rules, canonical fingerprint, allowable context values, historical revision and audit policy, and research cohort and evaluation rules are project choices. Cite their exact revisions when discussing the implementation. OWASP category definitions and FIRST metric semantics are external authority. Access authorization, observed behavior, containment and research outcomes require project evidence. [Project contract](../../AGENTS.md) · [Reviewed mapping and evidence](implementation-evidence.md) · [Proposed protocol](../research-protocol.md)


## M3 source pins

| ID | Title and organization | Version, access date and topic |
| --- | --- | --- |
| SRC-21 | FIRST CVSS v4 calculator, FIRST, Red Hat and contributors | Commit `c5b0d409ae9f57c44264c6ce5f27d89298e1d32a`, checked 2026-11-24. BSD-2-Clause scoring sources and attribution. [Exact upstream tree](https://github.com/FIRSTdotorg/cvss-v4-calculator/tree/c5b0d409ae9f57c44264c6ce5f27d89298e1d32a) |
| SRC-22 | Goja, Goja contributors | Commit `35fdc4b88e393df269ccbfb4183c5490558eafe9`, checked 2026-11-24. MIT-licensed pure-Go ECMAScript execution. [Exact upstream tree](https://github.com/dop251/goja/tree/35fdc4b88e393df269ccbfb4183c5490558eafe9) |
