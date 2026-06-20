package notification

import (
	"context"

	"github.com/anomalyco/story/internal/domain"
)

type Service struct {
	provider domain.NotificationProvider
}

func NewService(provider domain.NotificationProvider) *Service {
	return &Service{provider: provider}
}

func (s *Service) Send(ctx context.Context, req domain.NotificationRequest) error {
	return s.provider.Notify(ctx, req)
}
