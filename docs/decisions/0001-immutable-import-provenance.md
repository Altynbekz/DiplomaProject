# 0001 Keep operator provenance separate from report claims

M1. Adopted 2026-11-23.

ZAP Traditional JSON reports declare their own version. An imported file can provide that claim, while only the operator can supply provenance that links the file to the authorized lab. New imports keep those sources separate. `report_declared_version` is a nullable report claim. `parser_detected_format` records the format recognized by the application. The operator provenance object preserves unknown values as null and marks the new behavior with `report-provenance/v2`.

An exact duplicate with omitted or matching provenance returns the existing import. An exact duplicate with different supplied provenance now returns `409 conflict` at `/provenance`. This replaces the earlier behavior that ignored the differences. It prevents a successful response from implying new provenance was stored. No edit or revision route exists in M1.

Migration 4 adds the nullable report version column and an index for ordered observation receipt. Existing import records are kept intact. They return a null report-declared version and continue to carry their original provenance schema version. The UI identifies legacy provenance that may contain report-derived defaults. Clients should omit provenance when retrieving the existing import via duplicate upload and handle the conflict when supplying different metadata.
