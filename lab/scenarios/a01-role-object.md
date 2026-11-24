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

## Evidence to supply

For independent review, place an existing, legitimately obtained record under ignored `artifacts/a01/<run-id>/` or provide its private local path. Include the following items.

- The fixture origin and build digest, documented policy and procedure, disposable role and object aliases, and object ownership.
- The operator and authorization reference, allowed operations and limits, UTC validity window and actual run times, isolation results, and shutdown record.
- Expected and observed decisions per applicable case, including any unperformed case. Include UTC timestamps, policy log or response references, and SHA-256 hashes. A correct denial is a valid result.
- The original and redacted record hashes and a description of redactions. Keep authorization records, credentials, sessions, private originals and payloads outside Git.

Link an observed result to a finding only when the tested target, policy and operation substantiate its exact observation. Keep results from an unrelated fixture in a standalone record.

## Observed results

The passive Juice Shop session captured CORS scanner assertions. A later authorized standalone fixture attempt was blocked by automatic approval review before Docker started. Both the primary run and independent replay remained unstarted, with zero live requests and no observed decisions. The passed in-memory handler test exercised fictional cases. The independent reviewer determined that this leaves the role/object evidence gate open. The Juice Shop CORS finding remains separate.
