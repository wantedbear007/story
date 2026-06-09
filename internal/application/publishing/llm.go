package publishing

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
)

// LLMProvider is the interface the publisher needs from the LLM system.
// Defined here in the application layer so infrastructure can implement it.
type LLMProvider interface {
	Complete(ctx context.Context, prompt string, maxTokens int) (string, error)
	Name() string
}

// LLMPublisher implements the Publisher interface by using an LLM
// to transform entries before publishing. This separation keeps
// transformation logic (LLM calls) separate from delivery logic.
type LLMPublisher struct {
	provider LLMProvider
}

func NewPublisher(provider LLMProvider) *LLMPublisher {
	return &LLMPublisher{provider: provider}
}

func (p *LLMPublisher) Publish(ctx context.Context, entry *domain.Entry, target *domain.PublishingTarget) (string, error) {
	switch target.Type {
	case domain.PublishTargetTwitter:
		return p.publishTwitter(ctx, entry)
	case domain.PublishTargetBlog:
		return p.publishBlog(ctx, entry)
	case domain.PublishTargetMarkdown:
		return entry.Content, nil
	default:
		return "", fmt.Errorf("publishing to %s is not yet supported", target.Type)
	}
}

func (p *LLMPublisher) publishTwitter(ctx context.Context, entry *domain.Entry) (string, error) {
	prompt := fmt.Sprintf(
		"Summarize the following into a tweet under 280 characters:\n\nTitle: %s\n\nContent: %s",
		entry.Title, entry.Content,
	)

	result, err := p.provider.Complete(ctx, prompt, 100)
	if err != nil {
		return "", fmt.Errorf("LLM summarization failed: %w", err)
	}

	return result, nil
}

func (p *LLMPublisher) publishBlog(ctx context.Context, entry *domain.Entry) (string, error) {
	prompt := fmt.Sprintf(
		"Convert this note into a polished blog post with proper structure:\n\nTitle: %s\n\nContent: %s",
		entry.Title, entry.Content,
	)

	result, err := p.provider.Complete(ctx, prompt, 2048)
	if err != nil {
		return "", fmt.Errorf("LLM blog generation failed: %w", err)
	}

	return result, nil
}
