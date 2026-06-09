package main

import (
	"context"
	"os"

	"github.com/anomalyco/story/internal/infrastructure/bootstrap"
	"github.com/anomalyco/story/internal/pkg/logger"
)

func main() {
	cfgPath := os.Getenv("STORY_CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	if err := bootstrap.Run(context.Background(), cfgPath, start); err != nil {
		// bootstrap.Run already logs the error; a non-zero exit communicates
		// the failure to the shell/container orchestrator.
		os.Exit(1)
	}
}

// start is the application entry point after bootstrap completes.
// It receives the fully initialized Application with config, logger,
// and database pool ready to use. Business logic wiring happens here.
func start(ctx context.Context, app *bootstrap.Application) error {
	log := app.Logger.With(logger.F("component", "main"))

	// -----------------------------------------------------------------------
	// Business-layer wiring
	// Future: as the application grows, separate wiring into dedicated
	// modules (e.g., cmd/story/wire.go) or adopt Google Wire for codegen.
	// -----------------------------------------------------------------------
	// The code below is placeholder for actual service registration.
	// Each domain's services, repositories, and handlers get wired here.
	// See existing business packages for reference:
	//   internal/application/...
	//   internal/infrastructure/auth/...
	//   internal/infrastructure/repository/...
	//   internal/interfaces/cli/...

	_ = app.DB           // database pool ready

	log.Info("application initialized")

	// Future: CLI.Execute() or server.ListenAndServe() goes here.
	// For now, the app runs until a shutdown signal is received.
	<-ctx.Done()
	log.Info("application shutting down")

	return nil
}
