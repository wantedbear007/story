package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/application/collection"
	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/application/daemon"
	"github.com/anomalyco/story/internal/application/entry"
	appnotif "github.com/anomalyco/story/internal/application/notification"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/application/raw_entry"
	"github.com/anomalyco/story/internal/application/resource"
	"github.com/anomalyco/story/internal/application/tag"
	"github.com/anomalyco/story/internal/application/user"
	"github.com/anomalyco/story/internal/infrastructure/config"
	"github.com/anomalyco/story/internal/interfaces/api"
)

type Dependencies struct {
	Cfg               *config.Config
	UserService       *user.Service
	EntryService      *entry.Service
	CollectionService *collection.Service
	TagService        *tag.Service
	PublishingService *publishing.Service
	AuthService       *auth.Service
	ResourceService   *resource.Service
	TweetService      *content.Service
	RawEntryService   *raw_entry.Service
	ApiServer         *api.Server
	NotifService      *appnotif.Service
	DaemonService     *daemon.Service
}

func NewRootCommand(deps *Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:   "story",
		Short: "A CLI-first second brain for developers",
		Long: `Story captures learning, work logs, resources, and engineering notes,
transforms them into structured knowledge, and publishes to your favorite platforms.

Story helps you build your personal knowledge graph from the command line.

GitHub: https://github.com/wantedbear007/story
Founder: https://github.com/wantedbear007`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if deps != nil && deps.Cfg != nil {
				if err := deps.Cfg.Validate(); err != nil {
					return fmt.Errorf("invalid configuration: %w", err)
				}
			}
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(newAuthCommand(deps))
	root.AddCommand(newEntryCommand(deps))
	root.AddCommand(newCaptureCommand(deps))
	root.AddCommand(newQueryCommand(deps))
	root.AddCommand(newCollectionCommand(deps))
	root.AddCommand(newTagCommand(deps))
	root.AddCommand(newPublishCommand(deps))
	root.AddCommand(newTargetCommand(deps))
	root.AddCommand(newConfigCommand(deps))
	root.AddCommand(newLLMConfigCommand())
	root.AddCommand(newLLMCheckCommand())
	root.AddCommand(newResourceCommand(deps))
	root.AddCommand(newSearchCommand(deps))
	root.AddCommand(newTweetCommand(deps))
	root.AddCommand(newWebCommand(deps))

	topLevelRegister(deps, root)
	topLevelLogin(deps, root)
	topLevelLogout(deps, root)
	topLevelPassword(deps, root)
	topLevelForgotPassword(deps, root)
	topLevelWhoami(deps, root)

	root.AddCommand(newRawCommand(deps))
	root.AddCommand(newProcessCommand(deps))
	root.AddCommand(newResetCommand(deps))

	root.AddCommand(newStartCommand(deps))
	root.AddCommand(newStopCommand(deps))
	root.AddCommand(newRestartCommand(deps))
	root.AddCommand(newStartStatusCommand(deps))
	root.AddCommand(newTestNotiCommand(deps))

	return root
}

func topLevelRegister(deps *Dependencies, root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "register",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegister(deps, cmd)
		},
	})
}

func topLevelLogin(deps *Dependencies, root *cobra.Command) {
	var sessions bool
	var revoke string
	var revokeAll bool

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sessions {
				return runListSessions(deps, cmd)
			}
			if revokeAll {
				return runRevokeAllSessions(deps, cmd)
			}
			if revoke != "" {
				return runRevokeSession(deps, cmd, revoke)
			}
			return runLogin(deps, cmd)
		},
	}

	loginCmd.Flags().BoolVar(&sessions, "sessions", false, "List all active sessions")
	loginCmd.Flags().StringVar(&revoke, "revoke", "", "Revoke a specific session by ID")
	loginCmd.Flags().BoolVar(&revokeAll, "revoke-all", false, "Revoke all sessions except current")

	root.AddCommand(loginCmd)
}

func topLevelLogout(deps *Dependencies, root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Logout and revoke the current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID, err := resolveCurrentSessionID(deps)
			if err != nil {
				// No local session to revoke — still try to clean up
				if err := clearSession(); err != nil {
					return fmt.Errorf("clearing session: %w", err)
				}
				fmt.Println("Not logged in")
				return nil
			}
			_ = deps.AuthService.Logout(cmd.Context(), sessionID)
			if err := clearSession(); err != nil {
				return fmt.Errorf("clearing session: %w", err)
			}
			fmt.Println("Logged out")
			return nil
		},
	})
}

func topLevelPassword(deps *Dependencies, root *cobra.Command) {
	passwordCmd := &cobra.Command{
		Use:   "password",
		Short: "Manage your password",
	}
	passwordCmd.AddCommand(&cobra.Command{
		Use:   "change",
		Short: "Change your password",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangePassword(deps, cmd)
		},
	})
	root.AddCommand(passwordCmd)
}

func topLevelForgotPassword(deps *Dependencies, root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "forgot-password",
		Short: "Request a password reset email",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runForgotPassword(deps, cmd)
		},
	})
}

