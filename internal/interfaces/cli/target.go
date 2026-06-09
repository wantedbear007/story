package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/domain"
)

func newTargetCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target",
		Short: "Manage publishing targets",
		Long: `Configure publishing destinations for your entries.

Supported targets: twitter, notion, google_doc, blog, markdown`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newTargetCreateCommand(deps))
	cmd.AddCommand(newTargetListCommand(deps))

	return cmd
}

func newTargetCreateCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Configure a new publishing target",
		Long: `Create a new publishing target interactively.

Example:
  story target create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetType := promptTargetType()
			name := promptRequired("Target name")

			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			resp, err := deps.PublishingService.CreateTarget(cmd.Context(), userID, publishing.CreateTargetRequest{
				Type:   domain.PublishingTargetType(targetType),
				Name:   name,
				Config: map[string]interface{}{"placeholder": true},
			})
			if err != nil {
				return fmt.Errorf("creating target: %w", err)
			}

			fmt.Printf("Created target: %s (%s) — ID: %s\n", resp.Name, resp.Type, resp.ID)
			return nil
		},
	}

	return cmd
}

func promptTargetType() string {
	return promptDefault("Target type (twitter, notion, google_doc, blog, markdown)", "twitter",
		func(v string) string {
			switch v {
			case "twitter", "notion", "google_doc", "blog", "markdown":
				return v
			default:
				return ""
			}
		})
}

func newTargetListCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured publishing targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			targets, err := deps.PublishingService.ListTargets(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("listing targets: %w", err)
			}

			if len(targets) == 0 {
				fmt.Println("No publishing targets configured.")
				fmt.Println("Use 'story target create' to add one.")
				return nil
			}

			for _, t := range targets {
				fmt.Printf("  %s (%s) — ID: %s\n", t.Name, t.Type, t.ID)
			}
			return nil
		},
	}
}
