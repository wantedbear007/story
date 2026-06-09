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

// TagRepository implements domain.TagRepository using PostgreSQL via pgx.
type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

func (r *TagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	query := `
		INSERT INTO tags (id, user_id, name, color, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, name) DO NOTHING
		RETURNING id
	`
	now := time.Now()
	tag.CreatedAt = now

	err := r.pool.QueryRow(ctx, query,
		tag.ID, tag.UserID, tag.Name, tag.Color, tag.CreatedAt,
	).Scan(&tag.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: tag '%s' already exists", domain.ErrAlreadyExists, tag.Name)
		}
		return fmt.Errorf("inserting tag: %w", err)
	}
	return nil
}

func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	query := `SELECT id, user_id, name, color, created_at FROM tags WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	tag := &domain.Tag{}
	err := row.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning tag: %w", err)
	}
	return tag, nil
}

func (r *TagRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Tag, error) {
	query := `SELECT id, user_id, name, color, created_at FROM tags WHERE user_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if tags == nil {
		tags = make([]*domain.Tag, 0)
	}
	return tags, rows.Err()
}

func (r *TagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	query := `UPDATE tags SET name = $1, color = $2 WHERE id = $3`
	ct, err := r.pool.Exec(ctx, query, tag.Name, tag.Color, tag.ID)
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM entry_tags WHERE tag_id = $1`, id)
	if err != nil {
		return fmt.Errorf("removing tag associations: %w", err)
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TagRepository) AddTagToEntry(ctx context.Context, entryID, tagID uuid.UUID) error {
	query := `INSERT INTO entry_tags (entry_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, entryID, tagID)
	if err != nil {
		return fmt.Errorf("adding tag to entry: %w", err)
	}
	return nil
}

func (r *TagRepository) RemoveTagFromEntry(ctx context.Context, entryID, tagID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM entry_tags WHERE entry_id = $1 AND tag_id = $2`, entryID, tagID)
	if err != nil {
		return fmt.Errorf("removing tag from entry: %w", err)
	}
	return nil
}

func (r *TagRepository) GetEntryTags(ctx context.Context, entryID uuid.UUID) ([]*domain.Tag, error) {
	query := `
		SELECT t.id, t.user_id, t.name, t.color, t.created_at
		FROM tags t
		INNER JOIN entry_tags et ON et.tag_id = t.id
		WHERE et.entry_id = $1
		ORDER BY t.name
	`
	rows, err := r.pool.Query(ctx, query, entryID)
	if err != nil {
		return nil, fmt.Errorf("querying entry tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if tags == nil {
		tags = make([]*domain.Tag, 0)
	}
	return tags, rows.Err()
}
