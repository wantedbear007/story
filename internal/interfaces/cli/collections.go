package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/collection"
)

func newCollectionCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Manage collections",
		Long:  "Create, list, and manage collections to organize your entries.",
	}

	cmd.AddCommand(newCollectionCreateCommand(deps))
	cmd.AddCommand(newCollectionListCommand(deps))

	return cmd
}

func newCollectionCreateCommand(deps *Dependencies) *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			resp, err := deps.CollectionService.Create(cmd.Context(), userID, collection.CreateCollectionRequest{
				Name:        name,
				Description: description,
			})
			if err != nil {
				return fmt.Errorf("creating collection: %w", err)
			}

			fmt.Printf("Created collection: %s (ID: %s)\n", resp.Name, resp.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Collection name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Collection description")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newCollectionListCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			cols, err := deps.CollectionService.List(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("listing collections: %w", err)
			}

			if len(cols) == 0 {
				fmt.Println("No collections found. Create one with 'story collection create'.")
				return nil
			}

			for _, c := range cols {
				desc := c.Description
				if desc == "" {
					desc = "(no description)"
				}
				fmt.Printf("  %s — %s\n", c.Name, desc)
				fmt.Printf("    ID: %s\n", c.ID)
			}
			return nil
		},
	}
}
