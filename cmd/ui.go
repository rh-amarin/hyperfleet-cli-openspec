package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/rh-amarin/hyperfleet-cli/internal/server"
	uiassets "github.com/rh-amarin/hyperfleet-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	uiPort int
	uiOpen bool
)

func newUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Start the HyperFleet browser dashboard",
		Long: `Start an HTTP server that serves the HyperFleet browser dashboard.

The dashboard provides a live view of clusters and node pools with condition
status dots, auto-polling, and a drill-down side panel. Open your browser
at the printed URL.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := loadConfig()
			if err != nil {
				return err
			}
			if _, err := s.RequireActiveEnvironment(); err != nil {
				return fmt.Errorf("[ERROR] no active environment\n  → run 'hf env create <name>' to create one\n  → run 'hf env activate <name>' to activate an existing one")
			}

			client := newAPIClient(s)

			indexHTML, err := uiassets.StaticFS.ReadFile("static/index.html")
			if err != nil {
				return fmt.Errorf("reading embedded UI: %w", err)
			}

			srv := server.New(client, uiPort, indexHTML)

			if uiOpen {
				url := fmt.Sprintf("http://localhost:%d", uiPort)
				openBrowser(url)
			}

			return srv.Listen()
		},
	}

	cmd.Flags().IntVarP(&uiPort, "port", "p", 8088, "port to listen on")
	cmd.Flags().BoolVar(&uiOpen, "open", false, "open browser automatically on start")

	return cmd
}

// openBrowser attempts to open url in the system default browser.
func openBrowser(url string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	_ = c.Start()
}

func init() {
	rootCmd.AddCommand(newUICmd())
}
