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

// SessionRepository implements domain.SessionRepository using PostgreSQL via pgx.
type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if token.ExpiresAt.IsZero() {
		token.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	token.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting refresh token: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	token := &domain.RefreshToken{}
	err := row.Scan(
		&token.ID, &token.UserID, &token.TokenHash,
		&token.ExpiresAt, &token.RevokedAt, &token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning refresh token: %w", err)
	}
	return token, nil
}

func (r *SessionRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SessionRepository) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("revoking user refresh tokens: %w", err)
	}
	return nil
}

func (r *SessionRepository) CreatePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	token.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting password reset token: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens WHERE token_hash = $1
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	token := &domain.PasswordResetToken{}
	err := row.Scan(
		&token.ID, &token.UserID, &token.TokenHash,
		&token.ExpiresAt, &token.UsedAt, &token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning password reset token: %w", err)
	}
	return token, nil
}

func (r *SessionRepository) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2 AND used_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("marking password reset token as used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
