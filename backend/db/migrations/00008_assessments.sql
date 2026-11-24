-- +goose Up
ALTER TABLE verification_reviews ADD CONSTRAINT verification_review_finding_identity UNIQUE (id,finding_id);
ALTER TABLE classification_revisions ADD CONSTRAINT classification_revision_finding_identity UNIQUE (id,finding_id);
ALTER TABLE context_revisions ADD CONSTRAINT context_revision_target_identity UNIQUE (id,target_id);

CREATE TABLE assessment_snapshots (
 id uuid PRIMARY KEY,
 finding_id uuid NOT NULL,
 target_id uuid NOT NULL,
 kind text NOT NULL CHECK (kind IN ('B','BE')),
 status text NOT NULL CHECK (status IN ('incomplete','complete')),
 base_assessment_id uuid,
 verification_review_id uuid,
 classification_revision_id uuid,
 context_revision_id uuid,
 context_snapshot_json jsonb NOT NULL,
 metrics_json jsonb NOT NULL,
 metric_rationales_json jsonb NOT NULL,
 factor_dispositions_json jsonb NOT NULL,
 notes text NOT NULL CHECK (char_length(notes)<=4096),
 vector text,
 score numeric(3,1),
 calculator_json jsonb,
 schema_version text NOT NULL CHECK (schema_version='assessment-snapshot/v1'),
 policy_version text NOT NULL CHECK (char_length(policy_version) BETWEEN 1 AND 64),
 actor text NOT NULL CHECK (char_length(actor) BETWEEN 1 AND 200),
 created_at timestamptz NOT NULL,
 parent_kind text GENERATED ALWAYS AS (CASE WHEN kind='BE' THEN 'B' END) STORED,
 parent_status text GENERATED ALWAYS AS (CASE WHEN kind='BE' THEN 'complete' END) STORED,
 UNIQUE(id,finding_id,kind,status),
 FOREIGN KEY (finding_id,target_id) REFERENCES findings(id,target_id),
 FOREIGN KEY (verification_review_id,finding_id) REFERENCES verification_reviews(id,finding_id),
 FOREIGN KEY (classification_revision_id,finding_id) REFERENCES classification_revisions(id,finding_id),
 FOREIGN KEY (context_revision_id,target_id) REFERENCES context_revisions(id,target_id),
 FOREIGN KEY (base_assessment_id,finding_id,parent_kind,parent_status) REFERENCES assessment_snapshots(id,finding_id,kind,status),
 CONSTRAINT assessment_parent_required CHECK ((kind='BE')=(base_assessment_id IS NOT NULL)),
 CONSTRAINT assessment_score_state CHECK (
   (status='incomplete' AND score IS NULL AND vector IS NULL AND calculator_json IS NULL)
   OR (status='complete' AND score IS NOT NULL AND score BETWEEN 0 AND 10
       AND vector IS NOT NULL AND vector LIKE 'CVSS:4.0/%' AND char_length(vector)<=256
       AND calculator_json IS NOT NULL AND jsonb_typeof(calculator_json)='object'
       AND octet_length(calculator_json::text)<=4096
       AND verification_review_id IS NOT NULL AND classification_revision_id IS NOT NULL
       AND (kind='B' OR context_revision_id IS NOT NULL))
 ),
 CONSTRAINT assessment_snapshot_size CHECK (
   jsonb_typeof(context_snapshot_json)='object' AND octet_length(context_snapshot_json::text)<=131072
   AND jsonb_typeof(metrics_json)='object' AND octet_length(metrics_json::text)<=4096
   AND jsonb_typeof(metric_rationales_json)='object' AND octet_length(metric_rationales_json::text)<=65536
   AND jsonb_typeof(factor_dispositions_json)='object' AND octet_length(factor_dispositions_json::text)<=65536
 )
);
CREATE INDEX assessment_finding_history ON assessment_snapshots(finding_id,created_at DESC,id DESC);
CREATE INDEX assessment_finding_kind_history ON assessment_snapshots(finding_id,kind,created_at DESC,id DESC);
CREATE INDEX assessment_complete_base_history ON assessment_snapshots(finding_id,created_at DESC,id DESC)
 WHERE kind='B' AND status='complete';

-- Staleness describes current inputs. It never changes the persisted snapshot.
CREATE VIEW assessment_status AS
SELECT a.id,a.finding_id,a.kind,a.status,a.score,a.vector,a.created_at,
 a.verification_review_id IS DISTINCT FROM (
   SELECT v.id FROM verification_reviews v WHERE v.finding_id=a.finding_id ORDER BY v.created_at DESC,v.id DESC LIMIT 1
 ) OR a.classification_revision_id IS DISTINCT FROM (
   SELECT c.id FROM classification_revisions c WHERE c.finding_id=a.finding_id ORDER BY c.created_at DESC,c.id DESC LIMIT 1
 ) OR (a.kind='BE' AND (
   a.context_revision_id IS DISTINCT FROM (
     SELECT r.id FROM context_revisions r WHERE r.target_id=a.target_id
      AND ((r.method=f.method AND r.path=f.path) OR (r.method IS NULL AND r.path IS NULL))
     ORDER BY (r.method IS NOT NULL) DESC,r.revision DESC LIMIT 1
   ) OR a.base_assessment_id IS DISTINCT FROM (
     SELECT b.id FROM assessment_snapshots b WHERE b.finding_id=a.finding_id AND b.kind='B' AND b.status='complete'
     ORDER BY b.created_at DESC,b.id DESC LIMIT 1
   )
 )) AS stale
FROM assessment_snapshots a JOIN findings f ON f.id=a.finding_id;

GRANT SELECT, INSERT ON assessment_snapshots TO diploma_runtime;
GRANT SELECT ON assessment_status TO diploma_runtime;

-- +goose Down
DROP VIEW assessment_status;
DROP TABLE assessment_snapshots;
ALTER TABLE context_revisions DROP CONSTRAINT context_revision_target_identity;
ALTER TABLE classification_revisions DROP CONSTRAINT classification_revision_finding_identity;
ALTER TABLE verification_reviews DROP CONSTRAINT verification_review_finding_identity;
