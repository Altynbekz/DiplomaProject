package scope

import "testing"

func TestNormalization(t *testing.T) {
	for _, tc := range []struct {
		raw, origin, path string
		names             []string
	}{
		{"HTTP://EXAMPLE.com/a%7eb%2fc/?z=2&a=secret&a=other", "http://example.com:80", "/a~b%2Fc/", []string{"a", "z"}},
		{"https://example.com/1/", "https://example.com:443", "/1/", []string{}},
		{"http://[::1]:3000/A", "http://[::1]:3000", "/A", []string{}},
	} {
		v, e := Parse(tc.raw)
		if e != nil || v.Origin != tc.origin || v.Path != tc.path || len(v.QueryNames) != len(tc.names) {
			t.Fatalf("%s: %+v %v", tc.raw, v, e)
		}
		for i, n := range tc.names {
			if v.QueryNames[i] != n {
				t.Fatal(v)
			}
		}
	}
}
func TestRejectURLs(t *testing.T) {
	for _, v := range []string{"http://user:pass@example.com/", "http://example.com/#", "http://example.com/%zz", "http://127.1/", "http://0177.0.0.1/", "http://2130706433/", "http://0x7f000001/", "http://[fe80::1%25en0]/", "ftp://example.com/", "//example.com/", "http://example.com:0/", "http://example.com:65536/", "http://example.com:/", "http://example.com/?q%zz=1", "http://é.com/"} {
		if _, e := Parse(v); e == nil {
			t.Errorf("accepted %q", v)
		}
	}
}
func FuzzParse(f *testing.F) {
	f.Add("http://juice-shop:3000/a%2fb?q=1")
	f.Fuzz(func(t *testing.T, s string) { _, _ = Parse(s) })
}

func TestDecodedQueryNames(t *testing.T) {
	for _, raw := range []string{
		"http://juice-shop:3000/path?%00=x",
		"http://juice-shop:3000/path?%FF=x",
	} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("accepted query name that cannot be stored as PostgreSQL text: %s", raw)
		}
	}
	loc, err := Parse("http://juice-shop:3000/path?caf%C3%A9=%FF&café=%00")
	if err != nil || len(loc.QueryNames) != 1 || loc.QueryNames[0] != "café" {
		t.Fatalf("expected one valid decoded name regardless of query values: %+v, %v", loc, err)
	}
}

func TestQueryValuesDoNotAffectIdentity(t *testing.T) {
	for _, tc := range []struct{ query, name string }{
		{"q=1;2", "q"},
		{"q=100%", "q"},
		{"q=%27%3B--", "q"},
		{"=x", ""},
		{"q=1;other=2&q=3", "q"},
	} {
		loc, err := Parse("http://juice-shop:3000/path?" + tc.query)
		if err != nil || len(loc.QueryNames) != 1 || loc.QueryNames[0] != tc.name {
			t.Errorf("query %q: expected one name %q, got %+v, %v", tc.query, tc.name, loc, err)
		}
	}
}
