# Fixture provenance

The [synthetic fixture](../backend/testdata/zap/synthetic-traditional.json) exercises grouping, unknown severity and escaped display. The [genuine fixture](../backend/testdata/zap/juice-shop-traditional.json) is the reviewed derivative of an authorized passive Juice Shop session. The original and authorization record stay in private lab artifacts. See [status](status.md) for the collection and verification results.

Select the genuine fixture on the target's import form. Paste the contents of [juice-shop-provenance.json](../backend/testdata/zap/juice-shop-provenance.json) into **Optional report provenance**. Keep `original.json` as the untouched private export.

With access to the private export, reproduce the exact derivative and validate both report digests and the source layout offline. Compare the completed plan digest separately against the private plan. To reproduce the independent oracle, use the fixed target UUID seeded in the disposable integration test. The Python derivation uses the documented fingerprint identity and reads no application or database state.

```sh
python3 -B scripts/zap-fixture.py sanitize "$ORIGINAL_REPORT" /tmp/reproduced-juice-shop.json
cmp /tmp/reproduced-juice-shop.json backend/testdata/zap/juice-shop-traditional.json
python3 -B scripts/check-zap-fixture.py "$ORIGINAL_REPORT" backend/testdata/zap/juice-shop-traditional.json backend/testdata/zap/juice-shop-provenance.json
python3 -B scripts/zap-fixture.py groups backend/testdata/zap/juice-shop-traditional.json /tmp/reproduced-groups.json --target 00000000-0000-4000-8000-000000000001
cmp /tmp/reproduced-groups.json backend/testdata/zap/juice-shop-expected.json
```

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

## Collect a report in an isolated lab

Each new collection requires a named authorization and fresh isolation checks. The application imports reports offline.

1. **Authorize the target.** Use one disposable local `juice-shop` instance at `http://juice-shop:3000`, with paths under `/` on that origin. Record the operator, the lab instance and its fixed network identity, the Juice Shop image digest and version, the passive-only mode, and the UTC validity window in the private authorization record. Use its non-secret reference in provenance.
2. **Verify isolation.** Pin every lab container image by version and SHA-256 digest, and pin the ZAP and add-on versions. Keep the exact platform image identities with the private run evidence. Put the lab on a dedicated internal network without host-published ports or routes to the app, database or Internet. Restrict browser connections to the private ZAP proxy; restrict ZAP connections to an allowlisting proxy that serves only `http://juice-shop:3000` through the pinned Juice Shop IP on TCP 3000. Disable the ZAP management API, external telemetry and callbacks. Block external DNS, alternate addresses and ports, host and metadata routes, and out-of-scope redirects and subresources. Disable IPv6 if it is untested. Use controlled local sinks to check the actual browser and ZAP paths before collection. Record that the allowed path works, denied sinks receive zero requests, and request limits are enforced. Stop if any check fails. [Project isolation requirements](../AGENTS.md#10-mandatory-lab-authorization-ssrf-controls-and-secrets)
3. **Capture normal browsing.** Use a fresh ZAP session in Safe mode and a fresh browser profile. Open the Juice Shop home page, browse the catalogue and open a product detail by hand. ZAP observes those requests passively. Keep automatic discovery, active scanning, fuzzing, forced browsing and request replay disabled. Allow one session at a time, at most 10 minutes, 1,000 requests and 2 requests per second through the verified limiter. Stop browsing at the first limit or loss of authorization or isolation. Let the passive queue drain within the same window before calling the capture complete. [ZAP modes](https://www.zaproxy.org/docs/desktop/start/features/modes/) · [Passive Scanner](https://www.zaproxy.org/docs/desktop/addons/passive-scanner/)
4. **Export and record.** Export the Juice Shop site with the `traditional-json` template. Keep the untouched file in a private folder such as ignored `artifacts/zap/<collection-id>/original.json`. Record its SHA-256, actual scan time, exact ZAP and add-on versions, Juice Shop version and image digest, authorization reference, the SHA-256 of the completed plan, and the isolation and completion evidence. Preserve partial or empty results accurately. An export with no located observations cannot demonstrate findings persistence. [Traditional JSON format](https://www.zaproxy.org/docs/desktop/addons/report-generation/report-traditional-json/)
5. **Stop and preserve evidence.** Close the lab browser, stop the disposable ZAP and Juice Shop instances and the lab-only limiter, and verify the lab has stopped. Keep original evidence and authorization records private. Follow the checklist above to prepare a reviewed derivative and test import and persistence using a disposable application database.
