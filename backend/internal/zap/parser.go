// Package zap imports Traditional JSON; there is deliberately no scanner client.
package zap

import (
	"bytes"
	"crypto/sha256"
	"diploma/internal/domain"
	"diploma/internal/scope"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const Version = "zap-traditional/v1"
const ReportFormat = "traditional-json"
const MaxBytes = 20 << 20

type Warning struct {
	Code    string `json:"code"`
	Pointer string `json:"pointer"`
	Message string `json:"message"`
}
type Observation struct {
	Fingerprint, PluginID, Title, Method, Origin, Path, Parameter, Pointer, RawURI, SeverityRaw, Severity, Confidence, Description, Remediation string
	QueryNames                                                                                                                                  []string
	CWE                                                                                                                                         []int32
	Index                                                                                                                                       int32
	Evidence                                                                                                                                    json.RawMessage
}

// UnlocatedAlert keeps report extensions reversible without coercing their
// numbers into PostgreSQL JSONB's smaller numeric range.
type UnlocatedAlert struct {
	SchemaVersion string `json:"schema_version"`
	SourcePointer string `json:"source_pointer"`
	AlertJSON     string `json:"alert_json"`
}

type Report struct {
	Hash               string
	Observations       []Observation
	Warnings           []Warning
	Unlocated          []UnlocatedAlert
	Version, Generated string
}

// Decode bounds depth, per-field bytes and cardinality before allocating array elements.
// Mutation decoders reuse it for duplicate-key and trailing-document rejection.
func Decode(raw []byte, report bool) (any, error) {
	if !utf8.Valid(raw) {
		return nil, domain.Fail(400, "invalid_request", "", "JSON must be UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	counts := map[string]int{}
	alertArray := regexp.MustCompile(`^/site/\d+/alerts$`)
	instanceArray := regexp.MustCompile(`^/site/\d+/alerts/\d+/instances$`)
	var read func(int, string) (any, error)
	read = func(depth int, p string) (any, error) {
		if depth > 64 {
			return nil, domain.Fail(413, "limit_exceeded", p, "JSON nesting exceeds 64")
		}
		t, e := d.Token()
		if e != nil {
			return nil, domain.Fail(400, "invalid_request", p, "Malformed JSON")
		}
		fieldBytes := 0
		switch v := t.(type) {
		case string:
			fieldBytes = len(v)
		case json.Number:
			fieldBytes = len(v.String())
		}
		if fieldBytes > 65536 {
			return nil, domain.Fail(413, "limit_exceeded", p, "Scalar field exceeds 64 KiB")
		}
		switch t {
		case json.Delim('{'):
			m := map[string]any{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return nil, domain.Fail(400, "invalid_request", p, "Malformed JSON object")
				}
				s, ok := k.(string)
				if !ok {
					return nil, domain.Fail(400, "invalid_request", p, "Invalid key")
				}
				if len(s) > 65536 {
					return nil, domain.Fail(413, "limit_exceeded", p, "Key too long")
				}
				if _, ok := m[s]; ok {
					return nil, domain.Fail(400, "invalid_request", p, "Duplicate JSON key")
				}
				v, e := read(depth+1, p+"/"+strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1"))
				if e != nil {
					return nil, e
				}
				m[s] = v
			}
			if _, e = d.Token(); e != nil {
				return nil, domain.Fail(400, "invalid_request", p, "Malformed JSON")
			}
			return m, nil
		case json.Delim('['):
			a := []any{}
			for d.More() {
				if report {
					key := ""
					limit := 0
					switch {
					case p == "/site":
						key = "site"
						limit = 10
					case alertArray.MatchString(p):
						key = "alerts"
						limit = 5000
					case instanceArray.MatchString(p):
						key = "instances"
						limit = 20000
					}
					if key != "" {
						counts[key]++
						if counts[key] > limit {
							return nil, domain.Fail(413, "limit_exceeded", p, "Report object count limit exceeded")
						}
					}
				}
				v, e := read(depth+1, fmt.Sprintf("%s/%d", p, len(a)))
				if e != nil {
					return nil, e
				}
				a = append(a, v)
			}
			if _, e = d.Token(); e != nil {
				return nil, domain.Fail(400, "invalid_request", p, "Malformed JSON")
			}
			return a, nil
		case json.Delim('}'), json.Delim(']'):
			return nil, domain.Fail(400, "invalid_request", p, "Malformed JSON")
		default:
			return t, nil
		}
	}
	v, e := read(0, "")
	if e != nil {
		return nil, e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, domain.Fail(400, "invalid_request", "", "Trailing JSON data")
	}
	return v, nil
}
func text(m map[string]any, k string) string {
	switch v := m[k].(type) {
	case string:
		return v
	case json.Number:
		return string(v)
	}
	return ""
}
func object(v any, p string) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, domain.Fail(422, "validation_failed", p, "Expected an object")
	}
	return m, nil
}
func array(v any, p string, optional bool) ([]any, error) {
	if v == nil && optional {
		return []any{}, nil
	}
	a, ok := v.([]any)
	if !ok {
		return nil, domain.Fail(422, "validation_failed", p, "Expected an array")
	}
	return a, nil
}

var methodPattern = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Z-]+$")

