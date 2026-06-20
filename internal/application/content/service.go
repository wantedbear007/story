package content

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

const (
	defaultPromptName = "tweet-summarize"
	maxRetries        = 3
)

var (
	// Model cost per 1K tokens in USD (input, output).
	modelCosts = map[string][2]float64{
		"gpt-4":                {0.03, 0.06},
		"gpt-4-turbo":          {0.01, 0.03},
		"gpt-3.5-turbo":        {0.0015, 0.002},
		"gemini-pro":           {0.0005, 0.0015},
		"gemini-1.5-pro":       {0.0035, 0.0105},
		"claude-3-opus-20240229": {0.015, 0.075},
		"claude-3-sonnet-20240229": {0.003, 0.015},
		"claude-3-haiku-20240307": {0.00025, 0.00125},
	}
	defaultCost = [2]float64{0.01, 0.03}
)

type LLMProvider interface {
	Complete(ctx context.Context, prompt string, maxTokens int) (string, error)
	Name() string
}

type Service struct {
	tweetRepo  domain.TweetRepository
	promptRepo domain.PromptTemplateRepository
	entryRepo  domain.EntryRepository
	provider   LLMProvider

	mu               sync.Mutex
	lastHealthy      bool
	lastCheckTime    time.Time
	healthyStreak    int
	unhealthyStreak  int
	onHealthy        chan struct{}
}

func NewService(
	tweetRepo domain.TweetRepository,
	promptRepo domain.PromptTemplateRepository,
	entryRepo domain.EntryRepository,
	provider LLMProvider,
) *Service {
	return &Service{
		tweetRepo:  tweetRepo,
		promptRepo: promptRepo,
		entryRepo:  entryRepo,
		provider:   provider,
		onHealthy:  make(chan struct{}, 1),
	}
}

func (s *Service) IsLLMConfigured() bool {
	return s.provider != nil
}

func (s *Service) HealthyChan() <-chan struct{} {
	return s.onHealthy
}

func (s *Service) IsHealthy(ctx context.Context) bool {
	s.mu.Lock()
	if s.provider == nil {
		s.mu.Unlock()
		return false
	}
	if time.Since(s.lastCheckTime) < 30*time.Second {
		ok := s.lastHealthy
		s.mu.Unlock()
		return ok
	}
	s.mu.Unlock()

	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_, err := s.provider.Complete(probeCtx, "Say 'Hello from Story!' in one short sentence.", 50)

	s.mu.Lock()
	s.lastCheckTime = time.Now()
	probeOk := err == nil

	const threshold = 2
	if probeOk {
		s.healthyStreak++
		s.unhealthyStreak = 0
		if s.healthyStreak >= threshold && !s.lastHealthy {
			s.lastHealthy = true
			select {
			case s.onHealthy <- struct{}{}:
			default:
			}
		}
	} else {
		s.unhealthyStreak++
		s.healthyStreak = 0
		if s.unhealthyStreak >= threshold && s.lastHealthy {
			s.lastHealthy = false
		}
	}

	ok := s.lastHealthy
	s.mu.Unlock()
	return ok
}

type GenerateResult struct {
	Content      string
	ProviderName string
	ModelName    string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	RetryCount   int
	LatencyMs    int
}

