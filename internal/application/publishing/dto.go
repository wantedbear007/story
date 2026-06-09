package publishing

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type CreateTargetRequest struct {
	Type   domain.PublishingTargetType `json:"type" validate:"required"`
	Name   string                      `json:"name" validate:"required,min=1,max=100"`
	Config map[string]interface{}      `json:"config" validate:"required"`
}

type UpdateTargetRequest struct {
	Name   *string                     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Config map[string]interface{}      `json:"config,omitempty"`
}

type TargetResponse struct {
	ID        uuid.UUID                    `json:"id"`
	Type      domain.PublishingTargetType  `json:"type"`
	Name      string                       `json:"name"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
}

type PublishRequest struct {
	EntryID  uuid.UUID `json:"entry_id" validate:"required"`
	TargetID uuid.UUID `json:"target_id" validate:"required"`
}

type PublishedEntryResponse struct {
	ID          uuid.UUID       `json:"id"`
	EntryID     uuid.UUID       `json:"entry_id"`
	TargetID    uuid.UUID       `json:"target_id"`
	ExternalURL string          `json:"external_url,omitempty"`
	Status      domain.PublishStatus `json:"status"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

func TargetToResponse(t *domain.PublishingTarget) TargetResponse {
	return TargetResponse{
		ID:        t.ID,
		Type:      t.Type,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func PublishedEntryToResponse(pe *domain.PublishedEntry) PublishedEntryResponse {
	return PublishedEntryResponse{
		ID:          pe.ID,
		EntryID:     pe.EntryID,
		TargetID:    pe.TargetID,
		ExternalURL: pe.ExternalURL,
		Status:      pe.Status,
		PublishedAt: pe.PublishedAt,
		CreatedAt:   pe.CreatedAt,
	}
}
