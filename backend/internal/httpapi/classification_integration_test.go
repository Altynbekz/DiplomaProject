package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"diploma/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkClassificationWorkflow(t *testing.T, ctx context.Context, pool *pgxpool.Pool, request func(string, string, string, []byte, bool) *httptest.ResponseRecorder, targetPath string, raw []byte) {
	t.Helper()
	findings := func(filter string) []service.Finding {
		t.Helper()
		w := request("GET", targetPath+"/findings"+filter, "", nil, false)
		if w.Code != 200 {
			t.Fatalf("classification filter returned %d", w.Code)
		}
		var page service.Page[service.Finding]
		decodeFixtureJSON(t, w.Body.Bytes(), &page)
		return page.Items
	}
	one := findings("?category=A01%3A2025")
	two := findings("?category=A02%3A2025")
	if len(one) != 5 || len(two) != 1 || len(findings("?category=A05%3A2025")) != 0 {
		t.Fatal("reviewed rules gave unexpected classifications for the genuine fixture")
	}
	categories := map[string]int{}
	for _, finding := range findings("") {
		categories[finding.Classification.Status]++
	}
	if categories["mapped"] != 6 || categories["out_of_scope"] != 1 || categories["unmapped"] != 10 {
		t.Fatalf("unexpected classification states: %v", categories)
	}
	first := one[0]
	path := "/findings/" + first.ID + "/classifications"
	getHistory := func() []service.ClassificationRevision {
		t.Helper()
		var page service.Page[service.ClassificationRevision]
		w := request("GET", path, "", nil, false)
		if w.Code != 200 {
			t.Fatalf("classification history returned %d", w.Code)
		}
		decodeFixtureJSON(t, w.Body.Bytes(), &page)
		return page.Items
	}
	initial := getHistory()
	if len(initial) != 1 || initial[0].Version != "owasp-2025-v1" || len(initial[0].Rules) != 1 ||
		len(initial[0].EvidenceRefs) != 1 || len(initial[0].Rules[0].Sources) < 2 || initial[0].Rules[0].ReviewRecord == "" || len(initial[0].Rules[0].PolicySHA256) != 64 {
		t.Fatal("missing original rule and source history")
	}
	upload := func(content []byte) {
		t.Helper()
		var b bytes.Buffer
		form := multipart.NewWriter(&b)
		part, err := form.CreateFormFile("report", "synthetic-classification-exercise.json")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := form.Close(); err != nil {
			t.Fatal(err)
		}
		if w := request("POST", targetPath+"/imports/zap", form.FormDataContentType(), b.Bytes(), false); w.Code != 201 {
			t.Fatalf("test-only duplicate layout import returned %d: %s", w.Code, w.Body)
		}
	}
	// A contradictory authored scanner assertion on the same finding tests the
	// rule conflict without asserting any new observed lab result.
	conflict := bytes.Replace(raw, []byte(`"cweid": "264"`), []byte(`"cweid": "89"`), 1)
	if bytes.Equal(raw, conflict) {
		t.Fatal("the fixture no longer contains the expected CORS assertion")
	}
	upload(conflict)
	history := getHistory()
	if len(history) != 2 || history[0].Status != "ambiguous" || history[0].Category != nil ||
		len(history[0].Candidates) != 2 || history[0].Candidates[0] != "A01:2025" || history[0].Candidates[1] != "A05:2025" ||
		len(history[0].EvidenceRefs) != 2 || history[1].ID != initial[0].ID {
		t.Fatal("conflicting source assertions did not leave an ambiguous append-only history")
	}
	if len(findings("?category=A01%3A2025")) != 0 || len(findings("?category=A05%3A2025")) != 0 {
		t.Fatal("ambiguous findings entered a mapped category")
	}
	requestBody := func(predecessor string, refs []service.EvidenceRef) []byte {
		t.Helper()
		data, err := json.Marshal(map[string]any{
			"status": "out_of_scope", "category": nil, "candidate_categories": []string{},
			"rationale":     "Exercise an explicit classification revision on disposable test data",
			"evidence_refs": refs, "expected_previous_id": predecessor,
		})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	if w := request("POST", path, "application/json", requestBody(initial[0].ID, history[0].EvidenceRefs), false); w.Code != 409 {
		t.Fatal("stale classification predecessor accepted")
	}
	var other service.Page[service.Observation]
	w := request("GET", "/findings/"+two[0].ID+"/observations", "", nil, false)
	decodeFixtureJSON(t, w.Body.Bytes(), &other)
	wrong := service.EvidenceRef{ObservationID: other.Items[0].ID, SourcePointer: other.Items[0].Pointer}
	if w := request("POST", path, "application/json", requestBody(history[0].ID, []service.EvidenceRef{wrong}), false); w.Code != 422 {
		t.Fatal("cross-finding classification reference accepted")
	}
	missingCategory := bytes.Replace(requestBody(history[0].ID, history[0].EvidenceRefs), []byte(`"category":null,`), nil, 1)
	if w := request("POST", path, "application/json", missingCategory, false); w.Code != 400 {
		t.Fatalf("missing required category returned %d", w.Code)
	}
	w = request("POST", path, "application/json", requestBody(history[0].ID, history[0].EvidenceRefs), false)
	if w.Code != 201 {
		t.Fatalf("manual classification returned %d", w.Code)
	}
	var manual service.ClassificationRevision
	decodeFixtureJSON(t, w.Body.Bytes(), &manual)
	if manual.Actor != "fixture-operator" || manual.Version != "manual-v1" || manual.Status != "out_of_scope" {
		t.Fatal("manual classification has incorrect provenance")
	}
	upload(append(bytes.Clone(raw), '\n'))
	history = getHistory()
	if len(history) != 3 || history[0].ID != manual.ID {
		t.Fatal("a later import superseded the manual classification")
	}
	var detail service.Finding
	decodeFixtureJSON(t, request("GET", "/findings/"+first.ID, "", nil, false).Body.Bytes(), &detail)
	if detail.Severity != first.Severity || detail.Verification != first.Verification || detail.Classification.Status != "out_of_scope" {
		t.Fatal("verification or severity changed with classification")
	}
	var audits int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE action='classification.append' AND entity_id=$1 AND actor='fixture-operator'", manual.ID).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("manual classification audit missing: %v", err)
	}
	for _, statement := range []string{
		"UPDATE classification_revisions SET rationale='modified'",
		"DELETE FROM classification_revisions",
		"TRUNCATE classification_revisions",
	} {
		_, err := pool.Exec(ctx, statement)
		var failure *pgconn.PgError
		if !errors.As(err, &failure) || failure.Code != "42501" {
			t.Fatalf("classification runtime mutation permission expected to fail: %v", err)
		}
	}
	body := requestBody(manual.ID, manual.EvidenceRefs)
	results := make(chan int, 2)
	for range 2 {
		go func() {
			results <- request("POST", path, "application/json", body, false).Code
		}()
	}
	counts := map[int]int{}
	for range 2 {
		counts[<-results]++
	}
	if counts[201] != 1 || counts[409] != 1 {
		t.Fatalf("concurrent classification revisions did not serialize: %v", counts)
	}
	history = getHistory()
	if len(history) != 4 || history[0].ID == manual.ID || history[1].ID != manual.ID {
		t.Fatal("concurrent classification revisions changed prior history or appended twice")
	}
}
