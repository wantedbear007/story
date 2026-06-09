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

type ResourceRepository struct {
	pool *pgxpool.Pool
}

func NewResourceRepository(pool *pgxpool.Pool) *ResourceRepository {
	return &ResourceRepository{pool: pool}
}

func (r *ResourceRepository) Create(ctx context.Context, resource *domain.Resource) error {
	query := `
		INSERT INTO resources (id, user_id, type, title, url, description, metadata, content_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now()
	resource.CreatedAt = now
	resource.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		resource.ID, resource.UserID, string(resource.Type), resource.Title, resource.URL,
		resource.Description, resource.Metadata, resource.ContentHash,
		resource.CreatedAt, resource.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting resource: %w", err)
	}
	return nil
}

func (r *ResourceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Resource, error) {
	query := `
		SELECT id, user_id, type, title, url, description, metadata, content_hash, created_at, updated_at
		FROM resources WHERE id = $1
	`
	return r.scanResource(ctx, query, id)
}

func (r *ResourceRepository) List(ctx context.Context, filter domain.ResourceFilter) ([]*domain.Resource, error) {
	args := make([]interface{}, 0)
	conditions := make([]string, 0)
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
	args = append(args, filter.UserID)
	argIdx++

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
			"(to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $%d))",
			argIdx,
		))
		args = append(args, filter.Query)
		argIdx++
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, type, title, url, description, metadata, content_hash, created_at, updated_at
		FROM resources WHERE %s ORDER BY created_at DESC`,
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
		return nil, fmt.Errorf("querying resources: %w", err)
	}
	defer rows.Close()

	var resources []*domain.Resource
	for rows.Next() {
		res, err := r.scanResourceFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning resource row: %w", err)
		}
		resources = append(resources, res)
	}

	if resources == nil {
		resources = make([]*domain.Resource, 0)
	}
	return resources, rows.Err()
}

func (r *ResourceRepository) Update(ctx context.Context, resource *domain.Resource) error {
	query := `
		UPDATE resources SET title = $1, description = $2, metadata = $3, updated_at = $4
		WHERE id = $5
	`
	resource.UpdatedAt = time.Now()

	tag, err := r.pool.Exec(ctx, query,
		resource.Title, resource.Description, resource.Metadata, resource.UpdatedAt, resource.ID,
	)
	if err != nil {
		return fmt.Errorf("updating resource: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ResourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM entry_resources WHERE resource_id = $1`, id)
	if err != nil {
		return fmt.Errorf("removing resource associations: %w", err)
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM resources WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting resource: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ResourceRepository) AttachToEntry(ctx context.Context, entryID, resourceID uuid.UUID) error {
	query := `INSERT INTO entry_resources (entry_id, resource_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, entryID, resourceID)
	if err != nil {
		return fmt.Errorf("attaching resource to entry: %w", err)
	}
	return nil
}

func (r *ResourceRepository) DetachFromEntry(ctx context.Context, entryID, resourceID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM entry_resources WHERE entry_id = $1 AND resource_id = $2`, entryID, resourceID)
	if err != nil {
		return fmt.Errorf("detaching resource from entry: %w", err)
	}
	return nil
}

func (r *ResourceRepository) GetEntryResources(ctx context.Context, entryID uuid.UUID) ([]*domain.Resource, error) {
	query := `
		SELECT res.id, res.user_id, res.type, res.title, res.url, res.description,
		       res.metadata, res.content_hash, res.created_at, res.updated_at
		FROM resources res
		INNER JOIN entry_resources er ON er.resource_id = res.id
		WHERE er.entry_id = $1
		ORDER BY res.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, entryID)
	if err != nil {
		return nil, fmt.Errorf("querying entry resources: %w", err)
	}
	defer rows.Close()

	var resources []*domain.Resource
	for rows.Next() {
		res, err := r.scanResourceFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning resource row: %w", err)
		}
		resources = append(resources, res)
	}

	if resources == nil {
		resources = make([]*domain.Resource, 0)
	}
	return resources, rows.Err()
}

func (r *ResourceRepository) scanResource(ctx context.Context, query string, args ...interface{}) (*domain.Resource, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	return r.scanResourceFromRow(row)
}

func (r *ResourceRepository) scanResourceFromRow(row scannable) (*domain.Resource, error) {
	res := &domain.Resource{}
	var typeStr string
	err := row.Scan(
		&res.ID, &res.UserID, &typeStr, &res.Title, &res.URL,
		&res.Description, &res.Metadata, &res.ContentHash,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning resource: %w", err)
	}
	res.Type = domain.ResourceType(typeStr)
	return res, nil
}