func (s *Service) Generate(ctx context.Context, userID uuid.UUID, req GenerateRequest) (*TweetResponse, error) {
	entry, err := s.entryRepo.GetByID(ctx, req.EntryID)
	if err != nil {
		return nil, fmt.Errorf("entry not found: %w", err)
	}
	if entry.UserID != userID {
		return nil, fmt.Errorf("%w: entry does not belong to user", domain.ErrForbidden)
	}

	promptName := req.PromptName
	if promptName == "" {
		promptName = defaultPromptName
	}
	prompt, err := s.promptRepo.GetLatestByName(ctx, promptName)
	if err != nil {
		return nil, fmt.Errorf("prompt template %q not found: %w", promptName, err)
	}

	rendered, err := renderPrompt(prompt.Template, entry)
	if err != nil {
		return nil, fmt.Errorf("rendering prompt template: %w", err)
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 100
	}

	result, err := s.generateWithRetry(ctx, rendered, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("generation failed after %d retries: %w", maxRetries, err)
	}

	tweet := &domain.Tweet{
		ID:           uuid.New(),
		EntryID:      req.EntryID,
		UserID:       userID,
		Content:      result.Content,
		Status:       domain.TweetStatusDraft,
		Version:      1,
		PromptID:     prompt.ID,
		ProviderName: result.ProviderName,
		ModelName:    result.ModelName,
		InputTokens:  result.InputTokens,
		OutputTokens: result.OutputTokens,
		CostUSD:      result.CostUSD,
		RetryCount:   result.RetryCount,
		LatencyMs:    result.LatencyMs,
	}

	if err := s.tweetRepo.Create(ctx, tweet); err != nil {
		return nil, fmt.Errorf("saving tweet: %w", err)
	}

	audit := &domain.GenerationAudit{
		ID:         uuid.New(),
		TweetID:    tweet.ID,
		Action:     "generated",
		NewContent: result.Content,
		NewStatus:  ptr(domain.TweetStatusDraft),
	}

	if err := s.tweetRepo.CreateAudit(ctx, audit); err != nil {
		return nil, fmt.Errorf("saving audit: %w", err)
	}

	resp := tweetToResponse(tweet, promptName, prompt.Version)
	return &resp, nil
}

func (s *Service) Regenerate(ctx context.Context, userID uuid.UUID, req RegenerateRequest) (*TweetResponse, error) {
	existing, err := s.tweetRepo.GetByID(ctx, req.TweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if existing.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}

	if existing.Status == domain.TweetStatusPosted {
		return nil, fmt.Errorf("%w: cannot regenerate a posted tweet", domain.ErrInvalidInput)
	}

	entry, err := s.entryRepo.GetByID(ctx, existing.EntryID)
	if err != nil {
		return nil, fmt.Errorf("entry not found: %w", err)
	}

	promptName := req.PromptName
	if promptName == "" {
		promptName = defaultPromptName
	}
	prompt, err := s.promptRepo.GetLatestByName(ctx, promptName)
	if err != nil {
		return nil, fmt.Errorf("prompt template %q not found: %w", promptName, err)
	}

	rendered, err := renderPrompt(prompt.Template, entry)
	if err != nil {
		return nil, fmt.Errorf("rendering prompt template: %w", err)
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 100
	}

	result, err := s.generateWithRetry(ctx, rendered, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("regeneration failed after %d retries: %w", maxRetries, err)
	}

	oldContent := existing.Content
	oldStatus := existing.Status

	existing.Status = domain.TweetStatusArchived
	if err := s.tweetRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("archiving old version: %w", err)
	}

	tweet := &domain.Tweet{
		ID:           uuid.New(),
		EntryID:      existing.EntryID,
		UserID:       userID,
		Content:      result.Content,
		Status:       domain.TweetStatusDraft,
		Version:      existing.Version + 1,
		PromptID:     prompt.ID,
		ProviderName: result.ProviderName,
		ModelName:    result.ModelName,
		InputTokens:  result.InputTokens,
		OutputTokens: result.OutputTokens,
		CostUSD:      result.CostUSD,
		RetryCount:   result.RetryCount,
		LatencyMs:    result.LatencyMs,
	}

	if err := s.tweetRepo.Create(ctx, tweet); err != nil {
		return nil, fmt.Errorf("saving regenerated tweet: %w", err)
	}

	audit := &domain.GenerationAudit{
		ID:              uuid.New(),
		TweetID:         tweet.ID,
		Action:          "regenerated",
		PreviousContent: oldContent,
		NewContent:      result.Content,
		PreviousStatus:  &oldStatus,
		NewStatus:       ptr(domain.TweetStatusDraft),
	}

	if err := s.tweetRepo.CreateAudit(ctx, audit); err != nil {
		return nil, fmt.Errorf("saving audit: %w", err)
	}

	resp := tweetToResponse(tweet, promptName, prompt.Version)
	return &resp, nil
}

