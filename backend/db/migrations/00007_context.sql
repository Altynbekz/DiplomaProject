-- +goose Up
CREATE TABLE context_revisions (
 id uuid PRIMARY KEY,
 target_id uuid NOT NULL REFERENCES targets(id),
 method text,
 path text,
 revision integer NOT NULL CHECK (revision>=1),
 exposure text NOT NULL CHECK (exposure IN ('public','internal','restricted','unknown')),
 required_privileges text NOT NULL CHECK (required_privileges IN ('anonymous','user','admin','unknown')),
 data_sensitivity text NOT NULL CHECK (data_sensitivity IN ('public','internal','personal','highly_sensitive','unknown')),
 business_criticality text NOT NULL CHECK (business_criticality IN ('low','medium','high','critical','unknown')),
 factor_provenance_json jsonb NOT NULL CHECK (
   jsonb_typeof(factor_provenance_json)='object' AND octet_length(factor_provenance_json::text)<=65536
 ),
 actor text NOT NULL CHECK (char_length(actor) BETWEEN 1 AND 200),
 created_at timestamptz NOT NULL,
 CHECK ((method IS NULL AND path IS NULL) OR (method IS NOT NULL AND path IS NOT NULL
   AND char_length(method) BETWEEN 1 AND 64 AND char_length(path) BETWEEN 1 AND 4096
   AND left(path,1)='/' AND position(chr(92) in path)=0))
);
-- Exact comparisons in lookups still use the full path. Index its fixed
-- SHA-256 representation so valid endpoint paths fit PostgreSQL B-tree keys.
CREATE UNIQUE INDEX context_scope_revision ON context_revisions
 (target_id,method,(sha256(path::bytea)),revision) NULLS NOT DISTINCT;
CREATE INDEX context_target_history ON context_revisions (target_id,created_at DESC,id DESC);
GRANT SELECT, INSERT ON context_revisions TO diploma_runtime;

-- +goose Down
DROP TABLE context_revisions;