func Fingerprint(target string, o Observation) string {
	names := append([]string{}, o.QueryNames...)
	slices.Sort(names)
	names = slices.Compact(names)
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	_ = e.Encode([]any{target, "zap", o.PluginID, o.Method, o.Origin, o.Path, names, o.Parameter})
	return domain.Hash(sha256.Sum256(bytes.TrimSuffix(b.Bytes(), []byte("\n"))))
}
func Parse(raw []byte, target, origin string) (Report, error) {
	r := Report{Observations: []Observation{}, Warnings: []Warning{}, Unlocated: []UnlocatedAlert{}}
	if len(raw) > MaxBytes {
		return r, domain.Fail(413, "limit_exceeded", "", "Report exceeds 20 MiB")
	}
	r.Hash = domain.Hash(sha256.Sum256(raw))
	v, e := Decode(raw, true)
	if e != nil {
		return r, e
	}
	root, ok := v.(map[string]any)
	if !ok || text(root, "@version") == "" || root["site"] == nil {
		return r, domain.Fail(422, "unsupported_report", "", "Expected ZAP Traditional JSON @version and site array")
	}
	r.Version = text(root, "@version")
	if err := domain.ValidateStorageText(r.Version, "/@version"); err != nil {
		return r, err
	}
	r.Generated = text(root, "@generated")
	sites, e := array(root["site"], "/site", false)
	if e != nil {
		return r, e
	}
	if len(sites) == 0 {
		return r, domain.Fail(422, "unsupported_report", "/site", "At least one site is required")
	}
	for si, sv := range sites {
		sp := fmt.Sprintf("/site/%d", si)
		s, e := object(sv, sp)
		if e != nil {
			return r, e
		}
		loc, e := scope.Parse(text(s, "@name"))
		if e != nil {
			return r, domain.Fail(422, "validation_failed", sp+"/@name", "Invalid site URI")
		}
		if loc.Origin != origin {
			return r, domain.Fail(422, "out_of_scope", sp+"/@name", "Site origin differs from registered target")
		}
		alerts, e := array(s["alerts"], sp+"/alerts", false)
		if e != nil {
			return r, e
		}
		for ai, av := range alerts {
			ap := fmt.Sprintf("%s/alerts/%d", sp, ai)
			a, e := object(av, ap)
			if e != nil {
				return r, e
			}
			instances, e := array(a["instances"], ap+"/instances", true)
			if e != nil {
				return r, e
			}
			if len(instances) == 0 {
				if err := domain.ValidateStorageJSONText(a, ap); err != nil {
					return r, err
				}
				encoded, err := json.Marshal(a)
				if err != nil {
					return r, fmt.Errorf("serialize unlocated alert: %w", err)
				}
				r.Unlocated = append(r.Unlocated, UnlocatedAlert{
					SchemaVersion: "unlocated-alert/v2",
					SourcePointer: ap,
					AlertJSON:     string(encoded),
				})
				r.Warnings = append(r.Warnings, Warning{"unlocated_alert", ap, "Alert has no instances; retained at report level"})
				continue
			}
			title := text(a, "name")
			titleField := "name"
			if title == "" {
				title = text(a, "alert")
				titleField = "alert"
			}
			if title == "" {
				title = "Untitled scanner alert"
			}
			if err := validateStoredFields(a, ap, titleField, "pluginid", "riskcode", "confidence", "desc", "solution", "alertRef"); err != nil {
				return r, err
			}
			plugin := text(a, "pluginid")
			if plugin == "" {
				return r, domain.Fail(422, "validation_failed", ap+"/pluginid", "Plugin ID is required for located alerts")
			}
			severity := "unknown"
			switch text(a, "riskcode") {
			case "0":
				severity = "informational"
			case "1":
				severity = "low"
			case "2":
				severity = "medium"
			case "3":
				severity = "high"
			}
			if severity == "unknown" {
				r.Warnings = append(r.Warnings, Warning{"unknown_severity", ap + "/riskcode", "Unrecognized or absent scanner severity"})
			}
			cwes := []int32{}
			if cw := text(a, "cweid"); cw != "" {
				n, e := strconv.ParseInt(cw, 10, 32)
				if e != nil || n < -1 {
					return r, domain.Fail(422, "validation_failed", ap+"/cweid", "Invalid CWE identifier")
				}
				if n > 0 {
					cwes = append(cwes, int32(n))
				}
			}
			for ii, iv := range instances {
				ip := fmt.Sprintf("%s/instances/%d", ap, ii)
				in, e := object(iv, ip)
				if e != nil {
					return r, e
				}
				uri, _ := in["uri"].(string)
				l, e := scope.Parse(uri)
				if e != nil {
					return r, domain.Fail(422, "validation_failed", ip+"/uri", "Invalid instance URI")
				}
				if l.Origin != origin {
					return r, domain.Fail(422, "out_of_scope", ip+"/uri", "Instance origin differs from registered target")
				}
				rawMethod, _ := in["method"].(string)
				method := strings.ToUpper(rawMethod)
				if len(rawMethod) != len(method) || !methodPattern.MatchString(method) || strings.IndexFunc(rawMethod, func(r rune) bool { return r > 127 }) >= 0 {
					return r, domain.Fail(422, "validation_failed", ip+"/method", "Valid HTTP method is required")
				}
				if err := validateStoredFields(in, ip, "param", "attack", "evidence", "otherinfo"); err != nil {
					return r, err
				}
				evidence, _ := json.Marshal(map[string]any{"schema_version": "zap-evidence/v1", "alert_ref": text(a, "alertRef"), "attack": text(in, "attack"), "evidence": text(in, "evidence"), "other_info": text(in, "otherinfo")})
				o := Observation{PluginID: plugin, Title: title, Method: method, Origin: l.Origin, Path: l.Path, QueryNames: l.QueryNames, Parameter: text(in, "param"), Pointer: ip, Index: int32(ii), RawURI: uri, SeverityRaw: text(a, "riskcode"), Severity: severity, Confidence: text(a, "confidence"), CWE: cwes, Evidence: evidence, Description: text(a, "desc"), Remediation: text(a, "solution")}
				o.Fingerprint = Fingerprint(target, o)
				r.Observations = append(r.Observations, o)
			}
		}
	}
	return r, nil
}

func validateStoredFields(values map[string]any, pointer string, fields ...string) error {
	for _, field := range fields {
		if err := domain.ValidateStorageText(text(values, field), pointer+"/"+field); err != nil {
			return err
		}
	}
	return nil
}
