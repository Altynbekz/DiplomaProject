package httpapi

import (
	"bytes"
	"context"
	"diploma/internal/service"
	"diploma/internal/store"
	"diploma/internal/zap"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const testToken = "synthetic-unit-test-token-not-a-deployment-secret"

func TestAuthentication(t *testing.T) {
	a := API{Token: testToken, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, path := range []string{"/api/v1/targets", "/api/v1/scans", "/unknown"} {
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 {
			t.Fatal(w.Code)
		}
	}
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/healthz", nil))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
func TestStrictMutations(t *testing.T) {
	for _, raw := range []string{`{"name":"a","name":"b"}`, `{"name":"a","unknown":1}`, `{"name":"a","Name":"b"}`, `{"Name":"a"}`, `null`, `{} {}`, `{"name":12}`} {
		var v struct {
			Name string `json:"name"`
		}
		if e := strict([]byte(raw), &v); e == nil {
			t.Fatal("accepted", raw)
		}
	}
}
func TestIntegrationImport(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL required for disposable PostgreSQL integration")
	}
	ctx := context.Background()
	pool, e := pgxpool.New(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	if _, e = pool.Exec(ctx, "SET ROLE diploma_runtime"); e != nil {
		t.Fatal(e)
	} // use per-connection role below as well
	pool.Close()
	cfg, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["role"] = "diploma_runtime"
	pool, e = pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	var logs bytes.Buffer
	svc := &service.Service{Store: store.New(pool), Origin: "http://juice-shop:3000", Actor: "test-operator", CursorKey: testToken, Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	a := API{Service: svc, Token: testToken, Log: slog.New(slog.NewTextHandler(&logs, nil))}
	h := a.Handler()
	call := func(method, path, ct string, b []byte) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewReader(b))
		r.Header.Set("Authorization", "Bearer "+testToken)
		if ct != "" {
			r.Header.Set("Content-Type", ct)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	w := call("POST", "/api/v1/targets", "application/json", []byte(`{"name":"Synthetic integration target","lab_target_key":"juice-shop"}`))
	var target service.Target
	if w.Code != 201 {
		t.Fatalf("target %d %s", w.Code, w.Body)
	}
	if json.Unmarshal(w.Body.Bytes(), &target) != nil {
		t.Fatal(w.Body)
	}
	raw, e := os.ReadFile("../../testdata/zap/synthetic-traditional.json")
	if e != nil {
		t.Fatal(e)
	}
	normalized, e := zap.Parse(raw, target.ID, target.Origin)
	if e != nil {
		t.Fatal(e)
	}
	fingerprints := map[string]bool{}
	for _, o := range normalized.Observations {
		fingerprints[o.Fingerprint] = true
	}
	upload := func(data []byte, provenance ...string) *httptest.ResponseRecorder {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		part, e := mw.CreateFormFile("report", "synthetic.json")
		if e != nil {
			t.Error(e)
		}
		_, _ = part.Write(data)
		if len(provenance) > 0 {
			if err := mw.WriteField("provenance", provenance[0]); err != nil {
				t.Error(err)
			}
		}

		_ = mw.Close()
		return call("POST", "/api/v1/targets/"+target.ID+"/imports/zap", mw.FormDataContentType(), b.Bytes())
	}
	// Transport rejects duplicate fields, wrong MIME, byte limits, and malformed queries before mutation.
	var duplicateBody bytes.Buffer
	mw := multipart.NewWriter(&duplicateBody)
	for range 2 {
		part, err := mw.CreateFormFile("report", "synthetic.json")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = part.Write(raw); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	invalidTransport := call("POST", "/api/v1/targets/"+target.ID+"/imports/zap", mw.FormDataContentType(), duplicateBody.Bytes())
	if invalidTransport.Code != 400 {
		t.Fatal("duplicate report fields", invalidTransport.Code)
	}
	invalidTransport = call("POST", "/api/v1/targets/"+target.ID+"/imports/zap", "application/json", raw)
	if invalidTransport.Code != 415 {
		t.Fatal("wrong media type", invalidTransport.Code)
	}
	oversized := httptest.NewRequest("POST", "/api/v1/targets/"+target.ID+"/imports/zap", strings.NewReader(""))
	oversized.ContentLength = zap.MaxBytes + 1
	oversized.Header.Set("Authorization", "Bearer "+testToken)
	oversized.Header.Set("Content-Type", "multipart/form-data; boundary=test")
	out := httptest.NewRecorder()
	h.ServeHTTP(out, oversized)
	if out.Code != 413 {
		t.Fatal("oversize request", out.Code)
	}
	for _, query := range []string{"limit=0", "limit=201", "limit=1&limit=2", "unexpected=1", "cursor=bad", "scanner_severity=critical", "import_id=not-uuid", "limit=%zz"} {
		out := call("GET", "/api/v1/targets/"+target.ID+"/findings?"+query, "", nil)
		if out.Code != 400 {
			t.Fatal("invalid query", query, out.Code, out.Body)
		}
	}
	// Exercise malformed imports before any valid insert.
	const uri = "http://juice-shop:3000/rest/products/search?q=secret1"
	for _, tc := range []struct {
		name, raw, code, pointer string
		status                   int
	}{
		{"malformed", "{", "invalid_request", "", 400},
		{"envelope", `{"alerts":[]}`, "unsupported_report", "", 422},
		{"trailing", string(raw) + " {}", "invalid_request", "", 400},
		{"other origin", strings.Replace(string(raw), uri, "http://other:3000/", 1), "out_of_scope", "/site/0/alerts/0/instances/0/uri", 422},
		{"other site origin", strings.Replace(string(raw), `"@name": "http://juice-shop:3000"`, `"@name": "http://other:3000"`, 1), "out_of_scope", "/site/0/@name", 422},
		{"field limit", `{"@version":` + strings.Repeat("1", 65537) + `}`, "limit_exceeded", "/@version", 413},
		{"depth limit", strings.Repeat("[", 66) + strings.Repeat("]", 66), "limit_exceeded", "", 413},
		{"site count limit", `{"site":[` + strings.Repeat("{},", 10) + `{}]}`, "limit_exceeded", "/site", 413},
		{"NUL query name", strings.Replace(string(raw), uri, "http://juice-shop:3000/rest/products/search?%00=x", 1), "validation_failed", "/site/0/alerts/0/instances/0/uri", 422},
		{"invalid UTF-8 query name", strings.Replace(string(raw), uri, "http://juice-shop:3000/rest/products/search?%FF=x", 1), "validation_failed", "/site/0/alerts/0/instances/0/uri", 422},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := upload([]byte(tc.raw))
			var problem struct {
				Code   string
				Errors []struct{ Pointer string }
			}
			if w.Code != tc.status || json.Unmarshal(w.Body.Bytes(), &problem) != nil || problem.Code != tc.code {
				t.Fatalf("expected %d/%s, got %d: %s", tc.status, tc.code, w.Code, w.Body)
			}
			if tc.pointer != "" && (len(problem.Errors) == 0 || problem.Errors[0].Pointer != tc.pointer) {
				t.Fatalf("expected pointer %s, got %s", tc.pointer, w.Body)
			}
			var imports, findings, observations int
			err := pool.QueryRow(ctx, `SELECT
				(SELECT count(id) FROM report_imports WHERE target_id=$1),
				(SELECT count(id) FROM findings WHERE target_id=$1),
				(SELECT count(id) FROM observations WHERE target_id=$1)`, target.ID).Scan(&imports, &findings, &observations)
			if err != nil || imports != 0 || findings != 0 || observations != 0 {
				t.Fatalf("expected no partial rows, got %d imports, %d findings, %d observations: %v", imports, findings, observations, err)
			}
		})
	}
	// JSON escapes can carry a NUL even though the uploaded bytes are valid UTF-8.
	const escapedNUL = `\u0000`
	for _, tc := range []struct{ name, report, provenance, pointer string }{
		{"NUL in finding title", strings.Replace(string(raw), "Synthetic <script>", "Synthetic "+escapedNUL+"<script>", 1), "", "/site/0/alerts/0/name"},
		{"NUL in parameter identity", strings.Replace(string(raw), `"param": "q"`, `"param": "q`+escapedNUL+`"`, 1), "", "/site/0/alerts/0/instances/0/param"},
		{"NUL in private evidence", strings.Replace(string(raw), "synthetic attack A", "synthetic "+escapedNUL+" attack A", 1), "", "/site/0/alerts/0/instances/0/attack"},
		{"NUL in unlocated JSONB", strings.Replace(string(raw), "Synthetic unlocated alert", "Synthetic "+escapedNUL+" unlocated alert", 1), "", "/site/0/alerts/2/name"},
		{"NUL in provenance value", string(raw), `{"authorization_ref":"reference` + escapedNUL + `"}`, "/provenance/authorization_ref"},
		{"NUL in provenance key", string(raw), `{"addon_versions":{"addon` + escapedNUL + `":"1"}}`, "/provenance/addon_versions"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response *httptest.ResponseRecorder
			if tc.provenance != "" {
				response = upload([]byte(tc.report), tc.provenance)
			} else {
				response = upload([]byte(tc.report))
			}
			var problem struct {
				Code   string
				Errors []struct{ Pointer string }
			}
			if response.Code != 422 || json.Unmarshal(response.Body.Bytes(), &problem) != nil || problem.Code != "validation_failed" ||
				len(problem.Errors) != 1 || problem.Errors[0].Pointer != tc.pointer {
				t.Errorf("wanted 422 validation_failed at %s, got %d %s", tc.pointer, response.Code, response.Body)
			}
			var imports, findings, observations int
			err := pool.QueryRow(ctx, `SELECT
    (SELECT count(id) FROM report_imports WHERE target_id=$1),
    (SELECT count(id) FROM findings WHERE target_id=$1),
    (SELECT count(id) FROM observations WHERE target_id=$1)`, target.ID).Scan(&imports, &findings, &observations)
			if err != nil || imports != 0 || findings != 0 || observations != 0 {
				t.Fatalf("partial import: counts %d/%d/%d: %v", imports, findings, observations, err)
			}
		})
	}

	t.Run("normalized provenance size limit", func(t *testing.T) {
		value := ""
		meta := service.Provenance{SchemaVersion: "report-provenance/v2", AuthorizationRef: &value, SanitizationSteps: []string{}}
		base, err := json.Marshal(meta)
		if err != nil {
			t.Fatal(err)
		}
		value = strings.Repeat("a", 65536-len(base))
		compact, err := json.Marshal(meta)
		if err != nil || len(compact) != 65536 {
			t.Fatal("test must exercise the exact compact size limit", len(compact), err)
		}
		response := upload(raw, `{"authorization_ref":"`+value+`"}`)
		if response.Code != 413 || !strings.Contains(response.Body.String(), `"code":"limit_exceeded"`) {
			t.Errorf("expected documented provenance size limit, got %d %s", response.Code, response.Body)
		}
		var count int
		if err := pool.QueryRow(ctx, "SELECT count(id) FROM report_imports WHERE target_id=$1", target.ID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("oversized provenance left partial import: %d %v", count, err)
		}
	})

	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 4)
	for range 4 {
		wg.Add(1)
		go func() { defer wg.Done(); results <- upload(raw) }()
	}
	wg.Wait()
	close(results)
	newCount := 0
	importID := ""
	for w := range results {
		if w.Code != 200 && w.Code != 201 {
			t.Fatalf("upload %d %s", w.Code, w.Body)
		}
		if w.Code == 201 {
			newCount++
		}
		var im service.Import
		if e = json.Unmarshal(w.Body.Bytes(), &im); e != nil {
			t.Fatal(e)
		}
		if importID != "" && importID != im.ID {
			t.Fatal("duplicate ID mismatch")
		}
		importID = im.ID
		if int(im.FindingCount) != len(fingerprints) || int(im.ObservationCount) != len(normalized.Observations) || im.UnlocatedCount != len(normalized.Unlocated) {
			t.Fatalf("count mismatch %+v", im)
		}
	}
	if newCount != 1 {
		t.Fatal("normalizations", newCount)
	}
	w = upload(raw, `{"authorization_ref":"synthetic-new-authorization"}`)
	if w.Code != 409 || !strings.Contains(w.Body.String(), `"code":"conflict"`) {
		t.Fatal("expected provenance conflict", w.Code, w.Body)
	}
	var stored service.Import
	w = call("GET", "/api/v1/imports/"+importID, "", nil)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &stored) != nil || stored.ID != importID ||
		stored.Provenance.ZAPVersion != nil || stored.Provenance.ReportTemplate != nil ||
		stored.Provenance.AuthorizationRef != nil || stored.ReportDeclaredVersion == nil ||
		*stored.ReportDeclaredVersion != "SYNTHETIC-TEST-ONLY" || stored.ParserDetectedFormat != "traditional-json" {
		t.Fatal("operator provenance must stay null and report claims remain separate", w.Code, w.Body)
	}

	var count int
	if e = pool.QueryRow(ctx, "SELECT count(id) FROM observations WHERE report_import_id=$1", importID).Scan(&count); e != nil || count != len(normalized.Observations) {
		t.Fatal(count, e)
	}
	var saved []byte
	if e = pool.QueryRow(ctx, "SELECT raw_bytes FROM report_imports WHERE id=$1", importID).Scan(&saved); e != nil || !bytes.Equal(saved, raw) {
		t.Fatal("original bytes not preserved", e)
	}
	seen := map[string]bool{}
	cursor := ""
	findingID := ""
	groupedFindingID := ""
	for {
		w = call("GET", "/api/v1/targets/"+target.ID+"/findings?limit=1&cursor="+cursor, "", nil)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body)
		}
		var page service.Page[service.Finding]
		if e = json.Unmarshal(w.Body.Bytes(), &page); e != nil {
			t.Fatal(e)
		}
		for _, f := range page.Items {
			if seen[f.ID] {
				t.Fatal("repeated page")
			}
			seen[f.ID] = true
			findingID = f.ID
			if f.ObservationCount == 2 {
				groupedFindingID = f.ID
			}
			if f.Base != nil || f.Environmental != nil || f.Verification != "unreviewed" || f.Classification.Status != "unmapped" {
				t.Fatal("invented assessment")
			}
		}
		if page.Next == nil {
			break
		}
		cursor = *page.Next
	}
	if len(seen) != len(fingerprints) {
		t.Fatal("pagination lost findings")
	}
	w = call("GET", "/api/v1/findings/"+findingID+"/observations", "", nil)
	if w.Code != 200 || strings.Contains(w.Body.String(), "synthetic-private-value") || strings.Contains(w.Body.String(), "raw_uri") {
		t.Fatal("unsafe observations", w.Code, w.Body)
	}
	if groupedFindingID == "" {
		t.Fatal("grouped finding missing")
	}
	observationPath := "/api/v1/findings/" + groupedFindingID + "/observations?limit=1"
	var page1, page2 service.Page[service.Observation]
	w = call("GET", observationPath, "", nil)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page1) != nil || len(page1.Items) != 1 || page1.Next == nil {
		t.Fatal("first observation page", w.Code, w.Body)
	}
	w = call("GET", observationPath+"&cursor="+*page1.Next, "", nil)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page2) != nil || len(page2.Items) != 1 || page2.Next != nil {
		t.Fatal("second observation page", w.Code, w.Body)
	}
	if page1.Items[0].ID == page2.Items[0].ID || page1.Items[0].Pointer == page2.Items[0].Pointer {
		t.Fatal("repeated observation")
	}
	for _, q := range []string{"category=A05:2025", "verification=confirmed", "scanner_severity=low"} {
		w = call("GET", "/api/v1/targets/"+target.ID+"/findings?"+q, "", nil)
		var page service.Page[service.Finding]
		_ = json.Unmarshal(w.Body.Bytes(), &page)
		if w.Code != 200 || len(page.Items) != 0 {
			t.Fatal("filter", q, w.Code, w.Body)
		}
	}
	if cursor != "" {
		w = call("GET", "/api/v1/targets/"+target.ID+"/findings?category=A05:2025&cursor="+cursor, "", nil)
		if w.Code != 400 {
			t.Fatal("mismatched cursor", w.Code)
		}
	}
	w = call("GET", "/api/v1/imports/"+importID+"/raw", "", nil)
	if w.Code != 422 {
		t.Fatal(w.Code)
	}
	r := httptest.NewRequest("GET", "/api/v1/imports/"+importID+"/raw", nil)
	r.Header.Set("Authorization", "Bearer "+testToken)
	r.Header.Set("X-Evidence-Acknowledgement", "private-lab-data")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), raw) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("raw mismatch")
	}
	for _, table := range []string{"report_imports", "observations", "findings", "audit_events"} {
		_, e = pool.Exec(ctx, "DELETE FROM "+table+" WHERE false")
		if e == nil {
			t.Fatal("runtime delete allowed", table)
		}
	}
	for _, table := range []string{"report_imports", "observations", "findings", "audit_events"} {
		_, err := pool.Exec(ctx, "UPDATE "+table+" SET created_at=created_at WHERE false")
		if err == nil {
			t.Fatal("runtime update allowed", table)
		}
	}
	var auditCount int
	if err := pool.QueryRow(ctx, "SELECT count(id) FROM audit_events WHERE entity_id=$1 AND action='evidence.raw.read' AND actor='test-operator'", importID).Scan(&auditCount); err != nil || auditCount != 1 {
		t.Fatal("missing raw-read audit", auditCount, err)
	}
	// A second synthetic report groups two alert variants with different severity.
	// A fixed clock exposes ties between observations and between imports.
	variant := []byte(`{"@version":"SYNTHETIC-SECOND","site":[{"@name":"http://juice-shop:3000","alerts":[
  {"pluginid":40018,"alertRef":"variant-a","riskcode":3,"instances":[{"uri":"http://juice-shop:3000/rest/products/search?q=100%","method":"GET","param":"q"}]},
  {"pluginid":40018,"alertRef":"variant-b","riskcode":1,"instances":[{"uri":"http://juice-shop:3000/rest/products/search?q=a;b","method":"GET","param":"q"}]}
 ]}]}`)
	const operatorMetadata = `{"zap_version":"synthetic-operator-version","report_template":"traditional-json"}`
	w = upload(variant, operatorMetadata)
	var variantImport service.Import
	if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &variantImport) != nil ||
		variantImport.FindingCount != 1 || variantImport.ObservationCount != 2 ||
		variantImport.Provenance.ZAPVersion == nil || *variantImport.Provenance.ZAPVersion != "synthetic-operator-version" ||
		variantImport.ReportDeclaredVersion == nil || *variantImport.ReportDeclaredVersion != "SYNTHETIC-SECOND" {
		t.Fatal("expected one grouped finding with separately reported provenance", w.Code, w.Body)
	}
	w = upload(variant, operatorMetadata)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"duplicate":true`) {
		t.Fatal("same provenance must remain idempotent", w.Code, w.Body)
	}
	var history service.Page[service.Observation]
	w = call("GET", "/api/v1/findings/"+groupedFindingID+"/observations", "", nil)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &history) != nil || len(history.Items) != 4 {
		t.Fatal("expected observations from both imports", w.Code, w.Body)
	}
	for i := 1; i < len(history.Items); i++ {
		if !history.Items[i].CreatedAt.After(history.Items[i-1].CreatedAt) {
			t.Fatal("source and import order should hold even with a fixed clock")
		}
	}
	latest := history.Items[3]
	if history.Items[2].ImportID != variantImport.ID || history.Items[2].Severity != "high" ||
		latest.ImportID != variantImport.ID || latest.Severity != "low" || latest.Pointer != "/site/0/alerts/1/instances/0" {
		t.Fatal("variant observations did not keep their source order", w.Body)
	}
	var detail service.Finding
	w = call("GET", "/api/v1/findings/"+groupedFindingID, "", nil)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &detail) != nil || detail.Severity != "low" || detail.SeverityObservationID != latest.ID {
		t.Fatal("finding summary must use the last variant observation", w.Code, w.Body)
	}
	w = call("GET", "/api/v1/targets/"+target.ID+"/findings?import_id="+importID+"&scanner_severity=high", "", nil)
	var originalPage service.Page[service.Finding]
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &originalPage) != nil {
		t.Fatal("filter on the original import failed", w.Code, w.Body)
	}
	foundOriginal := false
	for _, f := range originalPage.Items {
		if f.ID == groupedFindingID {
			foundOriginal = f.Severity == "high" && f.SeverityObservationID == page2.Items[0].ID
		}
	}
	if !foundOriginal {
		t.Fatal("the original import filter must ignore later observations", w.Body)
	}

	t.Run("unlocated alert retains a large exponent", func(t *testing.T) {
		unlocated := []byte(`{"@version":"synthetic","site":[{"@name":"http://juice-shop:3000","alerts":[{"pluginid":"1","extra":1e999999}]}]}`)
		response := upload(unlocated)
		var imported service.Import
		if response.Code != 201 || json.Unmarshal(response.Body.Bytes(), &imported) != nil ||
			imported.UnlocatedCount != 1 || imported.ObservationCount != 0 {
			t.Fatalf("expected a retained unlocated alert, got %d %s", response.Code, response.Body)
		}
		var encoded string
		var rawBytes []byte
		err := pool.QueryRow(ctx, `SELECT unlocated_alerts_json->0->>'alert_json', raw_bytes
    FROM report_imports WHERE id=$1`, imported.ID).Scan(&encoded, &rawBytes)
		if err != nil || !strings.Contains(encoded, `"extra":1e999999`) || !bytes.Equal(rawBytes, unlocated) {
			t.Fatalf("unlocated alert or original bytes changed: %v", err)
		}
	})

	// Composite FKs reject a report/finding combination from another target.
	secondTarget := "11111111-1111-4111-8111-111111111111"
	_, err := pool.Exec(ctx, "INSERT INTO targets(id,name,canonical_origin,lab_target_key,created_at) VALUES ($1,'Synthetic second target','http://synthetic-other:3000','juice-shop',now())", secondTarget)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO observations(id,target_id,finding_id,report_import_id,source_pointer,instance_index,raw_uri,scanner_severity_raw,scanner_severity,confidence_raw,cwe_ids,evidence_json,description,remediation,created_at)
 SELECT '22222222-2222-4222-8222-222222222222',$1,finding_id,report_import_id,'/mismatch',0,raw_uri,scanner_severity_raw,scanner_severity,confidence_raw,cwe_ids,evidence_json,description,remediation,created_at FROM observations WHERE report_import_id=$2 ORDER BY id LIMIT 1`, secondTarget, importID)
	if err == nil {
		t.Fatal("cross-target observation accepted")
	}
	t.Run("synthetic assessment workflow", func(t *testing.T) {
		checkAssessmentWorkflow(t, ctx, pool, call, target, groupedFindingID)
	})
	if strings.Contains(logs.String(), testToken) || strings.Contains(logs.String(), "synthetic-private-value") {
		t.Fatal("secret in logs")
	}
}
