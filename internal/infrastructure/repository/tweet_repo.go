package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type TweetRepository struct {
	pool *pgxpool.Pool
}

func NewTweetRepository(pool *pgxpool.Pool) *TweetRepository {
	return &TweetRepository{pool: pool}
}

func (r *TweetRepository) Create(ctx context.Context, tweet *domain.Tweet) error {
	query := `
		INSERT INTO tweets (
			id, entry_id, user_id, content, status, version, prompt_id,
			provider_name, model_name, input_tokens, output_tokens, cost_usd,
			retry_count, latency_ms, error_message, scheduled_for, posted_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19
		)
	`
	now := domain.Now()
	if tweet.CreatedAt.IsZero() {
		tweet.CreatedAt = now
	}
	tweet.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		tweet.ID, tweet.EntryID, tweet.UserID, tweet.Content, tweet.Status, tweet.Version, nullableUUID(tweet.PromptID),
		tweet.ProviderName, tweet.ModelName, tweet.InputTokens, tweet.OutputTokens, tweet.CostUSD,
		tweet.RetryCount, tweet.LatencyMs, tweet.ErrorMessage, tweet.ScheduledFor, tweet.PostedAt,
		tweet.CreatedAt, tweet.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting tweet: %w", err)
	}
	return nil
}

func (r *TweetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tweet, error) {
	query := `
		SELECT id, entry_id, user_id, content, status, version, prompt_id,
		       provider_name, model_name, input_tokens, output_tokens, cost_usd,
		       retry_count, latency_ms, error_message, scheduled_for, posted_at,
		       created_at, updated_at
		FROM tweets WHERE id = $1
	`
	return r.scanTweet(ctx, query, id)
}

func (r *TweetRepository) List(ctx context.Context, filter domain.TweetFilter) ([]*domain.Tweet, error) {
	args := []interface{}{filter.UserID}
	argIdx := 2

	query := `
		SELECT id, entry_id, user_id, content, status, version, prompt_id,
		       provider_name, model_name, input_tokens, output_tokens, cost_usd,
		       retry_count, latency_ms, error_message, scheduled_for, posted_at,
		       created_at, updated_at
		FROM tweets WHERE user_id = $1
	`

	if filter.EntryID != nil {
		query += fmt.Sprintf(" AND entry_id = $%d", argIdx)
		args = append(args, *filter.EntryID)
		argIdx++
	}
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	} else {
		query += " LIMIT 20"
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tweets: %w", err)
	}
	defer rows.Close()

	var tweets []*domain.Tweet
	for rows.Next() {
		t, err := r.scanTweetFromRow(rows)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tweets: %w", err)
	}
	if tweets == nil {
		tweets = []*domain.Tweet{}
	}
	return tweets, nil
}

func (r *TweetRepository) Update(ctx context.Context, tweet *domain.Tweet) error {
	query := `
		UPDATE tweets SET
			content = $1, status = $2, version = $3, prompt_id = $4,
			provider_name = $5, model_name = $6, input_tokens = $7, output_tokens = $8,
			cost_usd = $9, retry_count = $10, latency_ms = $11, error_message = $12,
			scheduled_for = $13, posted_at = $14, updated_at = $15
		WHERE id = $16
	`
	tweet.UpdatedAt = time.Now()

	tag, err := r.pool.Exec(ctx, query,
		tweet.Content, tweet.Status, tweet.Version, nullableUUID(tweet.PromptID),
		tweet.ProviderName, tweet.ModelName, tweet.InputTokens, tweet.OutputTokens,
		tweet.CostUSD, tweet.RetryCount, tweet.LatencyMs, tweet.ErrorMessage,
		tweet.ScheduledFor, tweet.PostedAt, tweet.UpdatedAt, tweet.ID,
	)
	if err != nil {
		return fmt.Errorf("updating tweet: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TweetRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TweetStatus) error {
	query := `UPDATE tweets SET status = $1, updated_at = $2 WHERE id = $3`
	now := time.Now()
	tag, err := r.pool.Exec(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("updating tweet status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TweetRepository) CreateAudit(ctx context.Context, audit *domain.GenerationAudit) error {
	query := `
		INSERT INTO generation_audits (
			id, tweet_id, action, user_id,
			previous_content, new_content, previous_status, new_status,
			metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = time.Now()
	}

	metaJSON, err := json.Marshal(audit.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = r.pool.Exec(ctx, query,
		audit.ID, audit.TweetID, audit.Action, audit.UserID,
		audit.PreviousContent, audit.NewContent, audit.PreviousStatus, audit.NewStatus,
		metaJSON, audit.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting audit: %w", err)
	}
	return nil
}

func (r *TweetRepository) ListAudits(ctx context.Context, tweetID uuid.UUID) ([]*domain.GenerationAudit, error) {
	query := `
		SELECT id, tweet_id, action, user_id,
		       previous_content, new_content, previous_status, new_status,
		       metadata, created_at
		FROM generation_audits WHERE tweet_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tweetID)
	if err != nil {
		return nil, fmt.Errorf("querying audits: %w", err)
	}
	defer rows.Close()

	var audits []*domain.GenerationAudit
	for rows.Next() {
		a, err := r.scanAuditFromRow(rows)
		if err != nil {
			return nil, err
		}
		audits = append(audits, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating audits: %w", err)
	}
	if audits == nil {
		audits = []*domain.GenerationAudit{}
	}
	return audits, nil
}

func (r *TweetRepository) scanTweet(ctx context.Context, query string, args ...interface{}) (*domain.Tweet, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	return r.scanTweetFromRow(row)
}

func (r *TweetRepository) scanTweetFromRow(row scannable) (*domain.Tweet, error) {
	t := &domain.Tweet{}
	var promptID *uuid.UUID
	err := row.Scan(
		&t.ID, &t.EntryID, &t.UserID, &t.Content, &t.Status, &t.Version, &promptID,
		&t.ProviderName, &t.ModelName, &t.InputTokens, &t.OutputTokens, &t.CostUSD,
		&t.RetryCount, &t.LatencyMs, &t.ErrorMessage, &t.ScheduledFor, &t.PostedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning tweet: %w", err)
	}
	if promptID != nil {
		t.PromptID = *promptID
	}
	return t, nil
}

func (r *TweetRepository) scanAuditFromRow(row scannable) (*domain.GenerationAudit, error) {
	a := &domain.GenerationAudit{}
	var metaJSON []byte
	err := row.Scan(
		&a.ID, &a.TweetID, &a.Action, &a.UserID,
		&a.PreviousContent, &a.NewContent, &a.PreviousStatus, &a.NewStatus,
		&metaJSON, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning audit: %w", err)
	}
	if len(metaJSON) > 0 {
		_ = json.Unmarshal(metaJSON, &a.Metadata)
	}
	return a, nil
}

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
