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
	var entryType, title, tags, resourceIDs string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new learning entry",
		Long: `Add a new learning entry. Content is read from stdin.
Supports types: learning, work_log, resource, engineering_note.

Examples:
  story entry add --type learning --title "Go Interfaces" --tags go,patterns
  story entry add --type work_log --title "Sprint Review" --resource "res-id"`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVarP(&entryType, "type", "t", string(domain.EntryTypeLearning), "Entry type")
	cmd.Flags().StringVarP(&title, "title", "", "", "Entry title (required)")
	cmd.Flags().StringVarP(&tags, "tags", "", "", "Comma-separated tags")
	cmd.Flags().StringVarP(&resourceIDs, "resource", "r", "", "Comma-separated resource IDs")
	cmd.MarkFlagRequired("title")

	return cmd
}

func newEditCommand(deps *Dependencies) *cobra.Command {
	var title, content, tags string
	var entryType string

	cmd := &cobra.Command{
		Use:   "edit <entry-id>",
		Short: "Edit an existing entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuidParse(args[0])
			if err != nil {
				return err
			}

			req := entry.UpdateEntryRequest{
				Tags: parseCommaList(tags),
			}

			if title != "" {
				req.Title = &title
			}
			if content != "" {
				req.Content = &content
			}
			if entryType != "" {
				t := domain.EntryType(entryType)
				req.Type = &t
			}

			resp, err := deps.EntryService.Update(cmd.Context(), id, req)
			if err != nil {
				return fmt.Errorf("editing entry: %w", err)
			}

			fmt.Printf("Updated [%s] %s\n", resp.Type, resp.Title)
			return nil
		},
	}

	cmd.Flags().StringVarP(&entryType, "type", "t", "", "Entry type")
	cmd.Flags().StringVarP(&title, "title", "", "", "New title")
	cmd.Flags().StringVarP(&content, "content", "c", "", "New content")
	cmd.Flags().StringVarP(&tags, "tags", "", "", "Comma-separated tags")

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
