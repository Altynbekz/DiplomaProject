# ADR 0002 — Offline CVSS v4 reference calculation

Accepted for the M3 foundation on 2026-11-24. The published UI and API must attribute CVSS to FIRST and keep the exact vector with its score. [FIRST CVSS v4.0 document v1.2](https://www.first.org/cvss/v4.0/specification-document)

## Calculator and parity

Use the BSD-2-Clause FIRST CVSS v4 calculator at commit `c5b0d409ae9f57c44264c6ce5f27d89298e1d32a`. The five inputs to the scoring function are included byte for byte with the original license. The Go adapter compiles trusted embedded scripts once and supplies only validated metric values to a fresh pure-Go Goja runtime per calculation. No browser UI, HTTP client, filesystem bridge, command execution or user-supplied JavaScript is available to that runtime. Goja is MIT-licensed, pinned at `v0.0.0-20261122185237-35fdc4b88e39`, source commit `35fdc4b88e393df269ccbfb4183c5490558eafe9`. [Pinned FIRST source and license](https://github.com/FIRSTdotorg/cvss-v4-calculator/tree/c5b0d409ae9f57c44264c6ce5f27d89298e1d32a) · [Pinned Goja source](https://github.com/dop251/goja/tree/35fdc4b88e393df269ccbfb4183c5490558eafe9)

The candidate Go-CVSS v0.6.4 implementation (MIT, commit `8df5dbd1cdb98348548087af3157a7f96d7521ba`) was rejected after independent differential tests found different rounding on valid B and BE vectors. Two of those input cases are in the permanent fixture. Keeping the FIRST arithmetic and lookup tables unchanged avoids maintaining a second formula.

[The parity generator](../../scripts/generate-cvss-parity.mjs) reads the separately checked-out clean FIRST revision in Node and produces [the fixture](../../backend/testdata/cvss/first-v4.json). Each case records the synthetic inputs, reference vector and reference score under the fixture's exact upstream revision. Runtime Goja and generator Node are separate execution paths. Replay locally with an offline checkout.

```sh
node scripts/generate-cvss-parity.mjs "$FIRST_REFERENCE_CHECKOUT" |
  cmp - backend/testdata/cvss/first-v4.json
cd backend && go test ./internal/cvss
```

## Planned snapshot and API contract

This section specifies design intent for later foundation slices. Their implementation and acceptance will be documented and verified separately from the calculator.

The server accepts metric dictionaries, reference IDs and justifications. It rejects duplicate JSON keys, unsupported metric groups and illegal values. It orders Base and Environmental abbreviations per FIRST and omits `X` from the canonical vector. BE copies the exact Base metric and rationale snapshot, materializes omitted Environmental choices as `X`, and requires at least one defined Environmental choice. Threat stays Not Defined. Partial and intentionally unfinished snapshots have null score, vector and calculator. The calculated score uses FIRST's one-decimal output in PostgreSQL `numeric(3,1)`, including genuine zero. Calculator, reference, execution-engine, fixture and specification revisions accompany every calculated vector.

Each save adds a snapshot and one audit event in the same transaction. Finalization checks current confirmed verification, mapped study classification and same-finding evidence. The server captures context, including factor provenance, source IDs and resolver version. For B, expected context IDs are empty and the actual context is retained for traceability without affecting the score. BE checks current resolved context under the target lock and requires known factors and documented dispositions. The only request extension is `factor_dispositions`, a factor-to-reason/metric-list object required for BE finalization. An empty metric list records that the factor justifies no change.

BE references an explicitly named complete same-finding B parent with the same review/classification revisions. The parent must be the most recent complete B for a new BE snapshot. A later incomplete B remains visible as the latest Base snapshot and leaves completed parent eligibility intact. Replacing the completed Base, the resolved context for BE, or review/classification for either mode marks affected assessments historical on reads. History keeps its exact vector, score and provenance. Storage and the read API enforce separate permissions and never recalculate a previous vector.
