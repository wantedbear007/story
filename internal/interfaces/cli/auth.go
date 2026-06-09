package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/user"
)

func newAuthCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Register, login, and manage your Story account.",
	}

	cmd.AddCommand(newRegisterCommand(deps))
	cmd.AddCommand(newLoginCommand(deps))
	cmd.AddCommand(newProfileCommand(deps))

	return cmd
}

func newRegisterCommand(deps *Dependencies) *cobra.Command {
	var email, password, displayName string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := deps.UserService.Register(cmd.Context(), user.RegisterRequest{
				Email:       email,
				Password:    password,
				DisplayName: displayName,
			})
			if err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}

			fmt.Printf("Registered as %s (%s)\n", resp.User.DisplayName, resp.User.Email)
			fmt.Printf("Access Token: %s\n", resp.AccessToken)
			fmt.Printf("Refresh Token: %s\n", resp.RefreshToken)
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (min 8 characters)")
	cmd.Flags().StringVarP(&displayName, "display-name", "n", "", "Display name (required)")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("display-name")

	return cmd
}

func newLoginCommand(deps *Dependencies) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := deps.UserService.Login(cmd.Context(), user.LoginRequest{
				Email:    email,
				Password: password,
			})
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			fmt.Printf("Logged in as %s (%s)\n", resp.User.DisplayName, resp.User.Email)
			fmt.Printf("Access Token: %s\n", resp.AccessToken)
			fmt.Printf("Refresh Token: %s\n", resp.RefreshToken)
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newProfileCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "profile",
		Short: "Show your profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			// In a real CLI, this would extract the user ID from the stored token.
			// For now, it requires the user to pass their user ID or have a session.
			fmt.Println("Profile command requires authenticated session")
			fmt.Println("Use 'story auth login' first")
			return nil
		},
	}
}
