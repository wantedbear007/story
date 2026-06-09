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

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.DeviceInfo, session.IPAddress,
		session.IsRevoked, session.ExpiresAt, session.LastUsedAt, session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, last_used_at, created_at
		FROM sessions WHERE id = $1
	`
	return r.scanSession(ctx, query, id)
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, last_used_at, created_at
		FROM sessions WHERE token_hash = $1
	`
	return r.scanSession(ctx, query, tokenHash)
}

func (r *SessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, last_used_at, created_at
		FROM sessions WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s, err := r.scanSessionFromRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating sessions: %w", err)
	}
	if sessions == nil {
		sessions = []*domain.Session{}
	}
	return sessions, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET is_revoked = TRUE WHERE id = $1 AND is_revoked = FALSE`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE sessions SET is_revoked = TRUE WHERE user_id = $1 AND is_revoked = FALSE`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoking all sessions: %w", err)
	}
	return nil
}

func (r *SessionRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET last_used_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("updating last used: %w", err)
	}
	return nil
}

func (r *SessionRepository) scanSession(ctx context.Context, query string, args ...interface{}) (*domain.Session, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	return r.scanSessionFromRow(row)
}

func (r *SessionRepository) scanSessionFromRow(row scannable) (*domain.Session, error) {
	s := &domain.Session{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.DeviceInfo, &s.IPAddress,
		&s.IsRevoked, &s.ExpiresAt, &s.LastUsedAt, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning session: %w", err)
	}
	return s, nil
}
