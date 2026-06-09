package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anomalyco/story/internal/application/collection"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/application/tag"
	appauth "github.com/anomalyco/story/internal/application/auth"
	appuser "github.com/anomalyco/story/internal/application/user"
	"github.com/anomalyco/story/internal/infrastructure/auth"
	"github.com/anomalyco/story/internal/infrastructure/config"
	"github.com/anomalyco/story/internal/infrastructure/database"
	"github.com/anomalyco/story/internal/infrastructure/email"
	"github.com/anomalyco/story/internal/infrastructure/llm"
	"github.com/anomalyco/story/internal/infrastructure/repository"
	"github.com/anomalyco/story/internal/interfaces/cli"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := run(ctx); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

// run is the composition root — it wires all dependencies together.
// Dependency injection is done manually for clarity and explicitness.
// As the project grows, this can be migrated to a DI framework (e.g., Google Wire).
func run(ctx context.Context) error {
	cfgPath := os.Getenv("STORY_CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	passwordHasher := auth.NewPasswordHasher()

	jwtService, err := auth.NewJWTTokenService(cfg.Auth)
	if err != nil {
		return fmt.Errorf("initializing JWT service: %w", err)
	}

	llmProvider, err := llm.NewProvider(cfg.LLM)
	if err != nil {
		return fmt.Errorf("initializing LLM provider: %w", err)
	}

	mailer := email.NewSMTPMailer(cfg.SMTP)

	userRepo := repository.NewUserRepository(pool)
	entryRepo := repository.NewEntryRepository(pool)
	tagRepo := repository.NewTagRepository(pool)
	collectionRepo := repository.NewCollectionRepository(pool)
	publishingTargetRepo := repository.NewPublishingTargetRepository(pool)
	publishedEntryRepo := repository.NewPublishedEntryRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)

	userService := appuser.NewService(userRepo, sessionRepo, passwordHasher, jwtService)
	entryService := entry.NewService(entryRepo, tagRepo)
	collectionService := collection.NewService(collectionRepo)
	tagService := tag.NewService(tagRepo)

	llmAdapter := llm.NewCompleteAdapter(llmProvider)
	publisher := publishing.NewPublisher(llmAdapter)
	publishingService := publishing.NewService(publishingTargetRepo, publishedEntryRepo, entryRepo, publisher)

	authService := appauth.NewService(userRepo, sessionRepo, passwordHasher, jwtService, mailer)

	deps := &cli.Dependencies{
		Cfg:               cfg,
		UserService:       userService,
		EntryService:      entryService,
		CollectionService: collectionService,
		TagService:        tagService,
		PublishingService: publishingService,
		AuthService:       authService,
	}

	cli.Execute(ctx, deps)
	return nil
}
