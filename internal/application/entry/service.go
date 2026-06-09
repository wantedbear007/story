package entry

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	resourcedto "github.com/anomalyco/story/internal/application/resource"
	"github.com/google/uuid"
)

type Service struct {
	entryRepo    domain.EntryRepository
	tagRepo      domain.TagRepository
	resourceRepo domain.ResourceRepository
}

func NewService(
	entryRepo domain.EntryRepository,
	tagRepo domain.TagRepository,
	resourceRepo domain.ResourceRepository,
) *Service {
	return &Service{
		entryRepo:    entryRepo,
		tagRepo:      tagRepo,
		resourceRepo: resourceRepo,
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

	tagNames := s.associateTags(ctx, entry.ID, userID, req.Tags)

	for _, resourceID := range req.Resources {
		if err := s.resourceRepo.AttachToEntry(ctx, entry.ID, resourceID); err != nil {
			return nil, fmt.Errorf("attaching resource: %w", err)
		}
	}

	resources := s.getEntryResources(ctx, entry.ID)
	resp := EntryToResponse(entry, tagNames, resources)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*EntryResponse, error) {
	entry, err := s.entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: entry not found", domain.ErrNotFound)
	}

	tags := s.getEntryTags(ctx, entry.ID)
	resources := s.getEntryResources(ctx, entry.ID)
	resp := EntryToResponse(entry, tags, resources)
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
		UserID: userID,
		Types:  req.Types,
		Query:  req.Query,
		Tags:   req.Tags,
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	entries, err := s.entryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing entries: %w", err)
	}

	responses := make([]EntryResponse, len(entries))
	for i, e := range entries {
		tags := s.getEntryTags(ctx, e.ID)
		resources := s.getEntryResources(ctx, e.ID)
		responses[i] = EntryToResponse(e, tags, resources)
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
		s.associateTags(ctx, entry.ID, entry.UserID, req.Tags)
	}

	if req.Resources != nil {
		currentResources, _ := s.resourceRepo.GetEntryResources(ctx, entry.ID)
		for _, r := range currentResources {
			_ = s.resourceRepo.DetachFromEntry(ctx, entry.ID, r.ID)
		}
		for _, resourceID := range req.Resources {
			_ = s.resourceRepo.AttachToEntry(ctx, entry.ID, resourceID)
		}
	}

	tags := s.getEntryTags(ctx, entry.ID)
	resources := s.getEntryResources(ctx, entry.ID)
	resp := EntryToResponse(entry, tags, resources)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.entryRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}
	return nil
}

func (s *Service) associateTags(ctx context.Context, entryID, userID uuid.UUID, tagNames []string) []string {
	result := make([]string, 0, len(tagNames))
	for _, tagName := range tagNames {
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
		if err := s.tagRepo.AddTagToEntry(ctx, entryID, tag.ID); err != nil {
			continue
		}
		result = append(result, tag.Name)
	}
	return result
}

func (s *Service) getEntryTags(ctx context.Context, entryID uuid.UUID) []string {
	tags, err := s.tagRepo.GetEntryTags(ctx, entryID)
	if err != nil {
		return make([]string, 0)
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}

func (s *Service) getEntryResources(ctx context.Context, entryID uuid.UUID) []resourcedto.ResourceResponse {
	resources, err := s.resourceRepo.GetEntryResources(ctx, entryID)
	if err != nil {
		return make([]resourcedto.ResourceResponse, 0)
	}
	responses := make([]resourcedto.ResourceResponse, len(resources))
	for i, r := range resources {
		responses[i] = resourcedto.ResourceToResponse(r)
	}
	return responses
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
