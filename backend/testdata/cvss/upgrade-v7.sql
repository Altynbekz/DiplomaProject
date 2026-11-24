-- Authored M2 records for the isolated v7 -> v8 upgrade test only.
INSERT INTO targets(id,name,canonical_origin,lab_target_key,created_at)
VALUES ('00000000-0000-4000-8000-000000000071','Synthetic upgrade target','http://juice-shop:3000','juice-shop','2026-01-01T00:00:00Z');
INSERT INTO report_imports(id,target_id,report_sha256,parser_version,raw_bytes,source_metadata_json,warnings_json,
 unlocated_alerts_json,finding_count,observation_count,created_at)
VALUES ('00000000-0000-4000-8000-000000000072','00000000-0000-4000-8000-000000000071',
 encode(sha256(decode('7b7d','hex')),'hex'),'zap-traditional/v1',decode('7b7d','hex'),'{}','[]','[]',1,1,'2026-01-01T00:00:00Z');
INSERT INTO findings(id,target_id,fingerprint_version,fingerprint,plugin_id,method,canonical_origin,path,
 query_parameter_names,parameter_name,title,created_at)
VALUES ('00000000-0000-4000-8000-000000000073','00000000-0000-4000-8000-000000000071',
 'v1',repeat('a',64),'10038','GET','http://juice-shop:3000','/','{}','','Synthetic upgrade observation','2026-01-01T00:00:00Z');
INSERT INTO observations(id,target_id,finding_id,report_import_id,source_pointer,instance_index,raw_uri,
 scanner_severity_raw,scanner_severity,confidence_raw,cwe_ids,evidence_json,description,remediation,created_at)
VALUES ('00000000-0000-4000-8000-000000000074','00000000-0000-4000-8000-000000000071',
 '00000000-0000-4000-8000-000000000073','00000000-0000-4000-8000-000000000072','/site/0/alerts/0/instances/0',
 0,'http://juice-shop:3000/','2','medium','2','{693}','{}','Synthetic fixture','','2026-01-01T00:00:00Z');
INSERT INTO verification_reviews(id,finding_id,state,procedure,rationale,evidence_refs_json,actor,created_at)
VALUES ('00000000-0000-4000-8000-000000000075','00000000-0000-4000-8000-000000000073','inconclusive',
 'Synthetic review','No live evidence',
 '[{"observation_id":"00000000-0000-4000-8000-000000000074","source_pointer":"/site/0/alerts/0/instances/0","note":""}]',
 'synthetic-operator','2026-01-01T00:00:00Z');
INSERT INTO classification_revisions(id,finding_id,status,category,candidate_categories,mapping_version,
 rule_refs_json,rationale,evidence_refs_json,actor,created_at)
VALUES ('00000000-0000-4000-8000-000000000076','00000000-0000-4000-8000-000000000073','mapped','A02:2025',
 '{A02:2025}','manual-v1','[]','Synthetic category',
 '[{"observation_id":"00000000-0000-4000-8000-000000000074","source_pointer":"/site/0/alerts/0/instances/0","note":""}]',
 'synthetic-operator','2026-01-01T00:00:00Z');
INSERT INTO context_revisions(id,target_id,method,path,revision,exposure,required_privileges,data_sensitivity,
 business_criticality,factor_provenance_json,actor,created_at)
VALUES ('00000000-0000-4000-8000-000000000077','00000000-0000-4000-8000-000000000071','GET','/',1,
 'unknown','unknown','unknown','unknown',
 '{"schema_version":"factor-provenance/v1","factors":{"exposure":{"rationale":"Synthetic context","sources":[]},"required_privileges":{"rationale":"Synthetic context","sources":[]},"data_sensitivity":{"rationale":"Synthetic context","sources":[]},"business_criticality":{"rationale":"Synthetic context","sources":[]}}}',
 'synthetic-operator','2026-01-01T00:00:00Z');
