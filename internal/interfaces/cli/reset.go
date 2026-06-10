package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newResetCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset all local config and session data",
		Long: `Remove all local configuration and session files from ~/.story/.

This deletes:
  - ~/.story/config.yaml   (configuration)
  - ~/.story/session.json   (login session)

Your database and remote data are NOT affected by this command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReset(deps)
		},
	}
}

func runReset(deps *Dependencies) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}
	storyDir := filepath.Join(home, ".story")

	if _, err := os.Stat(storyDir); os.IsNotExist(err) {
		fmt.Println("Nothing to reset — ~/.story/ does not exist")
		return nil
	}

	fmt.Print("This will delete all config and session data. Type 'yes' to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "yes" {
		fmt.Println("Reset cancelled")
		return nil
	}

	removed := false

	configPath := filepath.Join(storyDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Printf("Removed %s\n", configPath)
		removed = true
	}

	sessionPath := filepath.Join(storyDir, "session.json")
	if _, err := os.Stat(sessionPath); err == nil {
		if err := os.Remove(sessionPath); err != nil {
			return fmt.Errorf("removing session: %w", err)
		}
		fmt.Printf("Removed %s\n", sessionPath)
		removed = true
	}

	if removed {
		if entries, _ := os.ReadDir(storyDir); len(entries) == 0 {
			os.Remove(storyDir)
		}
		fmt.Println("Reset complete")
	} else {
		fmt.Println("Nothing to reset")
	}

	return nil
}
