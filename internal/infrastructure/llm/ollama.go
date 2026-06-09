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

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(cfg config.OllamaLLMConfig) (*OllamaProvider, error) {
	return &OllamaProvider{
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}, nil
}

func (p *OllamaProvider) Name() string { return "ollama" }

func (p *OllamaProvider) Complete(ctx context.Context, prompt string, opts *CompleteOptions) (*Result, error) {
	model := p.model
	if opts != nil && opts.Model != "" {
		model = opts.Model
	}

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	json.Unmarshal(respBody, &result)

	return &Result{
		Content: result.Response,
		Model:   model,
	}, nil
}
