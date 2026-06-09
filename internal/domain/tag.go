package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Tag is a user-defined label for organizing entries.
// Tags are lightweight, user-scoped, and enable flexible cross-cutting organization
// that complements the hierarchical collection system.
type Tag struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	Color     string     `json:"color,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TagRepository defines persistence contract for Tag entities.
type TagRepository interface {
	Create(ctx context.Context, tag *Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tag, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Entry-tag association
	AddTagToEntry(ctx context.Context, entryID, tagID uuid.UUID) error
	RemoveTagFromEntry(ctx context.Context, entryID, tagID uuid.UUID) error
	GetEntryTags(ctx context.Context, entryID uuid.UUID) ([]*Tag, error)
}
