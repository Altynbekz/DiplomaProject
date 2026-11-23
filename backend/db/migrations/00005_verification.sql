-- +goose Up
CREATE TABLE verification_reviews (
 id uuid PRIMARY KEY,
 finding_id uuid NOT NULL REFERENCES findings(id),
 state text NOT NULL CHECK (state IN ('unreviewed','confirmed','false_positive','inconclusive')),
 procedure text NOT NULL CHECK (char_length(procedure) BETWEEN 1 AND 4096),
 rationale text NOT NULL CHECK (char_length(rationale) BETWEEN 1 AND 4096),
 evidence_refs_json jsonb NOT NULL CHECK (jsonb_typeof(evidence_refs_json)='array' AND octet_length(evidence_refs_json::text)<=65536),
 actor text NOT NULL CHECK (char_length(actor) BETWEEN 1 AND 200),
 created_at timestamptz NOT NULL
);
CREATE INDEX verification_finding_history ON verification_reviews (finding_id,created_at DESC,id DESC);
GRANT SELECT, INSERT ON verification_reviews TO diploma_runtime;

-- Permit a row lock without granting any evidence column UPDATE privileges.
-- +goose StatementBegin
CREATE FUNCTION lock_finding_revision(finding uuid) RETURNS uuid
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog,pg_temp
AS $$
 SELECT f.id FROM public.findings f WHERE f.id=finding FOR UPDATE
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION lock_finding_revision(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION lock_finding_revision(uuid) TO diploma_runtime;

-- +goose Down
DROP FUNCTION lock_finding_revision(uuid);
DROP TABLE verification_reviews;
