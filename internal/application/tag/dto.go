package tag

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Color string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

type UpdateTagRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Color *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

type TagResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func TagToResponse(t *domain.Tag) TagResponse {
	return TagResponse{
		ID:        t.ID,
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: t.CreatedAt,
	}
}
