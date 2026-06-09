package raw_entry

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	repo domain.RawEntryRepository
}

func NewService(repo domain.RawEntryRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRawEntryRequest) (*RawEntryResponse, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("%w: content is required", domain.ErrInvalidInput)
	}
	if req.Source == "" {
		req.Source = domain.RawEntrySourceCLI
	}

	entry := &domain.RawEntry{
		ID:      uuid.New(),
		UserID:  userID,
		Content: req.Content,
		Status:  domain.RawEntryStatusRAW,
		Source:  req.Source,
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("creating raw entry: %w", err)
	}

	resp := EntryToResponse(entry)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, id, userID uuid.UUID) (*RawEntryResponse, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting raw entry: %w", err)
	}
	if entry.UserID != userID {
		return nil, fmt.Errorf("%w: raw entry not found", domain.ErrNotFound)
	}

	resp := EntryToResponse(entry)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, req ListRawEntriesRequest) (*ListRawEntriesResponse, error) {
	filter := domain.RawEntryFilter{
		UserID: req.UserID,
		Status: req.Status,
		Source: req.Source,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	entries, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing raw entries: %w", err)
	}

	resp := &ListRawEntriesResponse{
		Entries: make([]RawEntryResponse, len(entries)),
		Total:   len(entries),
	}
	for i, e := range entries {
		resp.Entries[i] = EntryToResponse(e)
	}
	return resp, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id, userID uuid.UUID, status domain.RawEntryStatus) (*RawEntryResponse, error) {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("updating raw entry status: %w", err)
	}
	if entry.UserID != userID {
		return nil, fmt.Errorf("%w: raw entry not found", domain.ErrNotFound)
	}

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return nil, fmt.Errorf("updating raw entry status: %w", err)
	}

	entry.Status = status
	resp := EntryToResponse(entry)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting raw entry: %w", err)
	}
	if entry.UserID != userID {
		return fmt.Errorf("%w: raw entry not found", domain.ErrNotFound)
	}

	return s.repo.Delete(ctx, id)
}
