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

type AnthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropicProvider(cfg config.AnthropicLLMConfig) (*AnthropicProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%w: Anthropic API key not configured", ErrProviderNotAvailable)
	}
	return &AnthropicProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Complete(ctx context.Context, prompt string, opts *CompleteOptions) (*Result, error) {
	model := p.model
	if opts != nil && opts.Model != "" {
		model = opts.Model
	}

	maxTokens := 2048
	if opts != nil && opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	json.Unmarshal(respBody, &result)

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no content returned from Anthropic")
	}

	return &Result{
		Content:      result.Content[0].Text,
		Model:        model,
		InputTokens:  result.Usage.InputTokens,
		OutputTokens: result.Usage.OutputTokens,
	}, nil
}
