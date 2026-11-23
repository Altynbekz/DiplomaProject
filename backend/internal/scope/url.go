// Package scope validates report locations offline. It never resolves or fetches URLs.
package scope

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Location struct {
	Origin, Path string
	QueryNames   []string
}

func Parse(raw string) (Location, error) {
	bad := func() (Location, error) { return Location{}, fmt.Errorf("invalid HTTP(S) location") }
	if strings.Contains(raw, "#") || strings.ContainsAny(raw, "\\\r\n\t ") {
		return bad()
	}
	u, e := url.Parse(raw)
	if e != nil || u.Opaque != "" || u.User != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return bad()
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || strings.Contains(host, "%") || strings.HasSuffix(host, ".") {
		return bad()
	}
	for _, r := range host {
		if r > 127 || !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || strings.ContainsRune("-.:", r)) {
			return bad()
		}
	}
	if strings.Contains(host, ":") {
		if net.ParseIP(host) == nil {
			return bad()
		}
	} else {
		if ip := net.ParseIP(host); ip != nil {
			if ip.String() != host {
				return bad()
			}
		} else {
			numeric := true
			for _, r := range host {
				if !(r >= '0' && r <= '9' || r == '.') {
					numeric = false
				}
			}
			if numeric || strings.HasPrefix(host, "0x") {
				return bad()
			}
			for _, label := range strings.Split(host, ".") {
				if label == "" || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
					return bad()
				}
			}
		}
	}
	port := u.Port()
	if port == "" {
		if strings.HasSuffix(u.Host, ":") {
			return bad()
		}
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	n, e := strconv.Atoi(port)
	if e != nil || n < 1 || n > 65535 {
		return bad()
	}
	port = strconv.Itoa(n)
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	var b strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] != '%' {
			b.WriteByte(path[i])
			continue
		}
		if i+2 >= len(path) {
			return bad()
		}
		v, e := strconv.ParseUint(path[i+1:i+3], 16, 8)
		if e != nil {
			return bad()
		}
		c := byte(v)
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("-._~", rune(c)) {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
		i += 2
	}
	// Split only on the query-pair delimiter. Query values are private
	// evidence and do not affect endpoint identity, even if malformed.
	seen := map[string]bool{}
	names := []string{}
	for pair := range strings.SplitSeq(u.RawQuery, "&") {
		if pair == "" {
			continue
		}
		encodedName, _, _ := strings.Cut(pair, "=")
		name, err := url.QueryUnescape(encodedName)
		if err != nil || !utf8.ValidString(name) || strings.ContainsRune(name, 0) {
			return bad()
		}
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	sort.Strings(names)
	return Location{u.Scheme + "://" + net.JoinHostPort(host, port), b.String(), names}, nil
}
