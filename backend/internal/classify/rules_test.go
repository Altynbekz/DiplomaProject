package classify

import (
	"reflect"
	"testing"
)

func TestReviewedRules(t *testing.T) {
	rules, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		plugin     string
		cwes       []int32
		status     string
		candidates []string
	}{
		{"10098", []int32{264}, "mapped", []string{"A01:2025"}},
		{"10038", []int32{693}, "mapped", []string{"A02:2025"}},
		{"40018", []int32{89}, "mapped", []string{"A05:2025"}},
		{"10098", []int32{89, 264}, "ambiguous", []string{"A01:2025", "A05:2025"}},
		{"10109", nil, "out_of_scope", []string{}},
		{"10109", []int32{89}, "ambiguous", []string{"A05:2025"}},
		{"10038", nil, "unmapped", []string{}},
		{"unknown", []int32{693}, "unmapped", []string{}},
	} {
		decision := rules.Match(tc.plugin, tc.cwes)
		if decision.Status != tc.status || !reflect.DeepEqual(decision.Candidates, tc.candidates) {
			t.Errorf("plugin %s: got %s %v", tc.plugin, decision.Status, decision.Candidates)
		}
		if decision.Status != "mapped" && decision.Category != nil {
			t.Error("non-mapped rule supplied a category")
		}
	}
}
