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
	}

	cmd.AddCommand(newTagCreateCommand(deps))
	cmd.AddCommand(newTagListCommand(deps))

	return cmd
}

func newTagCreateCommand(deps *Dependencies) *cobra.Command {
	var name, color string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tag",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVarP(&name, "name", "n", "", "Tag name (required)")
	cmd.Flags().StringVarP(&color, "color", "c", "", "Tag color (hex, e.g., #ff0000)")
	cmd.MarkFlagRequired("name")

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
