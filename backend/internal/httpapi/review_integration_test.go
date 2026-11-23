package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"diploma/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkVerificationWorkflow(t *testing.T, ctx context.Context, pool *pgxpool.Pool, request func(string, string, string, []byte, bool) *httptest.ResponseRecorder, targetPath string) {
	t.Helper()
	var findings service.Page[service.Finding]
	decodeFixtureJSON(t, request("GET", targetPath+"/findings?limit=2", "", nil, false).Body.Bytes(), &findings)
	if len(findings.Items) != 2 {
		t.Fatal("need two fixture findings")
	}
	first, other := findings.Items[0], findings.Items[1]
	historyPath := "/findings/" + first.ID + "/verifications"
	source := func(id string) service.EvidenceRef {
		t.Helper()
		var observations service.Page[service.Observation]
		decodeFixtureJSON(t, request("GET", "/findings/"+id+"/observations", "", nil, false).Body.Bytes(), &observations)
		if len(observations.Items) == 0 {
			t.Fatal("missing fixture observation")
		}
		return service.EvidenceRef{ObservationID: observations.Items[0].ID, SourcePointer: observations.Items[0].Pointer, Note: "Fixture observation used in a disposable test"}
	}
	ref := source(first.ID)
	post := func(state string, predecessor any, refs []service.EvidenceRef, want int) service.Verification {
		t.Helper()
		input := map[string]any{
			"state": state, "procedure": "Inspect the recorded fixture without contacting the lab",
			"rationale": "Exercise the workflow on disposable test data", "evidence_refs": refs,
			"expected_previous_id": predecessor,
		}
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		w := request("POST", historyPath, "application/json", raw, false)
		if w.Code != want {
			t.Fatalf("verification returned %d, expected %d: %s", w.Code, want, w.Body)
		}
		var review service.Verification
		if want == 201 {
			decodeFixtureJSON(t, w.Body.Bytes(), &review)
			if review.Actor != "fixture-operator" || review.State != state || review.FindingID != first.ID {
				t.Fatal("invalid stored review identity")
			}
		}
		return review
	}
	post("confirmed", nil, []service.EvidenceRef{source(other.ID)}, 422)
	wrongPointer := ref
	wrongPointer.SourcePointer = "/wrong"
	post("confirmed", nil, []service.EvidenceRef{wrongPointer}, 422)
	post("confirmed", nil, nil, 422)
	if w := request("POST", historyPath, "application/json", []byte(`{"actor":"client"}`), false); w.Code != 400 {
		t.Fatal("client-supplied actor accepted")
	}
	if w := request("POST", historyPath, "application/json", []byte(`{"state":"unreviewed"}`), false); w.Code != 409 {
		t.Fatal("missing predecessor accepted")
	}

	confirmed := post("confirmed", nil, []service.EvidenceRef{ref}, 201)
	post("false_positive", nil, []service.EvidenceRef{ref}, 409)
	falsePositive := post("false_positive", confirmed.ID, []service.EvidenceRef{ref}, 201)
	post("inconclusive", confirmed.ID, []service.EvidenceRef{ref}, 409)
	inconclusive := post("inconclusive", falsePositive.ID, []service.EvidenceRef{ref}, 201)
	ids := []string{inconclusive.ID, falsePositive.ID, confirmed.ID}
	query := url.Values{"limit": {"1"}}
	for i, want := range ids {
		var history service.Page[service.Verification]
		w := request("GET", historyPath+"?"+query.Encode(), "", nil, false)
		if w.Code != 200 {
			t.Fatalf("history response returned %d", w.Code)
		}
		decodeFixtureJSON(t, w.Body.Bytes(), &history)
		if len(history.Items) != 1 || history.Items[0].ID != want || history.Items[0].EvidenceRefs[0] != ref {
			t.Fatal("history order or evidence changed")
		}
		if (history.Next != nil) != (i < len(ids)-1) {
			t.Fatal("invalid history pagination")
		}
		if history.Next != nil {
			query.Set("cursor", *history.Next)
		}
	}
	var filtered service.Page[service.Finding]
	decodeFixtureJSON(t, request("GET", targetPath+"/findings?verification=inconclusive", "", nil, false).Body.Bytes(), &filtered)
	if len(filtered.Items) != 1 || filtered.Items[0].ID != first.ID || filtered.Items[0].Severity != first.Severity ||
		filtered.Items[0].ReviewID == nil || *filtered.Items[0].ReviewID != inconclusive.ID {
		t.Fatal("finding summary did not reflect the latest human review independently of severity")
	}
	raw, err := json.Marshal(map[string]any{
		"state": "unreviewed", "procedure": "Retain the test's review history",
		"rationale":     "A concurrent append must compare against the same predecessor",
		"evidence_refs": []service.EvidenceRef{ref}, "expected_previous_id": inconclusive.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan int, 2)
	for range 2 {
		go func() { results <- request("POST", historyPath, "application/json", raw, false).Code }()
	}
	statuses := map[int]int{}
	for range 2 {
		statuses[<-results]++
	}
	if statuses[201] != 1 || statuses[409] != 1 {
		t.Fatalf("concurrent reviews did not serialize: %v", statuses)
	}
	for _, command := range []string{"UPDATE verification_reviews SET state='unreviewed'", "DELETE FROM verification_reviews", "TRUNCATE verification_reviews", "UPDATE findings SET title='changed'"} {
		_, err := pool.Exec(ctx, command)
		var failure *pgconn.PgError
		if !errors.As(err, &failure) || failure.Code != "42501" {
			t.Fatalf("mutation permission expected to fail: %v", err)
		}
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE action='verification.append' AND actor='fixture-operator'").Scan(&count); err != nil || count != 4 {
		t.Fatalf("expected one audit row per accepted review, got %d: %v", count, err)
	}
}
