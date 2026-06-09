package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EntryType categorizes entries for organization and search filtering.
// Each type maps to a distinct knowledge domain in the developer's workflow.
type EntryType string

const (
	EntryTypeLearning        EntryType = "learning"
	EntryTypeWorkLog         EntryType = "work_log"
	EntryTypeResource        EntryType = "resource"
	EntryTypeEngineeringNote EntryType = "engineering_note"
)

// Entry represents a captured piece of knowledge.
// Entries are the primary unit of content in the system.
// Metadata stores flexible key-value pairs for extensibility
// (e.g., code snippets, URLs, duration for work logs).
type Entry struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Type      EntryType              `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

// EntryFilter defines search and filter criteria for querying entries.
type EntryFilter struct {
	UserID    uuid.UUID
	Types     []EntryType
	Query     string
	Tags      []string
	CollectionID *uuid.UUID
	Limit     int
	Offset    int
}

// EntryRepository defines persistence contract for Entry entities.
type EntryRepository interface {
	Create(ctx context.Context, entry *Entry) error
	GetByID(ctx context.Context, id uuid.UUID) (*Entry, error)
	List(ctx context.Context, filter EntryFilter) ([]*Entry, error)
	Update(ctx context.Context, entry *Entry) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
