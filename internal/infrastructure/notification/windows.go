//go:build windows

package notification

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/anomalyco/story/internal/domain"
)

func notifyPlatform(ctx context.Context, req domain.NotificationRequest) error {
	script := fmt.Sprintf(
		`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; `+
			`$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); `+
			`$textNodes = $template.GetElementsByTagName("text"); `+
			`$textNodes.Item(0).AppendChild($template.CreateTextNode(%q)) > $null; `+
			`$textNodes.Item(1).AppendChild($template.CreateTextNode(%q)) > $null; `+
			`$toast = [Windows.UI.Notifications.ToastNotification]::new($template); `+
			`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier().Show($toast);`,
		req.Title, req.Message,
	)
	cmd := exec.CommandContext(ctx, "powershell", "-Command", script)
	return cmd.Run()
}
