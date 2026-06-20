package cli

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/application/notification"
	"github.com/anomalyco/story/internal/domain"
	"github.com/spf13/cobra"
)

func newTestNotiCommand(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "test-noti",
		Short: "Send a test desktop notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil || deps.NotifService == nil {
				return fmt.Errorf("notification not configured")
			}
			captureURL := fmt.Sprintf("http://%s:%d/capture.html", deps.Cfg.Capture.Host, deps.Cfg.Capture.Port)
			return RunTestNoti(cmd.Context(), deps.NotifService, captureURL)
		},
	}
}

func RunTestNoti(ctx context.Context, svc *notification.Service, captureURL string) error {
	return svc.Send(ctx, domain.NotificationRequest{
		Title:   "Story",
		Message: "Time to capture what you learned today",
		URL:     captureURL,
	})
}
