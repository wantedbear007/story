package resource

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type CreateResourceRequest struct {
	Type        domain.ResourceType      `json:"type" validate:"required"`
	Title       string                   `json:"title" validate:"required,min=1,max=500"`
	URL         string                   `json:"url" validate:"required"`
	Description string                   `json:"description,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

type UpdateResourceRequest struct {
	Title       *string                  `json:"title,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

type ResourceFilterRequest struct {
	Types    []domain.ResourceType `json:"types,omitempty"`
	Query    string                `json:"query,omitempty"`
	Page     int                   `json:"page,omitempty"`
	PageSize int                   `json:"page_size,omitempty"`
}

type ResourceResponse struct {
	ID          uuid.UUID              `json:"id"`
	Type        domain.ResourceType    `json:"type"`
	Title       string                 `json:"title"`
	URL         string                 `json:"url"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ListResponse struct {
	Resources []ResourceResponse `json:"resources"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
}

func ResourceToResponse(r *domain.Resource) ResourceResponse {
	return ResourceResponse{
		ID:          r.ID,
		Type:        r.Type,
		Title:       r.Title,
		URL:         r.URL,
		Description: r.Description,
		Metadata:    r.Metadata,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
