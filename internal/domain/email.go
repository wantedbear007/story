package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EmailVerification tracks the email verification process.
// A token is created when a user registers or changes their email.
// The token has an expiration and is one-time-use (consumed on verify).
type EmailVerification struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Email      string     `json:"email"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IsExpired returns true if the verification window has passed.
func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.ExpiresAt)
}

// IsVerified returns true if the email has been confirmed.
func (ev *EmailVerification) IsVerified() bool {
	return ev.VerifiedAt != nil && !ev.VerifiedAt.IsZero()
}

// EmailVerificationRepository defines persistence for email verification tokens.
type EmailVerificationRepository interface {
	Create(ctx context.Context, ev *EmailVerification) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*EmailVerification, error)
	MarkVerified(ctx context.Context, id uuid.UUID) error
}
