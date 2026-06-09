package domain_test

import (
	"testing"
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

func TestRefreshToken_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired when expiration is in the past",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "not expired when expiration is in the future",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &domain.RefreshToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				TokenHash: "hash",
				ExpiresAt: tt.expiresAt,
				CreatedAt: time.Now(),
			}
			if got := token.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	t.Parallel()

	t.Run("revoked when revoked_at is set", func(t *testing.T) {
		now := time.Now()
		token := &domain.RefreshToken{
			RevokedAt: &now,
		}
		if !token.IsRevoked() {
			t.Error("IsRevoked() = false, want true")
		}
	})

	t.Run("not revoked when revoked_at is nil", func(t *testing.T) {
		token := &domain.RefreshToken{}
		if token.IsRevoked() {
			t.Error("IsRevoked() = true, want false")
		}
	})
}

func TestRefreshToken_IsValid(t *testing.T) {
	t.Parallel()

	t.Run("valid when not expired and not revoked", func(t *testing.T) {
		token := &domain.RefreshToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if !token.IsValid() {
			t.Error("IsValid() = false, want true")
		}
	})

	t.Run("invalid when expired", func(t *testing.T) {
		token := &domain.RefreshToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if token.IsValid() {
			t.Error("IsValid() = true, want false")
		}
	})

	t.Run("invalid when revoked", func(t *testing.T) {
		now := time.Now()
		token := &domain.RefreshToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
			RevokedAt: &now,
		}
		if token.IsValid() {
			t.Error("IsValid() = true, want false")
		}
	})
}