func (s *Service) Approve(ctx context.Context, userID uuid.UUID, req ApproveRequest) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, req.TweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	if tweet.Status != domain.TweetStatusDraft && tweet.Status != domain.TweetStatusReviewing {
		return nil, fmt.Errorf("%w: can only approve draft or reviewing tweets", domain.ErrInvalidInput)
	}

	oldStatus := tweet.Status
	tweet.Status = domain.TweetStatusApproved
	if err := s.tweetRepo.UpdateStatus(ctx, tweet.ID, domain.TweetStatusApproved); err != nil {
		return nil, fmt.Errorf("approving tweet: %w", err)
	}

	s.logAudit(ctx, tweet.ID, "approved", userID, "", tweet.Content, &oldStatus, ptr(domain.TweetStatusApproved))
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) Review(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID) (*TweetResponse, error) {
	return s.transitionStatus(ctx, userID, tweetID, domain.TweetStatusReviewing, []domain.TweetStatus{domain.TweetStatusDraft}, "review")
}

func (s *Service) Reject(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	if tweet.Status != domain.TweetStatusReviewing && tweet.Status != domain.TweetStatusApproved {
		return nil, fmt.Errorf("%w: can only reject reviewing or approved tweets", domain.ErrInvalidInput)
	}

	oldStatus := tweet.Status
	tweet.Status = domain.TweetStatusDraft
	if err := s.tweetRepo.UpdateStatus(ctx, tweet.ID, domain.TweetStatusDraft); err != nil {
		return nil, fmt.Errorf("rejecting tweet: %w", err)
	}

	s.logAudit(ctx, tweet.ID, "rejected", userID, "", tweet.Content, &oldStatus, ptr(domain.TweetStatusDraft))
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) Schedule(ctx context.Context, userID uuid.UUID, req ScheduleRequest) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, req.TweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	if tweet.Status != domain.TweetStatusApproved {
		return nil, fmt.Errorf("%w: can only schedule approved tweets", domain.ErrInvalidInput)
	}

	oldStatus := tweet.Status
	tweet.Status = domain.TweetStatusScheduled
	tweet.ScheduledFor = &req.ScheduledAt
	if err := s.tweetRepo.Update(ctx, tweet); err != nil {
		return nil, fmt.Errorf("scheduling tweet: %w", err)
	}

	s.logAudit(ctx, tweet.ID, "scheduled", userID, "", tweet.Content, &oldStatus, ptr(domain.TweetStatusScheduled))
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) Archive(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	if tweet.Status == domain.TweetStatusArchived || tweet.Status == domain.TweetStatusPosted {
		return nil, fmt.Errorf("%w: cannot archive a %s tweet", domain.ErrInvalidInput, tweet.Status)
	}

	oldStatus := tweet.Status
	tweet.Status = domain.TweetStatusArchived
	if err := s.tweetRepo.UpdateStatus(ctx, tweet.ID, domain.TweetStatusArchived); err != nil {
		return nil, fmt.Errorf("archiving tweet: %w", err)
	}

	s.logAudit(ctx, tweet.ID, "archived", userID, "", tweet.Content, &oldStatus, ptr(domain.TweetStatusArchived))
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	filter := domain.TweetFilter{
		UserID:  req.UserID,
		EntryID: req.EntryID,
		Status:  req.Status,
		Limit:   limit,
		Offset:  req.Offset,
	}
	tweets, err := s.tweetRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing tweets: %w", err)
	}
	responses := make([]TweetResponse, len(tweets))
	for i, t := range tweets {
		promptName, promptVer := s.getPromptInfo(ctx, t.PromptID)
		responses[i] = tweetToResponse(t, promptName, promptVer)
	}
	return &ListResponse{
		Tweets: responses,
		Total:  len(responses),
	}, nil
}

func (s *Service) UpdateContent(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID, content string) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	if tweet.Status == domain.TweetStatusPosted || tweet.Status == domain.TweetStatusArchived {
		return nil, fmt.Errorf("%w: cannot edit a %s tweet", domain.ErrInvalidInput, tweet.Status)
	}

	tweet.Content = content
	if err := s.tweetRepo.Update(ctx, tweet); err != nil {
		return nil, fmt.Errorf("updating tweet: %w", err)
	}

	s.logAudit(ctx, tweet.ID, "content_updated", userID, "", content, nil, nil)
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) GetAudits(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID) ([]AuditResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	audits, err := s.tweetRepo.ListAudits(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("listing audits: %w", err)
	}
	responses := make([]AuditResponse, len(audits))
	for i, a := range audits {
		responses[i] = auditToResponse(a)
	}
	return responses, nil
}

