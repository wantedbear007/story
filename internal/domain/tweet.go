package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TweetStatus string

const (
	TweetStatusDraft     TweetStatus = "draft"
	TweetStatusReviewing TweetStatus = "reviewing"
	TweetStatusApproved  TweetStatus = "approved"
	TweetStatusScheduled TweetStatus = "scheduled"
	TweetStatusPosted    TweetStatus = "posted"
	TweetStatusArchived  TweetStatus = "archived"
)

type Tweet struct {
	ID           uuid.UUID
	EntryID      uuid.UUID
	UserID       uuid.UUID
	Content      string
	Status       TweetStatus
	Version      int
	PromptID     uuid.UUID
	ProviderName string
	ModelName    string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	RetryCount   int
	LatencyMs    int
	ErrorMessage string
	ScheduledFor *time.Time
	PostedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PromptTemplate struct {
	ID          uuid.UUID
	Name        string
	Version     int
	Template    string
	Description string
	CreatedAt   time.Time
}

type GenerationAudit struct {
	ID              uuid.UUID
	TweetID         uuid.UUID
	Action          string
	UserID          *uuid.UUID
	PreviousContent string
	NewContent      string
	PreviousStatus  *TweetStatus
	NewStatus       *TweetStatus
	Metadata        map[string]interface{}
	CreatedAt       time.Time
}

type TweetFilter struct {
	UserID  uuid.UUID
	EntryID *uuid.UUID
	Status  *TweetStatus
	Limit   int
	Offset  int
}

type TweetRepository interface {
	Create(ctx context.Context, tweet *Tweet) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tweet, error)
	List(ctx context.Context, filter TweetFilter) ([]*Tweet, error)
	Update(ctx context.Context, tweet *Tweet) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status TweetStatus) error

	CreateAudit(ctx context.Context, audit *GenerationAudit) error
	ListAudits(ctx context.Context, tweetID uuid.UUID) ([]*GenerationAudit, error)
}

type PromptTemplateRepository interface {
	Create(ctx context.Context, prompt *PromptTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*PromptTemplate, error)
	GetLatestByName(ctx context.Context, name string) (*PromptTemplate, error)
	List(ctx context.Context) ([]*PromptTemplate, error)
}
