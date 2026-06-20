package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/raw_entry"
	"github.com/anomalyco/story/internal/domain"
	"github.com/anomalyco/story/internal/interfaces/api"
)

type mockRawEntryRepo struct {
	mu      sync.Mutex
	entries []*domain.RawEntry
}

func (m *mockRawEntryRepo) Create(_ context.Context, e *domain.RawEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, e)
	return nil
}

func (m *mockRawEntryRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.RawEntry, error) {
	return nil, nil
}

func (m *mockRawEntryRepo) List(_ context.Context, _ domain.RawEntryFilter) ([]*domain.RawEntry, error) {
	return nil, nil
}

func (m *mockRawEntryRepo) UpdateContent(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockRawEntryRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.RawEntryStatus) error {
	return nil
}

func (m *mockRawEntryRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func newTestCaptureServer(t *testing.T) *api.CaptureServer {
	t.Helper()
	rawSvc := raw_entry.NewService(&mockRawEntryRepo{})
	return api.NewCaptureServer("127.0.0.1", 0, rawSvc)
}

func TestCaptureHandler_MissingContent(t *testing.T) {
	t.Parallel()
	srv := newTestCaptureServer(t)

	body, _ := json.Marshal(map[string]string{
		"user_id": uuid.New().String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/capture", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCaptureHandler_MissingUserID(t *testing.T) {
	t.Parallel()
	srv := newTestCaptureServer(t)

	body, _ := json.Marshal(map[string]string{
		"content": "test content",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/capture", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCaptureHandler_Success(t *testing.T) {
	t.Parallel()
	srv := newTestCaptureServer(t)

	userID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{
		"content": "test content",
		"user_id": userID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/capture", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp raw_entry.RawEntryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Content != "test content" {
		t.Errorf("expected content 'test content', got %q", resp.Content)
	}
	if resp.Source != domain.RawEntrySourceNotificationCapture {
		t.Errorf("expected source notification_capture, got %s", resp.Source)
	}
}
