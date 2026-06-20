//go:build darwin

package notification

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/anomalyco/story/internal/domain"
)

func notifyPlatform(ctx context.Context, req domain.NotificationRequest) error {
	var script string
	if req.URL != "" {
		script = fmt.Sprintf(
			`try
	display dialog %q with title %q buttons {"OK"} default button "OK"
end try
open location %q`,
			req.Message, req.Title, req.URL,
		)
	} else {
		script = fmt.Sprintf(
			`try
	display dialog %q with title %q buttons {"OK"} default button "OK"
end try`,
			req.Message, req.Title,
		)
	}
	cmd := exec.CommandContext(ctx, "/usr/bin/osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("osascript: %w (stderr: %s)", err, stderr.String())
	}
	return nil
}
