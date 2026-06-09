package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/google/uuid"

	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
)

func newCaptureCommand(deps *Dependencies) *cobra.Command {
	var entryType string
	var title string
	var tags string

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a new entry to your second brain",
		Long: `Capture a new entry. Supports types: learning, work_log, resource, engineering_note.

Examples:
  story capture --type learning --title "Go Interfaces" --tags go,patterns
  story capture --type work_log --title "Sprint Review Prep"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := readContentFromStdin()
			if err != nil {
				return fmt.Errorf("reading content: %w", err)
			}

			tagList := strings.Split(tags, ",")
			if len(tagList) == 1 && tagList[0] == "" {
				tagList = nil
			}

			// In production, extract userID from the stored JWT session.
			// For now we use a placeholder; the real auth middleware will inject it.
			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			resp, err := deps.EntryService.Create(cmd.Context(), userID, entry.CreateEntryRequest{
				Type:    domain.EntryType(entryType),
				Title:   title,
				Content: content,
				Tags:    tagList,
			})
			if err != nil {
				return fmt.Errorf("capture failed: %w", err)
			}

			fmt.Printf("Captured [%s] %s\n", resp.Type, resp.Title)
			fmt.Printf("ID: %s\n", resp.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&entryType, "type", "t", string(domain.EntryTypeLearning), "Entry type (learning, work_log, resource, engineering_note)")
	cmd.Flags().StringVarP(&title, "title", "", "", "Entry title (required)")
	cmd.Flags().StringVarP(&tags, "tags", "", "", "Comma-separated tags")
	cmd.MarkFlagRequired("title")

	return cmd
}

func readContentFromStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}

// resolveCurrentUserID extracts the authenticated user ID.
// In production, this reads from a session file or env variable.
// Future: integrate with a keyring/token store for persistent sessions.
func resolveCurrentUserID(deps *Dependencies) (uuid.UUID, error) {
	return uuid.Nil, fmt.Errorf("not yet implemented: session management")
}
