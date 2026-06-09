package raw_entry

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type CreateRawEntryRequest struct {
	Content string               `json:"content"`
	Source  domain.RawEntrySource `json:"source"`
}

type RawEntryResponse struct {
	ID        uuid.UUID            `json:"id"`
	UserID    uuid.UUID            `json:"user_id"`
	Content   string               `json:"content"`
	Status    domain.RawEntryStatus `json:"status"`
	Source    domain.RawEntrySource `json:"source"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

type ListRawEntriesRequest struct {
	UserID uuid.UUID
	Status *domain.RawEntryStatus
	Source *domain.RawEntrySource
	Limit  int
	Offset int
}

type ListRawEntriesResponse struct {
	Entries []RawEntryResponse `json:"entries"`
	Total   int                `json:"total"`
}

func EntryToResponse(e *domain.RawEntry) RawEntryResponse {
	return RawEntryResponse{
		ID:        e.ID,
		UserID:    e.UserID,
		Content:   e.Content,
		Status:    e.Status,
		Source:    e.Source,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
