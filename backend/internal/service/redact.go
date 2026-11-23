package service

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization|cookie|set-cookie)\s*:[^\r\n]*`),
	regexp.MustCompile(`(?i)(password|passwd|secret|token|api[_-]?key|session|access_token|refresh_token)["'\s]*[:=]["'\s]*[^\s&,;<>]+`),
	regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`),
	regexp.MustCompile(`\?[^\s<>"']+`),
	regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`),
}

func Redact(v string) (string, bool) {
	orig := v
	for _, p := range secretPatterns {
		v = p.ReplaceAllString(v, "[REDACTED]")
	}
	if len(v) > 4096 {
		v = v[:4096]
		for !utf8.ValidString(v) {
			v = strings.TrimSuffix(v, v[len(v)-1:])
		}
		v += "…"
	}
	return v, v != orig
}
