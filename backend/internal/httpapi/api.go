package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"diploma/internal/domain"
	"diploma/internal/service"
	"diploma/internal/zap"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type API struct {
	Service *service.Service
	Token   string
	Log     *slog.Logger
}
type requestKey struct{}

func requestID(r *http.Request) string { v, _ := r.Context().Value(requestKey{}).(string); return v }
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (a *API) problem(w http.ResponseWriter, r *http.Request, e error) {
	p := &domain.Problem{Status: 500, Code: "internal_error", Detail: "The operation could not be completed", Errors: []domain.FieldError{}}
	var known *domain.Problem
	var pe *pgconn.PgError
	var max *http.MaxBytesError
	switch {
	case errors.As(e, &known):
		p = known
	case errors.Is(e, pgx.ErrNoRows):
		p.Status = 404
		p.Code = "not_found"
		p.Detail = "Resource not found"
	case errors.As(e, &max):
		p.Status = 413
		p.Code = "limit_exceeded"
		p.Detail = "Request exceeds byte limit"
	case errors.As(e, &pe) && pe.Code == "23505":
		p.Status = 409
		p.Code = "conflict"
		p.Detail = "Resource already exists"
	}
	if p.Status == 500 {
		a.Log.Error("request failed", "request_id", requestID(r), "cause_type", fmt.Sprintf("%T", e))
	}
	if p.Errors == nil {
		p.Errors = []domain.FieldError{}
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{"type": "urn:diploma:problem:" + p.Code, "title": http.StatusText(p.Status), "status": p.Status, "code": p.Code, "detail": p.Detail, "request_id": requestID(r), "errors": p.Errors})
}

type endpoint func(http.ResponseWriter, *http.Request) error

