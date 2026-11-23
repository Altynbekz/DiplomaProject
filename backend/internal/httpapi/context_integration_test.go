package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"diploma/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkContextWorkflow(t *testing.T, ctx context.Context, pool *pgxpool.Pool, request func(string, string, string, []byte, bool) *httptest.ResponseRecorder, targetPath string) {
	t.Helper()
	var findings service.Page[service.Finding]
	decodeFixtureJSON(t, request("GET", targetPath+"/findings", "", nil, false).Body.Bytes(), &findings)
	var root, other service.Finding
	for _, finding := range findings.Items {
		if finding.Method == "GET" && finding.Path == "/" {
			root = finding
		} else if finding.Path != "/" {
			other = finding
		}
	}
	if root.ID == "" || other.ID == "" {
		t.Fatal("genuine fixture needs two exact paths")
	}
	resolve := func(findingID string) service.ResolvedContext {
		t.Helper()
		w := request("GET", "/findings/"+findingID+"/resolved-context", "", nil, false)
		if w.Code != 200 {
			t.Fatalf("resolved context returned %d: %s", w.Code, w.Body)
		}
		var result service.ResolvedContext
		decodeFixtureJSON(t, w.Body.Bytes(), &result)
		if len(result.Factors) != 4 || len(result.FactorProvenance) != 4 || result.ResolverVersion != "exact-endpoint-v1" {
			t.Fatal("resolved context omitted provenance or factors")
		}
		return result
	}
	r := resolve(root.ID)
	if len(r.SourceRevisionIDs) != 0 || r.Factors["exposure"] != "unknown" ||
		r.FactorProvenance["data_sensitivity"].Rationale == "" {
		t.Fatal("unconfigured context should explain unknown values")
	}
	var obs service.Page[service.Observation]
	decodeFixtureJSON(t, request("GET", "/findings/"+root.ID+"/observations", "", nil, false).Body.Bytes(), &obs)
	source := obs.Items[0]
	doc, digest := "synthetic-policy-document", strings.Repeat("1", 64)
	docRef := service.ContextSource{DocumentRef: &doc, SHA256: &digest, Note: "Synthetic test source only"}
	observationRef := service.ContextSource{ObservationID: &source.ID, SourcePointer: &source.Pointer}
	contextBody := func(method, path, previous any, value string, sources []service.ContextSource) []byte {
		t.Helper()
		factors := service.ContextFactors{
			"exposure": value, "required_privileges": "unknown",
			"data_sensitivity": "unknown", "business_criticality": "unknown",
		}
		provenance := service.FactorProvenance{}
		for key := range factors {
			provenance[key] = service.ContextFactorNote{Rationale: "Synthetic rationale for the disposable workflow", Sources: []service.ContextSource{}}
		}
		provenance["exposure"] = service.ContextFactorNote{Rationale: "Fixture policy assumption", Sources: sources}
		data, err := json.Marshal(map[string]any{"method": method, "path": path, "factors": factors, "factor_provenance": provenance, "expected_previous_id": previous})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	post := func(route string, data []byte, want int) service.ContextRevision {
		t.Helper()
		w := request("POST", route+"/contexts", "application/json", data, false)
		if w.Code != want {
			t.Fatalf("context mutation wanted %d, got %d: %s", want, w.Code, w.Body)
		}
		var revision service.ContextRevision
		if want == 201 {
			decodeFixtureJSON(t, w.Body.Bytes(), &revision)
			if revision.Actor != "fixture-operator" || revision.ID == "" {
				t.Fatal("invalid context revision identity")
			}
		}
		return revision
	}
	post(targetPath, contextBody("GET", nil, nil, "internal", []service.ContextSource{docRef}), 422)
	post(targetPath, contextBody("GET", "/?q=1", nil, "internal", []service.ContextSource{docRef}), 422)
	post(targetPath, contextBody("poſt", "/", nil, "internal", []service.ContextSource{docRef}), 422)

	var oversized map[string]any
	decodeFixtureJSON(t, contextBody(nil, nil, nil, "internal", []service.ContextSource{docRef}), &oversized)
	escapedDocument := strings.Repeat("\x01", 128)
	longSources := make([]service.ContextSource, 4)
	for i := range longSources {
		longSources[i] = service.ContextSource{DocumentRef: &escapedDocument, SHA256: &digest, Note: strings.Repeat("\x01", 256)}
	}
	provenance := service.FactorProvenance{}
	for _, key := range []string{"exposure", "required_privileges", "data_sensitivity", "business_criticality"} {
		provenance[key] = service.ContextFactorNote{Rationale: strings.Repeat("\x01", 2048), Sources: longSources}
	}
	oversized["factor_provenance"] = provenance
	oversizedBody, err := json.Marshal(oversized)
	if err != nil {
		t.Fatal(err)
	}
	post(targetPath, oversizedBody, 413)

	post(targetPath, contextBody(nil, nil, nil, "unsupported", []service.ContextSource{docRef}), 422)
	post(targetPath, contextBody(nil, nil, nil, "internal", []service.ContextSource{}), 422)
	badSource := observationRef
	wrong := "/site/0/wrong"
	badSource.SourcePointer = &wrong
	post(targetPath, contextBody(nil, nil, nil, "internal", []service.ContextSource{badSource}), 422)

	foreignID := "00000000-0000-4000-8000-000000000002"
	if _, err := pool.Exec(ctx, "INSERT INTO targets(id,name,canonical_origin,lab_target_key,created_at) VALUES ($1,$2,$3,$4,now())",
		foreignID, "Disposable other target", "http://other-fixture:3000", "juice-shop"); err != nil {
		t.Fatal(err)
	}
	post("/targets/"+foreignID, contextBody(nil, nil, nil, "internal", []service.ContextSource{observationRef}), 422)
	first := post(targetPath, contextBody(nil, nil, nil, "internal", []service.ContextSource{docRef}), 201)
	if first.Revision != 1 || first.Method != nil || first.Path != nil {
		t.Fatal("initial default saved incorrectly")
	}
	r = resolve(root.ID)
	if r.Factors["exposure"] != "internal" || !reflect.DeepEqual(r.SourceRevisionIDs, []string{first.ID}) {
		t.Fatal("target default was not resolved")
	}
	exact := post(targetPath, contextBody("get", "/", nil, "restricted", []service.ContextSource{observationRef}), 201)
	if exact.Method == nil || *exact.Method != "GET" || exact.Revision != 1 {
		t.Fatal("endpoint normalization or revision number differs")
	}
	post(targetPath, contextBody("GET", "/", first.ID, "public", []service.ContextSource{docRef}), 409)
	post(targetPath, contextBody(nil, nil, nil, "public", []service.ContextSource{docRef}), 409)
	newDefault := post(targetPath, contextBody(nil, nil, first.ID, "public", []service.ContextSource{docRef}), 201)
	if newDefault.Revision != 2 || resolve(other.ID).Factors["exposure"] != "public" {
		t.Fatal("target default revision did not advance")
	}
	r = resolve(root.ID)
	if r.Factors["exposure"] != "restricted" || r.Factors["data_sensitivity"] != "unknown" ||
		!reflect.DeepEqual(r.SourceRevisionIDs, []string{exact.ID}) || r.FactorProvenance["exposure"].Sources[0].ObservationID == nil {
		t.Fatal("exact endpoint did not override the newer default with its full snapshot")
	}

	data := contextBody("GET", "/", exact.ID, "internal", []service.ContextSource{observationRef})
	results := make(chan int, 2)
	for range 2 {
		go func() { results <- request("POST", targetPath+"/contexts", "application/json", data, false).Code }()
	}
	codes := map[int]int{}
	for range 2 {
		codes[<-results]++
	}
	if codes[201] != 1 || codes[409] != 1 || resolve(root.ID).Factors["exposure"] != "internal" {
		t.Fatalf("endpoint concurrency was not serialized: %v", codes)
	}
	query := url.Values{"limit": {"1"}}
	byID := map[string]service.ContextRevision{}
	for range 4 {
		var history service.Page[service.ContextRevision]
		w := request("GET", targetPath+"/contexts?"+query.Encode(), "", nil, false)
		if w.Code != 200 {
			t.Fatalf("history returned %d", w.Code)
		}
		decodeFixtureJSON(t, w.Body.Bytes(), &history)
		if len(history.Items) != 1 {
			t.Fatal("context pagination lost history")
		}
		revision := history.Items[0]
		byID[revision.ID] = revision
		if history.Next == nil {
			break
		}
		query.Set("cursor", *history.Next)
	}
	if len(byID) != 4 || byID[first.ID].Factors["exposure"] != "internal" ||
		byID[exact.ID].Factors["exposure"] != "restricted" {
		t.Fatal("historic context factors changed")
	}
	for _, command := range []string{"UPDATE context_revisions SET exposure='unknown'", "DELETE FROM context_revisions", "TRUNCATE context_revisions"} {
		_, err := pool.Exec(ctx, command)
		var failure *pgconn.PgError
		if !errors.As(err, &failure) || failure.Code != "42501" {
			t.Fatalf("context mutation permission expected to fail: %v", err)
		}
	}
	var audits, count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE action='context.append' AND actor='fixture-operator'").Scan(&audits); err != nil || audits != 4 {
		t.Fatalf("unexpected context audit count %d: %v", audits, err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM context_revisions").Scan(&count); err != nil || count != 4 {
		t.Fatalf("rejected operations or reads wrote context rows: %v", err)
	}
	// Vary the bytes so PostgreSQL cannot fit the original long key by compression.
	var long strings.Builder
	long.WriteByte('/')
	for i := range 60 {
		fmt.Fprintf(&long, "%x", sha256.Sum256([]byte(fmt.Sprint(i))))
	}
	saved := post(targetPath, contextBody("GET", long.String(), nil, "internal", []service.ContextSource{docRef}), 201)
	if saved.Path == nil || *saved.Path != long.String() {
		t.Fatal("long path did not round-trip")
	}
	_, err = pool.Exec(ctx, `INSERT INTO context_revisions
		(id,target_id,method,path,revision,exposure,required_privileges,data_sensitivity,business_criticality,factor_provenance_json,actor,created_at)
		SELECT '00000000-0000-4000-8000-000000000003',target_id,method,path,revision,exposure,required_privileges,
		          data_sensitivity,business_criticality,factor_provenance_json,actor,created_at
		FROM context_revisions WHERE id=$1`, saved.ID)
	var duplicate *pgconn.PgError
	if !errors.As(err, &duplicate) || duplicate.Code != "23505" {
		t.Fatalf("long exact scope did not retain database uniqueness: %v", err)
	}
}
