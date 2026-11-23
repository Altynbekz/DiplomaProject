# Fixture provenance

The only fixture currently available is [synthetic-traditional.json](../backend/testdata/zap/synthetic-traditional.json). It was authored to exercise the Traditional JSON schema, grouping, unknown severity and escaped display. Its version field identifies it as synthetic. It was never produced by a scanner. See [status](status.md) for its hash and observed counts.

A genuine M1 fixture requires an existing Traditional JSON report from the authorized local Juice Shop lab, a provenance record and a verified sanitized derivative. Keep the original privately and compute its SHA-256 before transformations. Record:

- `zap_version`, `addon_versions` and `report_template`.
- `juice_shop_image` and `juice_shop_version`.
- `authorization_ref` as an evidence reference rather than a credential, plus RFC3339 `scan_time` and `plan_sha256` if available.
- `raw_sha256`, `sanitized_sha256` and ordered `sanitization_steps`. Use null for unknown metadata. The uploaded derivative must match its recorded sanitized digest.

For a candidate fixture, keep the original bytes immutable and apply a reviewed, reproducible transformation to a separate file. Inventory report fields that can carry secrets, including URI query values, scanner evidence, attack/other information, diagnostic extensions and report metadata. Use synthetic identities and retain array order, plugin identifiers, methods, paths, parameter names and query parameter names whenever they are safe and necessary for identity. Record any unavoidable identity changes. Review the exact derivative for leakage and record the transformation command or script and its revision. A transformation script should reflect the fields actually encountered in the real input. None is recorded yet.

Compute the expected instance count from the report's instance arrays and the unlocated count from alert objects with empty or absent instances. Independently derive fingerprint v1 by hashing the compact UTF-8 JSON array of target UUID, scanner name `zap`, plugin ID, uppercased method, canonical origin, normalized path, sorted unique query parameter names and parameter name. Compare distinct fingerprints against the API and database rows. Grouping can conflate different alerts with the same identity, so the fingerprint is a grouping key rather than proof of a common cause.

[ZAP Traditional JSON schema example](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/).
