# A01 access-control evidence

This is a test protocol for an operator-owned access-control test fixture. Record an authorization reference and the documented access policy before using a live lab. The genuine passive Juice Shop report and the automated workflow tests provide scanner assertions and application behavior. They contain no observed role or object authorization decision.

## Fixture and expected decisions

Create two disposable identities, **owner** and **peer**, in the fixture's ordinary user role. Assign a private object **A** to the owner and a private object **B** to the peer. Use fictitious contents and keep the fixture account secrets private. Configure the fixture policy to permit only the owner to see a private object. Use the fixture's supported object access interface with both identities and without a session. Where the fixture has an administrator role, document its privileges and assign a disposable administrator only if its policy permits the check.

| Identity and object | Expected decision |
| --- | --- |
| Owner, A | Allow |
| Peer, B | Allow |
| Peer, A | Deny |
| Owner, B | Deny |
| Anonymous, A | Deny |
| Authorized administrator, A | Follow the documented administrator policy |

Record each actual decision with the scenario identifier, identity role, object label, time, and an evidence reference. A denial can be represented by a policy response, a lack of access through the supported interface, or another documented result. Record any mismatch as an observation with its own evidence. Separate legitimate role differences from unexpected access. Use an authorized operator's reproducible procedure and linked evidence before saving a confirmed review. Leave a finding inconclusive if its evidence is insufficient.

## Observed results

No live role/object result was collected for this project run. The previous authorization covered a single passive browser session and expired. The reviewed CORS mapping classifies a scanner assertion under A01 without confirming access by any role.
