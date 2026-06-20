package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration structure.
// Values are loaded from YAML files and overridden by environment variables.
// Environment variable overrides follow the pattern: STORY_<SECTION>_<KEY>
// e.g., STORY_DATABASE_HOST overrides config.database.host
type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	LLM      LLMConfig           `yaml:"llm"`
	SMTP     SMTPConfig          `yaml:"smtp"`
	Capture  CaptureConfig       `yaml:"capture"`
	Notify   NotificationConfig  `yaml:"notify"`
	Scheduler SchedulerConfig    `yaml:"scheduler"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"` // development, staging, production
	LogLevel    string `yaml:"log_level"`    // debug, info, warn, error
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Name         string `yaml:"name"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	SSLMode      string `yaml:"ssl_mode"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type AuthConfig struct {
	JWTSecret            string        `yaml:"jwt_secret"`
	AccessTokenTTL       time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `yaml:"refresh_token_ttl"`
	PasswordResetTTL     time.Duration `yaml:"password_reset_ttl"`
	EmailVerificationTTL time.Duration `yaml:"email_verification_ttl"`
}

type LLMConfig struct {
	Provider  string          `yaml:"provider"`
	OpenAI    OpenAILLMConfig `yaml:"openai"`
	Gemini    GeminiLLMConfig `yaml:"gemini"`
	Ollama    OllamaLLMConfig `yaml:"ollama"`
	Anthropic AnthropicLLMConfig `yaml:"anthropic"`
}

type OpenAILLMConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type GeminiLLMConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type OllamaLLMConfig struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type AnthropicLLMConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// DSN builds the SMTP connection string.
func (s SMTPConfig) DSN() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type CaptureConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type NotificationConfig struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Message string `yaml:"message"`
}

type SchedulerConfig struct {
	Enabled bool `yaml:"enabled"`
	Hour    int  `yaml:"hour"`
	Minute  int  `yaml:"minute"`
}

// Load reads configuration from a YAML file and applies environment variable overrides.
// Environment variables take precedence over file values.
// The secrets pattern (STORY_*) is used for sensitive values.
func Load(path string) (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        "story",
			Environment: "development",
			LogLevel:    "info",
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			Name:         "story",
			User:         "story",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		Auth: AuthConfig{
			AccessTokenTTL:       15 * time.Minute,
			RefreshTokenTTL:      7 * 24 * time.Hour,
			PasswordResetTTL:     1 * time.Hour,
			EmailVerificationTTL: 24 * time.Hour,
		},
		LLM: LLMConfig{
			Provider: "openai",
			OpenAI: OpenAILLMConfig{
				Model: "gpt-4",
			},
			Gemini: GeminiLLMConfig{
				Model: "gemini-pro",
			},
			Ollama: OllamaLLMConfig{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
			},
			Anthropic: AnthropicLLMConfig{
				Model: "claude-3-opus-20240229",
			},
		},
		SMTP: SMTPConfig{
			Host: "localhost",
			Port: 1025,
		},
		Capture: CaptureConfig{
			Host: "127.0.0.1",
			Port: 8081,
		},
		Notify: NotificationConfig{
			Enabled: true,
			Title:   "Story",
			Message: "Time to capture what you learned today",
		},
		Scheduler: SchedulerConfig{
			Enabled: true,
			Hour:    19,
			Minute:  0,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides reads STORY_* environment variables and overrides config values.
// Secret values like passwords and API keys should only be set via environment.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("STORY_DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("STORY_AUTH_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("STORY_LLM_OPENAI_API_KEY"); v != "" {
		cfg.LLM.OpenAI.APIKey = v
	}
	if v := os.Getenv("STORY_LLM_GEMINI_API_KEY"); v != "" {
		cfg.LLM.Gemini.APIKey = v
	}
	if v := os.Getenv("STORY_LLM_OLLAMA_BASE_URL"); v != "" {
		cfg.LLM.Ollama.BaseURL = v
	}
	if v := os.Getenv("STORY_LLM_ANTHROPIC_API_KEY"); v != "" {
		cfg.LLM.Anthropic.APIKey = v
	}
	if v := os.Getenv("STORY_SMTP_PASSWORD"); v != "" {
		cfg.SMTP.Password = v
	}
	if v := os.Getenv("STORY_APP_ENVIRONMENT"); v != "" {
		cfg.App.Environment = v
	}
	if v := os.Getenv("STORY_DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("STORY_DATABASE_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Database.Port)
	}
	if v := os.Getenv("STORY_DATABASE_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("STORY_DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("STORY_DATABASE_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := os.Getenv("STORY_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("STORY_SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("STORY_CAPTURE_HOST"); v != "" {
		cfg.Capture.Host = v
	}
	if v := os.Getenv("STORY_CAPTURE_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Capture.Port)
	}
	if v := os.Getenv("STORY_NOTIFY_ENABLED"); v != "" {
		cfg.Notify.Enabled = v == "true"
	}
	if v := os.Getenv("STORY_SCHEDULER_ENABLED"); v != "" {
		cfg.Scheduler.Enabled = v == "true"
	}
}

// Validate checks the configuration for required values.
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required (set STORY_DATABASE_PASSWORD)")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required (set STORY_AUTH_JWT_SECRET)")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}
	switch c.App.Environment {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("invalid environment: %q", c.App.Environment)
	}
	return nil
}
