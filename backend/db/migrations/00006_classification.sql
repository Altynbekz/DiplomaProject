-- +goose Up
CREATE TABLE classification_revisions (
 id uuid PRIMARY KEY,
 finding_id uuid NOT NULL REFERENCES findings(id),
 status text NOT NULL CHECK (status IN ('mapped','unmapped','ambiguous','out_of_scope')),
 category text,
 candidate_categories text[] NOT NULL,
 mapping_version text NOT NULL CHECK (char_length(mapping_version) BETWEEN 1 AND 64),
 rule_refs_json jsonb NOT NULL CHECK (jsonb_typeof(rule_refs_json)='array' AND octet_length(rule_refs_json::text)<=65536),
 rationale text NOT NULL CHECK (char_length(rationale) BETWEEN 1 AND 4096),
 evidence_refs_json jsonb NOT NULL CHECK (jsonb_typeof(evidence_refs_json)='array' AND octet_length(evidence_refs_json::text)<=65536),
 actor text NOT NULL CHECK (char_length(actor) BETWEEN 1 AND 200),
 created_at timestamptz NOT NULL,
 CHECK (candidate_categories <@ ARRAY['A01:2025','A02:2025','A05:2025']::text[]
   AND array_position(candidate_categories,NULL) IS NULL),
 CHECK ((status='mapped' AND category IS NOT NULL AND candidate_categories=ARRAY[category])
   OR (status='ambiguous' AND category IS NULL AND cardinality(candidate_categories)>=1)
   OR (status IN ('unmapped','out_of_scope') AND category IS NULL AND cardinality(candidate_categories)=0))
);
CREATE INDEX classification_finding_history ON classification_revisions (finding_id,created_at DESC,id DESC);
GRANT SELECT, INSERT ON classification_revisions TO diploma_runtime;

-- +goose Down
DROP TABLE classification_revisions;