func (a *API) wrap(fn endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := fn(w, r); e != nil {
			a.problem(w, r, e)
		}
	}
}
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", a.wrap(func(w http.ResponseWriter, r *http.Request) error {
		write(w, 200, map[string]string{"status": "ok"})
		return nil
	}))
	m.HandleFunc("GET /readyz", a.wrap(func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if e := a.Service.Ready(ctx); e != nil {
			return domain.Fail(503, "not_ready", "", "Database or schema is unavailable")
		}
		write(w, 200, map[string]string{"status": "ready"})
		return nil
	}))
	m.HandleFunc("POST /api/v1/targets", a.wrap(a.createTarget))
	m.HandleFunc("GET /api/v1/targets", a.wrap(a.targets))
	m.HandleFunc("GET /api/v1/targets/{id}", a.wrap(a.target))
	m.HandleFunc("POST /api/v1/targets/{id}/imports/zap", a.wrap(a.importReport))
	m.HandleFunc("GET /api/v1/imports/{id}", a.wrap(a.getImport))
	m.HandleFunc("GET /api/v1/imports/{id}/raw", a.wrap(a.raw))
	m.HandleFunc("GET /api/v1/targets/{id}/findings", a.wrap(a.findings))
	m.HandleFunc("GET /api/v1/findings/{id}", a.wrap(a.finding))
	m.HandleFunc("GET /api/v1/findings/{id}/observations", a.wrap(a.observations))
	m.HandleFunc("/", a.wrap(func(w http.ResponseWriter, r *http.Request) error {
		return domain.Fail(404, "not_found", "", "Route not found")
	}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if !regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`).MatchString(id) {
			id = uuid.NewString()
		}
		r = r.WithContext(context.WithValue(r.Context(), requestKey{}, id))
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.URL.Path != "/healthz" && r.URL.Path != "/readyz" {
			actual := sha256.Sum256([]byte(r.Header.Get("Authorization")))
			expected := sha256.Sum256([]byte("Bearer " + a.Token))
			if subtle.ConstantTimeCompare(actual[:], expected[:]) != 1 {
				a.problem(w, r, domain.Fail(401, "unauthorized", "", "A valid bearer token is required"))
				return
			}
		}
		m.ServeHTTP(w, r)
	})
}
func id(r *http.Request) (string, error) {
	v := r.PathValue("id")
	p, e := uuid.Parse(v)
	if e != nil || p.String() != v {
		return "", domain.Fail(400, "invalid_request", "/id", "Expected canonical UUID")
	}
	return v, nil
}
func strict(raw []byte, v any) error {
	decoded, err := zap.Decode(raw, false)
	if err != nil {
		return err
	}
	input, ok := decoded.(map[string]any)
	if !ok {
		return domain.Fail(400, "invalid_request", "", "Expected a JSON object")
	}

	// encoding/json matches struct fields case-insensitively. Compare keys
	// against the DTO's serialized names first to enforce the exact contract.
	shape, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var keys map[string]json.RawMessage
	if err = json.Unmarshal(shape, &keys); err != nil {
		return err
	}
	for k := range input {
		if _, known := keys[k]; !known {
			return domain.Fail(400, "invalid_request", "", "Unknown JSON field")
		}
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return domain.Fail(400, "invalid_request", "", "Invalid JSON field value")
	}
	return nil
}
func body(w http.ResponseWriter, r *http.Request, v any) error {
	ct, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || ct != "application/json" {
		return domain.Fail(415, "unsupported_media_type", "", "Expected application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	raw, e := io.ReadAll(r.Body)
	if e != nil {
		return e
	}
	return strict(raw, v)
}
func pagination(r *http.Request, allowed ...string) (int, string, error) {
	ok := map[string]bool{"limit": true, "cursor": true}
	for _, v := range allowed {
		ok[v] = true
	}
	q, e := url.ParseQuery(r.URL.RawQuery)
	if e != nil {
		return 0, "", domain.Fail(400, "invalid_request", "", "Malformed query encoding")
	}
	for k, v := range q {
		if !ok[k] || len(v) != 1 {
			return 0, "", domain.Fail(400, "invalid_request", "", "Unknown or repeated query field")
		}
	}
	n := 50
	if q.Has("limit") {
		n, e = strconv.Atoi(q.Get("limit"))
		if e != nil || n < 1 || n > 200 {
			return 0, "", domain.Fail(400, "invalid_request", "/limit", "Limit must be between 1 and 200")
		}
	}
	return n, q.Get("cursor"), nil
}
func (a *API) createTarget(w http.ResponseWriter, r *http.Request) error {
	var b struct {
		Name string `json:"name"`
		Key  string `json:"lab_target_key"`
	}
	if e := body(w, r, &b); e != nil {
		return e
	}
	v, e := a.Service.CreateTarget(r.Context(), b.Name, b.Key, requestID(r))
	if e != nil {
		return e
	}
	w.Header().Set("Location", "/api/v1/targets/"+v.ID)
	write(w, 201, v)
	return nil
}
func (a *API) targets(w http.ResponseWriter, r *http.Request) error {
	n, c, e := pagination(r)
	if e != nil {
		return e
	}
	v, e := a.Service.Targets(r.Context(), n, c)
	if e != nil {
		return e
	}
	write(w, 200, v)
	return nil
}
func (a *API) target(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	v, e := a.Service.Target(r.Context(), i)
	if e != nil {
		return e
	}
	write(w, 200, v)
	return nil
}
func (a *API) importReport(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	ct, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || ct != "multipart/form-data" {
		return domain.Fail(415, "unsupported_media_type", "", "Expected multipart/form-data")
	}
	if r.ContentLength > zap.MaxBytes {
		return domain.Fail(413, "limit_exceeded", "", "Request exceeds 20 MiB")
	}
	r.Body = http.MaxBytesReader(w, r.Body, zap.MaxBytes)
	mr, e := r.MultipartReader()
	if e != nil {
		return domain.Fail(400, "invalid_request", "", "Invalid multipart envelope")
	}
	seen := map[string]bool{}
	var raw []byte
	var prov *service.Provenance
	for {
		part, e := mr.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			var max *http.MaxBytesError
			if errors.As(e, &max) {
				return e
			}
			return domain.Fail(400, "invalid_request", "", "Invalid multipart body")
		}
		name := part.FormName()
		if seen[name] || (name != "report" && name != "provenance") {
			return domain.Fail(400, "invalid_request", "", "Unknown or duplicate multipart field")
		}
		seen[name] = true
		cap := int64(zap.MaxBytes)
		if name == "provenance" {
			cap = 65536
		}
		b, e := io.ReadAll(io.LimitReader(part, cap+1))
		if e != nil {
			return e
		}
		if int64(len(b)) > cap {
			return domain.Fail(413, "limit_exceeded", "/"+name, "Multipart field exceeds limit")
		}
		if name == "report" {
			raw = b
		} else {
			prov = &service.Provenance{}
			if e = strict(b, prov); e != nil {
				return e
			}
		}
	}
	if _, e := io.Copy(io.Discard, r.Body); e != nil {
		return e
	}
	if !seen["report"] {
		return domain.Fail(400, "invalid_request", "/report", "Exactly one report field is required")
	}
	v, e := a.Service.Import(r.Context(), i, requestID(r), raw, prov)
	if e != nil {
		return e
	}
	status := 201
	if v.Duplicate {
		status = 200
	}
	w.Header().Set("Location", "/api/v1/imports/"+v.ID)
	write(w, status, v)
	return nil
}
func (a *API) getImport(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	v, e := a.Service.GetImport(r.Context(), i)
	if e != nil {
		return e
	}
	write(w, 200, v)
	return nil
}
func (a *API) raw(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	if r.Header.Get("X-Evidence-Acknowledgement") != "private-lab-data" {
		return domain.Fail(422, "validation_failed", "/X-Evidence-Acknowledgement", "Explicit private-lab-data acknowledgement required")
	}
	raw, e := a.Service.Raw(r.Context(), i, requestID(r))
	if e != nil {
		return e
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="zap-report.json"`)
	w.WriteHeader(200)
	_, e = w.Write(raw)
	return e
}
func enum(value string, allowed ...string) bool {
	if value == "" {
		return true
	}
	for _, v := range allowed {
		if v == value {
			return true
		}
	}
	return false
}
func (a *API) findings(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	n, c, e := pagination(r, "import_id", "category", "verification", "scanner_severity")
	if e != nil {
		return e
	}
	q := r.URL.Query()
	f := service.FindingFilter{TargetID: i, ImportID: q.Get("import_id"), Category: q.Get("category"), Verification: q.Get("verification"), Severity: q.Get("scanner_severity")}
	if !enum(f.Category, "A01:2025", "A02:2025", "A05:2025") || !enum(f.Verification, "unreviewed", "confirmed", "false_positive", "inconclusive") || !enum(f.Severity, "informational", "low", "medium", "high", "unknown") {
		return domain.Fail(400, "invalid_request", "", "Invalid filter enum")
	}
	if f.ImportID != "" {
		u, e := uuid.Parse(f.ImportID)
		if e != nil || u.String() != f.ImportID {
			return domain.Fail(400, "invalid_request", "/import_id", "Expected UUID")
		}
	}
	v, e := a.Service.Findings(r.Context(), f, n, c)
	if e != nil {
		return e
	}
	write(w, 200, v)
	return nil
}
func (a *API) finding(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	v, e := a.Service.Finding(r.Context(), i)
	if e != nil {
		return e
	}
	write(w, 200, v)
	return nil
}
func (a *API) observations(w http.ResponseWriter, r *http.Request) error {
	i, e := id(r)
	if e != nil {
		return e
	}
	n, c, e := pagination(r)
	if e != nil {
		return e
	}
	v, e := a.Service.Observations(r.Context(), i, n, c)
	if e != nil {
		return e
	}
	write(w, 200, v)
	return nil
}
