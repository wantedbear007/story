package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/domain"
)

func newTweetCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tweet",
		Short: "Generate and manage tweets from entries",
		Long: `Generate tweets from your learning entries using AI.
Tweets follow a lifecycle: draft -> reviewing -> approved -> scheduled -> posted.
Use 'story tweet generate <entry-id>' to create a draft tweet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newTweetGenerateCommand(deps))
	cmd.AddCommand(newTweetRegenerateCommand(deps))
	cmd.AddCommand(newTweetListCommand(deps))
	cmd.AddCommand(newTweetGetCommand(deps))
	cmd.AddCommand(newTweetApproveCommand(deps))
	cmd.AddCommand(newTweetReviewCommand(deps))
	cmd.AddCommand(newTweetRejectCommand(deps))
	cmd.AddCommand(newTweetScheduleCommand(deps))
	cmd.AddCommand(newTweetArchiveCommand(deps))
	cmd.AddCommand(newTweetAuditCommand(deps))

	return cmd
}

func newTweetGenerateCommand(deps *Dependencies) *cobra.Command {
	var promptName string
	var temperature float64
	var maxTokens int

	cmd := &cobra.Command{
		Use:   "generate <entry-id>",
		Short: "Generate a tweet draft from an entry",
		Long: `Generate a tweet draft from an existing entry using AI.
By default uses the "tweet-summarize" prompt template.

Examples:
  story tweet generate <entry-id>
  story tweet generate <entry-id> --prompt tweet-thread`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			entryID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Generate(cmd.Context(), userID, content.GenerateRequest{
				EntryID:     entryID,
				PromptName:  promptName,
				Temperature: temperature,
				MaxTokens:   maxTokens,
			})
			if err != nil {
				return fmt.Errorf("generating tweet: %w", err)
			}

			printTweet(resp)
			return nil
		},
	}

	cmd.Flags().StringVarP(&promptName, "prompt", "p", "", "Prompt template name (default: tweet-summarize)")
	cmd.Flags().Float64VarP(&temperature, "temperature", "t", 0.7, "LLM temperature")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "m", 100, "Max output tokens")

	return cmd
}

func newTweetRegenerateCommand(deps *Dependencies) *cobra.Command {
	var promptName string
	var temperature float64
	var maxTokens int

	cmd := &cobra.Command{
		Use:   "regenerate <tweet-id>",
		Short: "Regenerate a tweet (archives old, creates new draft)",
		Long: `Regenerate an existing tweet. The current version is archived and a new
draft is created with an incremented version number.

Examples:
  story tweet regenerate <tweet-id>
  story tweet regenerate <tweet-id> --prompt tweet-thread`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Regenerate(cmd.Context(), userID, content.RegenerateRequest{
				TweetID:     tweetID,
				PromptName:  promptName,
				Temperature: temperature,
				MaxTokens:   maxTokens,
			})
			if err != nil {
				return fmt.Errorf("regenerating tweet: %w", err)
			}

			printTweet(resp)
			return nil
		},
	}

	cmd.Flags().StringVarP(&promptName, "prompt", "p", "", "Prompt template name")
	cmd.Flags().Float64VarP(&temperature, "temperature", "t", 0.7, "LLM temperature")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "m", 100, "Max output tokens")

	return cmd
}

func newTweetListCommand(deps *Dependencies) *cobra.Command {
	var entryID string
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tweets",
		Long: `List tweets with optional filtering by entry or status.

Examples:
  story tweet list
  story tweet list --status draft
  story tweet list --entry <entry-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			var parsedEntryID *uuid.UUID
			if entryID != "" {
				eid, err := uuidParse(entryID)
				if err != nil {
					return err
				}
				parsedEntryID = &eid
			}

			var parsedStatus *domain.TweetStatus
			if status != "" {
				s := domain.TweetStatus(status)
				parsedStatus = &s
			}

			if limit <= 0 || limit > 100 {
				limit = 20
			}

			resp, err := deps.TweetService.List(cmd.Context(), content.ListRequest{
				UserID:  userID,
				EntryID: parsedEntryID,
				Status:  parsedStatus,
				Limit:   limit,
			})
			if err != nil {
				return fmt.Errorf("listing tweets: %w", err)
			}

			if len(resp.Tweets) == 0 {
				fmt.Println("No tweets found")
				return nil
			}

			for _, t := range resp.Tweets {
				prefix := t.ID[:8]
				fmt.Printf("  %s [%s] (v%d) via %s\n", prefix, t.Status, t.Version, t.ProviderName)
				preview := t.Content
				if len(preview) > 80 {
					preview = preview[:80] + "..."
				}
				fmt.Printf("    %s\n", preview)
				fmt.Printf("    Entry: %s | Tokens: %d/%d | Cost: $%.6f\n",
					t.EntryID[:8], t.InputTokens, t.OutputTokens, t.CostUSD)
			}
			fmt.Printf("\n%d tweets\n", len(resp.Tweets))

			return nil
		},
	}

	cmd.Flags().StringVarP(&entryID, "entry", "e", "", "Filter by entry ID")
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (draft, reviewing, approved, scheduled, posted, archived)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Max results")

	return cmd
}

func newTweetGetCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get <tweet-id>",
		Short: "Show tweet details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Get(cmd.Context(), userID, tweetID)
			if err != nil {
				return fmt.Errorf("getting tweet: %w", err)
			}

			printTweet(resp)
			return nil
		},
	}
}

func newTweetApproveCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "approve <tweet-id>",
		Short: "Approve a tweet draft for publishing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Approve(cmd.Context(), userID, content.ApproveRequest{
				TweetID: tweetID,
			})
			if err != nil {
				return fmt.Errorf("approving tweet: %w", err)
			}

			fmt.Printf("Tweet %s approved\n", resp.ID[:8])
			return nil
		},
	}
}

func newTweetReviewCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "review <tweet-id>",
		Short: "Move a tweet to reviewing status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Review(cmd.Context(), userID, tweetID)
			if err != nil {
				return fmt.Errorf("reviewing tweet: %w", err)
			}

			fmt.Printf("Tweet %s moved to review\n", resp.ID[:8])
			return nil
		},
	}
}

func newTweetRejectCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "reject <tweet-id>",
		Short: "Reject a tweet and return it to draft",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Reject(cmd.Context(), userID, tweetID)
			if err != nil {
				return fmt.Errorf("rejecting tweet: %w", err)
			}

			fmt.Printf("Tweet %s returned to draft\n", resp.ID[:8])
			return nil
		},
	}
}

func newTweetScheduleCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "schedule <tweet-id> <datetime>",
		Short: "Schedule an approved tweet for posting",
		Long: `Schedule an approved tweet for posting at a specific time.
DateTime formats: "2006-01-02 15:04:05" or "2006-01-02T15:04:05"

Example:
  story tweet schedule <tweet-id> "2026-06-10 09:00:00"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			scheduledAt, err := parseDateTime(args[1])
			if err != nil {
				return fmt.Errorf("invalid datetime: %w", err)
			}

			resp, err := deps.TweetService.Schedule(cmd.Context(), userID, content.ScheduleRequest{
				TweetID:     tweetID,
				ScheduledAt: scheduledAt,
			})
			if err != nil {
				return fmt.Errorf("scheduling tweet: %w", err)
			}

			fmt.Printf("Tweet %s scheduled for %s\n", resp.ID[:8], scheduledAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}

func newTweetArchiveCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <tweet-id>",
		Short: "Archive a tweet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			resp, err := deps.TweetService.Archive(cmd.Context(), userID, tweetID)
			if err != nil {
				return fmt.Errorf("archiving tweet: %w", err)
			}

			fmt.Printf("Tweet %s archived\n", resp.ID[:8])
			return nil
		},
	}
}

func newTweetAuditCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "audit <tweet-id>",
		Short: "Show audit trail for a tweet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			tweetID, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			audits, err := deps.TweetService.GetAudits(cmd.Context(), userID, tweetID)
			if err != nil {
				return fmt.Errorf("getting audit trail: %w", err)
			}

			if len(audits) == 0 {
				fmt.Println("No audit records found")
				return nil
			}

			for _, a := range audits {
				userStr := "system"
				if a.UserID != nil {
					userStr = a.UserID.String()[:8]
				}
				fmt.Printf("  %s [%s] by %s\n", a.CreatedAt.Format("2006-01-02 15:04:05"), a.Action, userStr)
				if a.PreviousStatus != nil && a.NewStatus != nil {
					fmt.Printf("    Status: %s -> %s\n", *a.PreviousStatus, *a.NewStatus)
				}
				if a.PreviousContent != "" && a.PreviousContent != a.NewContent {
					prev := a.PreviousContent
					if len(prev) > 100 {
						prev = prev[:100] + "..."
					}
					new := a.NewContent
					if len(new) > 100 {
						new = new[:100] + "..."
					}
					fmt.Printf("    Previous: %s\n", prev)
					fmt.Printf("    New: %s\n", new)
				}
			}

			return nil
		},
	}
}

func printTweet(t *content.TweetResponse) {
	fmt.Printf("ID:      %s\n", t.ID)
	fmt.Printf("Entry:   %s\n", t.EntryID)
	fmt.Printf("Status:  %s (v%d)\n", t.Status, t.Version)
	fmt.Printf("Content: %s\n", t.Content)
	fmt.Printf("Prompt:  %s (v%d)\n", t.PromptName, t.PromptVer)
	fmt.Printf("Model:   %s / %s\n", t.ProviderName, t.ModelName)
	fmt.Printf("Tokens:  %d in / %d out\n", t.InputTokens, t.OutputTokens)
	fmt.Printf("Cost:    $%.6f\n", t.CostUSD)
	if t.RetryCount > 0 {
		fmt.Printf("Retries: %d\n", t.RetryCount)
	}
	fmt.Printf("Latency: %dms\n", t.LatencyMs)
	if t.ErrorMessage != "" {
		fmt.Printf("Error:   %s\n", t.ErrorMessage)
	}
	if t.ScheduledFor != nil {
		fmt.Printf("Scheduled: %s\n", t.ScheduledFor.Format("2006-01-02 15:04:05"))
	}
	if t.PostedAt != nil {
		fmt.Printf("Posted:    %s\n", t.PostedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Created: %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
}

func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q as datetime", s)
}
