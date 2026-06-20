package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/anomalyco/story/internal/application/daemon"
)

type PipelineStats struct {
	Raw        int
	Processing int
	Structured int
	Failed     int
}

func newStartCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the background daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil || deps.DaemonService == nil {
				return fmt.Errorf("daemon not configured")
			}
			return deps.DaemonService.Start(cmd.Context())
		},
	}
}

func newStopCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the background daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil || deps.DaemonService == nil {
				return fmt.Errorf("daemon not configured")
			}
			return deps.DaemonService.Stop(cmd.Context())
		},
	}
}

func newRestartCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the background daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil || deps.DaemonService == nil {
				return fmt.Errorf("daemon not configured")
			}
			if err := deps.DaemonService.Stop(cmd.Context()); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			return deps.DaemonService.Start(cmd.Context())
		},
	}
}

func newStartStatusCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon and pipeline status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil || deps.DaemonService == nil {
				return fmt.Errorf("daemon not configured")
			}
			return RunDaemonStatus(cmd.Context(), deps.DaemonService, nil)
		},
	}
}

func RunDaemonStatus(ctx context.Context, svc *daemon.Service, stats *PipelineStats) error {
	info, err := svc.Status(ctx)
	if err != nil {
		return err
	}
	if info.Status == "running" {
		fmt.Printf("Daemon is running (PID %d)\n", info.PID)
	} else {
		fmt.Println("Daemon is not running")
	}

	if stats != nil {
		fmt.Println()
		fmt.Println("Pipeline:")
		fmt.Printf("  Captured:    %d\n", stats.Raw)
		fmt.Printf("  Processing:  %d\n", stats.Processing)
		fmt.Printf("  Summarized:  %d\n", stats.Structured)
		if stats.Failed > 0 {
			fmt.Printf("  Failed:      %d ⚠\n", stats.Failed)
		} else {
			fmt.Printf("  Failed:      %d\n", stats.Failed)
		}
	}

	return nil
}
