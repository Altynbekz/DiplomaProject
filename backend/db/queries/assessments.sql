-- name: CreateAssessment :exec
INSERT INTO assessment_snapshots (
 id,finding_id,target_id,kind,status,base_assessment_id,verification_review_id,classification_revision_id,
 context_revision_id,context_snapshot_json,metrics_json,metric_rationales_json,factor_dispositions_json,notes,
 vector,score,calculator_json,schema_version,policy_version,actor,created_at
) VALUES (sqlc.arg(id),sqlc.arg(finding_id),sqlc.arg(target_id),sqlc.arg(kind),sqlc.arg(status),nullif(sqlc.arg(base_id)::text,'')::uuid,
 nullif(sqlc.arg(verification_id)::text,'')::uuid,
 nullif(sqlc.arg(classification_id)::text,'')::uuid,
 nullif(sqlc.arg(context_id)::text,'')::uuid,sqlc.arg(context_snapshot_json),sqlc.arg(metrics_json),sqlc.arg(metric_rationales_json),sqlc.arg(factor_dispositions_json),sqlc.arg(notes),sqlc.narg(vector),sqlc.narg(score),sqlc.narg(calculator_json),sqlc.arg(schema_version),sqlc.arg(policy_version),sqlc.arg(actor),sqlc.arg(created_at));

-- name: GetAssessment :one
SELECT sqlc.embed(a),s.stale
FROM assessment_snapshots a JOIN assessment_status s ON s.id=a.id
WHERE a.id=$1;

-- name: LatestAssessmentTime :one
SELECT created_at FROM assessment_snapshots WHERE finding_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1;

-- name: LatestCompleteBaseID :one
SELECT id FROM assessment_snapshots WHERE finding_id=$1 AND kind='B' AND status='complete'
ORDER BY created_at DESC,id DESC LIMIT 1;

-- name: ListAssessments :many
SELECT sqlc.embed(a),s.stale
FROM assessment_snapshots a JOIN assessment_status s ON s.id=a.id
WHERE a.finding_id=sqlc.arg(finding_id)
 AND (sqlc.arg(first_page)::boolean OR (a.created_at,a.id)<(sqlc.arg(after_time)::timestamptz,sqlc.arg(after_id)::uuid))
ORDER BY a.created_at DESC,a.id DESC LIMIT sqlc.arg(page_limit);
