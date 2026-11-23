-- +goose Up
ALTER TABLE report_imports ADD COLUMN report_declared_version text
  CHECK (octet_length(report_declared_version) <= 65536);
CREATE INDEX observations_target_time ON observations(target_id,created_at DESC,id DESC);

-- +goose Down
DROP INDEX observations_target_time;
ALTER TABLE report_imports DROP COLUMN report_declared_version;
