package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/google/uuid"
)

func newPublishCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish entries to external platforms",
		Long:  "Publish entries to configured targets (Twitter, Notion, Google Docs, blog).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newPublishEntryCommand(deps))
	cmd.AddCommand(newPublishStatusCommand(deps))

	return cmd
}

func newPublishEntryCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry",
		Short: "Publish an entry to a target",
		Long: `Publish an entry to a publishing target interactively.

Example:
  story publish entry`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entryID := promptRequired("Entry ID")
			targetID := promptRequired("Target ID")

			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			eid, err := uuid.Parse(entryID)
			if err != nil {
				return fmt.Errorf("invalid entry ID: %w", err)
			}

			tid, err := uuid.Parse(targetID)
			if err != nil {
				return fmt.Errorf("invalid target ID: %w", err)
			}

			resp, err := deps.PublishingService.Publish(cmd.Context(), userID, publishing.PublishRequest{
				EntryID:  eid,
				TargetID: tid,
			})
			if err != nil {
				return fmt.Errorf("publish failed: %w", err)
			}

			fmt.Printf("Published! Status: %s\n", resp.Status)
			if resp.ExternalURL != "" {
				fmt.Printf("URL: %s\n", resp.ExternalURL)
			}
			return nil
		},
	}

	return cmd
}

func newPublishStatusCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check publish status of an entry",
		Long: `Check the publish status of an entry interactively.

Example:
  story publish status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entryID := promptRequired("Entry ID")

			eid, err := uuid.Parse(entryID)
			if err != nil {
				return fmt.Errorf("invalid entry ID: %w", err)
			}

			entries, err := deps.PublishingService.ListPublished(cmd.Context(), eid)
			if err != nil {
				return fmt.Errorf("fetching publish status: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("Entry has not been published yet.")
				return nil
			}

			for _, pe := range entries {
				urlInfo := ""
				if pe.ExternalURL != "" {
					urlInfo = fmt.Sprintf(" -> %s", pe.ExternalURL)
				}
				fmt.Printf("[%s] %s%s\n", pe.Status, pe.TargetID, urlInfo)
			}
			return nil
		},
	}

	return cmd
}
