package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type initConfig struct {
	App struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		LogLevel    string `yaml:"log_level"`
	} `yaml:"app"`
	Server struct {
		Host    string `yaml:"host"`
		Port    int    `yaml:"port"`
		BaseURL string `yaml:"base_url"`
	} `yaml:"server"`
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Name         string `yaml:"name"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		SSLMode      string `yaml:"ssl_mode"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
	} `yaml:"database"`
	Auth struct {
		JWTSecret            string `yaml:"jwt_secret"`
		AccessTokenTTL       string `yaml:"access_token_ttl"`
		RefreshTokenTTL      string `yaml:"refresh_token_ttl"`
		PasswordResetTTL     string `yaml:"password_reset_ttl"`
		EmailVerificationTTL string `yaml:"email_verification_ttl"`
	} `yaml:"auth"`
	LLM struct {
		Provider  string `yaml:"provider"`
		OpenAI    struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		} `yaml:"openai,omitempty"`
		Gemini struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		} `yaml:"gemini,omitempty"`
		Ollama struct {
			BaseURL string `yaml:"base_url,omitempty"`
			Model   string `yaml:"model,omitempty"`
		} `yaml:"ollama,omitempty"`
		Anthropic struct {
			APIKey string `yaml:"api_key,omitempty"`
			Model  string `yaml:"model,omitempty"`
		} `yaml:"anthropic,omitempty"`
	} `yaml:"llm"`
	SMTP struct {
		Host     string `yaml:"host,omitempty"`
		Port     int    `yaml:"port,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
		From     string `yaml:"from,omitempty"`
	} `yaml:"smtp,omitempty"`
}

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Story configuration",
		Long:  "Interactive setup to create your ~/.story/config.yaml configuration file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	cfg := initConfig{}

	cfg.App.Name = "story"
	cfg.App.Environment = promptDefault("Application environment", "development",
		func(v string) string {
			switch v {
			case "development", "staging", "production":
				return v
			default:
				return ""
			}
		})
	cfg.App.LogLevel = "debug"

	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.Server.BaseURL = "http://localhost:8080"

	fmt.Fprintln(os.Stderr, "\n── Database Configuration ──")

	cfg.Database.Host = promptDefault("Database host", "localhost", nil)
	portStr := promptDefault("Database port", "5432",
		func(v string) string {
			p, err := strconv.Atoi(v)
			if err != nil || p < 1 || p > 65535 {
				return ""
			}
			return v
		})
	cfg.Database.Port, _ = strconv.Atoi(portStr)
	cfg.Database.Name = promptDefault("Database name", "story", nil)
	cfg.Database.User = promptDefault("Database username", "story", nil)
	cfg.Database.Password = promptRequired("Database password")
	cfg.Database.SSLMode = promptDefault("SSL mode (disable/require/verify-full)", "disable",
		func(v string) string {
			switch v {
			case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
				return v
			default:
				return ""
			}
		})
	cfg.Database.MaxOpenConns = 25
	cfg.Database.MaxIdleConns = 5

	fmt.Fprintln(os.Stderr, "\n── Authentication ──")

	secret, err := generateJWTSecret()
	if err != nil {
		return fmt.Errorf("generating JWT secret: %w", err)
	}
	cfg.Auth.JWTSecret = secret
	cfg.Auth.AccessTokenTTL = "15m"
	cfg.Auth.RefreshTokenTTL = "720h"
	cfg.Auth.PasswordResetTTL = "1h"
	cfg.Auth.EmailVerificationTTL = "24h"

	fmt.Fprintln(os.Stderr, "\n── SMTP Configuration (optional) ──")

	smtpEnabled := promptDefault("Configure SMTP?", "no",
		func(v string) string {
			v = strings.ToLower(v)
			if v == "yes" || v == "y" || v == "no" || v == "n" {
				return v
			}
			return ""
		})

	if smtpEnabled == "yes" || smtpEnabled == "y" {
		cfg.SMTP.Host = promptRequired("SMTP host")
		portStr := promptDefault("SMTP port", "587",
			func(v string) string {
				p, err := strconv.Atoi(v)
				if err != nil || p < 1 || p > 65535 {
					return ""
				}
				return v
			})
		cfg.SMTP.Port, _ = strconv.Atoi(portStr)
		cfg.SMTP.Username = promptInput("SMTP username: ")
		cfg.SMTP.Password = promptPassword("SMTP password (input hidden): ")
		cfg.SMTP.From = promptDefault("SMTP from address", "story@example.com", nil)
	}

	fmt.Fprintln(os.Stderr, "\n── LLM Configuration (optional) ──")

	llmEnabled := promptDefault("Configure LLM provider?", "no",
		func(v string) string {
			v = strings.ToLower(v)
			if v == "yes" || v == "y" || v == "no" || v == "n" {
				return v
			}
			return ""
		})

	if llmEnabled == "yes" || llmEnabled == "y" {
		cfg.LLM.Provider = promptDefault("LLM provider (gemini/openai/ollama/anthropic)", "openai",
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
			cfg.LLM.OpenAI.APIKey = promptRequired("OpenAI API key")
			cfg.LLM.OpenAI.Model = promptDefault("OpenAI model", "gpt-4", nil)
		case "gemini":
			cfg.LLM.Gemini.APIKey = promptRequired("Gemini API key")
			cfg.LLM.Gemini.Model = promptDefault("Gemini model", "gemini-pro", nil)
		case "ollama":
			cfg.LLM.Ollama.BaseURL = promptDefault("Ollama base URL", "http://localhost:11434", nil)
			cfg.LLM.Ollama.Model = promptDefault("Ollama model", "llama2", nil)
		case "anthropic":
			cfg.LLM.Anthropic.APIKey = promptRequired("Anthropic API key")
			cfg.LLM.Anthropic.Model = promptDefault("Anthropic model", "claude-3-opus-20240229", nil)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configDir := filepath.Join(home, ".story")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Configuration written to %s\n", configPath)
	fmt.Fprintf(os.Stderr, "  JWT Secret: %s\n", cfg.Auth.JWTSecret)
	if cfg.SMTP.Password != "" {
		fmt.Fprintf(os.Stderr, "  SMTP Password: %s\n", cfg.SMTP.Password)
	}
	if cfg.LLM.OpenAI.APIKey != "" {
		fmt.Fprintf(os.Stderr, "  OpenAI API Key: %s\n", cfg.LLM.OpenAI.APIKey)
	}
	if cfg.LLM.Gemini.APIKey != "" {
		fmt.Fprintf(os.Stderr, "  Gemini API Key: %s\n", cfg.LLM.Gemini.APIKey)
	}
	if cfg.LLM.Anthropic.APIKey != "" {
		fmt.Fprintf(os.Stderr, "  Anthropic API Key: %s\n", cfg.LLM.Anthropic.APIKey)
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  All tokens and API keys are stored in the config file above.")
	fmt.Fprintln(os.Stderr, "  To override any value at runtime, set the corresponding environment variable:")
	fmt.Fprintln(os.Stderr, "    export STORY_AUTH_JWT_SECRET=<your-secret>")
	fmt.Fprintln(os.Stderr, "    export STORY_DATABASE_PASSWORD=<your-password>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Run 'story verify' to validate your setup.")
	return nil
}

func promptDefault(label, defaultValue string, validate func(string) string) string {
	for {
		prompt := fmt.Sprintf("%s [%s]: ", label, defaultValue)
		fmt.Fprint(os.Stderr, prompt)
		input, _ := lineReader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue
		}

		if validate != nil {
			if result := validate(input); result != "" {
				return result
			}
		} else {
			return input
		}

		fmt.Fprintf(os.Stderr, "  Invalid input. Please try again.\n")
	}
}

func promptRequired(label string) string {
	for {
		fmt.Fprintf(os.Stderr, "%s: ", label)
		input, _ := lineReader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
		fmt.Fprintf(os.Stderr, "  This value is required.\n")
	}
}

func generateJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}


