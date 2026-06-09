package publishing

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// Publisher abstracts external publishing platform integration.
type Publisher interface {
	Publish(ctx context.Context, entry *domain.Entry, target *domain.PublishingTarget) (string, error)
}

// Service implements publishing-related use cases.
// It coordinates between entry retrieval, target configuration,
// and external platform publishing via the Publisher abstraction.
type Service struct {
	targetRepo    domain.PublishingTargetRepository
	publishedRepo domain.PublishedEntryRepository
	entryRepo     domain.EntryRepository
	publisher     Publisher
}

func NewService(
	targetRepo domain.PublishingTargetRepository,
	publishedRepo domain.PublishedEntryRepository,
	entryRepo domain.EntryRepository,
	publisher Publisher,
) *Service {
	return &Service{
		targetRepo:    targetRepo,
		publishedRepo: publishedRepo,
		entryRepo:     entryRepo,
		publisher:     publisher,
	}
}

func (s *Service) CreateTarget(ctx context.Context, userID uuid.UUID, req CreateTargetRequest) (*TargetResponse, error) {
	target := &domain.PublishingTarget{
		ID:     uuid.New(),
		UserID: userID,
		Type:   req.Type,
		Name:   req.Name,
		Config: req.Config,
	}

	if err := s.targetRepo.Create(ctx, target); err != nil {
		return nil, fmt.Errorf("creating publishing target: %w", err)
	}

	resp := TargetToResponse(target)
	return &resp, nil
}

func (s *Service) ListTargets(ctx context.Context, userID uuid.UUID) ([]TargetResponse, error) {
	targets, err := s.targetRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing targets: %w", err)
	}

	responses := make([]TargetResponse, len(targets))
	for i, t := range targets {
		responses[i] = TargetToResponse(t)
	}
	return responses, nil
}

func (s *Service) Publish(ctx context.Context, userID uuid.UUID, req PublishRequest) (*PublishedEntryResponse, error) {
	entry, err := s.entryRepo.GetByID(ctx, req.EntryID)
	if err != nil {
		return nil, fmt.Errorf("%w: entry not found", domain.ErrNotFound)
	}

	if entry.UserID != userID {
		return nil, fmt.Errorf("%w: entry does not belong to user", domain.ErrForbidden)
	}

	target, err := s.targetRepo.GetByID(ctx, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("%w: publishing target not found", domain.ErrNotFound)
	}

	if target.UserID != userID {
		return nil, fmt.Errorf("%w: target does not belong to user", domain.ErrForbidden)
	}

	pe := &domain.PublishedEntry{
		ID:      uuid.New(),
		EntryID: entry.ID,
		TargetID: target.ID,
		Status:  domain.PublishStatusPending,
	}

	if err := s.publishedRepo.Create(ctx, pe); err != nil {
		return nil, fmt.Errorf("recording publish attempt: %w", err)
	}

	externalURL, err := s.publisher.Publish(ctx, entry, target)
	if err != nil {
		pe.Status = domain.PublishStatusFailed
		pe.ErrorMessage = err.Error()
	} else {
		pe.Status = domain.PublishStatusPublished
		pe.ExternalURL = externalURL
		now := domain.Now()
		pe.PublishedAt = &now
	}

	if updateErr := s.publishedRepo.Update(ctx, pe); updateErr != nil {
		return nil, fmt.Errorf("updating publish status: %w", updateErr)
	}

	if err != nil {
		return nil, fmt.Errorf("publishing failed: %w", err)
	}

	resp := PublishedEntryToResponse(pe)
	return &resp, nil
}

func (s *Service) ListPublished(ctx context.Context, entryID uuid.UUID) ([]PublishedEntryResponse, error) {
	entries, err := s.publishedRepo.ListByEntry(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("listing published entries: %w", err)
	}

	responses := make([]PublishedEntryResponse, len(entries))
	for i, pe := range entries {
		responses[i] = PublishedEntryToResponse(pe)
	}
	return responses, nil
}
