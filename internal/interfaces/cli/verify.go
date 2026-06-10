package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/anomalyco/story/internal/infrastructure/config"
)

func NewVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify Story configuration and service connections",
		Long:  "Validate that Story can communicate with all configured services (database, SMTP, LLM).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify()
		},
	}
}

func runVerify() error {
	ctx := context.Background()
	allPassed := true

	cfgPath := resolveConfigPath()
	cfg, err := loadInitConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Config Found")
	fmt.Printf("  File: %s\n", cfgPath)
	fmt.Printf("  Environment: %s\n", cfg.App.Environment)

	fmt.Print("\n  Database: ")
	if err := verifyDatabase(ctx, cfg); err != nil {
		fmt.Printf("✗ %v\n", err)
		allPassed = false
	} else {
		fmt.Println("✓ Connection Successful")
	}

	if cfg.SMTP.Host != "" {
		fmt.Print("  SMTP: ")
		if err := verifySMTP(cfg); err != nil {
			fmt.Printf("✗ %v\n", err)
			allPassed = false
		} else {
			fmt.Println("✓ Connection Successful")
		}
	} else {
		fmt.Println("  SMTP: – (not configured)")
	}

	if hasLLMConfig(cfg) {
		fmt.Print("  LLM: ")
		if err := verifyLLM(ctx, cfg); err != nil {
			fmt.Printf("✗ %v\n", err)
			allPassed = false
		} else {
			fmt.Println("✓ API Reachable")
		}
	} else {
		fmt.Println("  LLM: – (not configured)")
	}

	fmt.Println()
	if allPassed {
		fmt.Println("All checks passed.")
		return nil
	}

	os.Exit(1)
	return nil
}

func resolveConfigPath() string {
	if p := os.Getenv("STORY_CONFIG_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil {
		p := home + "/.story/config.yaml"
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "configs/config.yaml"
}

func loadInitConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	cfg := &config.Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if v := os.Getenv("STORY_DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
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

	return cfg, nil
}

func verifyDatabase(ctx context.Context, cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s&connect_timeout=5",
		cfg.Database.User, cfg.Database.Password,
		net.JoinHostPort(cfg.Database.Host, fmt.Sprintf("%d", cfg.Database.Port)),
		cfg.Database.Name, cfg.Database.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

func verifySMTP(cfg *config.Config) error {
	addr := net.JoinHostPort(cfg.SMTP.Host, fmt.Sprintf("%d", cfg.SMTP.Port))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", addr, err)
	}
	conn.Close()
	return nil
}

func hasLLMConfig(cfg *config.Config) bool {
	if cfg.LLM.Provider == "" {
		return false
	}
	switch cfg.LLM.Provider {
	case "openai":
		return cfg.LLM.OpenAI.APIKey != ""
	case "gemini":
		return cfg.LLM.Gemini.APIKey != ""
	case "anthropic":
		return cfg.LLM.Anthropic.APIKey != ""
	case "ollama":
		return cfg.LLM.Ollama.BaseURL != ""
	}
	return false
}

func verifyLLM(ctx context.Context, cfg *config.Config) error {
	switch cfg.LLM.Provider {
	case "openai":
		if cfg.LLM.OpenAI.APIKey == "" {
			return fmt.Errorf("OpenAI API key not configured")
		}
	case "gemini":
		if cfg.LLM.Gemini.APIKey == "" {
			return fmt.Errorf("Gemini API key not configured")
		}
	case "anthropic":
		if cfg.LLM.Anthropic.APIKey == "" {
			return fmt.Errorf("Anthropic API key not configured")
		}
	case "ollama":
		addr := cfg.LLM.Ollama.BaseURL
		if addr == "" {
			return fmt.Errorf("Ollama base URL not configured")
		}
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			return fmt.Errorf("cannot reach %s: %w", addr, err)
		}
		conn.Close()
		return nil
	default:
		return fmt.Errorf("unknown provider: %s", cfg.LLM.Provider)
	}

	return nil
}


