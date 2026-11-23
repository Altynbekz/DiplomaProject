# ZAP Traditional JSON fixtures

`synthetic-traditional.json` is authored test data, not ZAP output and not lab measurement. It cannot satisfy M1's genuine-report gate. No scan was performed. It exercises numeric-string fields, grouping, method/parameter identity, unknown severity, unlocated alerts, and unsafe scanner text. All identities and payload values are synthetic.

`juice-shop-traditional.json` is the reviewed derivative of an authorized passive ZAP 2.17.0 Juice Shop session. `juice-shop-provenance.json` records the source and derivative digests and the sanitization procedure. `juice-shop-expected.json` holds the independent exact fingerprint and pointer grouping for the seeded test target. The 21 observations form 17 findings. See [fixture provenance](../../../docs/fixture-provenance.md) for reproducibility.
