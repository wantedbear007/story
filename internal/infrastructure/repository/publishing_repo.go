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

// PublishingTargetRepository implements domain.PublishingTargetRepository.
type PublishingTargetRepository struct {
	pool *pgxpool.Pool
}

func NewPublishingTargetRepository(pool *pgxpool.Pool) *PublishingTargetRepository {
	return &PublishingTargetRepository{pool: pool}
}

func (r *PublishingTargetRepository) Create(ctx context.Context, target *domain.PublishingTarget) error {
	query := `
		INSERT INTO publishing_targets (id, user_id, type, name, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	target.CreatedAt = now
	target.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		target.ID, target.UserID, string(target.Type), target.Name, target.Config,
		target.CreatedAt, target.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting publishing target: %w", err)
	}
	return nil
}

func (r *PublishingTargetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PublishingTarget, error) {
	query := `SELECT id, user_id, type, name, config, created_at, updated_at FROM publishing_targets WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	target := &domain.PublishingTarget{}
	var typeStr string
	err := row.Scan(
		&target.ID, &target.UserID, &typeStr, &target.Name, &target.Config,
		&target.CreatedAt, &target.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning publishing target: %w", err)
	}
	target.Type = domain.PublishingTargetType(typeStr)
	return target, nil
}

func (r *PublishingTargetRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.PublishingTarget, error) {
	query := `SELECT id, user_id, type, name, config, created_at, updated_at FROM publishing_targets WHERE user_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying publishing targets: %w", err)
	}
	defer rows.Close()

	var targets []*domain.PublishingTarget
	for rows.Next() {
		target := &domain.PublishingTarget{}
		var typeStr string
		if err := rows.Scan(
			&target.ID, &target.UserID, &typeStr, &target.Name, &target.Config,
			&target.CreatedAt, &target.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning publishing target: %w", err)
		}
		target.Type = domain.PublishingTargetType(typeStr)
		targets = append(targets, target)
	}

	if targets == nil {
		targets = make([]*domain.PublishingTarget, 0)
	}
	return targets, rows.Err()
}

func (r *PublishingTargetRepository) Update(ctx context.Context, target *domain.PublishingTarget) error {
	query := `UPDATE publishing_targets SET type = $1, name = $2, config = $3, updated_at = $4 WHERE id = $5`
	target.UpdatedAt = time.Now()

	tag, err := r.pool.Exec(ctx, query, string(target.Type), target.Name, target.Config, target.UpdatedAt, target.ID)
	if err != nil {
		return fmt.Errorf("updating publishing target: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PublishingTargetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM publishing_targets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting publishing target: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// PublishedEntryRepository implements domain.PublishedEntryRepository.
type PublishedEntryRepository struct {
	pool *pgxpool.Pool
}

func NewPublishedEntryRepository(pool *pgxpool.Pool) *PublishedEntryRepository {
	return &PublishedEntryRepository{pool: pool}
}

func (r *PublishedEntryRepository) Create(ctx context.Context, pe *domain.PublishedEntry) error {
	query := `
		INSERT INTO published_entries (id, entry_id, target_id, external_url, status, error_message, published_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	pe.CreatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		pe.ID, pe.EntryID, pe.TargetID, pe.ExternalURL, string(pe.Status),
		pe.ErrorMessage, pe.PublishedAt, pe.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting published entry: %w", err)
	}
	return nil
}

func (r *PublishedEntryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PublishedEntry, error) {
	query := `SELECT id, entry_id, target_id, external_url, status, error_message, published_at, created_at FROM published_entries WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return r.scanPublishedEntry(row)
}

func (r *PublishedEntryRepository) ListByEntry(ctx context.Context, entryID uuid.UUID) ([]*domain.PublishedEntry, error) {
	query := `SELECT id, entry_id, target_id, external_url, status, error_message, published_at, created_at FROM published_entries WHERE entry_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, entryID)
	if err != nil {
		return nil, fmt.Errorf("querying published entries: %w", err)
	}
	defer rows.Close()

	return r.scanPublishedEntries(rows)
}

func (r *PublishedEntryRepository) ListByTarget(ctx context.Context, targetID uuid.UUID) ([]*domain.PublishedEntry, error) {
	query := `SELECT id, entry_id, target_id, external_url, status, error_message, published_at, created_at FROM published_entries WHERE target_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, targetID)
	if err != nil {
		return nil, fmt.Errorf("querying published entries by target: %w", err)
	}
	defer rows.Close()

	return r.scanPublishedEntries(rows)
}

func (r *PublishedEntryRepository) Update(ctx context.Context, pe *domain.PublishedEntry) error {
	query := `
		UPDATE published_entries
		SET external_url = $1, status = $2, error_message = $3, published_at = $4
		WHERE id = $5
	`
	tag, err := r.pool.Exec(ctx, query, pe.ExternalURL, string(pe.Status), pe.ErrorMessage, pe.PublishedAt, pe.ID)
	if err != nil {
		return fmt.Errorf("updating published entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PublishedEntryRepository) scanPublishedEntry(row scannable) (*domain.PublishedEntry, error) {
	pe := &domain.PublishedEntry{}
	var statusStr string
	err := row.Scan(
		&pe.ID, &pe.EntryID, &pe.TargetID, &pe.ExternalURL, &statusStr,
		&pe.ErrorMessage, &pe.PublishedAt, &pe.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning published entry: %w", err)
	}
	pe.Status = domain.PublishStatus(statusStr)
	return pe, nil
}

func (r *PublishedEntryRepository) scanPublishedEntries(rows pgx.Rows) ([]*domain.PublishedEntry, error) {
	var entries []*domain.PublishedEntry
	for rows.Next() {
		pe := &domain.PublishedEntry{}
		var statusStr string
		if err := rows.Scan(
			&pe.ID, &pe.EntryID, &pe.TargetID, &pe.ExternalURL, &statusStr,
			&pe.ErrorMessage, &pe.PublishedAt, &pe.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning published entry: %w", err)
		}
		pe.Status = domain.PublishStatus(statusStr)
		entries = append(entries, pe)
	}

	if entries == nil {
		entries = make([]*domain.PublishedEntry, 0)
	}
	return entries, rows.Err()
}
