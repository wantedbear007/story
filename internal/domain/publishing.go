package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PublishingTargetType enumerates supported publishing platforms.
type PublishingTargetType string

const (
	PublishTargetTwitter  PublishingTargetType = "twitter"
	PublishTargetNotion   PublishingTargetType = "notion"
	PublishTargetGDoc     PublishingTargetType = "google_doc"
	PublishTargetBlog     PublishingTargetType = "blog"
	PublishTargetMarkdown PublishingTargetType = "markdown"
)

// PublishingTarget stores platform connection configuration.
// Config is provider-specific and stored as JSON for extensibility.
// Secrets (API keys, tokens) should be encrypted at rest.
type PublishingTarget struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Type      PublishingTargetType   `json:"type"`
	Name      string                 `json:"name"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// PublishStatus tracks the lifecycle of a published piece.
type PublishStatus string

const (
	PublishStatusPending   PublishStatus = "pending"
	PublishStatusPublished PublishStatus = "published"
	PublishStatusFailed    PublishStatus = "failed"
)

// PublishedEntry records the result of publishing an entry to a target.
type PublishedEntry struct {
	ID           uuid.UUID     `json:"id"`
	EntryID      uuid.UUID     `json:"entry_id"`
	TargetID     uuid.UUID     `json:"target_id"`
	ExternalURL  string        `json:"external_url,omitempty"`
	Status       PublishStatus `json:"status"`
	ErrorMessage string        `json:"error_message,omitempty"`
	PublishedAt  *time.Time    `json:"published_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// PublishingTargetRepository defines persistence contract for publishing targets.
type PublishingTargetRepository interface {
	Create(ctx context.Context, target *PublishingTarget) error
	GetByID(ctx context.Context, id uuid.UUID) (*PublishingTarget, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*PublishingTarget, error)
	Update(ctx context.Context, target *PublishingTarget) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PublishedEntryRepository defines persistence contract for published entries.
type PublishedEntryRepository interface {
	Create(ctx context.Context, pe *PublishedEntry) error
	GetByID(ctx context.Context, id uuid.UUID) (*PublishedEntry, error)
	ListByEntry(ctx context.Context, entryID uuid.UUID) ([]*PublishedEntry, error)
	ListByTarget(ctx context.Context, targetID uuid.UUID) ([]*PublishedEntry, error)
	Update(ctx context.Context, pe *PublishedEntry) error
}
