package content

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type GenerateRequest struct {
	EntryID     uuid.UUID
	PromptName  string
	Temperature float64
	MaxTokens   int
}

type RegenerateRequest struct {
	TweetID     uuid.UUID
	PromptName  string
	Temperature float64
	MaxTokens   int
}

type ApproveRequest struct {
	TweetID uuid.UUID
}

type ScheduleRequest struct {
	TweetID     uuid.UUID
	ScheduledAt time.Time
}

type ListRequest struct {
	UserID  uuid.UUID
	EntryID *uuid.UUID
	Status  *domain.TweetStatus
	Limit   int
	Offset  int
}

type TweetResponse struct {
	ID           uuid.UUID                `json:"id"`
	EntryID      uuid.UUID                `json:"entry_id"`
	Content      string                   `json:"content"`
	Status       domain.TweetStatus       `json:"status"`
	Version      int                      `json:"version"`
	PromptName   string                   `json:"prompt_name,omitempty"`
	PromptVer    int                      `json:"prompt_version,omitempty"`
	ProviderName string                   `json:"provider_name,omitempty"`
	ModelName    string                   `json:"model_name,omitempty"`
	InputTokens  int                      `json:"input_tokens,omitempty"`
	OutputTokens int                      `json:"output_tokens,omitempty"`
	CostUSD      float64                  `json:"cost_usd,omitempty"`
	RetryCount   int                      `json:"retry_count,omitempty"`
	LatencyMs    int                      `json:"latency_ms,omitempty"`
	ErrorMessage string                   `json:"error_message,omitempty"`
	ScheduledFor *time.Time               `json:"scheduled_for,omitempty"`
	PostedAt     *time.Time               `json:"posted_at,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

type ListResponse struct {
	Tweets []TweetResponse `json:"tweets"`
	Total  int             `json:"total"`
}

type AuditResponse struct {
	ID              uuid.UUID                `json:"id"`
	TweetID         uuid.UUID                `json:"tweet_id"`
	Action          string                   `json:"action"`
	UserID          *uuid.UUID               `json:"user_id,omitempty"`
	PreviousContent string                   `json:"previous_content,omitempty"`
	NewContent      string                   `json:"new_content,omitempty"`
	PreviousStatus  *domain.TweetStatus      `json:"previous_status,omitempty"`
	NewStatus       *domain.TweetStatus      `json:"new_status,omitempty"`
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
}

func tweetToResponse(t *domain.Tweet, promptName string, promptVer int) TweetResponse {
	return TweetResponse{
		ID:           t.ID,
		EntryID:      t.EntryID,
		Content:      t.Content,
		Status:       t.Status,
		Version:      t.Version,
		PromptName:   promptName,
		PromptVer:    promptVer,
		ProviderName: t.ProviderName,
		ModelName:    t.ModelName,
		InputTokens:  t.InputTokens,
		OutputTokens: t.OutputTokens,
		CostUSD:      t.CostUSD,
		RetryCount:   t.RetryCount,
		LatencyMs:    t.LatencyMs,
		ErrorMessage: t.ErrorMessage,
		ScheduledFor: t.ScheduledFor,
		PostedAt:     t.PostedAt,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

func auditToResponse(a *domain.GenerationAudit) AuditResponse {
	return AuditResponse{
		ID:              a.ID,
		TweetID:         a.TweetID,
		Action:          a.Action,
		UserID:          a.UserID,
		PreviousContent: a.PreviousContent,
		NewContent:      a.NewContent,
		PreviousStatus:  a.PreviousStatus,
		NewStatus:       a.NewStatus,
		Metadata:        a.Metadata,
		CreatedAt:       a.CreatedAt,
	}
}
