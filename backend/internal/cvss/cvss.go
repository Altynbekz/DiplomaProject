// Package cvss validates the project's CVSS v4 scope and delegates scoring.
package cvss

import (
	"embed"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/dop251/goja"
)

const (
	SourceCommit    = "c5b0d409ae9f57c44264c6ce5f27d89298e1d32a"
	EngineVersion   = "v0.0.0-20261122185237-35fdc4b88e39"
	ReferenceCommit = SourceCommit
	FixtureVersion  = "first-cvss-v4/v1"
)

type metric struct{ name, values string }

var baseOrder = []metric{
	{"AV", "NALP"}, {"AC", "LH"}, {"AT", "NP"}, {"PR", "NLH"}, {"UI", "NPA"},
	{"VC", "HLN"}, {"VI", "HLN"}, {"VA", "HLN"}, {"SC", "HLN"}, {"SI", "HLN"}, {"SA", "HLN"},
}
var environmentalOrder = []metric{
	{"CR", "XHML"}, {"IR", "XHML"}, {"AR", "XHML"}, {"MAV", "XNALP"}, {"MAC", "XLH"},
	{"MAT", "XNP"}, {"MPR", "XNLH"}, {"MUI", "XNPA"}, {"MVC", "XHLN"}, {"MVI", "XHLN"},
	{"MVA", "XHLN"}, {"MSC", "XHLN"}, {"MSI", "XSHLN"}, {"MSA", "XSHLN"},
}

type MetricError struct {
	Name string
}

func (e *MetricError) Error() string { return "Invalid or unsupported CVSS metric or value" }

type Provenance struct {
	Name                  string `json:"name"`
	Version               string `json:"version"`
	SourceCommit          string `json:"source_commit"`
	SpecificationVersion  string `json:"specification_version"`
	SpecificationDocument string `json:"specification_document_version"`
	ReferenceCommit       string `json:"reference_commit"`
	FixtureVersion        string `json:"fixture_version"`
	EngineVersion         string `json:"engine_version"`
}

func Calculator() Provenance {
	return Provenance{
		Name: "FIRST CVSS v4 reference calculator", Version: "git:c5b0d409ae9f",
		SourceCommit: SourceCommit, SpecificationVersion: "4.0", SpecificationDocument: "1.2",
		ReferenceCommit: ReferenceCommit, FixtureVersion: FixtureVersion, EngineVersion: EngineVersion,
	}
}

// Validate accepts partial inputs. Environmental metrics are the only
// optional metrics in this assessment model.
func Validate(metrics map[string]string, environmental bool) error {
	order := baseOrder
	if environmental {
		order = environmentalOrder
	}
	for _, name := range slices.Sorted(maps.Keys(metrics)) {
		i := slices.IndexFunc(order, func(def metric) bool { return def.name == name })
		value := metrics[name]
		if i < 0 || len(value) != 1 || !strings.Contains(order[i].values, value) {
			return &MetricError{Name: name}
		}
	}
	return nil
}

// SnapshotMetrics copies the Base values and materializes all optional
// Environmental omissions as Not Defined. A nil environmental map denotes B.
func SnapshotMetrics(base, environmental map[string]string) map[string]string {
	metrics := make(map[string]string, len(base)+len(environmentalOrder))
	maps.Copy(metrics, base)
	if environmental != nil {
		for _, key := range environmentalOrder {
			metrics[key.name] = "X"
		}
		maps.Copy(metrics, environmental)
	}
	return metrics
}

type Result struct {
	Vector *string
	Score  *float64
}

// Evaluate produces a null vector and score for an incomplete input. A
// non-nil Environmental map requests BE and must define at least one metric.
func Evaluate(base, environmental map[string]string) (Result, error) {
	if err := Validate(base, false); err != nil {
		return Result{}, err
	}
	if err := Validate(environmental, true); err != nil {
		return Result{}, err
	}
	if len(base) != len(baseOrder) {
		return Result{}, nil
	}
	var vector strings.Builder
	vector.WriteString("CVSS:4.0")
	for _, key := range baseOrder {
		fmt.Fprintf(&vector, "/%s:%s", key.name, base[key.name])
	}
	defined := false
	for _, key := range environmentalOrder {
		if value := environmental[key.name]; value != "" && value != "X" {
			fmt.Fprintf(&vector, "/%s:%s", key.name, value)
			defined = true
		}
	}
	if environmental != nil && !defined {
		return Result{}, nil
	}
	canonical := vector.String()
	score, err := referenceScore(base, environmental)
	if err != nil {
		return Result{}, err
	}
	if math.IsNaN(score) || score < 0 || score > 10 || math.Abs(score*10-math.Round(score*10)) > 1e-9 {
		return Result{}, fmt.Errorf("calculator returned an invalid rounded score")
	}
	return Result{Vector: &canonical, Score: &score}, nil
}

//go:embed first/*.js
var reference embed.FS

var referencePrograms = sync.OnceValues(func() ([]*goja.Program, error) {
	files := []string{"metrics.js", "cvss_lookup.js", "max_composed.js", "max_severity.js", "cvss_score.js"}
	programs := make([]*goja.Program, 0, len(files)+1)
	for _, file := range files {
		source, err := reference.ReadFile("first/" + file)
		if err != nil {
			return nil, err
		}
		program, err := goja.Compile(file, string(source), false)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	entry, err := goja.Compile("diploma-score", "cvss_score(selected, cvssLookup_global, maxSeverity, macroVector(selected))", false)
	return append(programs, entry), err
})

func referenceScore(base, environmental map[string]string) (float64, error) {
	programs, err := referencePrograms()
	if err != nil {
		return 0, err
	}
	if environmental == nil {
		environmental = map[string]string{}
	}
	selected := SnapshotMetrics(base, environmental)
	selected["E"] = "X"
	runtime := goja.New()
	if err := runtime.Set("selected", selected); err != nil {
		return 0, err
	}
	var score goja.Value
	for _, program := range programs {
		score, err = runtime.RunProgram(program)
		if err != nil {
			return 0, err
		}
	}
	return score.ToFloat(), nil
}
