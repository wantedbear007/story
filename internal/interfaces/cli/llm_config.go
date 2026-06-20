package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newLLMConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "llm-config",
		Short: "Configure LLM provider settings",
		Long:  "Interactively configure the LLM provider (OpenAI, Gemini, Ollama, Anthropic) and save to your config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLLMConfig()
		},
	}
}

func configFilePath() string {
	if p := os.Getenv("STORY_CONFIG_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(home, ".story", "config.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "configs/config.yaml"
}

func runLLMConfig() error {
	configPath := configFilePath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config file %s: %w", configPath, err)
	}

	cfg := &initConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	fmt.Fprintln(os.Stderr, "── LLM Configuration ──")

	currentProvider := cfg.LLM.Provider
	if currentProvider == "" {
		currentProvider = "openai"
	}

	cfg.LLM.Provider = promptDefault("LLM provider (gemini/openai/ollama/anthropic)", currentProvider,
		func(v string) string {
			switch strings.ToLower(v) {
			case "gemini", "openai", "ollama", "anthropic":
				return strings.ToLower(v)
			default:
				return ""
			}
		})

	switch cfg.LLM.Provider {
	case "openai":
		defaultKey := cfg.LLM.OpenAI.APIKey
		if defaultKey == "" {
			cfg.LLM.OpenAI.APIKey = promptRequired("OpenAI API key")
		} else {
			fmt.Fprintf(os.Stderr, "OpenAI API key [%s]: ", maskString(defaultKey))
			input, _ := lineReader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				cfg.LLM.OpenAI.APIKey = input
			}
		}
		defaultModel := cfg.LLM.OpenAI.Model
		if defaultModel == "" {
			defaultModel = "gpt-4"
		}
		cfg.LLM.OpenAI.Model = promptDefault("OpenAI model", defaultModel, nil)
		cfg.LLM.Gemini = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Ollama = struct {
			BaseURL string `yaml:"base_url,omitempty"`
			Model   string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Anthropic = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}

	case "gemini":
		defaultKey := cfg.LLM.Gemini.APIKey
		if defaultKey == "" {
			cfg.LLM.Gemini.APIKey = promptRequired("Gemini API key")
		} else {
			fmt.Fprintf(os.Stderr, "Gemini API key [%s]: ", maskString(defaultKey))
			input, _ := lineReader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				cfg.LLM.Gemini.APIKey = input
			}
		}
		defaultModel := cfg.LLM.Gemini.Model
		if defaultModel == "" {
			defaultModel = "gemini-pro"
		}
		cfg.LLM.Gemini.Model = promptDefault("Gemini model", defaultModel, nil)
		cfg.LLM.OpenAI = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Ollama = struct {
			BaseURL string `yaml:"base_url,omitempty"`
			Model   string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Anthropic = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}

	case "ollama":
		defaultURL := cfg.LLM.Ollama.BaseURL
		if defaultURL == "" {
			defaultURL = "http://localhost:11434"
		}
		cfg.LLM.Ollama.BaseURL = promptDefault("Ollama base URL", defaultURL, nil)
		defaultModel := cfg.LLM.Ollama.Model
		if defaultModel == "" {
			defaultModel = "llama2"
		}
		cfg.LLM.Ollama.Model = promptDefault("Ollama model", defaultModel, nil)
		cfg.LLM.OpenAI = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Gemini = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Anthropic = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}

	case "anthropic":
		defaultKey := cfg.LLM.Anthropic.APIKey
		if defaultKey == "" {
			cfg.LLM.Anthropic.APIKey = promptRequired("Anthropic API key")
		} else {
			fmt.Fprintf(os.Stderr, "Anthropic API key [%s]: ", maskString(defaultKey))
			input, _ := lineReader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				cfg.LLM.Anthropic.APIKey = input
			}
		}
		defaultModel := cfg.LLM.Anthropic.Model
		if defaultModel == "" {
			defaultModel = "claude-3-opus-20240229"
		}
		cfg.LLM.Anthropic.Model = promptDefault("Anthropic model", defaultModel, nil)
		cfg.LLM.OpenAI = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Gemini = struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		}{}
		cfg.LLM.Ollama = struct {
			BaseURL string `yaml:"base_url,omitempty"`
			Model   string `yaml:"model,omitempty"`
		}{}
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(configPath, out, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ LLM configuration saved to %s\n", configPath)
	return nil
}
