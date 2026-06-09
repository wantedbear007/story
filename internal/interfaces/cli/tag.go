package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/tag"
)

func newTagCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
		Long:  "Create, list, and manage tags for organizing entries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newTagCreateCommand(deps))
	cmd.AddCommand(newTagListCommand(deps))

	return cmd
}

func newTagCreateCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tag",
		Long: `Create a new tag interactively.

Example:
  story tag create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := promptRequired("Tag name")
			color := promptInput("Color (hex, e.g. #ff0000): ")

			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			resp, err := deps.TagService.Create(cmd.Context(), userID, tag.CreateTagRequest{
				Name:  name,
				Color: color,
			})
			if err != nil {
				return fmt.Errorf("creating tag: %w", err)
			}

			fmt.Printf("Created tag: %s (ID: %s)\n", resp.Name, resp.ID)
			return nil
		},
	}

	return cmd
}

func newTagListCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			tags, err := deps.TagService.List(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("listing tags: %w", err)
			}

			if len(tags) == 0 {
				fmt.Println("No tags found.")
				return nil
			}

			for _, t := range tags {
				colorInfo := ""
				if t.Color != "" {
					colorInfo = fmt.Sprintf(" [%s]", t.Color)
				}
				fmt.Printf("  %s%s (ID: %s)\n", t.Name, colorInfo, t.ID)
			}
			return nil
		},
	}
}
