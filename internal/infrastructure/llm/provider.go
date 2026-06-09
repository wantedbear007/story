package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/anomalyco/story/internal/infrastructure/config"
)

var (
	ErrProviderNotAvailable = errors.New("LLM provider not available")
	ErrContextLengthExceeded = errors.New("context length exceeded")
	ErrRateLimited          = errors.New("rate limited by provider")
)

// Provider is the abstraction over LLM backends.
// Implementations exist for OpenAI, Gemini, Ollama, and Anthropic.
// The provider abstraction enables swapping LLM backends without
// changing business logic — an application concern, not domain.
type Provider interface {
	// Complete sends a prompt and returns the model's completion.
	Complete(ctx context.Context, prompt string, opts *CompleteOptions) (*Result, error)

	// Name returns the provider name for identification/logging.
	Name() string
}

// CompleteOptions controls model behavior at inference time.
type CompleteOptions struct {
	Model       string
	Temperature float64
	MaxTokens   int
	TopP        float64
}

// Result wraps the LLM response with usage metadata.
type Result struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int
}

// NewProvider creates the appropriate LLM provider based on configuration.
func NewProvider(cfg config.LLMConfig) (Provider, error) {
	switch cfg.Provider {
	case "openai":
		return NewOpenAIProvider(cfg.OpenAI)
	case "gemini":
		return NewGeminiProvider(cfg.Gemini)
	case "ollama":
		return NewOllamaProvider(cfg.Ollama)
	case "anthropic":
		return NewAnthropicProvider(cfg.Anthropic)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}
