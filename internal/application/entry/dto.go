package entry

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type CreateEntryRequest struct {
	Type     domain.EntryType        `json:"type" validate:"required"`
	Title    string                  `json:"title" validate:"required,min=1,max=500"`
	Content  string                  `json:"content" validate:"required"`
	Metadata map[string]interface{}  `json:"metadata,omitempty"`
	Tags     []string                `json:"tags,omitempty"`
}

type UpdateEntryRequest struct {
	Type     *domain.EntryType       `json:"type,omitempty"`
	Title    *string                 `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Content  *string                 `json:"content,omitempty"`
	Metadata map[string]interface{}  `json:"metadata,omitempty"`
	Tags     []string                `json:"tags,omitempty"`
}

type EntryFilterRequest struct {
	Types     []domain.EntryType `json:"types,omitempty"`
	Query     string             `json:"query,omitempty"`
	Tags      []string           `json:"tags,omitempty"`
	Page      int                `json:"page,omitempty"`
	PageSize  int                `json:"page_size,omitempty"`
}

type EntryResponse struct {
	ID        uuid.UUID              `json:"id"`
	Type      domain.EntryType       `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type ListResponse struct {
	Entries []EntryResponse `json:"entries"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
}

func EntryToResponse(e *domain.Entry, tags []string) EntryResponse {
	return EntryResponse{
		ID:        e.ID,
		Type:      e.Type,
		Title:     e.Title,
		Content:   e.Content,
		Metadata:  e.Metadata,
		Tags:      tags,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
