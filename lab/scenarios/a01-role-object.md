# A01 access-control evidence

This protocol defines the evidence required from an operator-owned access-policy test fixture. The registered Juice Shop origin is `http://juice-shop:3000`. The [passive report provenance](../../backend/testdata/zap/juice-shop-provenance.json) identifies the image used for the prior capture. A new scenario needs the operator's exact fixture, image or build digest, documented access policy and authorization for its own test interface. The passive report and the automated application tests supply no observed role or object decision.

## Fixture and expected decisions

The proposed policy allows each ordinary user to read only objects they own. Use two disposable identities, **owner** and **peer**, with private objects **A** and **B** respectively. Use fictitious contents. If the selected fixture has a documented administrator exception, give that check a separate disposable identity. The matrix records the expected outcomes for the fixture owner's authorization-policy test interface. It is not a claim about the behavior of the registered Juice Shop build.

| Identity and object | Expected decision |
| --- | --- |
| Owner, A | Allow |
| Peer, B | Allow |
| Peer, A | Deny |
| Owner, B | Deny |
| Anonymous, A | Deny |
| Authorized administrator, A | Follow the documented administrator policy |

## Scope and records

1. **Authorize the fixture.** Identify the exact origin, build digest, policy reference, identities and objects, supported operations, operator and UTC start and expiry. Seed accounts and objects through the fixture's documented setup on a disposable stack. The previous passive authorization is expired and applies only to that capture.
2. **Constrain traffic.** The proposed live observation scope is ordinary sign-in and read-only owner views of A and B at the named fixture origin, using the documented interface. Use the fixture owner's authorization-policy test record for the other matrix decisions. Bind the client to an isolated internal lab network with egress limited to the pinned fixture address and port. Verify scope and redirect enforcement with controlled sinks before target traffic. Stop on isolation failure, expiry, unexpected destination or unexpected changes to the fixture. Cap the live observation at 5 minutes and 60 total requests, whichever comes first. Use fresh accounts and browser sessions without retries outside the named window.
3. **Preserve evidence.** Keep the private authorization, scope and isolation result, fixture version, UTC time window, operator procedure and original logs in ignored `artifacts/a01/<run-id>/`. Record each policy decision with the role and object aliases, expected and observed outcomes, UTC time and a SHA-256 digest of the policy log or response record. A denial or an unperformed case has its own outcome. Keep credentials, sessions and private payloads outside Git. Review and hash any shareable derivative separately.
4. **Close the run.** Stop the dedicated lab services and verify they stopped. End the disposable sessions and retain evidence privately. Before linking the result to a finding, match the tested policy and object operation to that finding's target, endpoint and observation reference. Record unrelated policy results as a standalone scenario. A correct denial is a valid result. Record an inconclusive review if the evidence is insufficient to confirm the finding.

## Observed results

No live role/object result was collected for this project run. The previous authorization covered a single passive browser session and expired. The reviewed CORS mapping classifies a scanner assertion under A01 without confirming access by any role.
