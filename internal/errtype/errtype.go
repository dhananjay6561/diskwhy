package errtype

import "fmt"

// CodedError carries an exit code alongside its message so the top-level
// Execute function can exit with the right code without inspecting error strings.
type CodedError struct {
	Code int
	Msg  string
	Fix  string
}

func (e *CodedError) Error() string {
	if e.Fix != "" {
		return fmt.Sprintf("%s\nFix: %s", e.Msg, e.Fix)
	}
	return e.Msg
}

func New(code int, msg, fix string) *CodedError {
	return &CodedError{Code: code, Msg: msg, Fix: fix}
}
