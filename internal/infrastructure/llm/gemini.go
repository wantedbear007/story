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

type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiProvider(cfg config.GeminiLLMConfig) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%w: Gemini API key not configured", ErrProviderNotAvailable)
	}
	return &GeminiProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Complete(ctx context.Context, prompt string, opts *CompleteOptions) (*Result, error) {
	model := p.model
	if opts != nil && opts.Model != "" {
		model = opts.Model
	}

	maxTokens := 2048
	if opts != nil && opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": maxTokens,
		},
	}

	body, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, p.apiKey)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	json.Unmarshal(respBody, &result)

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	return &Result{
		Content: result.Candidates[0].Content.Parts[0].Text,
		Model:   model,
	}, nil
}
