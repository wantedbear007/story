package entry

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// Service implements entry-related use cases.
// It orchestrates EntryRepository and TagRepository to maintain
// the aggregate consistency of entries and their tags.
type Service struct {
	entryRepo domain.EntryRepository
	tagRepo   domain.TagRepository
}

func NewService(
	entryRepo domain.EntryRepository,
	tagRepo domain.TagRepository,
) *Service {
	return &Service{
		entryRepo: entryRepo,
		tagRepo:   tagRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateEntryRequest) (*EntryResponse, error) {
	entry := &domain.Entry{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     req.Type,
		Title:    req.Title,
		Content:  req.Content,
		Metadata: req.Metadata,
	}

	if err := s.entryRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("creating entry: %w", err)
	}

	tagNames := make([]string, 0, len(req.Tags))
	for _, tagName := range req.Tags {
		tag := &domain.Tag{
			ID:     uuid.New(),
			UserID: userID,
			Name:   tagName,
		}
		if err := s.tagRepo.Create(ctx, tag); err != nil {
			if err == domain.ErrAlreadyExists {
				existing, lookupErr := s.findTagByName(ctx, userID, tagName)
				if lookupErr != nil {
					continue
				}
				tag = existing
			} else {
				continue
			}
		}
		if err := s.tagRepo.AddTagToEntry(ctx, entry.ID, tag.ID); err != nil {
			return nil, fmt.Errorf("associating tag: %w", err)
		}
		tagNames = append(tagNames, tag.Name)
	}

	resp := EntryToResponse(entry, tagNames)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*EntryResponse, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: entry not found", domain.ErrNotFound)
	}

	tags, err := s.tagRepo.GetEntryTags(ctx, entry.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching entry tags: %w", err)
	}

	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	resp := EntryToResponse(entry, tagNames)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, req EntryFilterRequest) (*ListResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := domain.EntryFilter{
		UserID:   userID,
		Types:    req.Types,
		Query:    req.Query,
		Tags:     req.Tags,
		Limit:    pageSize,
		Offset:   (page - 1) * pageSize,
	}

	entries, err := s.entryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing entries: %w", err)
	}

	responses := make([]EntryResponse, len(entries))
	for i, e := range entries {
		tags, tagErr := s.tagRepo.GetEntryTags(ctx, e.ID)
		tagNames := make([]string, 0)
		if tagErr == nil {
			for _, t := range tags {
				tagNames = append(tagNames, t.Name)
			}
		}
		responses[i] = EntryToResponse(e, tagNames)
	}

	return &ListResponse{
		Entries: responses,
		Page:    page,
		Total:   len(responses),
	}, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateEntryRequest) (*EntryResponse, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: entry not found", domain.ErrNotFound)
	}

	if req.Type != nil {
		entry.Type = *req.Type
	}
	if req.Title != nil {
		entry.Title = *req.Title
	}
	if req.Content != nil {
		entry.Content = *req.Content
	}
	if req.Metadata != nil {
		entry.Metadata = req.Metadata
	}

	if err := s.entryRepo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("updating entry: %w", err)
	}

	if req.Tags != nil {
		currentTags, tagErr := s.tagRepo.GetEntryTags(ctx, entry.ID)
		if tagErr == nil {
			for _, t := range currentTags {
				_ = s.tagRepo.RemoveTagFromEntry(ctx, entry.ID, t.ID)
			}
		}

		for _, tagName := range req.Tags {
			tag := &domain.Tag{
				ID:     uuid.New(),
				UserID: entry.UserID,
				Name:   tagName,
			}
			if createErr := s.tagRepo.Create(ctx, tag); createErr != nil {
				if createErr == domain.ErrAlreadyExists {
					existing, lookupErr := s.findTagByName(ctx, entry.UserID, tagName)
					if lookupErr != nil {
						continue
					}
					tag = existing
				} else {
					continue
				}
			}
			_ = s.tagRepo.AddTagToEntry(ctx, entry.ID, tag.ID)
		}
	}

	tags, _ := s.tagRepo.GetEntryTags(ctx, entry.ID)
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	resp := EntryToResponse(entry, tagNames)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.entryRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}
	return nil
}

func (s *Service) findTagByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Tag, error) {
	tags, err := s.tagRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, domain.ErrNotFound
}
