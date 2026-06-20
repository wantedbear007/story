//go:build linux

package notification

import (
	"context"
	"os/exec"

	"github.com/anomalyco/story/internal/domain"
)

func notifyPlatform(ctx context.Context, req domain.NotificationRequest) error {
	cmd := exec.CommandContext(ctx, "notify-send", req.Title, req.Message)
	return cmd.Run()
}
