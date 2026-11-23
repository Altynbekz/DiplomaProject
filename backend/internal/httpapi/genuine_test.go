package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"diploma/internal/service"
	"diploma/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

const genuineFixtureTargetID = "00000000-0000-4000-8000-000000000001"

func TestIntegrationGenuineReport(t *testing.T) {
	dsn := os.Getenv("TEST_GENUINE_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_GENUINE_DATABASE_URL required for disposable genuine-report integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["role"] = "diploma_runtime"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	const origin = "http://juice-shop:3000"
	if _, err := pool.Exec(ctx, "INSERT INTO targets(id, name, canonical_origin, lab_target_key, created_at) VALUES ($1, $2, $3, $4, now())",
		genuineFixtureTargetID, "Genuine fixture", origin, "juice-shop"); err != nil {
		t.Fatal(err)
	}
	api := API{
		Service: &service.Service{Store: store.New(pool), Origin: origin, Actor: "fixture-operator", CursorKey: testToken, Now: time.Now},
		Token:   testToken, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	handler := api.Handler()
	request := func(method, path, media string, body []byte, acknowledge bool) *httptest.ResponseRecorder {
		r := httptest.NewRequestWithContext(ctx, method, "/api/v1"+path, bytes.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+testToken)
		r.Header.Set("Content-Type", media)
		if acknowledge {
			r.Header.Set("X-Evidence-Acknowledgement", "private-lab-data")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}

	const fixturePath = "../../testdata/zap/juice-shop-"
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(fixturePath + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	raw := read("traditional")
	provenance := read("provenance")
	var wantMetadata service.Provenance
	decodeFixtureJSON(t, provenance, &wantMetadata)
	digest := sha256.Sum256(raw)
	if wantMetadata.SanitizedSHA256 == nil || *wantMetadata.SanitizedSHA256 != hex.EncodeToString(digest[:]) ||
		wantMetadata.RawSHA256 == nil || *wantMetadata.RawSHA256 == *wantMetadata.SanitizedSHA256 {
		t.Fatal("fixture digests disagree with provenance")
	}
	var expected map[string][]string
	decodeFixtureJSON(t, read("expected"), &expected)
	count := 0
	for _, pointers := range expected {
		count += len(pointers)
		slices.Sort(pointers)
	}
	if len(expected) != 17 || count != 21 {
		t.Fatal("unexpected published fixture cardinality")
	}

	makeUpload := func(metadata []byte) (string, []byte) {
		t.Helper()
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("report", "juice-shop-traditional.json")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(raw); err != nil {
			t.Fatal(err)
		}
		if err := writer.WriteField("provenance", string(metadata)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return writer.FormDataContentType(), body.Bytes()
	}
	media, body := makeUpload(provenance)
	path := "/targets/" + genuineFixtureTargetID
	results := make(chan *httptest.ResponseRecorder, 4)
	begin := make(chan struct{})
	for range 4 {
		go func() {
			<-begin
			results <- request("POST", path+"/imports/zap", media, body, false)
		}()
	}
	close(begin)
	created, importID := 0, ""
	for range 4 {
		w := <-results
		var result service.Import
		switch w.Code {
		case 201:
			created++
		case 200:
		default:
			t.Fatalf("genuine import returned %d: %s", w.Code, w.Body)
		}
		decodeFixtureJSON(t, w.Body.Bytes(), &result)
		if importID == "" {
			importID = result.ID
		}
		if result.ID != importID || result.Duplicate != (w.Code == 200) ||
			result.FindingCount != int32(len(expected)) || result.ObservationCount != int32(count) ||
			result.UnlocatedCount != 0 || result.Hash != *wantMetadata.SanitizedSHA256 ||
			!reflect.DeepEqual(result.Provenance, wantMetadata) {
			t.Fatal("genuine import result disagrees with the fixture or duplicate result")
		}
	}
	if created != 1 {
		t.Fatalf("expected one creation, got %d", created)
	}
	if w := request("POST", path+"/imports/zap", media, body, false); w.Code != 200 {
		t.Fatalf("sequential duplicate returned %d", w.Code)
	} else {
		var duplicate service.Import
		decodeFixtureJSON(t, w.Body.Bytes(), &duplicate)
		if !duplicate.Duplicate || duplicate.ID != importID {
			t.Fatal("sequential duplicate had a different identity")
		}
	}
	conflicting := bytes.Replace(provenance, []byte(*wantMetadata.AuthorizationRef), []byte("different-fixture-authorization"), 1)
	conflictMedia, conflictBody := makeUpload(conflicting)
	if w := request("POST", path+"/imports/zap", conflictMedia, conflictBody, false); w.Code != 409 {
		t.Fatalf("conflicting provenance returned %d", w.Code)
	}
	if w := request("GET", "/imports/"+importID+"/raw", "", nil, true); w.Code != 200 || !bytes.Equal(w.Body.Bytes(), raw) {
		t.Fatal("explicit evidence download did not preserve exact submitted bytes")
	}
	var storedBytes, storedProvenance []byte
	var imports int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM report_imports").Scan(&imports); err != nil || imports != 1 {
		t.Fatalf("expected one database import, got %d: %v", imports, err)
	}
	if err := pool.QueryRow(ctx, "SELECT raw_bytes, source_metadata_json FROM report_imports WHERE id=$1", importID).Scan(&storedBytes, &storedProvenance); err != nil {
		t.Fatal(err)
	}
	var metadata service.Provenance
	decodeFixtureJSON(t, storedProvenance, &metadata)
	if !bytes.Equal(storedBytes, raw) || !reflect.DeepEqual(metadata, wantMetadata) {
		t.Fatal("PostgreSQL did not preserve the bytes and original provenance")
	}

	dbGroups := make(map[string][]string)
	rows, err := pool.Query(ctx, "SELECT f.fingerprint, o.source_pointer FROM findings f JOIN observations o ON o.finding_id=f.id ORDER BY f.fingerprint, o.source_pointer")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var fingerprint, pointer string
		if err := rows.Scan(&fingerprint, &pointer); err != nil {
			t.Fatal(err)
		}
		dbGroups[fingerprint] = append(dbGroups[fingerprint], pointer)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dbGroups, expected) {
		t.Fatal("database fingerprint groups differed from the independent oracle")
	}

	apiGroups := make(map[string][]string)
	query := url.Values{"limit": {"5"}, "import_id": {importID}}
	for pageCount := 0; ; pageCount++ {
		if pageCount >= len(expected) {
			t.Fatal("findings pagination did not terminate")
		}
		var page service.Page[service.Finding]
		w := request("GET", path+"/findings?"+query.Encode(), "", nil, false)
		if w.Code != 200 {
			t.Fatalf("findings list returned %d", w.Code)
		}
		decodeFixtureJSON(t, w.Body.Bytes(), &page)
		for _, item := range page.Items {
			if item.Verification != "unreviewed" || item.Base != nil || item.Environmental != nil {
				t.Fatal("genuine import assigned a review or score")
			}
			var finding service.Finding
			w := request("GET", "/findings/"+item.ID, "", nil, false)
			if w.Code != 200 {
				t.Fatalf("finding detail returned %d", w.Code)
			}
			decodeFixtureJSON(t, w.Body.Bytes(), &finding)
			var observations service.Page[service.Observation]
			w = request("GET", "/findings/"+item.ID+"/observations", "", nil, false)
			if w.Code != 200 {
				t.Fatalf("observation history returned %d", w.Code)
			}
			decodeFixtureJSON(t, w.Body.Bytes(), &observations)
			for _, observation := range observations.Items {
				if observation.ImportID != importID {
					t.Fatal("observation connected to the wrong import")
				}
				apiGroups[finding.Fingerprint] = append(apiGroups[finding.Fingerprint], observation.Pointer)
			}
			slices.Sort(apiGroups[finding.Fingerprint])
		}
		if page.Next == nil {
			break
		}
		query.Set("cursor", *page.Next)
	}
	if !reflect.DeepEqual(apiGroups, expected) {
		t.Fatal("API fingerprint groups differed from the independent oracle")
	}
	t.Run("verification workflow", func(t *testing.T) {
		checkVerificationWorkflow(t, ctx, pool, request, path)
	})
}

func decodeFixtureJSON(t *testing.T, data []byte, value any) {
	t.Helper()
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatal(err)
	}
}