func runRegister(deps *Dependencies, cmd *cobra.Command) error {
	email := promptInput("Email: ")
	password := promptPassword("Password: ")
	confirm := promptPassword("Retype password: ")

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	displayName := promptInput("Display name: ")

	_, err := deps.UserService.Register(cmd.Context(), &user.RegisterRequest{
		Email:       email,
		Password:    password,
		DisplayName: displayName,
	})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	loginResp, err := deps.AuthService.Login(cmd.Context(), &auth.LoginRequest{
		Email:      email,
		Password:   password,
		DeviceInfo: deviceInfo(),
		IPAddress:  "127.0.0.1",
	})
	if err != nil {
		return fmt.Errorf("registration succeeded but login failed: %w", err)
	}

	if err := saveLoginSession(loginResp); err != nil {
		return fmt.Errorf("registration succeeded but saving session failed: %w", err)
	}

	fmt.Printf("Registered and logged in as %s (%s)\n", loginResp.User.DisplayName, loginResp.User.Email)
	return nil
}

func runLogin(deps *Dependencies, cmd *cobra.Command) error {
	email := promptInput("Email: ")
	password := promptPassword("Password: ")

	resp, err := deps.AuthService.Login(cmd.Context(), &auth.LoginRequest{
		Email:      email,
		Password:   password,
		DeviceInfo: deviceInfo(),
		IPAddress:  "127.0.0.1",
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := saveLoginSession(resp); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	fmt.Printf("Logged in as %s (%s)\n", resp.User.DisplayName, resp.User.Email)
	return nil
}

func runListSessions(deps *Dependencies, cmd *cobra.Command) error {
	userID, err := resolveCurrentUserID(deps)
	if err != nil {
		return err
	}
	sessionID, err := resolveCurrentSessionID(deps)
	if err != nil {
		return err
	}

	sessions, err := deps.AuthService.ListSessions(cmd.Context(), userID, sessionID)
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No active sessions")
		return nil
	}

	fmt.Println("Current Sessions")
	fmt.Println()
	for i, s := range sessions {
		current := ""
		if s.IsCurrent {
			current = " (current)"
		}
		fmt.Printf("%d. %s%s\n", i+1, s.DeviceInfo, current)
		if s.LastUsedAt != nil {
			fmt.Printf("   Last Active: %s\n", formatDuration(time.Since(*s.LastUsedAt)))
		}
		if s.IPAddress != "" {
			fmt.Printf("   IP: %s\n", s.IPAddress)
		}
	}
	return nil
}

func runRevokeSession(deps *Dependencies, cmd *cobra.Command, sessionIDStr string) error {
	userID, err := resolveCurrentUserID(deps)
	if err != nil {
		return err
	}

	sessionID, err := uuidParse(sessionIDStr)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	if err := deps.AuthService.RevokeSession(cmd.Context(), userID, sessionID); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}

	fmt.Printf("Session %s revoked\n", sessionID)
	return nil
}

func runRevokeAllSessions(deps *Dependencies, cmd *cobra.Command) error {
	userID, err := resolveCurrentUserID(deps)
	if err != nil {
		return err
	}
	sessionID, err := resolveCurrentSessionID(deps)
	if err != nil {
		return err
	}

	if err := deps.AuthService.RevokeAllSessions(cmd.Context(), userID, sessionID); err != nil {
		return fmt.Errorf("revoking all sessions: %w", err)
	}

	fmt.Println("All other sessions revoked")
	return nil
}

func runChangePassword(deps *Dependencies, cmd *cobra.Command) error {
	userID, err := resolveCurrentUserID(deps)
	if err != nil {
		return err
	}

	currentPassword := promptPassword("Current password: ")
	newPassword := promptPassword("New password: ")

	if err := deps.UserService.ChangePassword(cmd.Context(), userID, &user.ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	}); err != nil {
		return fmt.Errorf("changing password: %w", err)
	}

	fmt.Println("Password changed successfully")
	return nil
}

func topLevelWhoami(deps *Dependencies, root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "whoami",
		Short: "Show current logged-in user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami()
		},
	})
}

func runWhoami() error {
	s, err := loadSession()
	if err != nil {
		fmt.Println("Not logged in")
		return nil
	}

	fmt.Printf("Logged in as %s (%s)\n", s.DisplayName, s.Email)
	fmt.Printf("User ID: %s\n", s.UserID)
	return nil
}

func runForgotPassword(deps *Dependencies, cmd *cobra.Command) error {
	email := promptInput("Email: ")

	if err := deps.AuthService.RequestPasswordReset(cmd.Context(), &auth.ForgotPasswordRequest{
		Email: email,
	}); err != nil {
		return fmt.Errorf("requesting password reset: %w", err)
	}

	fmt.Println("If the email is registered, a reset token has been sent")
	return nil
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

func Execute(ctx context.Context, deps *Dependencies) {
	rootCmd := NewRootCommand(deps)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
