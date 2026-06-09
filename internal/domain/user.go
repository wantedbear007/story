package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents the core user entity.
// Authentication is email/password based with Argon2 hashing.
// Soft deletion preserves referential integrity of user-created content.
// EmailVerifiedAt being nil means the user has not verified their email.
type User struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	DisplayName      string     `json:"display_name"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

// IsEmailVerified returns true if the user has completed email verification.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil && !u.EmailVerifiedAt.IsZero()
}

// UserRepository defines persistence contract for User entities.
// Implementations belong in the infrastructure layer.
// The domain layer defines the interface — dependencies point inward.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
