package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/anomalyco/story/internal/infrastructure/config"
	"github.com/anomalyco/story/internal/infrastructure/database"
	"github.com/anomalyco/story/internal/pkg/logger"
)

// Application holds all top-level dependencies.
// It is the composition root — every object the application needs
// is wired here, then passed down. This avoids global state and
// makes dependency graphs explicit.
type Application struct {
	Config *config.Config
	Logger logger.Logger
	DB     *pgxpool.Pool
}

// Run starts the application lifecycle:
// 1. Load configuration
// 2. Initialize logger
// 3. Connect to database
// 4. Run user-provided startup function
// 5. Wait for shutdown signal
func Run(ctx context.Context, cfgPath string, start func(context.Context, *Application) error) error {
	app, err := New(ctx, cfgPath)
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}
	defer app.Shutdown()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigCh
		app.Logger.Debug("received shutdown signal", logger.F("signal", sig.String())) //nolint:errcheck
		cancel()
	}()

	if err := start(ctx, app); err != nil {
		app.Logger.Error("application failed", logger.Err(err))
		return err
	}

	return nil
}

// New initializes all foundational dependencies.
// Returns an Application ready to start serving.
func New(ctx context.Context, cfgPath string) (*Application, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var log logger.Logger
	if cfg.App.Environment == "development" {
		log = logger.NewDev(logger.ParseLevel(cfg.App.LogLevel))
	} else {
		log = logger.New(logger.ParseLevel(cfg.App.LogLevel), os.Stderr)
	}

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &Application{
		Config: cfg,
		Logger: log,
		DB:     pool,
	}, nil
}

// Shutdown gracefully releases resources.
func (a *Application) Shutdown() {
	if a.DB != nil {
		a.DB.Close()
	}
	// shutdown complete
}
