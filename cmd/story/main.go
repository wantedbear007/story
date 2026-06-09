package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	appauth "github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/application/collection"
	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/application/resource"
	"github.com/anomalyco/story/internal/application/tag"
	appuser "github.com/anomalyco/story/internal/application/user"
	infraauth "github.com/anomalyco/story/internal/infrastructure/auth"
	"github.com/anomalyco/story/internal/infrastructure/bootstrap"
	"github.com/anomalyco/story/internal/infrastructure/email"
	"github.com/anomalyco/story/internal/infrastructure/llm"
	"github.com/anomalyco/story/internal/infrastructure/repository"
	"github.com/anomalyco/story/internal/interfaces/api"
	"github.com/anomalyco/story/internal/interfaces/cli"
	"github.com/anomalyco/story/internal/pkg/logger"
)

func main() {
	args := os.Args[1:]

	isHelp := len(args) == 0 || args[0] == "help" || args[0] == "--help"
	isOnboarding := len(args) > 0 && (args[0] == "init" || args[0] == "verify" || args[0] == "setup")

	if isHelp {
		if hasConfigFile() {
			showFullHelp(args)
			return
		}
		showOnboardingHelp()
		return
	}

	if isOnboarding {
		runOnboarding(args)
		return
	}

	if !hasConfigFile() {
		fmt.Fprintln(os.Stderr, "Story is not configured.")
		fmt.Fprintln(os.Stderr, "Run 'story init' to create your configuration.")
		os.Exit(1)
	}

	if err := bootstrap.Run(context.Background(), bootstrapConfigPath(), start); err != nil {
		os.Exit(1)
	}
}

type schemaChecker interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func checkDatabaseSetup(ctx context.Context, db schemaChecker) error {
	var tableCount int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("unable to verify database setup: %w", err)
	}
	if tableCount == 0 {
		return fmt.Errorf("database not initialized: run 'story setup' to create the schema")
	}
	return nil
}

func bootstrapConfigPath() string {
	if p := os.Getenv("STORY_CONFIG_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(home + "/.story/config.yaml"); err == nil {
			return home + "/.story/config.yaml"
		}
	}
	return "configs/config.yaml"
}

func hasConfigFile() bool {
	if os.Getenv("STORY_DATABASE_PASSWORD") != "" && os.Getenv("STORY_AUTH_JWT_SECRET") != "" {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(home + "/.story/config.yaml")
	return err == nil
}

func showFullHelp(args []string) {
	root := newOnboardingRoot()
	root.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	})
	root.AddCommand(&cobra.Command{
		Use:   "entry",
		Short: "Manage learning entries",
	})
	root.AddCommand(&cobra.Command{
		Use:   "capture",
		Short: "Capture a new entry to your second brain",
	})
	root.AddCommand(&cobra.Command{
		Use:   "query",
		Short: "Search and list captured entries",
	})
	root.AddCommand(&cobra.Command{
		Use:   "search",
		Short: "Search entries by title, content, and tags",
	})
	root.AddCommand(&cobra.Command{
		Use:   "collection",
		Short: "Manage collections",
	})
	root.AddCommand(&cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	})
	root.AddCommand(&cobra.Command{
		Use:   "publish",
		Short: "Publish entries to external platforms",
	})
	root.AddCommand(&cobra.Command{
		Use:   "target",
		Short: "Manage publishing targets",
	})
	root.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "View and manage configuration",
	})
	root.AddCommand(&cobra.Command{
		Use:   "resource",
		Short: "Manage resources",
	})
	root.AddCommand(&cobra.Command{
		Use:   "tweet",
		Short: "Generate and manage tweets from entries",
	})
	root.AddCommand(&cobra.Command{
		Use:   "web",
		Short: "Start the web dashboard",
	})
	root.AddCommand(&cobra.Command{
		Use:   "register",
		Short: "Create a new account",
	})
	root.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Login to your account",
	})
	root.AddCommand(&cobra.Command{
		Use:   "password",
		Short: "Manage your password",
	})
	root.AddCommand(&cobra.Command{
		Use:   "forgot-password",
		Short: "Request a password reset email",
	})

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showOnboardingHelp() {
	if err := newOnboardingRoot().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newOnboardingRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "story",
		Short:         "A CLI-first second brain for developers",
		Long: `Story captures learning, work logs, resources, and engineering notes,
transforms them into structured knowledge, and publishes to your favorite platforms.

Story helps you build your personal knowledge graph from the command line.

GitHub: https://github.com/wantedbear007/story
Founder: https://github.com/wantedbear007`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(cli.NewInitCommand())
	root.AddCommand(cli.NewVerifyCommand())
	root.AddCommand(cli.NewSetupCommand())
	return root
}

func runOnboarding(args []string) {
	root := newOnboardingRoot()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
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

	llmProvider, llmErr := llm.NewProvider(app.Config.LLM)
	var llmAdapter *llm.CompleteAdapter
	if llmErr != nil {
		log.Warn("LLM provider not configured, tweet generation will be unavailable", logger.F("error", llmErr.Error()))
	} else {
		llmAdapter = llm.NewCompleteAdapter(llmProvider)
	}

	tweetRepo := repository.NewTweetRepository(app.DB)
	promptRepo := repository.NewPromptTemplateRepository(app.DB)
	tweetSvc := content.NewService(tweetRepo, promptRepo, entryRepo, llmAdapter)

	apiServer := api.NewServer(app.Config.Server.Host, app.Config.Server.Port, tweetSvc, entrySvc, jwtService)

	deps := &cli.Dependencies{
		Cfg:               app.Config,
		UserService:       userSvc,
		EntryService:      entrySvc,
		CollectionService: collectionSvc,
		TagService:        tagSvc,
		PublishingService: publishingSvc,
		AuthService:       authSvc,
		ResourceService:   resourceSvc,
		TweetService:      tweetSvc,
		ApiServer:         apiServer,
	}

	log.Info("application initialized")

	if err := checkDatabaseSetup(ctx, app.DB); err != nil {
		return err
	}

	cli.Execute(ctx, deps)

	return nil
}


