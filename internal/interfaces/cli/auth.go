package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/application/user"
)

func newAuthCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Register, login, and manage your Story account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newRegisterCommand(deps))
	cmd.AddCommand(newLoginCommand(deps))
	cmd.AddCommand(newLogoutCommand(deps))
	cmd.AddCommand(newStatusCommand(deps))
	cmd.AddCommand(newSessionsCommand(deps))
	cmd.AddCommand(newRevokeCommand(deps))
	cmd.AddCommand(newVerifyEmailCommand(deps))
	cmd.AddCommand(newPasswordCommand(deps))

	return cmd
}

func newRegisterCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}

	return cmd
}

func newLoginCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your account",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}

	return cmd
}

func newLogoutCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and revoke the current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID, err := resolveCurrentSessionID(deps)
			if err != nil {
				return err
			}

			if err := deps.AuthService.Logout(cmd.Context(), sessionID); err != nil {
				return fmt.Errorf("logout failed: %w", err)
			}

			if err := clearSession(); err != nil {
				return fmt.Errorf("clearing session: %w", err)
			}

			fmt.Println("Logged out")
			return nil
		},
	}
}

func newStatusCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current login status",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := loadSession()
			if err != nil {
				fmt.Println("Not logged in")
				return nil
			}

			fmt.Printf("Logged in as %s (%s)\n", s.DisplayName, s.Email)
			fmt.Printf("Session ID: %s\n", s.SessionID)
			return nil
		},
	}
}

func newSessionsCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "sessions",
		Short: "List all active sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			for _, s := range sessions {
				current := ""
				if s.IsCurrent {
					current = " (current)"
				}
				fmt.Printf("  %s%s\n", s.ID, current)
				fmt.Printf("    Device: %s\n", s.DeviceInfo)
				fmt.Printf("    IP: %s\n", s.IPAddress)
				fmt.Printf("    Created: %s\n", s.CreatedAt.Format("2006-01-02 15:04:05"))
				if s.LastUsedAt != nil {
					fmt.Printf("    Last used: %s\n", s.LastUsedAt.Format("2006-01-02 15:04:05"))
				}
				fmt.Printf("    Expires: %s\n", s.ExpiresAt.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}
}

func newRevokeCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <session-id>",
		Short: "Revoke a specific session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			sessionID, err := uuidParse(args[0])
			if err != nil {
				return fmt.Errorf("invalid session ID: %w", err)
			}

			if err := deps.AuthService.RevokeSession(cmd.Context(), userID, sessionID); err != nil {
				return fmt.Errorf("revoking session: %w", err)
			}

			fmt.Printf("Session %s revoked\n", sessionID)
			return nil
		},
	}
}

func newVerifyEmailCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "verify <token>",
		Short: "Verify your email address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deps.UserService.VerifyEmail(cmd.Context(), &user.VerifyEmailRequest{
				Token: args[0],
			}); err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			fmt.Println("Email verified successfully")
			return nil
		},
	}
}

func newPasswordCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "Manage your password",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newChangePasswordCommand(deps))
	cmd.AddCommand(newForgotPasswordCommand(deps))
	cmd.AddCommand(newResetPasswordCommand(deps))

	return cmd
}

func newChangePasswordCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "change",
		Short: "Change your password (requires current password)",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
}

func newForgotPasswordCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "forgot",
		Short: "Request a password reset email",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := promptInput("Email: ")

			if err := deps.AuthService.RequestPasswordReset(cmd.Context(), &auth.ForgotPasswordRequest{
				Email: email,
			}); err != nil {
				return fmt.Errorf("requesting password reset: %w", err)
			}

			fmt.Println("If the email is registered, a reset token has been sent")
			return nil
		},
	}
}

func newResetPasswordCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "reset <token>",
		Short: "Reset your password using a reset token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			newPassword := promptPassword("New password: ")

			if err := deps.AuthService.ResetPassword(cmd.Context(), &auth.ResetPasswordRequest{
				Token:       args[0],
				NewPassword: newPassword,
			}); err != nil {
				return fmt.Errorf("resetting password: %w", err)
			}

			fmt.Println("Password reset successfully")
			fmt.Println("Please login again with your new password")
			return nil
		},
	}
}

// uuidParse wraps uuid.Parse with a cleaner error message.
func uuidParse(s string) (uuid.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %s", s)
	}
	return u, nil
}

// deviceInfo returns a string identifying the current device.
func deviceInfo() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("cli:%s", hostname)
}

// promptPassword reads a password from stdin (no echo).
func promptPassword(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	s, _ := lineReader.ReadString('\n')
	return strings.TrimSpace(s)
}

// promptInput reads a line from stdin.
func promptInput(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	s, _ := lineReader.ReadString('\n')
	return strings.TrimSpace(s)
}

// lineReader is a shared buffered reader for stdin.
var lineReader = bufio.NewReader(os.Stdin)