func (s *Service) transitionStatus(ctx context.Context, userID uuid.UUID, tweetID uuid.UUID, to domain.TweetStatus, allowed []domain.TweetStatus, action string) (*TweetResponse, error) {
	tweet, err := s.tweetRepo.GetByID(ctx, tweetID)
	if err != nil {
		return nil, fmt.Errorf("tweet not found: %w", err)
	}
	if tweet.UserID != userID {
		return nil, fmt.Errorf("%w: tweet does not belong to user", domain.ErrForbidden)
	}
	valid := false
	for _, st := range allowed {
		if tweet.Status == st {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", domain.ErrInvalidInput, tweet.Status, to)
	}
	oldStatus := tweet.Status
	if err := s.tweetRepo.UpdateStatus(ctx, tweet.ID, to); err != nil {
		return nil, fmt.Errorf("%s tweet: %w", action, err)
	}
	s.logAudit(ctx, tweet.ID, action, userID, "", tweet.Content, &oldStatus, ptr(to))
	promptName, promptVer := s.getPromptInfo(ctx, tweet.PromptID)
	resp := tweetToResponse(tweet, promptName, promptVer)
	return &resp, nil
}

func (s *Service) getPromptInfo(ctx context.Context, promptID uuid.UUID) (string, int) {
	p, err := s.promptRepo.GetByID(ctx, promptID)
	if err != nil {
		return "", 0
	}
	return p.Name, p.Version
}

func (s *Service) logAudit(ctx context.Context, tweetID uuid.UUID, action string, userID uuid.UUID, prevContent, newContent string, prevStatus, newStatus *domain.TweetStatus) {
	au := &domain.GenerationAudit{
		ID:              uuid.New(),
		TweetID:         tweetID,
		Action:          action,
		UserID:          &userID,
		PreviousContent: prevContent,
		NewContent:      newContent,
		PreviousStatus:  prevStatus,
		NewStatus:       newStatus,
	}
	_ = s.tweetRepo.CreateAudit(ctx, au)
}

func (s *Service) generateWithRetry(ctx context.Context, prompt string, maxTokens int) (*GenerateResult, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("LLM provider not configured: set STORY_LLM_OPENAI_API_KEY (or Gemini/Anthropic) in environment")
	}

	start := time.Now()
	var lastErr error
	var retries int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration(attempt)):
			}
		}

		content, err := s.provider.Complete(ctx, prompt, maxTokens)
		if err == nil {
			latency := time.Since(start)
			cost := estimateCost(s.provider.Name(), "", content, len(prompt), len(content))
			return &GenerateResult{
				Content:      strings.TrimSpace(content),
				ProviderName: s.provider.Name(),
				InputTokens:  estimateTokens(prompt),
				OutputTokens: estimateTokens(content),
				CostUSD:      cost,
				RetryCount:   retries,
				LatencyMs:    int(latency.Milliseconds()),
			}, nil
		}

		lastErr = err
		retries++
		if !isRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func renderPrompt(tmpl string, entry *domain.Entry) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	data := map[string]string{
		"Title":   entry.Title,
		"Content": entry.Content,
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return buf.String(), nil
}

func estimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / 4.0))
}

func estimateCost(providerName, modelName, content string, inputTokens, outputTokens int) float64 {
	key := modelName
	if key == "" {
		key = providerName
	}
	cost, ok := modelCosts[key]
	if !ok {
		cost = defaultCost
	}
	inputCost := (float64(inputTokens) / 1000.0) * cost[0]
	outputCost := (float64(outputTokens) / 1000.0) * cost[1]
	return inputCost + outputCost
}

func isRetryable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "5") ||
		strings.Contains(msg, "temporarily unavailable")
}

func backoffDuration(attempt int) time.Duration {
	base := time.Duration(math.Pow(2, float64(attempt))) * time.Second / 2
	jitter := time.Duration(rand.Intn(100)) * time.Millisecond
	return base + jitter
}

func ptr[T any](v T) *T {
	return &v
}
