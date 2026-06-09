package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/raw_entry"
	"github.com/anomalyco/story/internal/domain"
)

func newRawCommand(deps *Dependencies) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "raw",
		Short: "Capture raw notes quickly",
		Long: `Capture unstructured thoughts, notes, or debugging logs without any structure.
Raw entries are stored as-is and can be processed later into structured knowledge.

Interactive mode:
  story raw

File input:
  story raw --file notes.txt

Pipe input:
  cat notes.txt | story raw
  git diff | story raw`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			var source domain.RawEntrySource

			if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
				content = strings.TrimSpace(string(data))
				source = domain.RawEntrySourceFile
			} else if !isTerminal() {
				var err error
				content, err = readContentFromStdin()
				if err != nil {
					return fmt.Errorf("reading pipe input: %w", err)
				}
				content = strings.TrimSpace(content)
				source = domain.RawEntrySourcePipe
			} else {
				fmt.Fprintln(os.Stderr, "Enter raw notes. Press CTRL+D when finished.")
				content = readMultilineInput()
				source = domain.RawEntrySourceCLI
			}

			if content == "" {
				return fmt.Errorf("no content provided")
			}

			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			resp, err := deps.RawEntryService.Create(cmd.Context(), userID, raw_entry.CreateRawEntryRequest{
				Content: content,
				Source:  source,
			})
			if err != nil {
				return fmt.Errorf("saving raw entry: %w", err)
			}

			fmt.Printf("Raw entry stored successfully\n")
			fmt.Printf("ID: %s\n", resp.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Read content from a file")
	return cmd
}

func isTerminal() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func readMultilineInput() string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return strings.Join(lines, "\n")
}

func newProcessCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process raw entries into structured knowledge",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newProcessRawCommand(deps))
	return cmd
}

func newProcessRawCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw",
		Short: "Process raw entries into structured knowledge",
		Long: `Process raw entries into learning entries, work logs, and more.
Processing delegates to an AI provider and is not yet implemented.

Examples:
  story process raw <id>
  story process raw --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")

			if all {
				fmt.Println("Processing all raw entries...")
				fmt.Println("AI processing is not yet implemented.")
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("specify a raw entry ID or use --all")
			}

			fmt.Printf("Processing raw entry %s...\n", args[0])
			fmt.Println("AI processing is not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Process all unprocessed raw entries")
	return cmd
}
