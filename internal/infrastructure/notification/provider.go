package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/anomalyco/story/internal/infrastructure/config"
)

var ErrProviderNotAvailable = errors.New("notification provider not available")

func NewProvider(cfg config.NotificationConfig) (domain.NotificationProvider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("notification not enabled: %w", ErrProviderNotAvailable)
	}
	return &platformProvider{cfg: cfg}, nil
}

type platformProvider struct {
	cfg config.NotificationConfig
}

func (p *platformProvider) Name() string {
	return "platform"
}

func (p *platformProvider) Notify(ctx context.Context, req domain.NotificationRequest) error {
	title := req.Title
	if title == "" {
		title = p.cfg.Title
	}
	message := req.Message
	if message == "" {
		message = p.cfg.Message
	}
	return notifyPlatform(ctx, domain.NotificationRequest{
		Title:   title,
		Message: message,
		URL:     req.URL,
	})
}
