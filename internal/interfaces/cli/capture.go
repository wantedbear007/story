package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
)

func newCaptureCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a new entry to your second brain",
		Long: `Capture a new entry. By default opens the browser-based capture page.
Content piped via stdin is captured as a structured entry.

Examples:
  story capture
  echo "Go interfaces allow you to define behavior" | story capture --stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			useStdin, _ := cmd.Flags().GetBool("stdin")

			if useStdin {
				return captureFromStdin(cmd, deps)
			}

			return openCaptureBrowser(deps)
		},
	}

	cmd.Flags().Bool("stdin", false, "Capture content from stdin")
	return cmd
}

func captureFromStdin(cmd *cobra.Command, deps *Dependencies) error {
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
}

func openCaptureBrowser(deps *Dependencies) error {
	host := deps.Cfg.Capture.Host
	if host == "0.0.0.0" || host == "" {
		host = "127.0.0.1"
	}
	url := fmt.Sprintf("http://%s:%d/capture.html", host, deps.Cfg.Capture.Port)
	fmt.Printf("Opening %s\n", url)

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	if err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}
	return nil
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
