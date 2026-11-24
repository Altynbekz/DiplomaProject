# Limitations

## Technical scope

- M1 handles one registered origin per ZAP Traditional JSON import. The tested fingerprint groups scanner observations using endpoint identity and parameter. It can group separate root causes or separate occurrences of the same cause across paths.
- The reviewed mapping covers selected A01, A02 and A05 rules. Unmapped, ambiguous and out-of-scope alerts remain visible. Mapping, scanner severity, human verification and context represent separate judgments. Findings without assessments display Not assessed.
- The prototype is offline and single-operator. Local authentication, database append permissions and bounded imports are workflow controls within that setting. Public hosting, multi-user access and scanner orchestration are outside the verified baseline.
- Context scopes are exact method/path or target default. Factor labels require stated rationale and sources when known. Recorded lab reachability alone gives no production exposure or business impact.

## Empirical scope

- The genuine dataset is one sanitized passive capture of a particular Juice Shop build with 21 observations, grouped into 17 findings. It covers limited ordinary browsing and a GET/HEAD upstream policy. Three report entries identify blocked browser POST attempts. The capture omitted a total forwarded request count.
- The fixture lacks a completed authorized role/object decision record. A01 CORS classification is a scanner assertion. Demonstration Inconclusive history and authored test fixtures show workflow behavior only.
- The lab is artificial and the dataset is convenience-selected. Unknown ground truth and the lack of a paired set of independently judged cases prevent claims about detection coverage, precision, recall or comparative ranking accuracy.
- Normalization tests, browser flows, audit equivalence and CI runs verify software behavior for their exact scopes and revisions. They do not measure real operational impact, classifier generalization or performance improvements.

## Methodological scope

Independent priority judgments will depend on reviewer expertise and the provided scenario evidence. Blinding, judgment provenance, agreement and recorded adjudication can reduce and expose bias. Contextual knowledge may still leak between roles, and limited recruitment may constrain independence. Grouped observations and repeated captures are dependent, excluded or uncertain cases can bias the evaluated cohort, and ties and missing labels may leave scenario statistics undefined. Small numbers of independent groups limit uncertainty estimates. Even a favorable result would apply to the studied cohort and rules. [Proposed protocol](../research-protocol.md)

The pinned FIRST adapter and assessment storage are implemented. Parity covers the recorded and independently sampled vectors, and the completed software test snapshots use synthetic inputs. The UI is read-only for assessment history. Browser authoring, comparative ranking, export replay and empirical analysis remain planned. The paper can describe implemented software and the proposed research method. A comparative claim needs approved paired assessments, a frozen export, blinded reference judgments and independently reproducible analysis. [Claim–evidence map](claim-evidence-map.md)
