package domain_test

import (
	"testing"
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

func TestSession_IsExpired(t *testing.T) {
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
			s := &domain.Session{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				TokenHash: "hash",
				ExpiresAt: tt.expiresAt,
				CreatedAt: time.Now(),
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_IsActive(t *testing.T) {
	t.Parallel()

	t.Run("active when not expired and not revoked", func(t *testing.T) {
		s := &domain.Session{
			IsRevoked: false,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if !s.IsActive() {
			t.Error("IsActive() = false, want true")
		}
	})

	t.Run("not active when expired", func(t *testing.T) {
		s := &domain.Session{
			IsRevoked: false,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if s.IsActive() {
			t.Error("IsActive() = true, want false")
		}
	})

	t.Run("not active when revoked", func(t *testing.T) {
		s := &domain.Session{
			IsRevoked: true,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if s.IsActive() {
			t.Error("IsActive() = true, want false")
		}
	})
}

func TestPasswordResetToken_IsValid(t *testing.T) {
	t.Parallel()

	t.Run("valid when not expired and not used", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if !token.IsValid() {
			t.Error("IsValid() = false, want true")
		}
	})

	t.Run("invalid when expired", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if token.IsValid() {
			t.Error("IsValid() = true, want false")
		}
	})

	t.Run("invalid when used", func(t *testing.T) {
		now := time.Now()
		token := &domain.PasswordResetToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    &now,
		}
		if token.IsValid() {
			t.Error("IsValid() = true, want false")
		}
	})
}

func TestEmailVerification_IsExpired(t *testing.T) {
	t.Parallel()

	t.Run("expired when past expiration", func(t *testing.T) {
		ev := &domain.EmailVerification{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		if !ev.IsExpired() {
			t.Error("IsExpired() = false, want true")
		}
	})

	t.Run("not expired when future", func(t *testing.T) {
		ev := &domain.EmailVerification{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		if ev.IsExpired() {
			t.Error("IsExpired() = true, want false")
		}
	})
}

func TestEmailVerification_IsVerified(t *testing.T) {
	t.Parallel()

	t.Run("verified when verified_at is set", func(t *testing.T) {
		now := time.Now()
		ev := &domain.EmailVerification{
			VerifiedAt: &now,
		}
		if !ev.IsVerified() {
			t.Error("IsVerified() = false, want true")
		}
	})

	t.Run("not verified when verified_at is nil", func(t *testing.T) {
		ev := &domain.EmailVerification{}
		if ev.IsVerified() {
			t.Error("IsVerified() = true, want false")
		}
	})
}
