package domain

import "errors"

var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("resource conflict")
	ErrValidationFailed = errors.New("validation failed")
)

type ValidationErrors []ValidationError

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return "validation failed: " + ve[0].Message
}
