-- name: LatestClassification :one
SELECT id,finding_id,status,category,candidate_categories,mapping_version,rule_refs_json,rationale,evidence_refs_json,actor,created_at
FROM classification_revisions WHERE finding_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1;

-- name: CreateClassification :exec
INSERT INTO classification_revisions(id,finding_id,status,category,candidate_categories,mapping_version,rule_refs_json,rationale,evidence_refs_json,actor,created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);

-- name: ListClassifications :many
SELECT id,finding_id,status,category,candidate_categories,mapping_version,rule_refs_json,rationale,evidence_refs_json,actor,created_at
FROM classification_revisions
WHERE finding_id=sqlc.arg(finding_id)
 AND (sqlc.arg(first_page)::boolean OR (created_at,id)<(sqlc.arg(after_time)::timestamptz,sqlc.arg(after_id)::uuid))
ORDER BY created_at DESC,id DESC LIMIT sqlc.arg(page_limit);

-- name: ClassificationEvidence :many
SELECT DISTINCT ON (cwe_ids) id,source_pointer,cwe_ids
FROM observations WHERE finding_id=$1
ORDER BY cwe_ids,created_at,id;
