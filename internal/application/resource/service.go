package resource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/anomalyco/story/internal/domain"
	apperrors "github.com/anomalyco/story/internal/pkg/errors"
)

type Service struct {
	resourceRepo domain.ResourceRepository
}

func NewService(resourceRepo domain.ResourceRepository) *Service {
	return &Service{
		resourceRepo: resourceRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateResourceRequest) (*ResourceResponse, error) {
	resource := &domain.Resource{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        req.Type,
		Title:       req.Title,
		URL:         req.URL,
		Description: req.Description,
		Metadata:    req.Metadata,
		ContentHash: hashContent(req.URL),
	}

	if resource.Metadata == nil {
		resource.Metadata = make(map[string]interface{})
	}
	extractTypeMetadata(resource)

	if err := s.resourceRepo.Create(ctx, resource); err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	resp := ResourceToResponse(resource)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*ResourceResponse, error) {
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperrors.ErrNotFound("resource not found")
		}
		return nil, fmt.Errorf("fetching resource: %w", err)
	}

	resp := ResourceToResponse(resource)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, req ResourceFilterRequest) (*ListResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := domain.ResourceFilter{
		UserID: userID,
		Types:  req.Types,
		Query:  req.Query,
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	resources, err := s.resourceRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing resources: %w", err)
	}

	responses := make([]ResourceResponse, len(resources))
	for i, r := range resources {
		responses[i] = ResourceToResponse(r)
	}

	return &ListResponse{
		Resources: responses,
		Total:     len(responses),
		Page:      page,
	}, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateResourceRequest) (*ResourceResponse, error) {
	resource, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperrors.ErrNotFound("resource not found")
		}
		return nil, fmt.Errorf("fetching resource: %w", err)
	}

	if req.Title != nil {
		resource.Title = *req.Title
	}
	if req.Description != nil {
		resource.Description = *req.Description
	}
	if req.Metadata != nil {
		resource.Metadata = req.Metadata
	}

	if err := s.resourceRepo.Update(ctx, resource); err != nil {
		return nil, fmt.Errorf("updating resource: %w", err)
	}

	resp := ResourceToResponse(resource)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.resourceRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.ErrNotFound("resource not found")
		}
		return fmt.Errorf("deleting resource: %w", err)
	}
	return nil
}

func (s *Service) AttachToEntry(ctx context.Context, resourceID, entryID uuid.UUID) error {
	if err := s.resourceRepo.AttachToEntry(ctx, entryID, resourceID); err != nil {
		return fmt.Errorf("attaching resource to entry: %w", err)
	}
	return nil
}

func (s *Service) DetachFromEntry(ctx context.Context, resourceID, entryID uuid.UUID) error {
	if err := s.resourceRepo.DetachFromEntry(ctx, entryID, resourceID); err != nil {
		return fmt.Errorf("detaching resource from entry: %w", err)
	}
	return nil
}

func (s *Service) GetEntryResources(ctx context.Context, entryID uuid.UUID) ([]ResourceResponse, error) {
	resources, err := s.resourceRepo.GetEntryResources(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("fetching entry resources: %w", err)
	}

	responses := make([]ResourceResponse, len(resources))
	for i, r := range resources {
		responses[i] = ResourceToResponse(r)
	}
	return responses, nil
}

func hashContent(s string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h[:])
}

func extractTypeMetadata(r *domain.Resource) {
	switch r.Type {
	case domain.ResourceTypeGitHub:
		r.Metadata["source"] = "github"
		if parts := strings.Split(strings.TrimPrefix(r.URL, "https://github.com/"), "/"); len(parts) >= 2 {
			r.Metadata["owner"] = parts[0]
			r.Metadata["repo"] = strings.TrimSuffix(parts[1], ".git")
		}
	case domain.ResourceTypeYouTube:
		r.Metadata["source"] = "youtube"
		for _, prefix := range []string{"https://youtu.be/", "https://www.youtube.com/watch?v="} {
			if vid := strings.TrimPrefix(r.URL, prefix); vid != r.URL {
				r.Metadata["video_id"] = strings.SplitN(vid, "&", 2)[0]
				break
			}
		}
	case domain.ResourceTypeArticle:
		r.Metadata["source"] = "article"
	case domain.ResourceTypeURL:
		r.Metadata["source"] = "web"
	case domain.ResourceTypePDF:
		r.Metadata["source"] = "pdf"
	case domain.ResourceTypeMarkdown:
		r.Metadata["source"] = "markdown"
	}
}
