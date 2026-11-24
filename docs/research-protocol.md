# Proposed research protocol

**Version 0.1, authored 2026-11-24 (Asia/Almaty), awaiting independent research review.** The proposal precedes evaluation. It records no preregistration, reference judgments, scores, rankings or comparative results. Approval, freeze date, code and manifest digests will be recorded before held-out judgments or outcomes are examined. [Research contract](../AGENTS.md#12-research-protocol-and-python-analysis)

## Question and paired design

**RQ-01** asks whether a contextual ranking agrees more closely with independent reference prioritization for the studied web application scenarios. The hypothesis is that agreement increases for CVSS-BE compared with CVSS-B. Positive, null and negative differences are all possible. CVSS measures severity. Business consequences beyond the metric choices remain qualitative. [FIRST nomenclature and metrics](research/sources.md#standards-and-lab-sources)

| Mode | Selection and order |
| --- | --- |
| Technical-only | One complete CVSS v4.0 Base assessment (CVSS-B) per eligible finding/scenario, descending by the official rounded score |
| Contextual | A complete CVSS-BE assessment on the identical cohort, with the same frozen CVSS-B assessment as its parent, descending by the official rounded score |

Freeze Base assessments before environmental assessment. Record all Base and Environmental choices and their evidence, including contextual factors that justify no metric change. An unknown factor, incomplete Base, or missing required Environmental justification prevents a paired complete assessment. A CVSS-BE label requires at least one defined Environmental metric. Threat remains Not Defined. Freeze calculator version, source commit, parity fixture and output vectors before comparing ranks. Tied scores share a rank and stable display order adds no priority distinction. The displayed ZAP severity may be a separately labeled descriptive baseline.

## Corpus, eligibility and exclusions

The unit is a verified normalized finding in a named scenario. Predefine the inventory, intentional negative and unobservable cases, fixture versions and the grouping of cases by scenario and underlying weakness. Use Juice Shop as the primary lab. Record any other fixture as a separate source. Document actual authorized scope and intended application/business context independently. The existing passive report and demo rows establish engineering readiness. Any case used to design the rubric or calibrate assessors belongs to calibration, with its entire scenario/root-weakness group excluded from held-out evaluation.

Before review, freeze the candidate IDs, evidence cutoff, policy versions and inclusion/exclusion rules for both modes. Include only confirmed, mapped A01:2025, A02:2025 or A05:2025 cases with applicable evidence, a complete context snapshot and a complete paired assessment built from the same current review/classification state. Record exclusion codes for unreviewed, false positive, inconclusive, unmapped, ambiguous, out of scope, unresolved context, incomplete or stale assessment, unmatched Base parent, duplicate grouping, and unobservable cases as applicable. Report counts and reasons for all excluded candidates. Never use scanner severity, missing assessments or unrelated evidence to fill a numerical score. [Current claim map](research/claim-evidence-map.md)

## Independent reference judgments

Recruit at least two security-competent reviewers who did not author the ranking policy. Document roles and conflicts and keep assessors separate from reference reviewers where feasible. If only one reviewer is available, report the shortfall and treat any single-rater comparison as descriptive. Supervisory and applicable ethics approval, consent and data handling arrangements must be recorded before recruitment.

Provide identical sanitized, verified evidence and documented business scenarios in a randomized, recorded order with opaque IDs. The written rubric asks reviewers to order urgency within a scenario using demonstrated feasibility, prerequisites and evidenced consequences, allowing ties when the evidence cannot separate urgency. Collect each ranking, rationale, confidence, explicit ties and missing or insufficient evidence independently. Keep system modes, scores, vectors, displayed scanner priority and produced orders hidden until judgments and consensus are frozen.

Preserve each original response and report between-reviewer agreement using Kendall tau-b on commonly rated cases. Record each disagreement with case IDs, initial weak orders, rationale and final disposition. Use a separate blinded adjudicator to produce a coherent scenario order when the two reviews differ. Preserve the adjudication reasons. Unresolved or missing judgments stay missing, with reasons. The scoring cohort remains frozen. For agreement metrics, intersect it with available reference judgments once per scenario and use that same subset for both comparisons, reporting attrition and its reasons.

## Outcomes, ties and uncertainty

- **Primary** — Kendall tau-b for each method against the same frozen consensus order in each scenario. Report contextual minus technical agreement per scenario and the unweighted mean paired difference across scenarios with both defined. Report undefined scenarios separately.
- **Secondary** — Spearman correlation using average tie ranks and top-k agreement at proposed cutoffs `k = 1, 3, 5` for scenarios with at least k comparable cases. Allocate top-k membership fractionally across a tie spanning the cutoff and sum the minimum paired membership weights divided by k. Freeze the cutoffs before evaluating.
- **Conditional** — nDCG@k only with independently collected graded relevance under a reviewed, prespecified rubric. Use linear gain `g = r` and discount `1 / log2(position + 1)`. Average gain across tied positions, including at the cutoff. All-zero ideal relevance yields undefined nDCG and a reported count.
- Preserve ties in all inputs. Fewer than two comparable items or a constant ranking can make rank correlation undefined. Store null with a reason. Report paired differences, scenario counts, cohort sizes, missingness, exclusions and disagreement examples.
- With a defensible number of independent scenario/root-weakness clusters, estimate percentile 95% intervals for the mean paired delta with a seeded paired cluster bootstrap. Proposed analysis uses 10,000 resamples and seed 20261124. Freeze the cluster definition and minimum defensible cluster count before held-out analysis. With too few independent clusters, report descriptive differences and explain the uncertainty. Repeated imports and correlated observations are not independent clusters.

The proposed statistics draw on the ranking literature. The cohort rules, tie treatment at cutoffs and interpretation are study-specific. [Method sources](research/sources.md#analysis-sources)

## Reproduction and threats to validity

Freeze a `diploma-research/v1` export with export ID and UTC time, code commit or dirty-tree source digest, schema, parser, mapping, context policy/resolver and calculator versions, and lab/plan/report digests. Include candidate, included and excluded IDs with reasons, scenario/cohort IDs and sanitized evidence provenance references. Include the frozen review and classification records and the resolved context snapshot for each selected assessment, including factor values, rationales, source revision IDs and resolver version. Preserve paired assessment IDs, vectors, scores and ranking tie groups. Keep private originals and secrets outside the export. [Record template](research/experiment-record-template.md)

Freeze the blinded reference judgments separately before revealing system rankings. Bind the exact reference file's SHA-256 to the export manifest. The offline analysis checks that digest, scenario and case IDs, cohort equality and assessment parentage before comparing ranks. Stop on mismatches. Pin dependencies and seed and retain input/output hashes, machine-readable results, per-scenario tables, exclusions and figures. Validate the metrics against hand-checkable fixtures, including ties and undefined cases.

Discuss the artificial lab, small and convenience-selected corpus, limited passive coverage, grouping and mapping errors, contextual knowledge leakage, reviewer subjectivity and overlap, cohort attrition, ties and missingness, uncertainty, and external generalization. A separate detection-precision or recall claim requires independent known-issue and negative-case labels. [Limitations](research/limitations.md)

## Version history

| Date | Version | Change | Status |
| --- | --- | --- | --- |
| 2026-11-24 | 0.1 | Initial paired comparison, eligibility, blinded judgment and analysis proposal | Proposed, awaiting research review |

Append dated amendments with reasons, prior version and reviewer disposition. Identify any changes made after accessing held-out inputs or results and label their analyses exploratory.
