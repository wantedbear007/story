package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and manage configuration",
		Long:  "Display current configuration settings and manage config paths.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration (secrets masked)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := deps.Cfg

			fmt.Println("Application:")
			fmt.Printf("  Name:        %s\n", cfg.App.Name)
			fmt.Printf("  Environment: %s\n", cfg.App.Environment)
			fmt.Printf("  Log Level:   %s\n", cfg.App.LogLevel)
			fmt.Println()
			fmt.Println("Server:")
			fmt.Printf("  Host: %s\n", cfg.Server.Host)
			fmt.Printf("  Port: %d\n", cfg.Server.Port)
			fmt.Println()
			fmt.Println("Database:")
			fmt.Printf("  Host:     %s\n", cfg.Database.Host)
			fmt.Printf("  Port:     %d\n", cfg.Database.Port)
			fmt.Printf("  Name:     %s\n", cfg.Database.Name)
			fmt.Printf("  User:     %s\n", cfg.Database.User)
			fmt.Printf("  SSL Mode: %s\n", cfg.Database.SSLMode)
			fmt.Println()
			fmt.Println("Auth:")
			fmt.Printf("  JWT Secret:        %s\n", maskString(cfg.Auth.JWTSecret))
			fmt.Printf("  Access Token TTL:  %s\n", cfg.Auth.AccessTokenTTL)
			fmt.Printf("  Refresh Token TTL: %s\n", cfg.Auth.RefreshTokenTTL)
			fmt.Println()
			fmt.Println("LLM:")
			fmt.Printf("  Provider: %s\n", cfg.LLM.Provider)
			fmt.Println()
			fmt.Println("SMTP:")
			fmt.Printf("  Host: %s\n", cfg.SMTP.Host)
			fmt.Printf("  Port: %d\n", cfg.SMTP.Port)
			fmt.Printf("  From: %s\n", cfg.SMTP.From)

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deps.Cfg.Validate(); err != nil {
				return fmt.Errorf("configuration invalid: %w", err)
			}
			fmt.Println("Configuration is valid.")
			return nil
		},
	})

	cmd.AddCommand(newConfigSMTPCommand())

	return cmd
}

func newConfigSMTPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "smtp",
		Short: "Configure SMTP settings",
		Long:  "Interactively configure SMTP email settings and save them to your config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSMTP()
		},
	}
}

func configSMTPPath() string {
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

func runConfigSMTP() error {
	configPath := configSMTPPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config file %s: %w", configPath, err)
	}

	cfg := &initConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	fmt.Fprintln(os.Stderr, "── SMTP Configuration ──")

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

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(configPath, out, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ SMTP configuration saved to %s\n", configPath)
	return nil
}

func maskString(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
