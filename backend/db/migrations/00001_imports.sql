-- +goose Up
CREATE TABLE targets (
 id uuid PRIMARY KEY, name text NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
 canonical_origin text NOT NULL UNIQUE, lab_target_key text NOT NULL CHECK (lab_target_key='juice-shop'), created_at timestamptz NOT NULL
);
CREATE TABLE report_imports (
 id uuid PRIMARY KEY, target_id uuid NOT NULL REFERENCES targets(id), scan_run_id uuid,
 report_sha256 text NOT NULL CHECK (report_sha256 ~ '^[0-9a-f]{64}$'), parser_version text NOT NULL,
 raw_bytes bytea NOT NULL CHECK (octet_length(raw_bytes)<=20971520),
 source_metadata_json jsonb NOT NULL CHECK (octet_length(source_metadata_json::text)<=65536),
 warnings_json jsonb NOT NULL, unlocated_alerts_json jsonb NOT NULL,
 finding_count integer NOT NULL CHECK (finding_count>=0), observation_count integer NOT NULL CHECK (observation_count>=0),
 created_at timestamptz NOT NULL, UNIQUE(target_id,report_sha256,parser_version), UNIQUE(id,target_id)
);
CREATE INDEX report_target_time ON report_imports(target_id,created_at,id);
CREATE TABLE findings (
 id uuid PRIMARY KEY, target_id uuid NOT NULL REFERENCES targets(id), fingerprint_version text NOT NULL CHECK (fingerprint_version='v1'),
 fingerprint text NOT NULL CHECK (fingerprint ~ '^[0-9a-f]{64}$'), plugin_id text NOT NULL, method text NOT NULL,
 canonical_origin text NOT NULL, path text NOT NULL, query_parameter_names text[] NOT NULL,
 parameter_name text NOT NULL, title text NOT NULL, created_at timestamptz NOT NULL,
 UNIQUE(target_id,fingerprint_version,fingerprint), UNIQUE(id,target_id)
);
CREATE TABLE observations (
 id uuid PRIMARY KEY, target_id uuid NOT NULL, finding_id uuid NOT NULL, report_import_id uuid NOT NULL,
 source_pointer text NOT NULL, instance_index integer NOT NULL CHECK(instance_index>=0), raw_uri text NOT NULL,
 scanner_severity_raw text NOT NULL, scanner_severity text NOT NULL CHECK(scanner_severity IN ('informational','low','medium','high','unknown')),
 confidence_raw text NOT NULL, cwe_ids integer[] NOT NULL, evidence_json jsonb NOT NULL,
 description text NOT NULL, remediation text NOT NULL, created_at timestamptz NOT NULL,
 UNIQUE(report_import_id,source_pointer), FOREIGN KEY(finding_id,target_id) REFERENCES findings(id,target_id),
 FOREIGN KEY(report_import_id,target_id) REFERENCES report_imports(id,target_id)
);
CREATE INDEX observations_finding_time ON observations(finding_id,created_at,id);
CREATE INDEX observations_import ON observations(report_import_id);
CREATE TABLE audit_events (
 id uuid PRIMARY KEY, request_id text NOT NULL, actor text NOT NULL, action text NOT NULL,
 entity_type text NOT NULL, entity_id uuid NOT NULL, metadata_redacted_json jsonb NOT NULL, created_at timestamptz NOT NULL
);
-- +goose Down
DROP TABLE audit_events;
DROP TABLE observations;
DROP TABLE findings;
DROP TABLE report_imports;
DROP TABLE targets;
