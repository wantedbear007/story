package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// EntryRepository implements domain.EntryRepository using PostgreSQL via pgx.
type EntryRepository struct {
	pool *pgxpool.Pool
}

func NewEntryRepository(pool *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{pool: pool}
}

func (r *EntryRepository) Create(ctx context.Context, entry *domain.Entry) error {
	query := `
		INSERT INTO entries (id, user_id, type, title, content, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.UserID, string(entry.Type), entry.Title, entry.Content,
		entry.Metadata, entry.CreatedAt, entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting entry: %w", err)
	}
	return nil
}

func (r *EntryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Entry, error) {
	query := `
		SELECT id, user_id, type, title, content, metadata, created_at, updated_at, deleted_at
		FROM entries
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.scanEntry(ctx, query, id)
}

func (r *EntryRepository) List(ctx context.Context, filter domain.EntryFilter) ([]*domain.Entry, error) {
	args := make([]interface{}, 0)
	conditions := make([]string, 0)
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
	args = append(args, filter.UserID)
	argIdx++

	conditions = append(conditions, "deleted_at IS NULL")

	if len(filter.Types) > 0 {
		typePlaceholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			typePlaceholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, string(t))
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(typePlaceholders, ",")))
	}

	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $%d))",
			argIdx,
		))
		args = append(args, filter.Query)
		argIdx++
	}

	if filter.CollectionID != nil {
		conditions = append(conditions, fmt.Sprintf(
			"id IN (SELECT entry_id FROM entry_collections WHERE collection_id = $%d)",
			argIdx,
		))
		args = append(args, *filter.CollectionID)
		argIdx++
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, type, title, content, metadata, created_at, updated_at, deleted_at
		FROM entries WHERE %s ORDER BY created_at DESC`,
		strings.Join(conditions, " AND "),
	)

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.Entry
	for rows.Next() {
		entry, err := r.scanEntryFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning entry row: %w", err)
		}
		entries = append(entries, entry)
	}

	if entries == nil {
		entries = make([]*domain.Entry, 0)
	}

	return entries, rows.Err()
}

func (r *EntryRepository) Update(ctx context.Context, entry *domain.Entry) error {
	query := `
		UPDATE entries
		SET type = $1, title = $2, content = $3, metadata = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`
	entry.UpdatedAt = time.Now()

	tag, err := r.pool.Exec(ctx, query,
		string(entry.Type), entry.Title, entry.Content, entry.Metadata,
		entry.UpdatedAt, entry.ID,
	)
	if err != nil {
		return fmt.Errorf("updating entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *EntryRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE entries SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("soft deleting entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *EntryRepository) scanEntry(ctx context.Context, query string, args ...interface{}) (*domain.Entry, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	return r.scanEntryFromRow(row)
}

func (r *EntryRepository) scanEntryFromRow(row scannable) (*domain.Entry, error) {
	entry := &domain.Entry{}
	var typeStr string
	err := row.Scan(
		&entry.ID, &entry.UserID, &typeStr, &entry.Title, &entry.Content,
		&entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning entry: %w", err)
	}
	entry.Type = domain.EntryType(typeStr)
	return entry, nil
}
