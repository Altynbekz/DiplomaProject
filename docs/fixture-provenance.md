# Fixture provenance

The only fixture currently available is [synthetic-traditional.json](../backend/testdata/zap/synthetic-traditional.json). It was authored to exercise the Traditional JSON schema, grouping, unknown severity and escaped display. Its version field identifies it as synthetic. It was never produced by a scanner. See [status](status.md) for its hash and observed counts.

## M1 acceptance checklist

1. Confirm that the lab authorization covered the Juice Shop origin, scan mode and scan time. Keep the original report and the private authorization record outside Git. Record `authorization_ref` as a non-secret reference, `scan_time` in RFC3339, and `report_template` as `traditional-json`. Record `zap_version`, `addon_versions`, `juice_shop_image`, `juice_shop_version` and `plan_sha256` from the run when available. Use null for unknown values.
2. Hash the untouched original and record `raw_sha256`. Apply a candidate-specific, reproducible transformation to a separate file. Review URI query values, scanner evidence, attack and other information, descriptions, extensions and metadata for private data. Preserve array order, plugin IDs, method, path and parameter/query names wherever safe. Record any identity changes in the ordered `sanitization_steps` along with the transformation command or script revision. Review the entire derivative for leakage and record its `sanitized_sha256`.
3. Run the offline preflight with the two report paths and the import provenance JSON file. It checks the recorded digests, required provenance fields, report shape and per-alert instance counts at the original positions. It prints aggregate counts. A reviewer must still confirm authorization, corresponding alert order and sanitization before the derivative can be used as a fixture.

   ```sh
   python3 scripts/check-zap-fixture.py "$ORIGINAL_REPORT" "$SANITIZED_REPORT" "$PROVENANCE"
   ```

4. Derive the expected observation pointers (`/site/{site}/alerts/{alert}/instances/{instance}`) and group them independently of the importer using [fingerprint v1](../AGENTS.md#5-normalization-and-real-report-import). Use the registered target UUID, the derivative's normalized location and parameter fields, and SHA-256 of the compact UTF-8 JSON identity array. Count distinct fingerprints and compare them and their exact pointer groups with the API and PostgreSQL results. Grouping identifies shared endpoint and alert identities. Root-cause equivalence requires separate evidence.
5. Add the reviewed derivative and shareable provenance to the genuine integration fixture. On a disposable database, check creation, same-byte and concurrent duplicate reuse, counts, grouping, pointers, exact imported derivative bytes and immutable provenance. In a browser, verify the actual findings and evidence persist after API and database restart. Record the observed checks, versions and hashes in [implementation status](status.md). Keep the private original outside Git.

[ZAP Traditional JSON schema example](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/).
