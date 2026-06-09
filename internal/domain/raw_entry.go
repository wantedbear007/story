package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RawEntryStatus string

const (
	RawEntryStatusRAW        RawEntryStatus = "raw"
	RawEntryStatusProcessing RawEntryStatus = "processing"
	RawEntryStatusStructured RawEntryStatus = "structured"
	RawEntryStatusArchived   RawEntryStatus = "archived"
)

type RawEntrySource string

const (
	RawEntrySourceCLI    RawEntrySource = "cli"
	RawEntrySourceFile   RawEntrySource = "file"
	RawEntrySourcePipe   RawEntrySource = "pipe"
	RawEntrySourceImport RawEntrySource = "import"
	RawEntrySourceAPI    RawEntrySource = "api"
)

type RawEntry struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Content   string
	Status    RawEntryStatus
	Source    RawEntrySource
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *RawEntry) MarkProcessing() {
	e.Status = RawEntryStatusProcessing
	e.UpdatedAt = time.Now()
}

func (e *RawEntry) MarkStructured() {
	e.Status = RawEntryStatusStructured
	e.UpdatedAt = time.Now()
}

func (e *RawEntry) MarkArchived() {
	e.Status = RawEntryStatusArchived
	e.UpdatedAt = time.Now()
}

type RawEntryFilter struct {
	UserID uuid.UUID
	Status *RawEntryStatus
	Source *RawEntrySource
	Limit  int
	Offset int
}

type RawEntryRepository interface {
	Create(ctx context.Context, entry *RawEntry) error
	GetByID(ctx context.Context, id uuid.UUID) (*RawEntry, error)
	List(ctx context.Context, filter RawEntryFilter) ([]*RawEntry, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status RawEntryStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}
