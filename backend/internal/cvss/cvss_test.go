package cvss

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestFirstReferenceParity(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/cvss/first-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Version         string
		ReferenceCommit string `json:"reference_commit"`
		Cases           []struct {
			Name          string
			Base          map[string]string
			Environmental map[string]string
			Vector        string
			Score         float64
		}
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if fixtures.ReferenceCommit != ReferenceCommit || fixtures.Version != FixtureVersion || len(fixtures.Cases) < 10 {
		t.Fatal("incorrect or incomplete reference fixture")
	}
	for _, c := range fixtures.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := Evaluate(c.Base, c.Environmental)
			if err != nil || got.Vector == nil || got.Score == nil {
				t.Fatalf("calculation unavailable: %v (%+v)", err, got)
			}
			if *got.Vector != c.Vector || *got.Score != c.Score {
				t.Fatalf("got %s %v, FIRST returned %s %v", *got.Vector, *got.Score, c.Vector, c.Score)
			}
			snapshot := SnapshotMetrics(c.Base, c.Environmental)
			for name, value := range c.Base {
				if snapshot[name] != value {
					t.Fatalf("Base metric %s changed during inheritance", name)
				}
			}
			if c.Environmental != nil {
				for _, key := range environmentalOrder {
					want := c.Environmental[key.name]
					if want == "" {
						want = "X"
					}
					if snapshot[key.name] != want {
						t.Fatalf("snapshot did not keep the choice for %s", key.name)
					}
				}
			}
		})
	}
}

func TestMetricScopeAndIncompleteInputs(t *testing.T) {
	for _, tc := range []struct {
		name, value   string
		environmental bool
	}{
		{"AV", "X", false}, {"SI", "S", false}, {"E", "X", true}, {"S", "P", true},
		{"AV", "N", true}, {"MAV", "N", false}, {"MSI", "s", true}, {"MVA", "S", true},
		{"VC", "", false}, {"AV", "NN", false}, {"av", "N", false},
	} {
		err := Validate(map[string]string{tc.name: tc.value}, tc.environmental)
		var invalid *MetricError
		if !errors.As(err, &invalid) || invalid.Name != tc.name {
			t.Errorf("accepted invalid metric %s:%s in environmental=%v: %v", tc.name, tc.value, tc.environmental, err)
		}
	}
	base := map[string]string{"AV": "N", "AC": "L", "AT": "N", "PR": "N", "UI": "N", "VC": "L", "VI": "N", "VA": "N", "SC": "N", "SI": "N", "SA": "N"}
	for _, result := range []struct{ b, e map[string]string }{
		{map[string]string{"AV": "N"}, nil},
		{base, map[string]string{}},
		{base, map[string]string{"MAV": "X"}},
	} {
		got, err := Evaluate(result.b, result.e)
		if err != nil || got.Vector != nil || got.Score != nil {
			t.Fatalf("incomplete input produced output: %+v %v", got, err)
		}
	}
}
