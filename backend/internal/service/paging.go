package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"diploma/internal/domain"
	"encoding/base64"
	"encoding/json"
	"strings"
)

type Cursor struct {
	ID     string `json:"id"`
	Time   string `json:"time"`
	Filter string `json:"filter"`
}

func (s *Service) Cursor(raw, filter string) (Cursor, error) {
	bad := func() (Cursor, error) {
		return Cursor{}, domain.Fail(400, "invalid_request", "/cursor", "Invalid or mismatched cursor")
	}
	if raw == "" {
		return Cursor{}, nil
	}
	if len(raw) > 2048 {
		return bad()
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return bad()
	}
	b, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return bad()
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return bad()
	}
	m := hmac.New(sha256.New, []byte(s.CursorKey))
	m.Write(b)
	if !hmac.Equal(sig, m.Sum(nil)) {
		return bad()
	}
	var c Cursor
	if json.Unmarshal(b, &c) != nil || c.Filter != filter {
		return bad()
	}
	return c, nil
}
func (s *Service) Next(c Cursor) *string {
	b, _ := json.Marshal(c)
	m := hmac.New(sha256.New, []byte(s.CursorKey))
	m.Write(b)
	v := base64.RawURLEncoding.EncodeToString(b) + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
	return &v
}
func Filter(parts ...string) string {
	b, _ := json.Marshal(parts)
	return domain.Hash(sha256.Sum256(b))
}
