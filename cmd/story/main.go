package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	appauth "github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/application/collection"
	"github.com/anomalyco/story/internal/application/content"
	appdaemon "github.com/anomalyco/story/internal/application/daemon"
	"github.com/anomalyco/story/internal/application/entry"
	appnotif "github.com/anomalyco/story/internal/application/notification"
	"github.com/anomalyco/story/internal/application/processor"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/application/raw_entry"
	"github.com/anomalyco/story/internal/application/resource"
	appsched "github.com/anomalyco/story/internal/application/scheduler"
	"github.com/anomalyco/story/internal/application/tag"
	appuser "github.com/anomalyco/story/internal/application/user"
	infraauth "github.com/anomalyco/story/internal/infrastructure/auth"
	"github.com/anomalyco/story/internal/infrastructure/bootstrap"
	"github.com/anomalyco/story/internal/infrastructure/config"
	infradaemon "github.com/anomalyco/story/internal/infrastructure/daemon"
	"github.com/anomalyco/story/internal/infrastructure/email"
	"github.com/anomalyco/story/internal/infrastructure/llm"
	infranotif "github.com/anomalyco/story/internal/infrastructure/notification"
	"github.com/anomalyco/story/internal/infrastructure/repository"
	"github.com/anomalyco/story/internal/interfaces/api"
	"github.com/anomalyco/story/internal/interfaces/cli"
	"github.com/anomalyco/story/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

