package collection

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type CreateCollectionRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description string  `json:"description,omitempty" validate:"max=2000"`
	ParentID    *string `json:"parent_id,omitempty"`
}

type UpdateCollectionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
}

type CollectionResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	EntryCount  int        `json:"entry_count,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func CollectionToResponse(c *domain.Collection) CollectionResponse {
	return CollectionResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
