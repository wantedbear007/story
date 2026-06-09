package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
)

func newCaptureCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a new entry to your second brain",
		Long: `Capture a new entry interactively. Content is read from stdin.

Examples:
  echo "Go interfaces allow you to define behavior" | story capture`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entryType := promptEntryType()
			title := promptRequired("Title")
			tags := promptInput("Tags (comma-separated): ")

			content, err := readContentFromStdin()
			if err != nil {
				return fmt.Errorf("reading content: %w", err)
			}

			tagList := parseCommaList(tags)

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

	return cmd
}

func promptEntryType() string {
	return promptDefault("Entry type (learning, work_log, resource, engineering_note)", "learning",
		func(v string) string {
			switch v {
			case "learning", "work_log", "resource", "engineering_note":
				return v
			default:
				return ""
			}
		})
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
