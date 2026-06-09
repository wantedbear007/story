package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/domain"
)

func newEntryCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry",
		Short: "Manage learning entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newAddCommand(deps))
	cmd.AddCommand(newEditCommand(deps))
	cmd.AddCommand(newDeleteCommand(deps))
	cmd.AddCommand(newTimelineCommand(deps))

	return cmd
}

func newAddCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new learning entry",
		Long: `Add a new learning entry interactively. Content is read from stdin.

Examples:
  echo "Go interfaces allow you to define behavior" | story entry add`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entryType := promptEntryType()
			title := promptRequired("Title")
			tags := promptInput("Tags (comma-separated): ")
			resourceIDs := promptInput("Resource IDs (comma-separated): ")

			content, err := readContentFromStdin()
			if err != nil {
				return fmt.Errorf("reading content: %w", err)
			}

			userID, err := resolveCurrentUserID(deps)
			if err != nil {
				return err
			}

			resp, err := deps.EntryService.Create(cmd.Context(), userID, entry.CreateEntryRequest{
				Type:      domain.EntryType(entryType),
				Title:     title,
				Content:   content,
				Tags:      parseCommaList(tags),
				Resources: parseUUIDList(resourceIDs),
			})
			if err != nil {
				return fmt.Errorf("adding entry: %w", err)
			}

			fmt.Printf("Added [%s] %s\n", resp.Type, resp.Title)
			fmt.Printf("ID: %s\n", resp.ID)
			return nil
		},
	}

	return cmd
}

func newEditCommand(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <entry-id>",
		Short: "Edit an existing entry",
		Long: `Edit an entry interactively. Leave fields blank to keep current values.

Example:
  story entry edit <entry-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			req := entry.UpdateEntryRequest{}

			if title := promptInput("Title (leave blank to keep): "); title != "" {
				req.Title = &title
			}
			if entryType := promptInput("Type (leave blank to keep): "); entryType != "" {
				t := domain.EntryType(entryType)
				if t != domain.EntryTypeLearning && t != domain.EntryTypeWorkLog &&
					t != domain.EntryTypeResource && t != domain.EntryTypeEngineeringNote {
					return fmt.Errorf("invalid type: %s", entryType)
				}
				req.Type = &t
			}
			if content := promptInput("Content (leave blank to keep): "); content != "" {
				req.Content = &content
			}
			if tags := promptInput("Tags (comma-separated, leave blank to keep): "); tags != "" {
				req.Tags = parseCommaList(tags)
			}

			resp, err := deps.EntryService.Update(cmd.Context(), id, req)
			if err != nil {
				return fmt.Errorf("editing entry: %w", err)
			}

			fmt.Printf("Updated [%s] %s\n", resp.Type, resp.Title)
			return nil
		},
	}

	return cmd
}

func newDeleteCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <entry-id>",
		Short: "Delete an entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			if err := deps.EntryService.Delete(cmd.Context(), id); err != nil {
				return fmt.Errorf("deleting entry: %w", err)
			}

			fmt.Printf("Deleted entry %s\n", id)
			return nil
		},
	}
}

func parseCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func parseUUIDList(s string) []uuid.UUID {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		u, err := uuid.Parse(strings.TrimSpace(p))
		if err == nil {
			result = append(result, u)
		}
	}
	return result
}
