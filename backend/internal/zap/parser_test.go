package zap

import (
	"diploma/internal/domain"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../testdata/zap/synthetic-traditional.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestSyntheticGrouping(t *testing.T) {
	r, e := Parse(fixture(t), "target", "http://juice-shop:3000")
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Observations) != 5 || len(r.Unlocated) != 1 || len(r.Warnings) != 2 {
		t.Fatalf("counts %+v", r)
	}
	if r.Observations[0].Fingerprint != r.Observations[1].Fingerprint {
		t.Fatal("attack variations did not group")
	}
	for _, i := range []int{2, 3} {
		if r.Observations[i].Fingerprint == r.Observations[0].Fingerprint {
			t.Fatal("method/parameter incorrectly grouped")
		}
	}
	if r.Observations[4].Severity != "unknown" {
		t.Fatal("unknown severity lost")
	}
}
func TestFingerprintGolden(t *testing.T) {
	o := Observation{PluginID: "40018", Method: "GET", Origin: "http://juice-shop:3000", Path: "/rest/products/search", QueryNames: []string{"q"}, Parameter: "q"}
	const want = "15c1acc59c7bb1a04fc6b6ca6f97795e8b045bff34f8aebc47d572be603e6707"
	if got := Fingerprint("target", o); got != want {
		t.Fatalf("fingerprint %s expected %s", got, want)
	}
}
func TestInvalidReports(t *testing.T) {
	base := string(fixture(t))
	for _, tc := range []struct{ name, raw, code string }{
		{"syntax", "{", "invalid_request"}, {"unsupported", `{"alerts":[]}`, "unsupported_report"},
		{"trailing", base + " {}", "invalid_request"}, {"duplicate-key", `{"@version":"x","@version":"y","site":[]}`, "invalid_request"},
		{"cross-origin", strings.Replace(base, "http://juice-shop:3000/rest/products/search?q=secret1", "http://other:3000/", 1), "out_of_scope"},
		{"bad-method", strings.Replace(base, `"method": "get"`, `"method": "GET BAD"`, 1), "validation_failed"},
		{"bad-uri", strings.Replace(base, "http://juice-shop:3000/rest/products/search?q=secret1", "http://user:password@juice-shop:3000/", 1), "validation_failed"},
		{"field-limit", `{"@version":"` + strings.Repeat("x", 65537) + `"}`, "limit_exceeded"},
		{"depth", strings.Repeat("[", 66) + strings.Repeat("]", 66), "limit_exceeded"},
		{"numeric-limit", `{"@version":` + strings.Repeat("1", 65537) + `}`, "limit_exceeded"},
		{"bytes", strings.Repeat(" ", MaxBytes+1), "limit_exceeded"},
		{"utf8", string([]byte{0xff}), "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, e := Parse([]byte(tc.raw), "target", "http://juice-shop:3000")
			var p *domain.Problem
			if !errors.As(e, &p) || p.Code != tc.code {
				t.Fatalf("got %v", e)
			}
		})
	}
}
func TestCountLimits(t *testing.T) {
	for _, tc := range []struct {
		kind string
		n    int
	}{{"site", 11}, {"alerts", 5001}, {"instances", 20001}} {
		t.Run(tc.kind, func(t *testing.T) {
			var v any
			items := make([]any, tc.n)
			for i := range items {
				items[i] = map[string]any{}
			}
			switch tc.kind {
			case "site":
				v = map[string]any{"site": items}
			case "alerts":
				v = map[string]any{"site": []any{map[string]any{"alerts": items}}}
			default:
				v = map[string]any{"site": []any{map[string]any{"alerts": []any{map[string]any{"instances": items}}}}}
			}
			b, _ := json.Marshal(v)
			_, e := Decode(b, true)
			var p *domain.Problem
			if !errors.As(e, &p) || p.Status != 413 {
				t.Fatal(e)
			}
		})
	}
}
func FuzzDecode(f *testing.F) {
	f.Add([]byte(`{"@version":"x","site":[]}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) < 100000 {
			_, _ = Decode(b, true)
		}
	})
}

func TestAlertWithoutInstances(t *testing.T) {
	raw := []byte(`{"@version":"synthetic","site":[{"@name":"http://juice-shop:3000","alerts":[{"pluginid":"1","alert":"Synthetic unlocated"}]}]}`)
	report, err := Parse(raw, "target", "http://juice-shop:3000")
	if err != nil || len(report.Observations) != 0 || len(report.Unlocated) != 1 ||
		len(report.Warnings) != 1 || report.Warnings[0].Pointer != "/site/0/alerts/0" {
		t.Fatalf("expected one unlocated alert, got %+v, %v", report, err)
	}
}

func TestFingerprintCanonicalQueryNames(t *testing.T) {
	o := Observation{PluginID: "1", Method: "GET", Origin: "http://juice-shop:3000", Path: "/"}
	const empty = "6ab0f17d47d8f3143b1a07ae8cb3f795116fb1f41a9c6418618b7115423116f9"
	if got := Fingerprint("target", o); got != empty {
		t.Fatalf("empty identity hash: %s", got)
	}
	o.QueryNames = []string{}
	if got := Fingerprint("target", o); got != empty {
		t.Fatalf("empty slice identity hash: %s", got)
	}
	o.QueryNames = []string{"z", "a", "z"}
	const sorted = "18581812760e95328597222d34ce819172e103fa4829474f62e0d575de3aa389"
	if got := Fingerprint("target", o); got != sorted || o.QueryNames[0] != "z" || len(o.QueryNames) != 3 {
		t.Fatalf("canonical identity hash or caller slice: %s %v", got, o.QueryNames)
	}
}
