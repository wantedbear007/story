package notification_test

import (
	"context"
	"testing"

	"github.com/anomalyco/story/internal/application/notification"
	"github.com/anomalyco/story/internal/domain"
)

type mockProvider struct {
	lastReq *domain.NotificationRequest
}

func (m *mockProvider) Notify(_ context.Context, req domain.NotificationRequest) error {
	m.lastReq = &req
	return nil
}

func (m *mockProvider) Name() string { return "mock" }

func TestService_Send(t *testing.T) {
	t.Parallel()

	mock := &mockProvider{}
	svc := notification.NewService(mock)

	err := svc.Send(context.Background(), domain.NotificationRequest{
		Title:   "Test Title",
		Message: "Test Message",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastReq == nil {
		t.Fatal("expected notification to be sent")
	}
	if mock.lastReq.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %q", mock.lastReq.Title)
	}
	if mock.lastReq.Message != "Test Message" {
		t.Errorf("expected message 'Test Message', got %q", mock.lastReq.Message)
	}
}
