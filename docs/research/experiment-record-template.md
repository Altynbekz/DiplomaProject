# Experiment record template

Copy into a new, versioned record for one run. Keep sensitive fields in a private companion record referenced by opaque ID and SHA-256. Preserve this distinction when a run is blocked, partial, offline or synthetic.

| Identity field | Value |
| --- | --- |
| Record ID and revision | |
| Prior record and reason for revision, if applicable | |
| Record authoring time and actual run start/end, UTC | |
| Objective and research question or hypothesis | |
| Run type | Genuine lab observation, synthetic test, software verification, scenario assumption or planned work |
| Status | Planned, blocked, in progress, partial, completed or independently verified |
| Operator and reviewer aliases or opaque IDs | |

## Scenario and cohort

- Name the fixture, applicable target and exact build. State the access policy, identity and object aliases, documented expected decisions and the intended business scenario.
- Record exposure, required privileges, data sensitivity and business criticality with rationale, sources and unknowns. Distinguish assumptions from actual lab reachability and observations.
- Identify the dataset, versions, scenario/root-weakness grouping, calibration or evaluation partition, candidate IDs, inclusion rules, cutoff and exclusion codes.

## Environment, authorization and procedure

- Record code revision, dirty-tree flag and source digest if dirty, schema/parser/mapping and tool versions, OS/architecture, image IDs and configuration digest. List configuration keys without private values.
- For applicable live activity, record the authorization ID and digest, UTC validity window, allowed target, operations and limits, verified isolation record and stop conditions. Record offline-only status when applicable.
- List commands in execution order with working-directory or version references, inputs, expected behavior, observed exit status and error codes. Omit credentials, absolute private paths and session data.
- Record readiness, state changes, shutdown and the resources retained under the approved procedure.

## Input and output manifest

| Artifact ID | Role and provenance | Version or UTC time | SHA-256 | Disclosure class and location/reference |
| --- | --- | --- | --- | --- |
| | Input, original, derivative, output or verification record | | | |

Document derivations and sanitization separately. Hash output bytes and the manifest. Record seed, configuration and tool lock used to generate each analysis output.

## Expected and observed results

| Case or metric | Expected behavior or hypothesis | Observed result and UTC time | Evidence ID/digest | Interpretation and limits |
| --- | --- | --- | --- | --- |
| | | | | |

Use an explicit status for unperformed cases. Record failures, deviations, excluded candidates, missing values, unresolved decisions and the effect on acceptance or eligibility. A demonstration state or synthetic assertion requires that label.

## Independent verification and reproduction

- Reviewer role, independence, blinded material where applicable, verification date, frozen record and code hashes, checks, conclusions and action items.
- Reproduction from authorized, versioned inputs, including exact replay command, expected checks, required private references and cleanup.
- Remaining uncertainty and conditions under which reproducibility or generalization has not been demonstrated.
