package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Collection groups entries into a hierarchical folder-like structure.
// Collections can represent projects, topics, sprints, or any logical grouping.
type Collection struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// CollectionRepository defines persistence contract for Collection entities.
type CollectionRepository interface {
	Create(ctx context.Context, col *Collection) error
	GetByID(ctx context.Context, id uuid.UUID) (*Collection, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Collection, error)
	Update(ctx context.Context, col *Collection) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Entry-collection association
	AddEntry(ctx context.Context, collectionID, entryID uuid.UUID) error
	RemoveEntry(ctx context.Context, collectionID, entryID uuid.UUID) error
	GetEntries(ctx context.Context, collectionID uuid.UUID) ([]*Entry, error)
}
