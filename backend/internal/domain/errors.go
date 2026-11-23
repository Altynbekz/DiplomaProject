package domain

import "fmt"

type FieldError struct {
	Pointer string `json:"pointer"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Problem struct {
	Status int          `json:"status"`
	Code   string       `json:"code"`
	Detail string       `json:"detail"`
	Errors []FieldError `json:"errors"`
}

func (p *Problem) Error() string { return p.Code + ": " + p.Detail }
func Fail(status int, code, pointer, message string) error {
	return &Problem{status, code, message, []FieldError{{pointer, code, message}}}
}
func Hash(v [32]byte) string { return fmt.Sprintf("%x", v) }
