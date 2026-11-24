# Security and ethics

## Implemented controls

| Boundary | Control and evidence |
| --- | --- |
| Operator and local access | Single local bearer credential. The frontend holds it in memory and uses the same-origin proxy. Host bindings are loopback. Compose uses loopback frontend/API publishing and a separate internal storage network with no PostgreSQL host port by default. [Startup](../../README.md#start-locally) · [Compose](../../compose.yaml) · [API authentication](../../backend/internal/httpapi/api.go) |
| Offline target scope | Registration resolves one server-configured origin and imports validate same-origin report data without resolving or fetching report URLs. The application has no scan launcher, background scanner or public deployment. [Scope validation](../../backend/internal/scope/url.go) · [API contract](../../contracts/openapi.yaml) |
| Hostile and private content | Multipart, JSON and report parser limits bound storage and processing. Escaped scanner text is rendered as text. Default reads withhold sensitive payload fields and URL values. Acknowledged raw download requires an operator token, audits access, uses no-store and returns the exact uploaded bytes. [Parser](../../backend/internal/zap/parser.go) · [Redaction](../../backend/internal/service/redact.go) · [HTTP handlers](../../backend/internal/httpapi/api.go) |
| Persistent data | Separate schema ownership and runtime role. Runtime can read and append evidence and histories without update, delete or truncate grants on those records. Narrow database permissions support locking. Keep original reports, credentials, sessions and databases outside Git. Publish only reviewed sanitized derivatives and opaque references. [Runtime migration](../../backend/db/migrations/00002_runtime_privileges.sql) · [Revision privileges](../../backend/db/migrations/00005_verification.sql) · [Evidence register](artifact-register.md) |

This is a research prototype. Bearer access and local database controls support one operator. The project has no production hardening, multi-user authorization policy, or independent external security certification.

## Historical collection and authorization

The sole genuine report relied on the explicitly authorized local Juice Shop passive session recorded as ART-01, ART-06 and ART-07. ZAP ran separately from the application. Independent checkpoint 20 inspected the authorization window and isolation evidence, verified the allowlisted browser path and zero denied-canary requests, and checked shutdown and report integrity. The origin guard forwarded only GET and HEAD. Three reported POST cases represent browser attempts rather than forwarded transactions. The total forwarded request count was not recorded. This verifies the observed collection and controls at that run's scope. It provides no general guarantee for future scans. [Collection status](../status.md#genuine-report-acceptance-readiness) · [Independent capture review](../coordination/review.md#independent-verification--m1-20261123-20)

Later A01 role/object activity was authorized separately. Automatic approval review stopped it before execution, so neither run generated live evidence. The test handler's fictional decisions and the passive CORS assertion give no observed role/object decision. Reference private records by ART-05 and ART-08. Keep unrelated fixture or policy results separate from Juice Shop findings. [A01 evidence checklist](../../lab/scenarios/a01-role-object.md#evidence-to-supply)

## Retention and future research

Preserve immutable originals in restricted local storage and publish a separately hashed, checked derivative with provenance. Limit private retention and access to authorized operators and reviewers under the applicable institutional policy. Keep readouts, screenshots and exports free of credentials, sessions, participant identifiers, private paths and raw payloads. Use aliases for disposable accounts and independent reviewers. Record the redaction procedure and disclosure decision per artifact. [Sanitization procedure](../fixture-provenance.md) · [Record template](experiment-record-template.md)

Any future live collection requires explicit authorization for its own fixture and verified isolation. An offline import or a historical authorization supplies no new scope. The proposed evaluation also requires consent, appropriate institutional review, access arrangements and blinded judgments. These are future records. [Research protocol](../research-protocol.md) · [Lab requirements](../../AGENTS.md#10-mandatory-lab-authorization-ssrf-controls-and-secrets)
