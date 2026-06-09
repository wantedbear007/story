package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type EmailVerificationRepository struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{pool: pool}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, ev *domain.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (id, user_id, email, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		ev.ID, ev.UserID, ev.Email, ev.TokenHash, ev.ExpiresAt, ev.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting email verification: %w", err)
	}
	return nil
}

func (r *EmailVerificationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.EmailVerification, error) {
	query := `
		SELECT id, user_id, email, token_hash, expires_at, verified_at, created_at
		FROM email_verifications WHERE token_hash = $1
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	ev := &domain.EmailVerification{}
	err := row.Scan(
		&ev.ID, &ev.UserID, &ev.Email, &ev.TokenHash,
		&ev.ExpiresAt, &ev.VerifiedAt, &ev.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning email verification: %w", err)
	}
	return ev, nil
}

func (r *EmailVerificationRepository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE email_verifications SET verified_at = $1 WHERE id = $2 AND verified_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("marking email verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
