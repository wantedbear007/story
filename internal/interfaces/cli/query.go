package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
)

func newQueryCommand(deps *Dependencies) *cobra.Command {
	var query string
	var types []string
	var tags string
	var page int

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Search and list captured entries",
		Long: `Search your second brain with full-text search and filters.

Examples:
  story query --query "golang interfaces" --type learning
  story query --tag go,patterns --page 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			entryTypes := make([]domain.EntryType, len(types))
			for i, t := range types {
				entryTypes[i] = domain.EntryType(t)
			}

			tagList := strings.Split(tags, ",")
			if len(tagList) == 1 && tagList[0] == "" {
				tagList = nil
			}

			resp, err := deps.EntryService.List(cmd.Context(), userID, entry.EntryFilterRequest{
				Types:    entryTypes,
				Query:    query,
				Tags:     tagList,
				Page:     page,
				PageSize: 20,
			})
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}

			if len(resp.Entries) == 0 {
				fmt.Println("No entries found.")
				return nil
			}

			for _, e := range resp.Entries {
				tagStr := strings.Join(e.Tags, ", ")
				published := ""

				tagSection := ""
				if tagStr != "" {
					tagSection = fmt.Sprintf(" [%s]", tagStr)
				}

				fmt.Printf("[%s] %s%s\n", strings.ToUpper(string(e.Type)[:1]), e.Title, tagSection)
				fmt.Printf("  ID: %s | %s%s\n", e.ID, e.CreatedAt.Format("Jan 02, 2006 15:04"), published)

				// Show first line of content as preview
				if contentPreview, _, _ := strings.Cut(e.Content, "\n"); contentPreview != "" {
					maxLen := 80
					if len(contentPreview) > maxLen {
						contentPreview = contentPreview[:maxLen] + "..."
					}
					fmt.Printf("  %s\n", contentPreview)
				}
				fmt.Println()
			}

			fmt.Printf("Page %d — %d results\n", resp.Page, resp.Total)
			return nil
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "Full-text search query")
	cmd.Flags().StringArrayVarP(&types, "type", "t", nil, "Filter by type (can specify multiple: --type learning --type work_log)")
	cmd.Flags().StringVarP(&tags, "tags", "", "", "Filter by comma-separated tags")
	cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number")

	return cmd
}
