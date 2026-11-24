-- name: Ready :one
SELECT to_regclass('public.observations') IS NOT NULL AND to_regclass('public.verification_reviews') IS NOT NULL AND to_regclass('public.classification_revisions') IS NOT NULL AND to_regclass('public.context_revisions') IS NOT NULL AND to_regclass('public.assessment_snapshots') IS NOT NULL AND to_regclass('public.assessment_status') IS NOT NULL AND to_regclass('public.report_imports') IS NOT NULL AND has_column_privilege(current_user,'public.report_imports','report_declared_version','SELECT') AND has_column_privilege(current_user,'public.targets','name','UPDATE') AND NOT has_table_privilege(current_user,'public.observations','UPDATE,DELETE') AS ready;
-- name: CreateTarget :one
INSERT INTO targets (id,name,canonical_origin,lab_target_key,created_at) VALUES ($1,$2,$3,$4,$5)
RETURNING id,name,canonical_origin,lab_target_key,created_at;
-- name: GetTarget :one
SELECT id,name,canonical_origin,lab_target_key,created_at FROM targets WHERE id=$1;
-- name: ListTargets :many
SELECT id,name,canonical_origin,lab_target_key,created_at FROM targets WHERE id::text > $1 ORDER BY id LIMIT $2;
-- name: LockTarget :one
SELECT id FROM targets WHERE id=$1 FOR UPDATE;
-- name: FindImport :one
SELECT id,target_id,report_sha256,parser_version,source_metadata_json,warnings_json,unlocated_alerts_json,finding_count,observation_count,created_at,report_declared_version FROM report_imports WHERE target_id=$1 AND report_sha256=$2 AND parser_version=$3;
-- name: GetImport :one
SELECT id,target_id,report_sha256,parser_version,source_metadata_json,warnings_json,unlocated_alerts_json,finding_count,observation_count,created_at,report_declared_version FROM report_imports WHERE id=$1;
-- name: GetRaw :one
SELECT raw_bytes FROM report_imports WHERE id=$1;
-- name: CreateImport :exec
INSERT INTO report_imports(id,target_id,report_sha256,parser_version,raw_bytes,source_metadata_json,warnings_json,unlocated_alerts_json,finding_count,observation_count,created_at,report_declared_version) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);
-- name: FindFinding :one
SELECT id FROM findings WHERE target_id=$1 AND fingerprint_version='v1' AND fingerprint=$2;
-- name: CreateFinding :exec
INSERT INTO findings(id,target_id,fingerprint_version,fingerprint,plugin_id,method,canonical_origin,path,query_parameter_names,parameter_name,title,created_at) VALUES ($1,$2,'v1',$3,$4,$5,$6,$7,$8,$9,$10,$11);
-- name: CreateObservation :exec
INSERT INTO observations(id,target_id,finding_id,report_import_id,source_pointer,instance_index,raw_uri,scanner_severity_raw,scanner_severity,confidence_raw,cwe_ids,evidence_json,description,remediation,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15);
-- name: Audit :exec
INSERT INTO audit_events(id,request_id,actor,action,entity_type,entity_id,metadata_redacted_json,created_at) VALUES ($1,$2,$3,$4,$5,$6,'{}',$7);
-- name: ListFindings :many
SELECT f.id,f.target_id,f.fingerprint_version,f.fingerprint,f.plugin_id,f.method,f.canonical_origin,f.path,f.query_parameter_names,f.parameter_name,f.title,f.created_at,
 o.id AS severity_observation_id,o.scanner_severity,
 coalesce(v.state,'unreviewed')::text AS verification,coalesce(v.id::text,'')::text AS verification_id,
 coalesce(c.status,'unmapped')::text AS classification_status,
 coalesce(c.id::text,'')::text AS classification_id,coalesce(c.category,'')::text AS classification_category,
 coalesce(c.mapping_version,'')::text AS classification_version,
 coalesce(c.candidate_categories,ARRAY[]::text[])::text[] AS classification_candidates,
 coalesce(to_jsonb(b), 'null'::jsonb)::jsonb AS latest_base,
 coalesce(to_jsonb(be), 'null'::jsonb)::jsonb AS latest_environmental,
 (SELECT count(*) FROM observations n WHERE n.finding_id=f.id AND (sqlc.arg(import_id)::text='' OR n.report_import_id::text=sqlc.arg(import_id))) AS observation_count
FROM findings f
JOIN LATERAL (SELECT ob.id,ob.scanner_severity FROM observations ob WHERE ob.finding_id=f.id AND (sqlc.arg(import_id)::text='' OR ob.report_import_id::text=sqlc.arg(import_id)) ORDER BY ob.created_at DESC,ob.id DESC LIMIT 1) o ON true
LEFT JOIN LATERAL (SELECT vr.id,vr.state FROM verification_reviews vr WHERE vr.finding_id=f.id ORDER BY vr.created_at DESC,vr.id DESC LIMIT 1) v ON true
LEFT JOIN LATERAL (SELECT cr.id,cr.status,cr.category,cr.candidate_categories,cr.mapping_version
 FROM classification_revisions cr WHERE cr.finding_id=f.id ORDER BY cr.created_at DESC,cr.id DESC LIMIT 1) c ON true
LEFT JOIN LATERAL (
 SELECT s.id,s.kind,s.status,s.score,s.vector,s.stale,s.created_at FROM assessment_status s
 WHERE s.finding_id=f.id AND s.kind='B' ORDER BY s.created_at DESC,s.id DESC LIMIT 1
) b ON true
LEFT JOIN LATERAL (
 SELECT s.id,s.kind,s.status,s.score,s.vector,s.stale,s.created_at FROM assessment_status s
 WHERE s.finding_id=f.id AND s.kind='BE' ORDER BY s.created_at DESC,s.id DESC LIMIT 1
) be ON true
WHERE f.target_id=sqlc.arg(target_id) AND f.id::text>sqlc.arg(after_id)::text
 AND (sqlc.arg(finding_id)::text='' OR f.id::text=sqlc.arg(finding_id))
 AND (sqlc.arg(severity)::text='' OR o.scanner_severity=sqlc.arg(severity))
 AND (sqlc.arg(verification)::text='' OR sqlc.arg(verification)::text=coalesce(v.state,'unreviewed'))
 AND (sqlc.arg(category)::text='' OR c.category=sqlc.arg(category))
ORDER BY f.id LIMIT sqlc.arg(page_limit);
-- name: ListObservations :many
SELECT id,report_import_id,source_pointer,scanner_severity_raw,scanner_severity,confidence_raw,cwe_ids,evidence_json,description,remediation,created_at
FROM observations WHERE finding_id=$1 AND (created_at,id) > (sqlc.arg(after_time)::timestamptz,sqlc.arg(after_id)::uuid)
ORDER BY created_at,id LIMIT sqlc.arg(page_limit);

-- name: FindingTarget :one
SELECT target_id FROM findings WHERE id=$1;

-- name: LatestTargetObservationTime :one
SELECT created_at FROM observations WHERE target_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1;
