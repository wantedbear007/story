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

// CollectionRepository implements domain.CollectionRepository using PostgreSQL via pgx.
type CollectionRepository struct {
	pool *pgxpool.Pool
}

func NewCollectionRepository(pool *pgxpool.Pool) *CollectionRepository {
	return &CollectionRepository{pool: pool}
}

func (r *CollectionRepository) Create(ctx context.Context, col *domain.Collection) error {
	query := `
		INSERT INTO collections (id, user_id, name, description, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	col.CreatedAt = now
	col.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		col.ID, col.UserID, col.Name, col.Description, col.ParentID,
		col.CreatedAt, col.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Collection, error) {
	query := `
		SELECT id, user_id, name, description, parent_id, created_at, updated_at, deleted_at
		FROM collections WHERE id = $1 AND deleted_at IS NULL
	`
	row := r.pool.QueryRow(ctx, query, id)

	col := &domain.Collection{}
	err := row.Scan(
		&col.ID, &col.UserID, &col.Name, &col.Description, &col.ParentID,
		&col.CreatedAt, &col.UpdatedAt, &col.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning collection: %w", err)
	}
	return col, nil
}

func (r *CollectionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Collection, error) {
	query := `
		SELECT id, user_id, name, description, parent_id, created_at, updated_at, deleted_at
		FROM collections WHERE user_id = $1 AND deleted_at IS NULL ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying collections: %w", err)
	}
	defer rows.Close()

	var cols []*domain.Collection
	for rows.Next() {
		col := &domain.Collection{}
		if err := rows.Scan(
			&col.ID, &col.UserID, &col.Name, &col.Description, &col.ParentID,
			&col.CreatedAt, &col.UpdatedAt, &col.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning collection: %w", err)
		}
		cols = append(cols, col)
	}

	if cols == nil {
		cols = make([]*domain.Collection, 0)
	}
	return cols, rows.Err()
}

func (r *CollectionRepository) Update(ctx context.Context, col *domain.Collection) error {
	query := `
		UPDATE collections SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`
	col.UpdatedAt = time.Now()

	tag, err := r.pool.Exec(ctx, query, col.Name, col.Description, col.UpdatedAt, col.ID)
	if err != nil {
		return fmt.Errorf("updating collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CollectionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE collections SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("soft deleting collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CollectionRepository) AddEntry(ctx context.Context, collectionID, entryID uuid.UUID) error {
	query := `INSERT INTO entry_collections (collection_id, entry_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, collectionID, entryID)
	if err != nil {
		return fmt.Errorf("adding entry to collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) RemoveEntry(ctx context.Context, collectionID, entryID uuid.UUID) error {
	query := `DELETE FROM entry_collections WHERE collection_id = $1 AND entry_id = $2`
	_, err := r.pool.Exec(ctx, query, collectionID, entryID)
	if err != nil {
		return fmt.Errorf("removing entry from collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) GetEntries(ctx context.Context, collectionID uuid.UUID) ([]*domain.Entry, error) {
	query := `
		SELECT e.id, e.user_id, e.type, e.title, e.content, e.metadata,
		       e.created_at, e.updated_at, e.deleted_at
		FROM entries e
		INNER JOIN entry_collections ec ON ec.entry_id = e.id
		WHERE ec.collection_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, collectionID)
	if err != nil {
		return nil, fmt.Errorf("querying collection entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.Entry
	for rows.Next() {
		entry := &domain.Entry{}
		var typeStr string
		if err := rows.Scan(
			&entry.ID, &entry.UserID, &typeStr, &entry.Title, &entry.Content,
			&entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning entry: %w", err)
		}
		entry.Type = domain.EntryType(typeStr)
		entries = append(entries, entry)
	}

	if entries == nil {
		entries = make([]*domain.Entry, 0)
	}
	return entries, rows.Err()
}
