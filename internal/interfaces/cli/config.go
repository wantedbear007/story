package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and manage configuration",
		Long:  "Display current configuration settings and manage config paths.",
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

	return cmd
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
