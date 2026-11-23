// Package classify matches reviewed OWASP rules against scanner assertions.
package classify

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
)

//go:embed owasp-mapping.v1.yaml
var reviewedData []byte

//go:embed owasp-mapping.v1.review.json
var reviewData []byte

type reviewEvidence struct {
	Version      string `json:"version"`
	PolicySHA256 string `json:"policy_sha256"`
	Record       string `json:"review_record"`
}

type Rule struct {
	ID           string   `json:"id"`
	PluginID     string   `json:"plugin_id,omitempty"`
	CWEID        int32    `json:"cwe_id,omitempty"`
	Category     *string  `json:"category"`
	Rationale    string   `json:"rationale"`
	Sources      []string `json:"sources"`
	ReviewRecord string   `json:"review_record"`
	PolicySHA256 string   `json:"policy_sha256"`
}

type Ruleset struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

type Decision struct {
	Status     string
	Category   *string
	Candidates []string
	Rules      []Rule
	Rationale  string
}

func ValidCategory(category string) bool {
	switch category {
	case "A01:2025", "A02:2025", "A05:2025":
		return true
	default:
		return false
	}
}

func Load() (Ruleset, error) {
	var rules Ruleset
	if err := json.Unmarshal(reviewedData, &rules); err != nil {
		return Ruleset{}, err
	}
	if rules.Version == "" || len(rules.Rules) == 0 {
		return Ruleset{}, fmt.Errorf("missing reviewed OWASP rule data")
	}
	var review reviewEvidence
	if err := json.Unmarshal(reviewData, &review); err != nil {
		return Ruleset{}, err
	}
	if review.Version != rules.Version || review.PolicySHA256 != fmt.Sprintf("%x", sha256.Sum256(reviewedData)) || review.Record == "" {
		return Ruleset{}, fmt.Errorf("OWASP policy does not match the independently reviewed version")
	}
	seen := map[string]bool{}
	for index, rule := range rules.Rules {
		if rule.ID == "" || seen[rule.ID] || rule.PluginID == "" && rule.CWEID <= 0 ||
			rule.Rationale == "" || len(rule.Sources) == 0 ||
			rule.Category != nil && !ValidCategory(*rule.Category) {
			return Ruleset{}, fmt.Errorf("invalid reviewed OWASP rule data")
		}
		seen[rule.ID] = true
		rules.Rules[index].ReviewRecord = review.Record
		rules.Rules[index].PolicySHA256 = review.PolicySHA256
	}
	return rules, nil
}

func (r Rule) Matches(plugin string, cwes []int32) bool {
	return (r.PluginID == "" || r.PluginID == plugin) && (r.CWEID == 0 || slices.Contains(cwes, r.CWEID))
}

func (rules Ruleset) Match(plugin string, cwes []int32) Decision {
	r := Decision{Status: "unmapped", Rules: []Rule{}, Candidates: []string{}}
	categories := map[string]bool{}
	rationales := []string{}
	outside := false
	for _, rule := range rules.Rules {
		if !rule.Matches(plugin, cwes) {
			continue
		}
		r.Rules = append(r.Rules, rule)
		rationales = append(rationales, rule.Rationale)
		if rule.Category != nil {
			categories[*rule.Category] = true
		} else {
			outside = true
		}
	}
	r.Candidates = append(r.Candidates, slices.Sorted(maps.Keys(categories))...)
	r.Rationale = strings.Join(rationales, " ")
	switch {
	case len(r.Candidates) > 1 || len(r.Candidates) == 1 && outside:
		r.Status = "ambiguous"
	case len(r.Candidates) == 1:
		r.Status, r.Category = "mapped", &r.Candidates[0]
	case len(r.Rules) > 0:
		r.Status = "out_of_scope"
	}
	return r
}
