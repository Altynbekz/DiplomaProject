# Diploma research artifacts

**Paper title** — Development of a Context-Aware Web Application Vulnerability Assessment System with OWASP Top 10 Risk Prioritization

The M0–M2 evidence baseline is `f1613e6fd5d64625dc239d2673cb019ecb964efd`. M1 genuine-report acceptance is complete. The M2 software is independently verified and its live A01 role/object evidence gate remains open. M3 provides offline scoring, analyst authoring and immutable assessment history. Its completed software demonstrations use synthetic assessments. Ranking, scanner orchestration and research evaluation are planned. The research question, **RQ-01**, asks whether contextual CVSS-BE ranking agrees more closely with independent reference prioritization than CVSS-B ranking in the studied scenarios. Current artifacts support writing about the system and its verification. The comparative outcome is unknown. [Implementation status](../status.md) · [Independent software review](../coordination/review.md#independent-verification--m2-20261124-31)

## Ready for writing

| Artifact | Use now | When to update |
| --- | --- | --- |
| [Artifact register](artifact-register.md) | Identify evidence, versions, capture dates, hashes, disclosure classifications and figures | Add a dated record for each capture, derivative, verification run or released dataset |
| [System design](system-design.md) | Describe the implemented browser, API, service and database boundaries | Update after a verified contract, schema or architecture change |
| [Implementation evidence](implementation-evidence.md) | Trace software claims to code, tests, independent checkpoints and exact CI runs | Add evidence for a verified slice and retain the baseline references |
| [Security and ethics](security-and-ethics.md) | Explain offline scope, authorization, isolation and private evidence handling | Record actual changes to controls, disclosure or approval separately from a proposed procedure |
| [Limitations](limitations.md) | Bound technical and empirical claims | Reassess after a new corpus, evidence gate, reviewer process or evaluation |
| [Experiment record template](experiment-record-template.md) | Capture reproduction details for a future authorized run | Instantiate once per run and version amendments |
| [Proposed research protocol](../research-protocol.md) | Describe the comparison and the planned analysis | Date and justify amendments, then freeze the reviewed protocol before seeing evaluation data or results |
| [Source register](sources.md) and [claim–evidence map](claim-evidence-map.md) | Cite primary sources and check the support for each sentence of the paper | Verify new sources and add claim rows as corroborating evidence arrives |
| [Paper outline](paper-outline.md) | Draft the introduction, design, implementation and methods structure | Replace result placeholders only after verified data and analysis exist |
| [Application figures](artifact-register.md#application-figures) | Illustrate the present UI using the genuine sanitized report and scoped demo inputs | Preserve these PNGs and add new figure IDs for later captures |

## Evidence still to collect

| Stage | Record needed |
| --- | --- |
| Open M2 gate | Authorized, independently reviewable observed A01 role/object decisions for the documented fixture and policy. Keep standalone scenarios separate from the existing Juice Shop CORS assertion. [Operator checklist](../../lab/scenarios/a01-role-object.md#evidence-to-supply) |
| M3 | Human metric justifications and authorized paired B/BE assessments. The pinned calculator and reference fixtures are recorded in the implementation evidence. |
| M4 | Frozen cohort with candidate and excluded IDs, paired assessment IDs, tie groups, cutoffs and a sanitized export manifest |
| M5 | Applicable approved lab scope, isolation and lifecycle records if using newly collected scanner data. Keep scanner orchestration outside the current application. |
| M6 | Reviewed protocol, frozen calibration/evaluation split, independent blinded judgments and disagreements, replayable analysis outputs and limitations |

Human judgment is needed for lab authorization, vulnerability confirmation, context and CVSS rationale, reference priorities, and disclosure review. Independent software verification tests the workflow. It supplies no evaluation labels or reference rankings.

## Preservation rules

Assign stable IDs and hash immutable bytes. Distinguish actual UTC activity and capture times from document authoring dates. Use `unknown` when the record does not supply a value. Keep private originals, authorization records, identity mappings, tokens and sessions outside Git, with only opaque IDs and digests in shareable research records. Publish a reviewed derivative under its own ID and record the transformation. Preserve superseded experiment records. An amendment gets a date, reason and predecessor ID, and a replay gets its own run ID and output hashes. [Fixture provenance](../fixture-provenance.md)
