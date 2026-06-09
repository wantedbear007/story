package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newWebCommand(deps *Dependencies) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start the web dashboard",
		Long: `Start the web dashboard server and open the browser.
The dashboard provides a GUI for managing tweets, viewing resources, and more.

Examples:
  story web
  story web --port 8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if port > 0 {
				deps.Cfg.Server.Port = port
				deps.ApiServer.SetPort(port)
			}

			s, err := loadSession()
			if err == nil {
				if err := deps.ApiServer.ValidateToken(s.AccessToken); err != nil {
					fmt.Fprintf(os.Stderr, "Session expired. Run 'story auth login' first.\n")
				} else {
					code, cerr := deps.ApiServer.CreateLoginCode(s.AccessToken)
					if cerr == nil {
						host := deps.Cfg.Server.Host
						if host == "0.0.0.0" {
							host = "localhost"
						}
						url := fmt.Sprintf("http://%s:%d/?code=%s", host, deps.Cfg.Server.Port, code)
						deps.ApiServer.SetAuthURL(url)
						fmt.Fprintf(os.Stderr, "Open this URL: %s\n", url)
					}
				}
			}

			if err := deps.ApiServer.Start(cmd.Context()); err != nil {
				return fmt.Errorf("web server error: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "Port to listen on (default: from config)")
	return cmd
}
