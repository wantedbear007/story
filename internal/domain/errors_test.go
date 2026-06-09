package domain_test

import (
	"errors"
	"testing"

	"github.com/anomalyco/story/internal/domain"
)

func TestDomainErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", domain.ErrNotFound},
		{"ErrAlreadyExists", domain.ErrAlreadyExists},
		{"ErrInvalidInput", domain.ErrInvalidInput},
		{"ErrUnauthorized", domain.ErrUnauthorized},
		{"ErrForbidden", domain.ErrForbidden},
		{"ErrConflict", domain.ErrConflict},
		{"ErrValidationFailed", domain.ErrValidationFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
				return
			}
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is(%s, %s) should be true", tt.name, tt.name)
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty validation errors returns generic message", func(t *testing.T) {
		ve := domain.ValidationErrors{}
		if ve.Error() != "validation failed" {
			t.Errorf("Error() = %q, want %q", ve.Error(), "validation failed")
		}
	})

	t.Run("non-empty validation errors returns first message", func(t *testing.T) {
		ve := domain.ValidationErrors{
			{Field: "email", Message: "email is required"},
			{Field: "name", Message: "name is required"},
		}
		if ve.Error() != "validation failed: email is required" {
			t.Errorf("Error() = %q, want %q", ve.Error(), "validation failed: email is required")
		}
	})
}
