package errors_test

import (
	"testing"

	"github.com/anomalyco/story/internal/pkg/errors"
)

func TestNew(t *testing.T) {
	t.Parallel()

	err := errors.New(errors.CodeNotFound, "user not found")
	if err.Code != errors.CodeNotFound {
		t.Errorf("Code = %q, want %q", err.Code, errors.CodeNotFound)
	}
	if err.Message != "user not found" {
		t.Errorf("Message = %q, want %q", err.Message, "user not found")
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	inner := errors.New(errors.CodeInvalidInput, "bad request")
	wrapped := errors.Wrap(errors.CodeInternal, "process", inner)

	if wrapped.Code != errors.CodeInternal {
		t.Errorf("Code = %q, want %q", wrapped.Code, errors.CodeInternal)
	}
	if wrapped.Op != "process" {
		t.Errorf("Op = %q, want %q", wrapped.Op, "process")
	}
	if wrapped.Err != inner {
		t.Error("Err should be the wrapped error")
	}
}

func TestError_Is(t *testing.T) {
	t.Parallel()

	err := errors.New(errors.CodeNotFound, "resource not found")

	if !errors.IsNotFound(err) {
		t.Error("IsNotFound should return true")
	}
	if errors.IsAlreadyExists(err) {
		t.Error("IsAlreadyExists should return false")
	}
}

func TestExtractCode(t *testing.T) {
	t.Parallel()

	inner := errors.New(errors.CodeUnauthorized, "not allowed")
	wrapped := errors.Wrap(errors.CodeInternal, "handler", inner)

	code := errors.ExtractCode(wrapped)
	if code != errors.CodeInternal {
		t.Errorf("ExtractCode = %q, want %q", code, errors.CodeInternal)
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		ve := errors.ValidationErrors{}
		if ve.Error() != "validation failed" {
			t.Errorf("Error() = %q, want %q", ve.Error(), "validation failed")
		}
	})

	t.Run("with items", func(t *testing.T) {
		ve := errors.ValidationErrors{
			{Field: "email", Message: "email is required"},
		}
		if ve.Error() != "validation failed: email is required" {
			t.Errorf("Error() = %q, want %q", ve.Error(), "validation failed: email is required")
		}
	})
}
