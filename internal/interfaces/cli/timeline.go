package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/entry"
)

func newTimelineCommand(deps *Dependencies) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show recent entries in chronological order",
		Long: `Display a timeline of your most recent entries.
Use --limit to control how many entries are shown.

Examples:
  story timeline
  story timeline --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			resp, err := deps.EntryService.List(cmd.Context(), userID, entry.EntryFilterRequest{
				Page:     1,
				PageSize: limit,
			})
			if err != nil {
				return fmt.Errorf("fetching timeline: %w", err)
			}

			if len(resp.Entries) == 0 {
				fmt.Println("No entries yet. Use 'story entry add' to create one.")
				return nil
			}

			for _, e := range resp.Entries {
				fmt.Printf("  %s [%s] %s\n", e.CreatedAt.Format("2006-01-02 15:04"), e.Type, e.Title)
			}
			fmt.Printf("\n%d entries\n", len(resp.Entries))

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Number of entries to show")

	return cmd
}
