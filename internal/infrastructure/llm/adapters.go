package llm

import "context"

// CompleteAdapter wraps the Provider to implement the application layer's
// LLMProvider interface, which uses a simpler method signature.
// This adapter follows Clean Architecture: infrastructure adapts to application needs.
type CompleteAdapter struct {
	provider Provider
}

func NewCompleteAdapter(provider Provider) *CompleteAdapter {
	return &CompleteAdapter{provider: provider}
}

func (a *CompleteAdapter) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	result, err := a.provider.Complete(ctx, prompt, &CompleteOptions{MaxTokens: maxTokens})
	if err != nil {
		return "", err
	}
	return result.Content, nil
}

func (a *CompleteAdapter) Name() string {
	return a.provider.Name()
}
