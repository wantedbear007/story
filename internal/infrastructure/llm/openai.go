package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anomalyco/story/internal/infrastructure/config"
)

// OpenAIProvider communicates with the OpenAI API.
type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIProvider(cfg config.OpenAILLMConfig) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%w: OpenAI API key not configured", ErrProviderNotAvailable)
	}
	return &OpenAIProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Complete(ctx context.Context, prompt string, opts *CompleteOptions) (*Result, error) {
	model := p.model
	if opts != nil && opts.Model != "" {
		model = opts.Model
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2048,
	}
	if opts != nil {
		if opts.Temperature > 0 {
			requestBody["temperature"] = opts.Temperature
		}
		if opts.MaxTokens > 0 {
			requestBody["max_tokens"] = opts.MaxTokens
		}
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, ErrRateLimited
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	return &Result{
		Content:      result.Choices[0].Message.Content,
		Model:        model,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
	}, nil
}
