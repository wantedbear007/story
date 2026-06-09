package cli

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/infrastructure/config"
	"github.com/anomalyco/story/internal/infrastructure/setup"
)

func NewSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Prepare the database for Story",
		Long:  "Run migrations, create required extensions, and seed system data. Safe to run multiple times.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
}

func runSetup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfgPath := resolveConfigPath()
	cfg, err := loadInitConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.Database.Password == "" {
		return fmt.Errorf("database password not configured in %s", cfgPath)
	}

	pool, err := connectForSetup(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	fmt.Println("Running migrations...")

	if err := setup.RunMigrations(ctx, pool); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	fmt.Println()
	fmt.Println("Setup completed successfully.")

	return nil
}

func connectForSetup(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s&connect_timeout=10",
		cfg.Database.User, cfg.Database.Password,
		net.JoinHostPort(cfg.Database.Host, fmt.Sprintf("%d", cfg.Database.Port)),
		cfg.Database.Name, cfg.Database.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return pool, nil
}


