package tag

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	tagRepo domain.TagRepository
}

func NewService(tagRepo domain.TagRepository) *Service {
	return &Service{
		tagRepo: tagRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateTagRequest) (*TagResponse, error) {
	tag := &domain.Tag{
		ID:     uuid.New(),
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("creating tag: %w", err)
	}

	resp := TagToResponse(tag)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*TagResponse, error) {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: tag not found", domain.ErrNotFound)
	}

	resp := TagToResponse(tag)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]TagResponse, error) {
	tags, err := s.tagRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}

	responses := make([]TagResponse, len(tags))
	for i, t := range tags {
		responses[i] = TagToResponse(t)
	}
	return responses, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateTagRequest) (*TagResponse, error) {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: tag not found", domain.ErrNotFound)
	}

	if req.Name != nil {
		tag.Name = *req.Name
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}

	if err := s.tagRepo.Update(ctx, tag); err != nil {
		return nil, fmt.Errorf("updating tag: %w", err)
	}

	resp := TagToResponse(tag)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.tagRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	return nil
}
