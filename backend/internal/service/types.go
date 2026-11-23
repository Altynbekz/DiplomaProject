package service

import (
	"diploma/internal/zap"
	"time"
)

type Target struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Origin    string    `json:"canonical_origin"`
	Key       string    `json:"lab_target_key"`
	CreatedAt time.Time `json:"created_at"`
}
type Provenance struct {
	SchemaVersion     string            `json:"schema_version"`
	ZAPVersion        *string           `json:"zap_version"`
	AddonVersions     map[string]string `json:"addon_versions"`
	ReportTemplate    *string           `json:"report_template"`
	JuiceShopVersion  *string           `json:"juice_shop_version"`
	JuiceShopImage    *string           `json:"juice_shop_image"`
	AuthorizationRef  *string           `json:"authorization_ref"`
	ScanTime          *string           `json:"scan_time"`
	PlanSHA256        *string           `json:"plan_sha256"`
	RawSHA256         *string           `json:"raw_sha256"`
	SanitizedSHA256   *string           `json:"sanitized_sha256"`
	SanitizationSteps []string          `json:"sanitization_steps"`
}
type Import struct {
	ID                    string        `json:"id"`
	TargetID              string        `json:"target_id"`
	Hash                  string        `json:"report_sha256"`
	ReportDeclaredVersion *string       `json:"report_declared_version"`
	ParserDetectedFormat  string        `json:"parser_detected_format"`
	ParserVersion         string        `json:"parser_version"`
	Duplicate             bool          `json:"duplicate"`
	FindingCount          int32         `json:"finding_count"`
	ObservationCount      int32         `json:"observation_count"`
	UnlocatedCount        int           `json:"unlocated_alert_count"`
	Warnings              []zap.Warning `json:"warnings"`
	Provenance            Provenance    `json:"provenance"`
	CreatedAt             time.Time     `json:"created_at"`
}
type Classification struct {
	RevisionID     *string  `json:"revision_id"`
	Status         string   `json:"status"`
	Category       *string  `json:"category"`
	Candidates     []string `json:"candidate_categories"`
	MappingVersion *string  `json:"mapping_version"`
}
type Assessment struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	Score     *float64  `json:"score"`
	Vector    *string   `json:"vector"`
	Stale     bool      `json:"stale"`
	CreatedAt time.Time `json:"created_at"`
}
type Finding struct {
	ID                    string         `json:"id"`
	TargetID              string         `json:"target_id"`
	Title                 string         `json:"title"`
	Method                string         `json:"method"`
	Origin                string         `json:"origin"`
	Path                  string         `json:"path"`
	Parameter             string         `json:"parameter_name"`
	QueryNames            []string       `json:"query_parameter_names"`
	ObservationCount      int64          `json:"observation_count"`
	Severity              string         `json:"scanner_severity"`
	SeverityObservationID string         `json:"severity_observation_id"`
	Verification          string         `json:"verification"`
	ReviewID              *string        `json:"verification_review_id"`
	Classification        Classification `json:"classification"`
	Base                  *Assessment    `json:"latest_base"`
	Environmental         *Assessment    `json:"latest_environmental"`
	PluginID              string         `json:"plugin_id"`
	FingerprintVersion    string         `json:"fingerprint_version"`
	Fingerprint           string         `json:"fingerprint"`
	CreatedAt             time.Time      `json:"created_at"`
}
type Observation struct {
	ID          string    `json:"id"`
	ImportID    string    `json:"import_id"`
	Pointer     string    `json:"source_pointer"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Parameter   string    `json:"parameter_name"`
	Severity    string    `json:"scanner_severity"`
	SeverityRaw string    `json:"scanner_severity_raw"`
	Confidence  string    `json:"confidence_raw"`
	CWE         []int32   `json:"cwe_ids"`
	Description string    `json:"description"`
	Remediation string    `json:"remediation"`
	Excerpt     string    `json:"evidence_excerpt"`
	Redactions  []string  `json:"redactions"`
	CreatedAt   time.Time `json:"created_at"`
}
type Page[T any] struct {
	Items []T     `json:"items"`
	Next  *string `json:"next_cursor"`
}
