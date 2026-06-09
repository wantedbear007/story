package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ResourceType string

const (
	ResourceTypeURL      ResourceType = "url"
	ResourceTypeGitHub   ResourceType = "github"
	ResourceTypeArticle  ResourceType = "article"
	ResourceTypeYouTube  ResourceType = "youtube"
	ResourceTypePDF      ResourceType = "pdf"
	ResourceTypeMarkdown ResourceType = "markdown"
)

type Resource struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Type        ResourceType           `json:"type"`
	Title       string                 `json:"title"`
	URL         string                 `json:"url"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ContentHash string                 `json:"-"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ResourceFilter struct {
	UserID uuid.UUID
	Types  []ResourceType
	Query  string
	Limit  int
	Offset int
}

type ResourceRepository interface {
	Create(ctx context.Context, resource *Resource) error
	GetByID(ctx context.Context, id uuid.UUID) (*Resource, error)
	List(ctx context.Context, filter ResourceFilter) ([]*Resource, error)
	Update(ctx context.Context, resource *Resource) error
	Delete(ctx context.Context, id uuid.UUID) error

	AttachToEntry(ctx context.Context, entryID, resourceID uuid.UUID) error
	DetachFromEntry(ctx context.Context, entryID, resourceID uuid.UUID) error
	GetEntryResources(ctx context.Context, entryID uuid.UUID) ([]*Resource, error)
}
