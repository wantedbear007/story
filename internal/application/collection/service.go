package collection

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// Service implements collection-related use cases.
type Service struct {
	collectionRepo domain.CollectionRepository
}

func NewService(collectionRepo domain.CollectionRepository) *Service {
	return &Service{
		collectionRepo: collectionRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateCollectionRequest) (*CollectionResponse, error) {
	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid parent_id", domain.ErrInvalidInput)
		}
		parentID = &pid
	}

	col := &domain.Collection{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
	}

	if err := s.collectionRepo.Create(ctx, col); err != nil {
		return nil, fmt.Errorf("creating collection: %w", err)
	}

	resp := CollectionToResponse(col)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*CollectionResponse, error) {
	col, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: collection not found", domain.ErrNotFound)
	}

	resp := CollectionToResponse(col)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]CollectionResponse, error) {
	cols, err := s.collectionRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing collections: %w", err)
	}

	responses := make([]CollectionResponse, len(cols))
	for i, c := range cols {
		responses[i] = CollectionToResponse(c)
	}
	return responses, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateCollectionRequest) (*CollectionResponse, error) {
	col, err := s.collectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: collection not found", domain.ErrNotFound)
	}

	if req.Name != nil {
		col.Name = *req.Name
	}
	if req.Description != nil {
		col.Description = *req.Description
	}

	if err := s.collectionRepo.Update(ctx, col); err != nil {
		return nil, fmt.Errorf("updating collection: %w", err)
	}

	resp := CollectionToResponse(col)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.collectionRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("deleting collection: %w", err)
	}
	return nil
}
