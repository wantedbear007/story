package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type RawEntryRepository struct {
	pool *pgxpool.Pool
}

func NewRawEntryRepository(pool *pgxpool.Pool) *RawEntryRepository {
	return &RawEntryRepository{pool: pool}
}

func (r *RawEntryRepository) Create(ctx context.Context, entry *domain.RawEntry) error {
	query := `
		INSERT INTO raw_entries (id, user_id, content, status, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.UserID, entry.Content, string(entry.Status), string(entry.Source),
		entry.CreatedAt, entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting raw entry: %w", err)
	}
	return nil
}

func (r *RawEntryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RawEntry, error) {
	query := `
		SELECT id, user_id, content, status, source, created_at, updated_at
		FROM raw_entries
		WHERE id = $1
	`
	return r.scanRawEntry(ctx, query, id)
}

func (r *RawEntryRepository) List(ctx context.Context, filter domain.RawEntryFilter) ([]*domain.RawEntry, error) {
	args := make([]interface{}, 0)
	conditions := make([]string, 0)
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
	args = append(args, filter.UserID)
	argIdx++

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.Source != nil {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIdx))
		args = append(args, string(*filter.Source))
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, content, status, source, created_at, updated_at
		FROM raw_entries
		WHERE %s
		ORDER BY created_at DESC
	`, joinConditions(conditions))

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
		return nil, fmt.Errorf("listing raw entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.RawEntry
	for rows.Next() {
		entry, err := scanRawEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning raw entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if entries == nil {
		entries = make([]*domain.RawEntry, 0)
	}
	return entries, nil
}

func (r *RawEntryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RawEntryStatus) error {
	query := `
		UPDATE raw_entries
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.pool.Exec(ctx, query, string(status), time.Now(), id)
	if err != nil {
		return fmt.Errorf("updating raw entry status: %w", err)
	}
	return nil
}

func (r *RawEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM raw_entries WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting raw entry: %w", err)
	}
	return nil
}

func (r *RawEntryRepository) scanRawEntry(ctx context.Context, query string, args ...interface{}) (*domain.RawEntry, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	entry, err := scanRawEntry(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning raw entry: %w", err)
	}
	return entry, nil
}

func scanRawEntry(row interface{ Scan(dest ...interface{}) error }) (*domain.RawEntry, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		content   string
		status    string
		source    string
		createdAt time.Time
		updatedAt time.Time
	)
	err := row.Scan(&id, &userID, &content, &status, &source, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return &domain.RawEntry{
		ID:        id,
		UserID:    userID,
		Content:   content,
		Status:    domain.RawEntryStatus(status),
		Source:    domain.RawEntrySource(source),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func joinConditions(conditions []string) string {
	if len(conditions) == 0 {
		return "TRUE"
	}
	result := conditions[0]
	for _, c := range conditions[1:] {
		result += " AND " + c
	}
	return result
}
