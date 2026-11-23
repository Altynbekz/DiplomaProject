-- name: FindingEndpoint :one
SELECT target_id,method,path FROM findings WHERE id=$1;

-- name: LatestContext :one
SELECT id,target_id,method,path,revision,exposure,required_privileges,data_sensitivity,business_criticality,factor_provenance_json,actor,created_at
FROM context_revisions WHERE target_id=sqlc.arg(target_id)
 AND method IS NOT DISTINCT FROM sqlc.narg(method)::text AND path IS NOT DISTINCT FROM sqlc.narg(path)::text
ORDER BY revision DESC LIMIT 1;

-- name: CreateContext :exec
INSERT INTO context_revisions(id,target_id,method,path,revision,exposure,required_privileges,data_sensitivity,business_criticality,factor_provenance_json,actor,created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);

-- name: ListContexts :many
SELECT id,target_id,method,path,revision,exposure,required_privileges,data_sensitivity,business_criticality,factor_provenance_json,actor,created_at
FROM context_revisions WHERE target_id=sqlc.arg(target_id)
 AND (sqlc.arg(first_page)::boolean OR (created_at,id)<(sqlc.arg(after_time)::timestamptz,sqlc.arg(after_id)::uuid))
ORDER BY created_at DESC,id DESC LIMIT sqlc.arg(page_limit);

-- name: ResolveContext :one
SELECT id,target_id,method,path,revision,exposure,required_privileges,data_sensitivity,business_criticality,factor_provenance_json,actor,created_at
FROM context_revisions WHERE target_id=sqlc.arg(target_id)
 AND ((method=sqlc.arg(method)::text AND path=sqlc.arg(path)::text) OR (method IS NULL AND path IS NULL))
ORDER BY (method IS NOT NULL) DESC,revision DESC LIMIT 1;
