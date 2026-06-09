package entry_test

import (
	"context"
	"testing"

	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type mockEntryRepo struct {
	entries map[uuid.UUID]*domain.Entry
}

func newMockEntryRepo() *mockEntryRepo {
	return &mockEntryRepo{entries: make(map[uuid.UUID]*domain.Entry)}
}

func (m *mockEntryRepo) Create(ctx context.Context, e *domain.Entry) error {
	m.entries[e.ID] = e
	return nil
}

func (m *mockEntryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Entry, error) {
	e, exists := m.entries[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return e, nil
}

func (m *mockEntryRepo) List(ctx context.Context, f domain.EntryFilter) ([]*domain.Entry, error) {
	var result []*domain.Entry
	for _, e := range m.entries {
		if e.UserID != f.UserID {
			continue
		}
		result = append(result, e)
	}
	if result == nil {
		result = make([]*domain.Entry, 0)
	}
	return result, nil
}

func (m *mockEntryRepo) Update(ctx context.Context, e *domain.Entry) error {
	if _, exists := m.entries[e.ID]; !exists {
		return domain.ErrNotFound
	}
	m.entries[e.ID] = e
	return nil
}

func (m *mockEntryRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.entries[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.entries, id)
	return nil
}

type mockTagRepo struct {
	tags    map[uuid.UUID]*domain.Tag
	entries map[uuid.UUID][]uuid.UUID // entryID -> tagIDs
}

func newMockTagRepo() *mockTagRepo {
	return &mockTagRepo{
		tags:    make(map[uuid.UUID]*domain.Tag),
		entries: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockTagRepo) Create(ctx context.Context, t *domain.Tag) error {
	m.tags[t.ID] = t
	return nil
}

func (m *mockTagRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	t, exists := m.tags[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockTagRepo) ListByUser(ctx context.Context, uid uuid.UUID) ([]*domain.Tag, error) {
	var result []*domain.Tag
	for _, t := range m.tags {
		if t.UserID == uid {
			result = append(result, t)
		}
	}
	if result == nil {
		result = make([]*domain.Tag, 0)
	}
	return result, nil
}

func (m *mockTagRepo) Update(ctx context.Context, t *domain.Tag) error {
	if _, exists := m.tags[t.ID]; !exists {
		return domain.ErrNotFound
	}
	m.tags[t.ID] = t
	return nil
}

func (m *mockTagRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.tags, id)
	return nil
}

func (m *mockTagRepo) AddTagToEntry(ctx context.Context, eid, tid uuid.UUID) error {
	m.entries[eid] = append(m.entries[eid], tid)
	return nil
}

func (m *mockTagRepo) RemoveTagFromEntry(ctx context.Context, eid, tid uuid.UUID) error {
	tags := m.entries[eid]
	for i, id := range tags {
		if id == tid {
			m.entries[eid] = append(tags[:i], tags[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockTagRepo) GetEntryTags(ctx context.Context, eid uuid.UUID) ([]*domain.Tag, error) {
	tagIDs, exists := m.entries[eid]
	if !exists {
		return make([]*domain.Tag, 0), nil
	}
	var result []*domain.Tag
	for _, tid := range tagIDs {
		if t, ok := m.tags[tid]; ok {
			result = append(result, t)
		}
	}
	if result == nil {
		result = make([]*domain.Tag, 0)
	}
	return result, nil
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	svc := entry.NewService(newMockEntryRepo(), newMockTagRepo())
	userID := uuid.New()

	t.Run("creates entry with basic fields", func(t *testing.T) {
		resp, err := svc.Create(context.Background(), userID, entry.CreateEntryRequest{
			Type:    domain.EntryTypeLearning,
			Title:   "Go Interfaces",
			Content: "Interfaces in Go are implicitly satisfied.",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if resp.Title != "Go Interfaces" {
			t.Errorf("Title = %q, want %q", resp.Title, "Go Interfaces")
		}
		if resp.Type != domain.EntryTypeLearning {
			t.Errorf("Type = %q, want %q", resp.Type, domain.EntryTypeLearning)
		}
	})

	t.Run("creates entry with tags", func(t *testing.T) {
		resp, err := svc.Create(context.Background(), userID, entry.CreateEntryRequest{
			Type:    domain.EntryTypeResource,
			Title:   "Awesome Go",
			Content: "A curated list of Go frameworks.",
			Tags:    []string{"go", "resources"},
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if len(resp.Tags) != 2 {
			t.Errorf("Tags count = %d, want 2", len(resp.Tags))
		}
	})
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	svc := entry.NewService(newMockEntryRepo(), newMockTagRepo())
	userID := uuid.New()

	created, _ := svc.Create(context.Background(), userID, entry.CreateEntryRequest{
		Type:    domain.EntryTypeLearning,
		Title:   "Test Entry",
		Content: "Test content",
	})

	t.Run("existing entry returns correctly", func(t *testing.T) {
		got, err := svc.Get(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Title != "Test Entry" {
			t.Errorf("Title = %q, want %q", got.Title, "Test Entry")
		}
	})

	t.Run("non-existent entry returns error", func(t *testing.T) {
		_, err := svc.Get(context.Background(), uuid.New())
		if err == nil {
			t.Fatal("Get() expected error for non-existent entry")
		}
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	svc := entry.NewService(newMockEntryRepo(), newMockTagRepo())
	userID := uuid.New()

	created, _ := svc.Create(context.Background(), userID, entry.CreateEntryRequest{
		Type:    domain.EntryTypeWorkLog,
		Title:   "To Delete",
		Content: "This will be deleted.",
	})

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := svc.Get(context.Background(), created.ID)
	if err == nil {
		t.Fatal("Get() after Delete() should return error")
	}
}
