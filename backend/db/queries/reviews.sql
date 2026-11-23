-- name: LockFindingRevision :one
WITH locked AS MATERIALIZED (SELECT public.lock_finding_revision(sqlc.arg(finding_id)::uuid) AS id)
SELECT f.id FROM findings f JOIN locked l ON l.id=f.id;

-- name: EvidenceObservation :one
SELECT finding_id,target_id,source_pointer FROM observations WHERE id=$1;

-- name: LatestVerification :one
SELECT id,finding_id,state,procedure,rationale,evidence_refs_json,actor,created_at
FROM verification_reviews WHERE finding_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1;

-- name: CreateVerification :exec
INSERT INTO verification_reviews(id,finding_id,state,procedure,rationale,evidence_refs_json,actor,created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8);

-- name: ListVerifications :many
SELECT id,finding_id,state,procedure,rationale,evidence_refs_json,actor,created_at
FROM verification_reviews
WHERE finding_id=sqlc.arg(finding_id)
 AND (sqlc.arg(first_page)::boolean OR (created_at,id)<(sqlc.arg(after_time)::timestamptz,sqlc.arg(after_id)::uuid))
ORDER BY created_at DESC,id DESC LIMIT sqlc.arg(page_limit);
