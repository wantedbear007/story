package domain_test

import (
	"testing"
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

func TestRawEntry_MarkProcessing(t *testing.T) {
	t.Parallel()

	entry := &domain.RawEntry{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Content:   "some raw notes",
		Status:    domain.RawEntryStatusRAW,
		Source:    domain.RawEntrySourceCLI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	entry.MarkProcessing()

	if entry.Status != domain.RawEntryStatusProcessing {
		t.Errorf("expected status processing, got %s", entry.Status)
	}
}

func TestRawEntry_MarkStructured(t *testing.T) {
	t.Parallel()

	entry := &domain.RawEntry{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Content:   "some raw notes",
		Status:    domain.RawEntryStatusRAW,
		Source:    domain.RawEntrySourceCLI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	entry.MarkStructured()

	if entry.Status != domain.RawEntryStatusStructured {
		t.Errorf("expected status structured, got %s", entry.Status)
	}
}

func TestRawEntry_MarkArchived(t *testing.T) {
	t.Parallel()

	entry := &domain.RawEntry{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Content:   "some raw notes",
		Status:    domain.RawEntryStatusRAW,
		Source:    domain.RawEntrySourceCLI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	entry.MarkArchived()

	if entry.Status != domain.RawEntryStatusArchived {
		t.Errorf("expected status archived, got %s", entry.Status)
	}
}
