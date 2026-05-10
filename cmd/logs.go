// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
	"github.com/spf13/cobra"
)

var logsClusterID string

// logsCmd streams pod logs matching a pattern (replaces stern).
var logsCmd = &cobra.Command{
	Use:   "logs [pattern]",
	Short: "Stream pod logs matching a pattern",
	Long: `Stream pod logs from one or more pods matching a pattern.

If stern is available in PATH it is used; otherwise goroutine fan-out is used.

Subcommands: adapter.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}
		s, err := loadConfig()
		if err != nil {
			return err
		}
		namespace := s.Get("kubernetes", "namespace")
		if namespace == "" {
			namespace = "amarin-ns1"
		}

		// Delegate to stern if available.
		if sternPath, err := exec.LookPath("stern"); err == nil {
			sternArgs := []string{pattern, "-n", namespace}
			c := exec.CommandContext(context.Background(), sternPath, sternArgs...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin
			return c.Run()
		}

		// Fallback: goroutine fan-out via client-go.
		cs, err := kube.NewClientset(resolvedKubeconfig(s))
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}
		return kube.StreamLogs(context.Background(), cs, namespace, pattern, os.Stdout)
	},
}

// logsAdapterCmd tails adapter logs filtered by cluster ID.
var logsAdapterCmd = &cobra.Command{
	Use:   "adapter [pattern]",
	Short: "Stream adapter logs filtered by cluster ID",
	Long: `Stream adapter pod logs filtered to lines containing cluster_id=<id>.

JSON/OpenTelemetry span lines (starting with '{') are skipped.
Lines are displayed as: [pod] <time>  <LEVEL>  <msg>`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		namespace := s.Get("kubernetes", "namespace")
		if namespace == "" {
			namespace = "amarin-ns1"
		}

		podPattern := "adapter"
		if len(args) > 0 && args[0] != "" {
			podPattern = "adapter-" + args[0]
		}

		clusterID := logsClusterID
		if clusterID == "" {
			clusterID = s.GetState("cluster-id")
		}

		cs, err := kube.NewClientset(resolvedKubeconfig(s))
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}
		return kube.StreamLogsFiltered(context.Background(), cs, namespace, podPattern, clusterID, os.Stdout)
	},
}

func init() {
	logsAdapterCmd.Flags().StringVar(&logsClusterID, "cluster-id", "",
		"filter by cluster ID (default: active cluster from state)")
	logsCmd.AddCommand(logsAdapterCmd)
	rootCmd.AddCommand(logsCmd)
}
