package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshToken enables stateless JWT access tokens with stateful refresh.
// TokenHash stores Argon2 hash of the raw token — never the raw value.
// Revocation supports explicit logout and compromise response.
type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsExpired returns true if the token has passed its expiration time.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked returns true if the token has been explicitly revoked.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid returns true if the token is neither expired nor revoked.
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}

// PasswordResetToken facilitates secure password reset flows.
type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type SessionRepository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error

	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error
}
