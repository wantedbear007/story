package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type PromptTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewPromptTemplateRepository(pool *pgxpool.Pool) *PromptTemplateRepository {
	return &PromptTemplateRepository{pool: pool}
}

func (r *PromptTemplateRepository) Create(ctx context.Context, prompt *domain.PromptTemplate) error {
	query := `
		INSERT INTO prompt_templates (id, name, version, template, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if prompt.CreatedAt.IsZero() {
		prompt.CreatedAt = domain.Now()
	}
	_, err := r.pool.Exec(ctx, query,
		prompt.ID, prompt.Name, prompt.Version, prompt.Template, prompt.Description, prompt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting prompt template: %w", err)
	}
	return nil
}

func (r *PromptTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PromptTemplate, error) {
	query := `
		SELECT id, name, version, template, description, created_at
		FROM prompt_templates WHERE id = $1
	`
	return r.scanPrompt(ctx, query, id)
}

func (r *PromptTemplateRepository) GetLatestByName(ctx context.Context, name string) (*domain.PromptTemplate, error) {
	query := `
		SELECT id, name, version, template, description, created_at
		FROM prompt_templates WHERE name = $1
		ORDER BY version DESC
		LIMIT 1
	`
	return r.scanPrompt(ctx, query, name)
}

func (r *PromptTemplateRepository) List(ctx context.Context) ([]*domain.PromptTemplate, error) {
	query := `
		SELECT id, name, version, template, description, created_at
		FROM prompt_templates
		ORDER BY name, version DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying prompt templates: %w", err)
	}
	defer rows.Close()

	var prompts []*domain.PromptTemplate
	for rows.Next() {
		p, err := r.scanPromptFromRow(rows)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating prompt templates: %w", err)
	}
	if prompts == nil {
		prompts = []*domain.PromptTemplate{}
	}
	return prompts, nil
}

func (r *PromptTemplateRepository) scanPrompt(ctx context.Context, query string, args ...interface{}) (*domain.PromptTemplate, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	return r.scanPromptFromRow(row)
}

func (r *PromptTemplateRepository) scanPromptFromRow(row scannable) (*domain.PromptTemplate, error) {
	p := &domain.PromptTemplate{}
	err := row.Scan(&p.ID, &p.Name, &p.Version, &p.Template, &p.Description, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning prompt template: %w", err)
	}
	return p, nil
}
