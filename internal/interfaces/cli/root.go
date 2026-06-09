package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/collection"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/application/publishing"
	"github.com/anomalyco/story/internal/application/tag"
	"github.com/anomalyco/story/internal/application/user"
	"github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/infrastructure/config"
)

// Dependencies holds all service dependencies for CLI commands.
// This is the composition root for CLI dependency injection.
type Dependencies struct {
	Cfg               *config.Config
	UserService       *user.Service
	EntryService      *entry.Service
	CollectionService *collection.Service
	TagService        *tag.Service
	PublishingService *publishing.Service
	AuthService       *auth.Service
}

func NewRootCommand(deps *Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:   "story",
		Short: "A CLI-first second brain for developers",
		Long: `Story captures learning, work logs, resources, and engineering notes,
transforms them into structured knowledge, and publishes to your favorite platforms.

Story helps you build your personal knowledge graph from the command line.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := deps.Cfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(newAuthCommand(deps))
	root.AddCommand(newCaptureCommand(deps))
	root.AddCommand(newQueryCommand(deps))
	root.AddCommand(newCollectionCommand(deps))
	root.AddCommand(newTagCommand(deps))
	root.AddCommand(newPublishCommand(deps))
	root.AddCommand(newTargetCommand(deps))
	root.AddCommand(newConfigCommand(deps))

	return root
}

func Execute(ctx context.Context, deps *Dependencies) {
	rootCmd := NewRootCommand(deps)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
