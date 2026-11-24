# Context review aid, version 1

The assessment policy identifier is `context-policy/v1`. The API captures the resolver output and its provenance. Analysts choose the CVSS metrics and justify Environmental choices within the authorized scenario. The policy does not convert context labels into scores or default metric selections. [Assessment design](decisions/0002-cvss-v4-calculator.md) · [FIRST CVSS 4.0](https://www.first.org/cvss/v4.0/specification-document)

| Factor | Review prompt |
| --- | --- |
| Exposure | Consider which evidenced changes affect environmental feasibility. Intended reachability and observed lab isolation have different meanings. |
| Required privileges | Consider the prerequisites of the documented operation. The assessor's current login role does not establish a metric. |
| Data sensitivity | Evaluate whether the evidence justifies a Confidentiality Requirement or another relevant Environmental choice. |
| Business criticality | Evaluate scenario-specific Integrity and Availability Requirements and relevant Environmental choices. Record other consequences as qualitative notes. |

Each BE factor needs a disposition with a reason. A disposition can list no metric changes. For any factor used by a metric, the metric rationale and factor disposition name one another. A defined Environmental choice needs a rationale and matching observation evidence or a reviewed context factor. An unknown factor keeps BE incomplete. Undefined modified metrics inherit the Base choice. Undefined Environmental requirements use FIRST defaults. The server preserves explicit or materialized `X` in the snapshot and omits those defaults from the canonical vector. Threat is left Not Defined.

A complete B requires Base metrics and observation evidence for each choice. A BE inherits the current complete B parent and its full Base evidence. An assessor must gather applicable lab evidence before changing a finding's review to confirmed. Existing scanner severity and classification remain separate from that review and from CVSS.
