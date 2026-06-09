package main

import (
	"context"
	"os"

	appauth "github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/application/collection"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/application/resource"
	"github.com/anomalyco/story/internal/application/tag"
	appuser "github.com/anomalyco/story/internal/application/user"
	infraauth "github.com/anomalyco/story/internal/infrastructure/auth"
	"github.com/anomalyco/story/internal/infrastructure/bootstrap"
	"github.com/anomalyco/story/internal/infrastructure/email"
	"github.com/anomalyco/story/internal/infrastructure/repository"
	"github.com/anomalyco/story/internal/interfaces/cli"
	"github.com/anomalyco/story/internal/pkg/logger"
)

func main() {
	cfgPath := os.Getenv("STORY_CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	if err := bootstrap.Run(context.Background(), cfgPath, start); err != nil {
		os.Exit(1)
	}
}

func start(ctx context.Context, app *bootstrap.Application) error {
	log := app.Logger.With(logger.F("component", "main"))

	passwordHasher := infraauth.NewPasswordHasher()

	jwtService, err := infraauth.NewJWTTokenService(app.Config.Auth)
	if err != nil {
		return err
	}

	mailer := email.NewSMTPMailer(app.Config.SMTP)

	userRepo := repository.NewUserRepository(app.DB)
	sessionRepo := repository.NewSessionRepository(app.DB)
	passwordResetRepo := repository.NewPasswordResetRepository(app.DB)
	emailVerificationRepo := repository.NewEmailVerificationRepository(app.DB)
	tagRepo := repository.NewTagRepository(app.DB)
	entryRepo := repository.NewEntryRepository(app.DB)
	collectionRepo := repository.NewCollectionRepository(app.DB)
	resourceRepo := repository.NewResourceRepository(app.DB)

	userSvc := appuser.NewService(
		userRepo,
		emailVerificationRepo,
		sessionRepo,
		passwordHasher,
		mailer,
		app.Config.Auth.EmailVerificationTTL,
	)

	authSvc := appauth.NewService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		jwtService,
		passwordHasher,
		mailer,
		app.Config.Auth.RefreshTokenTTL,
		app.Config.Auth.PasswordResetTTL,
	)

	tagSvc := tag.NewService(tagRepo)
	entrySvc := entry.NewService(entryRepo, tagRepo, resourceRepo)
	collectionSvc := collection.NewService(collectionRepo)
	publishingTargetRepo := repository.NewPublishingTargetRepository(app.DB)
	publishedEntryRepo := repository.NewPublishedEntryRepository(app.DB)
	publishingSvc := publishing.NewService(publishingTargetRepo, publishedEntryRepo, entryRepo, nil)
	resourceSvc := resource.NewService(resourceRepo)

	deps := &cli.Dependencies{
		Cfg:               app.Config,
		UserService:       userSvc,
		EntryService:      entrySvc,
		CollectionService: collectionSvc,
		TagService:        tagSvc,
		PublishingService: publishingSvc,
		AuthService:       authSvc,
		ResourceService:   resourceSvc,
	}

	log.Info("application initialized")

	cli.Execute(ctx, deps)

	return nil
}
