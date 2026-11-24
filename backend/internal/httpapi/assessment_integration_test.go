package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"diploma/internal/cvss"
	"diploma/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkAssessmentWorkflow(t *testing.T, ctx context.Context, pool *pgxpool.Pool, call func(string, string, string, []byte) *httptest.ResponseRecorder, target service.Target, findingID string) {
	t.Helper()
	const decl = "Authored synthetic assessment workflow only"
	path := "/api/v1/findings/" + findingID
	get := func(route string, out any) {
		t.Helper()
		w := call("GET", route, "", nil)
		if w.Code != 200 {
			t.Fatalf("read %s: %d %s", route, w.Code, w.Body)
		}
		decodeFixtureJSON(t, w.Body.Bytes(), out)
	}
	post := func(route string, input any, want int, out any) *httptest.ResponseRecorder {
		t.Helper()
		data, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		w := call("POST", route, "application/json", data)
		if w.Code != want {
			t.Fatalf("append %s: got %d want %d: %s", route, w.Code, want, w.Body)
		}
		if out != nil {
			decodeFixtureJSON(t, w.Body.Bytes(), out)
		}
		return w
	}
	var observationPage service.Page[service.Observation]
	get(path+"/observations", &observationPage)
	obs := observationPage.Items[0]
	ref := service.EvidenceRef{ObservationID: obs.ID, SourcePointer: obs.Pointer, Note: decl}
	var finding service.Finding
	get(path, &finding)
	var reviewID, classID *string
	reviewID, classID = finding.ReviewID, finding.Classification.RevisionID
	var parity struct {
		Cases []struct {
			Name                string
			Base, Environmental map[string]string
			Vector              string
			Score               float64
		}
	}
	raw, err := os.ReadFile("../../testdata/cvss/first-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	decodeFixtureJSON(t, raw, &parity)
	bCase := parity.Cases[2]
	beCase := parity.Cases[4]
	if !reflect.DeepEqual(bCase.Base, beCase.Base) {
		t.Fatal("parity cases no longer share the same Base inputs")
	}
	bReasons := map[string]service.MetricRationale{}
	for name := range bCase.Base {
		bReasons[name] = service.MetricRationale{Reason: decl, EvidenceRefs: []service.EvidenceRef{ref}, ContextFactorRefs: []string{}}
	}
	baseInput := func(finalize bool) map[string]any {
		return map[string]any{
			"kind": "B", "finalize": finalize, "base_assessment_id": nil,
			"expected_verification_review_id": reviewID, "expected_classification_revision_id": classID,
			"expected_context_revision_ids": []string{}, "metrics": bCase.Base, "metric_rationales": bReasons, "notes": decl,
		}
	}
	save := func(input map[string]any) service.AssessmentSnapshot {
		t.Helper()
		var snapshot service.AssessmentSnapshot
		w := post(path+"/assessments", input, 201, &snapshot)
		if w.Header().Get("Location") != "/api/v1/assessments/"+snapshot.ID || snapshot.Actor != "test-operator" {
			t.Fatal("missing snapshot identity or authenticated actor")
		}
		return snapshot
	}
	assertProblem := func(input map[string]any, status int, pointer, code string) {
		t.Helper()
		var failure struct {
			Code   string
			Errors []struct{ Pointer string }
		}
		post(path+"/assessments", input, status, &failure)
		if failure.Code != code || len(failure.Errors) == 0 || failure.Errors[0].Pointer != pointer {
			t.Fatalf("unexpected error: %+v", failure)
		}
	}
	assertScore := func(snapshot service.AssessmentSnapshot, kind, vector string, score float64) {
		t.Helper()
		if snapshot.Kind != kind || snapshot.Status != "complete" || snapshot.Score == nil || *snapshot.Score != score ||
			snapshot.Vector == nil || *snapshot.Vector != vector || snapshot.Calculator == nil || *snapshot.Calculator != cvss.Calculator() {
			t.Fatalf("snapshot disagrees with independent reference: %+v", snapshot)
		}
	}
	assertProblem(baseInput(true), 422, "/expected_verification_review_id", "incomplete_assessment")
	draftInput := baseInput(false)
	draftInput["metrics"] = map[string]string{"AV": "N"}
	draftInput["metric_rationales"] = map[string]service.MetricRationale{
		"AV": {EvidenceRefs: []service.EvidenceRef{}, ContextFactorRefs: []string{}},
	}
	draft := save(draftInput)
	if draft.Status != "incomplete" || draft.Score != nil || draft.Vector != nil || draft.Calculator != nil || draft.Context.Factors["exposure"] != "unknown" {
		t.Fatal("draft must stay incomplete with an honest context snapshot")
	}
	invalid := maps.Clone(draftInput)
	invalid["score"] = 10
	post(path+"/assessments", invalid, 400, nil)
	raw, err = json.Marshal(draftInput)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := strings.Replace(string(raw), `"metrics":{"AV":"N"}`, `"metrics":{"AV":"N","AV":"L"}`, 1)
	if duplicate == string(raw) || call("POST", path+"/assessments", "application/json", []byte(duplicate)).Code != 400 {
		t.Fatal("duplicate metric was not rejected")
	}

	reviewInput := func(state string) map[string]any {
		return map[string]any{"state": state, "procedure": decl, "rationale": decl, "evidence_refs": []service.EvidenceRef{ref}, "expected_previous_id": reviewID}
	}
	var review service.Verification
	post(path+"/verifications", reviewInput("confirmed"), 201, &review)
	reviewID = &review.ID
	var classification service.ClassificationRevision
	post(path+"/classifications", map[string]any{
		"status": "mapped", "category": "A02:2025", "candidate_categories": []string{"A02:2025"},
		"rationale": decl, "evidence_refs": []service.EvidenceRef{ref}, "expected_previous_id": classID,
	}, 201, &classification)
	classID = &classification.ID
	bad := baseInput(true)
	badReasons := maps.Clone(bReasons)
	badReasons["AV"] = service.MetricRationale{Reason: decl, EvidenceRefs: []service.EvidenceRef{}, ContextFactorRefs: []string{}}
	bad["metric_rationales"] = badReasons
	assertProblem(bad, 422, "/metric_rationales/AV", "incomplete_assessment")
	badReasons["AV"] = service.MetricRationale{Reason: decl, EvidenceRefs: []service.EvidenceRef{{ObservationID: obs.ID, SourcePointer: "/wrong"}}, ContextFactorRefs: []string{}}
	assertProblem(bad, 422, "/metric_rationales/AV/evidence_refs/0/source_pointer", "validation_failed")
	stale := baseInput(true)
	stale["expected_verification_review_id"] = nil
	assertProblem(stale, 409, "/expected_verification_review_id", "conflict")
	stale = baseInput(true)
	stale["expected_classification_revision_id"] = nil
	assertProblem(stale, 409, "/expected_classification_revision_id", "conflict")
	base := save(baseInput(true))
	assertScore(base, "B", bCase.Vector, bCase.Score)

	factors := map[string]string{"exposure": "restricted", "required_privileges": "user", "data_sensitivity": "internal", "business_criticality": "low"}
	notes := map[string]service.ContextFactorNote{}
	dispositions := map[string]service.FactorDisposition{}
	for factor := range factors {
		doc, digest := "synthetic-context-policy", strings.Repeat("1", 64)
		notes[factor] = service.ContextFactorNote{Rationale: decl, Sources: []service.ContextSource{{DocumentRef: &doc, SHA256: &digest}}}
		dispositions[factor] = service.FactorDisposition{Reason: decl, Metrics: []string{}}
	}
	dispositions["data_sensitivity"] = service.FactorDisposition{Reason: decl, Metrics: []string{"CR"}}
	beReasons := map[string]service.MetricRationale{}
	for name := range beCase.Environmental {
		beReasons[name] = service.MetricRationale{Reason: decl, EvidenceRefs: []service.EvidenceRef{ref}, ContextFactorRefs: []string{}}
	}
	beReasons["CR"] = service.MetricRationale{Reason: decl, EvidenceRefs: []service.EvidenceRef{}, ContextFactorRefs: []string{"data_sensitivity"}}
	var contextIDs = []string{}
	beInput := func() map[string]any {
		in := baseInput(true)
		in["kind"], in["base_assessment_id"], in["expected_context_revision_ids"] = "BE", base.ID, contextIDs
		in["metrics"], in["metric_rationales"], in["factor_dispositions"] = beCase.Environmental, beReasons, dispositions
		return in
	}
	assertProblem(beInput(), 422, "/context_snapshot/factors/exposure", "incomplete_assessment")
	var contextRevision service.ContextRevision
	contextRoute := "/api/v1/targets/" + target.ID + "/contexts"
	contextPayload := map[string]any{"method": nil, "path": nil, "factors": factors, "factor_provenance": notes, "expected_previous_id": nil}
	post(contextRoute, contextPayload, 201, &contextRevision)
	assertProblem(beInput(), 409, "/expected_context_revision_ids", "conflict")
	contextIDs = []string{contextRevision.ID}
	missing := beInput()
	delete(missing, "factor_dispositions")
	assertProblem(missing, 422, "/factor_dispositions/exposure", "incomplete_assessment")
	wrongParent := beInput()
	wrongParent["base_assessment_id"] = draft.ID
	assertProblem(wrongParent, 422, "/base_assessment_id", "validation_failed")
	be := save(beInput())
	assertScore(be, "BE", beCase.Vector, beCase.Score)
	if be.BaseAssessmentID == nil || *be.BaseAssessmentID != base.ID ||
		be.Metrics["MAV"] != "X" || !reflect.DeepEqual(be.Context.SourceRevisionIDs, contextIDs) {
		t.Fatal("BE did not retain the paired Base, inherited choice and context")
	}
	for name, value := range base.Metrics {
		if be.Metrics[name] != value || !reflect.DeepEqual(be.MetricRationales[name], base.MetricRationales[name]) {
			t.Fatalf("BE changed inherited metric or rationale %s", name)
		}
	}

	latest := save(baseInput(false))
	get(path, &finding)
	if finding.Base == nil || finding.Base.ID != latest.ID || finding.Base.Score != nil ||
		finding.Environmental == nil || finding.Environmental.ID != be.ID {
		t.Fatal("finding must surface latest incomplete and complete assessment summaries separately")
	}
	var page service.Page[service.AssessmentSnapshot]
	get(path+"/assessments?limit=1", &page)
	if len(page.Items) != 1 || page.Items[0].ID != latest.ID || page.Next == nil {
		t.Fatal("assessment history did not page newest first")
	}
	get(path+"/assessments?limit=1&cursor="+*page.Next, &page)
	if len(page.Items) != 1 || page.Items[0].ID != be.ID {
		t.Fatal("assessment cursor skipped a snapshot")
	}

	contextPayload["expected_previous_id"] = contextRevision.ID
	factors["exposure"] = "internal"
	post(contextRoute, contextPayload, 201, &contextRevision)
	assertProblem(beInput(), 409, "/expected_context_revision_ids", "conflict")
	contextIDs = []string{contextRevision.ID}
	var historic service.AssessmentSnapshot
	get("/api/v1/assessments/"+be.ID, &historic)
	if !historic.Stale {
		t.Fatal("BE should become stale after context replacement")
	}
	historic.Stale = false
	if !reflect.DeepEqual(historic, be) {
		t.Fatal("context changes altered the immutable snapshot")
	}
	var latestBase service.AssessmentSnapshot
	get("/api/v1/assessments/"+base.ID, &latestBase)
	if latestBase.Stale || !reflect.DeepEqual(latestBase, base) {
		t.Fatal("context-only change altered or invalidated B")
	}
	replacement := save(baseInput(true))
	assertProblem(beInput(), 409, "/base_assessment_id", "conflict")
	beFromReplacement := beInput()
	beFromReplacement["base_assessment_id"] = replacement.ID
	newBE := save(beFromReplacement)
	assertScore(newBE, "BE", beCase.Vector, beCase.Score)
	post(path+"/verifications", reviewInput("inconclusive"), 201, &review)
	get("/api/v1/assessments/"+replacement.ID, &historic)
	if !historic.Stale || historic.Score == nil || *historic.Score != bCase.Score {
		t.Fatal("review change should mark B stale without changing its score")
	}

	for _, statement := range []string{
		"UPDATE assessment_snapshots SET notes='changed'",
		"DELETE FROM assessment_snapshots",
		"TRUNCATE assessment_snapshots",
	} {
		_, err := pool.Exec(ctx, statement)
		assessmentSQLState(t, err, "42501")
	}
	for _, column := range []string{"finding_id", "verification_review_id", "classification_revision_id", "context_revision_id", "base_assessment_id"} {
		args := []any{newBE.ID}
		if column == "base_assessment_id" {
			args = append(args, draft.ID)
		}
		_, err := pool.Exec(ctx, `INSERT INTO assessment_snapshots
		(id,finding_id,target_id,kind,status,base_assessment_id,verification_review_id,classification_revision_id,
		context_revision_id,context_snapshot_json,metrics_json,metric_rationales_json,factor_dispositions_json,
		notes,vector,score,calculator_json,schema_version,policy_version,actor,created_at)
		SELECT gen_random_uuid(),`+replacementColumns(column)+`,
		context_snapshot_json,metrics_json,metric_rationales_json,factor_dispositions_json,notes,vector,score,calculator_json,
		schema_version,policy_version,actor,created_at FROM assessment_snapshots WHERE id=$1`, args...)
		assessmentSQLState(t, err, "23503")
	}

	owner, err := pgxpool.New(ctx, os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if _, err := owner.Exec(ctx, "ALTER TABLE audit_events ADD CONSTRAINT test_assessment_audit_failure CHECK (action<>'assessment.append') NOT VALID"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := owner.Exec(ctx, "ALTER TABLE audit_events DROP CONSTRAINT test_assessment_audit_failure"); err != nil {
			t.Error(err)
		}
	}()
	reviewID = &review.ID
	post(path+"/assessments", baseInput(false), 500, nil)
	var stored, audit int
	if err := pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM assessment_snapshots WHERE finding_id=$1),
		(SELECT count(*) FROM audit_events WHERE action='assessment.append')`, findingID).Scan(&stored, &audit); err != nil || stored != 6 || audit != stored {
		t.Fatalf("failed append must leave only six snapshots and matching audits: %d %d %v", stored, audit, err)
	}
}

func assessmentSQLState(t *testing.T, err error, want string) {
	t.Helper()
	var dbError *pgconn.PgError
	if !errors.As(err, &dbError) || dbError.Code != want {
		t.Fatalf("expected SQLSTATE %s: %v", want, err)
	}
}

// The column name comes from the fixed constraint cases above, never a request.
func replacementColumns(replace string) string {
	cols := []string{"finding_id", "target_id", "kind", "status", "base_assessment_id", "verification_review_id", "classification_revision_id", "context_revision_id"}
	for i, col := range cols {
		if col == replace {
			cols[i] = "gen_random_uuid()"
			if col == "base_assessment_id" {
				cols[i] = "$2::uuid"
			}
		}
	}
	return strings.Join(cols, ",")
}
