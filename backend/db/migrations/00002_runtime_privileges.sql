-- +goose Up
-- Provision LOGIN/password separately, never in a migration.
-- +goose StatementBegin
DO $$ BEGIN
 IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='diploma_runtime') THEN
  CREATE ROLE diploma_runtime NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
 END IF;
END $$;
-- +goose StatementEnd
GRANT USAGE ON SCHEMA public TO diploma_runtime;
GRANT SELECT, INSERT ON targets,report_imports,findings,observations,audit_events TO diploma_runtime;
REVOKE UPDATE, DELETE, TRUNCATE ON targets,report_imports,findings,observations,audit_events FROM diploma_runtime;
-- +goose Down
REVOKE ALL ON targets,report_imports,findings,observations,audit_events FROM diploma_runtime;
