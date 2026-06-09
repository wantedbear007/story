package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/resource"
	"github.com/anomalyco/story/internal/domain"
)

func newResourceCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Manage resources",
		Long:  "Track and manage external resources like URLs, articles, videos, and more.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newResourceAddCommand(deps))
	cmd.AddCommand(newResourceListCommand(deps))
	cmd.AddCommand(newResourceSearchCommand(deps))
	cmd.AddCommand(newResourceAttachCommand(deps))

	return cmd
}

func newResourceAddCommand(deps *Dependencies) *cobra.Command {
	var rtype, title, url, description string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new resource",
		Example: `  story resource add --type url --url "https://example.com" --title "Example"
  story resource add --type github --url "https://github.com/user/repo" --title "My Repo"
  story resource add --type youtube --url "https://youtu.be/dQw4w9WgXcQ" --title "Video"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			resp, err := deps.ResourceService.Create(cmd.Context(), userID, resource.CreateResourceRequest{
				Type:        domain.ResourceType(rtype),
				Title:       title,
				URL:         url,
				Description: description,
			})
			if err != nil {
				return fmt.Errorf("adding resource: %w", err)
			}

			fmt.Printf("Added resource [%s] %s\n", resp.Type, resp.Title)
			fmt.Printf("ID: %s\n", resp.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&rtype, "type", "t", "url", "Resource type (url, github, article, youtube, pdf, markdown)")
	cmd.Flags().StringVarP(&title, "title", "", "", "Resource title (required)")
	cmd.Flags().StringVarP(&url, "url", "u", "", "Resource URL (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Resource description")
	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("url")

	return cmd
}

func newResourceListCommand(deps *Dependencies) *cobra.Command {
	var rtype, search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			var types []domain.ResourceType
			if rtype != "" {
				types = []domain.ResourceType{domain.ResourceType(rtype)}
			}

			resp, err := deps.ResourceService.List(cmd.Context(), userID, resource.ResourceFilterRequest{
				Types:    types,
				Query:    search,
				Page:     1,
				PageSize: 50,
			})
			if err != nil {
				return fmt.Errorf("listing resources: %w", err)
			}

			if len(resp.Resources) == 0 {
				fmt.Println("No resources found")
				return nil
			}

			for _, r := range resp.Resources {
				fmt.Printf("  %s [%s] %s\n", r.ID[:8], r.Type, r.Title)
				fmt.Printf("    %s\n", r.URL)
				if r.Description != "" {
					fmt.Printf("    %s\n", r.Description)
				}
			}
			fmt.Printf("\n%d resources\n", len(resp.Resources))

			return nil
		},
	}

	cmd.Flags().StringVarP(&rtype, "type", "t", "", "Filter by type")
	cmd.Flags().StringVarP(&search, "search", "s", "", "Search query")

	return cmd
}

func newResourceSearchCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search resources by title or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			resp, err := deps.ResourceService.List(cmd.Context(), userID, resource.ResourceFilterRequest{
				Query:    args[0],
				Page:     1,
				PageSize: 50,
			})
			if err != nil {
				return fmt.Errorf("searching resources: %w", err)
			}

			if len(resp.Resources) == 0 {
				fmt.Println("No resources found")
				return nil
			}

			for _, r := range resp.Resources {
				fmt.Printf("  %s [%s] %s\n", r.ID[:8], r.Type, r.Title)
				fmt.Printf("    %s\n", r.URL)
			}
			fmt.Printf("\n%d results\n", len(resp.Resources))

			return nil
		},
	}
}

func newResourceAttachCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "attach <resource-id> <entry-id>",
		Short: "Attach a resource to an entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceID, err := uuidParse(args[0])
			if err != nil {
				return err
			}
			entryID, err := uuidParse(args[1])
			if err != nil {
				return err
			}

			if err := deps.ResourceService.AttachToEntry(cmd.Context(), resourceID, entryID); err != nil {
				return fmt.Errorf("attaching resource: %w", err)
			}

			fmt.Printf("Attached resource %s to entry %s\n", resourceID[:8], entryID[:8])
			return nil
		},
	}
}
