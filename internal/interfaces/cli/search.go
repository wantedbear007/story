package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
)

func newSearchCommand(deps *Dependencies) *cobra.Command {
	var entryType, tags string
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search entries by title, content, and tags",
		Long: `Full-text search across all entries.
Supports filtering by type and tags.

Examples:
  story search "Go interfaces"
  story search "deployment" --type learning
  story search "testing" --tags go,testing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			var types []domain.EntryType
			if entryType != "" {
				types = []domain.EntryType{domain.EntryType(entryType)}
			}

			resp, err := deps.EntryService.List(cmd.Context(), userID, entry.EntryFilterRequest{
				Query:    args[0],
				Types:    types,
				Tags:     parseCommaList(tags),
				Page:     1,
				PageSize: limit,
			})
			if err != nil {
				return fmt.Errorf("searching entries: %w", err)
			}

			if len(resp.Entries) == 0 {
				fmt.Println("No results found")
				return nil
			}

			for _, e := range resp.Entries {
				fmt.Printf("  %s [%s] %s\n", e.ID[:8], e.Type, e.Title)
				if len(e.Tags) > 0 {
					fmt.Printf("    Tags: %s\n", strings.Join(e.Tags, ", "))
				}
				preview := e.Content
				if len(preview) > 120 {
					preview = preview[:120] + "..."
				}
				if preview != "" {
					fmt.Printf("    %s\n", preview)
				}
			}
			fmt.Printf("\n%d results\n", len(resp.Entries))

			return nil
		},
	}

	cmd.Flags().StringVarP(&entryType, "type", "t", "", "Filter by type")
	cmd.Flags().StringVarP(&tags, "tags", "", "", "Filter by comma-separated tags")
	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Max results")

	return cmd
}
