package errors

import (
	"fmt"
	"strings"
)

// Code classifies the type of error for programmatic handling.
type Code string

const (
	CodeNotFound        Code = "NOT_FOUND"
	CodeAlreadyExists   Code = "ALREADY_EXISTS"
	CodeInvalidInput    Code = "INVALID_INPUT"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeConflict        Code = "CONFLICT"
	CodeInternal        Code = "INTERNAL"
	CodeUnavailable     Code = "UNAVAILABLE"
	CodeRateLimited     Code = "RATE_LIMITED"
	CodeNotImplemented  Code = "NOT_IMPLEMENTED"
)

// Severity indicates how critical the error is.
type Severity int

const (
	SeverityDebug   Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// Error is the canonical application error type.
// It wraps an underlying error with classification, operation context,
// and human-readable message — all without leaking implementation details.
type Error struct {
	Code     Code
	Message  string
	Op       string  // The operation being performed when the error occurred
	Severity Severity
	Err      error   // The wrapped underlying error
}

func (e *Error) Error() string {
	var b strings.Builder
	if e.Op != "" {
		b.WriteString(e.Op)
		b.WriteString(": ")
	}
	if e.Message != "" {
		b.WriteString(e.Message)
	}
	if e.Err != nil {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is allows errors.Is to match on Code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New builds a classified error.
func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf builds a classified error with a formatted message.
func Newf(code Code, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with classification and context.
func Wrap(code Code, op string, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code: code,
		Op:   op,
		Err:  err,
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(code Code, op string, err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Op:      op,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// IsNotFound checks if the error has a NOT_FOUND code.
func IsNotFound(err error) bool {
	return hasCode(err, CodeNotFound)
}

// IsAlreadyExists checks if the error has an ALREADY_EXISTS code.
func IsAlreadyExists(err error) bool {
	return hasCode(err, CodeAlreadyExists)
}

// IsInvalidInput checks if the error has an INVALID_INPUT code.
func IsInvalidInput(err error) bool {
	return hasCode(err, CodeInvalidInput)
}

// IsUnauthorized checks if the error has an UNAUTHORIZED code.
func IsUnauthorized(err error) bool {
	return hasCode(err, CodeUnauthorized)
}

// IsInternal checks if the error has an INTERNAL code.
func IsInternal(err error) bool {
	return hasCode(err, CodeInternal)
}

func hasCode(err error, code Code) bool {
	if err == nil {
		return false
	}
	if as := asError(err); as != nil {
		return as.Code == code
	}
	return false
}

// ExtractCode extracts the error code from an error chain.
// Returns CodeInternal if no classified error is found.
func ExtractCode(err error) Code {
	if err == nil {
		return ""
	}
	if as := asError(err); as != nil {
		return as.Code
	}
	return CodeInternal
}

// ErrNotFound creates a NOT_FOUND error with the given message.
func ErrNotFound(message string) *Error {
	return New(CodeNotFound, message)
}

// ErrAlreadyExists creates an ALREADY_EXISTS error.
func ErrAlreadyExists(message string) *Error {
	return New(CodeAlreadyExists, message)
}

// ErrInvalidInput creates an INVALID_INPUT error.
func ErrInvalidInput(message string) *Error {
	return New(CodeInvalidInput, message)
}

// ErrUnauthorized creates an UNAUTHORIZED error.
func ErrUnauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

// ErrForbidden creates a FORBIDDEN error.
func ErrForbidden(message string) *Error {
	return New(CodeForbidden, message)
}

// ErrConflict creates a CONFLICT error.
func ErrConflict(message string) *Error {
	return New(CodeConflict, message)
}

// ErrInternal creates an INTERNAL error.
func ErrInternal(message string) *Error {
	return New(CodeInternal, message)
}

// ValidationErrors aggregates multiple validation failures.
type ValidationErrors []ValidationError

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", ve[0].Message)
}

func asError(err error) *Error {
	var e *Error
	// Walk the chain looking for an *Error.
	for err != nil {
		var ok bool
		if e, ok = err.(*Error); ok {
			return e
		}
		err = unwrap(err)
	}
	return nil
}

func unwrap(err error) error {
	u, ok := err.(interface{ Unwrap() error })
	if !ok {
		return nil
	}
	return u.Unwrap()
}
