package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/application/raw_entry"
	"github.com/anomalyco/story/internal/domain"
)

type Service struct {
	rawEntrySvc *raw_entry.Service
	entrySvc    *entry.Service
	tweetSvc    *content.Service
}

func NewService(rawEntrySvc *raw_entry.Service, entrySvc *entry.Service, tweetSvc *content.Service) *Service {
	return &Service{
		rawEntrySvc: rawEntrySvc,
		entrySvc:    entrySvc,
		tweetSvc:    tweetSvc,
	}
}

func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processPending(ctx)
		}
	}
}

func (s *Service) processPending(ctx context.Context) {
	if !s.tweetSvc.IsHealthy(ctx) {
		return
	}

	entries, err := s.rawEntrySvc.List(ctx, raw_entry.ListRawEntriesRequest{
		Status: ptr(domain.RawEntryStatusRAW),
		Limit:  5,
	})
	if err != nil {
		return
	}

	for _, re := range entries.Entries {
		if err := s.processOne(ctx, re); err != nil {
			continue
		}
	}
}

func (s *Service) processOne(ctx context.Context, re raw_entry.RawEntryResponse) error {
	id := re.ID

	if _, err := s.rawEntrySvc.UpdateStatus(ctx, id, re.UserID, domain.RawEntryStatusProcessing); err != nil {
		return fmt.Errorf("mark processing: %w", err)
	}

	title := firstLine(re.Content)
	entryResp, err := s.entrySvc.Create(ctx, re.UserID, entry.CreateEntryRequest{
		Type:    domain.EntryTypeLearning,
		Title:   title,
		Content: re.Content,
	})
	if err != nil {
		s.rawEntrySvc.UpdateStatus(ctx, id, re.UserID, domain.RawEntryStatusFailed)
		return fmt.Errorf("create entry: %w", err)
	}

	_, err = s.tweetSvc.Generate(ctx, re.UserID, content.GenerateRequest{
		EntryID: entryResp.ID,
	})
	if err != nil {
		s.rawEntrySvc.UpdateStatus(ctx, id, re.UserID, domain.RawEntryStatusFailed)
		return fmt.Errorf("generate tweet: %w", err)
	}

	if _, err := s.rawEntrySvc.UpdateStatus(ctx, id, re.UserID, domain.RawEntryStatusStructured); err != nil {
		return fmt.Errorf("mark structured: %w", err)
	}

	return nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	idx := strings.Index(s, "\n")
	if idx == -1 || idx > 60 {
		if len(s) > 60 {
			return s[:60] + "..."
		}
		return s
	}
	return s[:idx]
}

func ptr[T any](v T) *T {
	return &v
}
