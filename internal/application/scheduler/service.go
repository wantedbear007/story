package scheduler

import (
	"context"
	"time"

	"github.com/anomalyco/story/internal/application/notification"
	"github.com/anomalyco/story/internal/domain"
)

type Service struct {
	notif     *notification.Service
	cfg       Config
	ticker    *time.Ticker
	done      chan struct{}
	sentToday bool
}

type Config struct {
	Enabled    bool
	Hour       int
	Minute     int
	CaptureURL string
}

func NewService(notif *notification.Service, cfg Config) *Service {
	return &Service{
		notif: notif,
		cfg:   cfg,
		done:  make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) {
	if !s.cfg.Enabled {
		return
	}
	s.ticker = time.NewTicker(30 * time.Second)
	go func() {
		defer s.ticker.Stop()
		for {
			select {
			case <-s.ticker.C:
				s.checkAndNotify(ctx)
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Service) Stop() {
	close(s.done)
}

func (s *Service) checkAndNotify(ctx context.Context) {
	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), s.cfg.Hour, s.cfg.Minute, 0, 0, now.Location())
	if now.After(target) || now.Equal(target) {
		if !s.sentToday {
			s.sentToday = true
			_ = s.notif.Send(ctx, domain.NotificationRequest{
				URL: s.cfg.CaptureURL,
			})
		}
	}
	if now.Day() != target.Day() || now.Month() != target.Month() || now.Year() != target.Year() {
		s.sentToday = false
	}
}