func main() {
	args := os.Args[1:]

	isHelp := len(args) == 0 || args[0] == "help" || args[0] == "--help"
	isOnboarding := len(args) > 0 && (args[0] == "init" || args[0] == "verify" || args[0] == "setup")
	isOffline := len(args) > 0 && (args[0] == "logout" || args[0] == "reset")
	isDaemonCmd := len(args) > 0 && (args[0] == "start" || args[0] == "stop" || args[0] == "restart" || args[0] == "status" || args[0] == "test-noti")
	isDaemonRun := len(args) > 0 && args[0] == "_daemon"

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

	if isOffline {
		runOffline(args)
		return
	}

	if !hasConfigFile() {
		fmt.Fprintln(os.Stderr, "Story is not configured.")
		fmt.Fprintln(os.Stderr, "Run 'story init' to create your configuration.")
		os.Exit(1)
	}

	if isDaemonRun {
		if err := bootstrap.Run(context.Background(), bootstrapConfigPath(), runDaemon); err != nil {
			os.Exit(1)
		}
		return
	}

	if isDaemonCmd {
		runDaemonCommand(context.Background(), args)
		return
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
		Use:   "raw",
		Short: "Capture raw notes without structure",
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
		Use:   "llm-config",
		Short: "Configure LLM provider settings",
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
		Use:   "logout",
		Short: "Logout and revoke the current session",
	})
	root.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset all local config and session data",
	})
	root.AddCommand(&cobra.Command{
		Use:   "password",
		Short: "Manage your password",
	})
	root.AddCommand(&cobra.Command{
		Use:   "forgot-password",
		Short: "Request a password reset email",
	})
	root.AddCommand(&cobra.Command{
		Use:   "whoami",
		Short: "Show current logged-in user",
	})
	root.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the background daemon",
	})
	root.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop the background daemon",
	})
	root.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart the background daemon",
	})
	root.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
	})
	root.AddCommand(&cobra.Command{
		Use:   "test-noti",
		Short: "Send a test desktop notification",
	})

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showOnboardingHelp() {
	root := newOnboardingRoot()
	root.AddCommand(&cobra.Command{Use: "logout", Short: "Logout and revoke the current session"})
	root.AddCommand(&cobra.Command{Use: "reset", Short: "Reset all local config and session data"})
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newOnboardingRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "story",
		Short: "A CLI-first second brain for developers",
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

func runOffline(args []string) {
	root := newOnboardingRoot()
	root.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Logout and revoke the current session",
		RunE: func(_ *cobra.Command, _ []string) error {
			cleared := clearLocalSession()
			if cleared {
				fmt.Println("Logged out")
			}
			return nil
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset all local config and session data",
		Long: `Remove all local configuration and session files from ~/.story/.

This deletes:
  - ~/.story/config.yaml   (configuration)
  - ~/.story/session.json   (login session)

Your database and remote data are NOT affected by this command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResetOffline()
		},
	})
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func clearLocalSession() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return false
	}
	sessionPath := home + "/.story/session.json"
	if err := os.Remove(sessionPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Not logged in")
			return false
		}
		fmt.Fprintf(os.Stderr, "Error removing session file: %v\n", err)
		return false
	}
	return true
}

func runResetOffline() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}
	storyDir := home + "/.story"

	if _, err := os.Stat(storyDir); os.IsNotExist(err) {
		fmt.Println("Nothing to reset — ~/.story/ does not exist")
		return nil
	}

	fmt.Print("This will delete all config and session data. Type 'yes' to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "yes" {
		fmt.Println("Reset cancelled")
		return nil
	}

	removed := false

	configPath := home + "/.story/config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Printf("Removed %s\n", configPath)
		removed = true
	}

	sessionPath := home + "/.story/session.json"
	if _, err := os.Stat(sessionPath); err == nil {
		if err := os.Remove(sessionPath); err != nil {
			return fmt.Errorf("removing session: %w", err)
		}
		fmt.Printf("Removed %s\n", sessionPath)
		removed = true
	}

	if removed {
		if entries, _ := os.ReadDir(storyDir); len(entries) == 0 {
			os.Remove(storyDir)
		}
		fmt.Println("Reset complete")
	} else {
		fmt.Println("Nothing to reset")
	}

	return nil
}

func runDaemonCommand(ctx context.Context, args []string) {
	cfg, err := config.Load(bootstrapConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	notifProvider, err := infranotif.NewProvider(cfg.Notify)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	var notifSvc *appnotif.Service
	if notifProvider != nil {
		notifSvc = appnotif.NewService(notifProvider)
	}

	captureURL := fmt.Sprintf("http://%s:%d/capture.html", cfg.Capture.Host, cfg.Capture.Port)

	schedCfg := appsched.Config{
		Enabled:    cfg.Scheduler.Enabled,
		Hour:       cfg.Scheduler.Hour,
		Minute:     cfg.Scheduler.Minute,
		CaptureURL: captureURL,
	}
	sched := appsched.NewService(notifSvc, schedCfg)

	daemonStore, err := infradaemon.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating daemon store: %v\n", err)
		os.Exit(1)
	}

	daemonSvc := appdaemon.NewService(daemonStore, notifSvc, sched, cfg.Capture)

	switch args[0] {
	case "start":
		err = daemonSvc.Start(ctx)
	case "stop":
		err = daemonSvc.Stop(ctx)
	case "restart":
		if e := daemonSvc.Stop(ctx); e != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
		}
		err = daemonSvc.Start(ctx)
	case "status":
		err = cli.RunDaemonStatus(ctx, daemonSvc)
	case "test-noti":
		if notifSvc == nil {
			err = fmt.Errorf("notification not configured: enable notifications in config or set STORY_NOTIFY_ENABLED=true")
		} else {
			err = cli.RunTestNoti(ctx, notifSvc, captureURL)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDaemon(ctx context.Context, app *bootstrap.Application) error {
	log := app.Logger.With(logger.F("component", "daemon"))
	log.Debug("daemon starting")

	notifProvider, err := infranotif.NewProvider(app.Config.Notify)
	if err != nil {
		log.Debug("notification not available")
	}
	var notifSvc *appnotif.Service
	if notifProvider != nil {
		notifSvc = appnotif.NewService(notifProvider)
	}

	captureURL := fmt.Sprintf("http://%s:%d/capture.html", app.Config.Capture.Host, app.Config.Capture.Port)

	schedCfg := appsched.Config{
		Enabled:    app.Config.Scheduler.Enabled,
		Hour:       app.Config.Scheduler.Hour,
		Minute:     app.Config.Scheduler.Minute,
		CaptureURL: captureURL,
	}
	sched := appsched.NewService(notifSvc, schedCfg)

	rawEntryRepo := repository.NewRawEntryRepository(app.DB)
	rawEntrySvc := raw_entry.NewService(rawEntryRepo)

	captureSvr := api.NewCaptureServer(app.Config.Capture.Host, app.Config.Capture.Port, rawEntrySvc)

	go func() {
		if err := captureSvr.Start(ctx); err != nil {
			log.Error("capture server error", logger.F("error", err.Error()))
		}
	}()

	tagRepo := repository.NewTagRepository(app.DB)
	resourceRepo := repository.NewResourceRepository(app.DB)
	entryRepo := repository.NewEntryRepository(app.DB)
	entrySvc := entry.NewService(entryRepo, tagRepo, resourceRepo)

	llmProvider, llmErr := llm.NewProvider(app.Config.LLM)
	var llmAdapter *llm.CompleteAdapter
	if llmErr == nil {
		llmAdapter = llm.NewCompleteAdapter(llmProvider)
	}
	tweetRepo := repository.NewTweetRepository(app.DB)
	promptRepo := repository.NewPromptTemplateRepository(app.DB)
	tweetSvc := content.NewService(tweetRepo, promptRepo, entryRepo, llmAdapter)

	proc := processor.NewService(rawEntrySvc, entrySvc, tweetSvc)
	go proc.Start(ctx)

	sched.Start(ctx)

	log.Debug("daemon ready")

	<-ctx.Done()
	sched.Stop()
	log.Debug("daemon stopped")
	return nil
}

func start(ctx context.Context, app *bootstrap.Application) error {
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
	rawEntryRepo := repository.NewRawEntryRepository(app.DB)

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
	rawEntrySvc := raw_entry.NewService(rawEntryRepo)

	llmProvider, llmErr := llm.NewProvider(app.Config.LLM)
	var llmAdapter *llm.CompleteAdapter
	if llmErr != nil {
		// LLM not configured — tweet generation unavailable
	} else {
		llmAdapter = llm.NewCompleteAdapter(llmProvider)
	}

	tweetRepo := repository.NewTweetRepository(app.DB)
	promptRepo := repository.NewPromptTemplateRepository(app.DB)
	tweetSvc := content.NewService(tweetRepo, promptRepo, entryRepo, llmAdapter)

	apiServer := api.NewServer(app.Config.Server.Host, app.Config.Server.Port, tweetSvc, entrySvc, rawEntrySvc, jwtService)

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
		RawEntryService:   rawEntrySvc,
		ApiServer:         apiServer,
	}

	if err := checkDatabaseSetup(ctx, app.DB); err != nil {
		return err
	}

	cli.Execute(ctx, deps)

	return nil
}
