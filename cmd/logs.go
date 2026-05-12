// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/insights"
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
			namespace = "my-namespace"
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
			namespace = "my-namespace"
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

var logsSince string

// logsInsightsCmd summarises what the HyperFleet system did over a recent time window.
var logsInsightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Summarise recent HyperFleet system activity from pod logs",
	Long: `Fetch logs from api, sentinel, and adapter pods and produce a human-readable
summary of what the system has been doing.

  -s / --since   time window to look back (Go duration: 30s, 5m, 1h). Default: 1m.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dur, err := time.ParseDuration(logsSince)
		if err != nil {
			return fmt.Errorf("[ERROR] invalid --since value %q: %v", logsSince, err)
		}
		sinceSeconds := int64(dur.Seconds())
		if sinceSeconds < 1 {
			sinceSeconds = 1
		}

		s, err := loadConfig()
		if err != nil {
			return err
		}
		namespace := s.Get("kubernetes", "namespace")
		if namespace == "" {
			namespace = "my-namespace"
		}

		cs, err := kube.NewClientset(resolvedKubeconfig(s))
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}

		ctx := context.Background()
		var (
			apiLines      []string
			sentinelLines []string
			adapterLines  []string
			apiErr        error
			sentinelErr   error
			adapterErr    error
			wg            sync.WaitGroup
		)
		wg.Add(3)
		go func() { defer wg.Done(); apiLines, apiErr = kube.CollectLogs(ctx, cs, namespace, "api", sinceSeconds) }()
		go func() {
			defer wg.Done()
			sentinelLines, sentinelErr = kube.CollectLogs(ctx, cs, namespace, "sentinel", sinceSeconds)
		}()
		go func() {
			defer wg.Done()
			adapterLines, adapterErr = kube.CollectLogs(ctx, cs, namespace, "adapter", sinceSeconds)
		}()
		wg.Wait()

		w := cmd.OutOrStdout()
		label := fmt.Sprintf("(last %s)", logsSince)

		// ---- API section ----
		fmt.Fprintf(w, "API  %s\n", label)
		if apiErr != nil {
			fmt.Fprintf(w, "  [ERROR] %v\n", apiErr)
		} else {
			ai := insights.ParseAPILogs(apiLines)
			if len(ai.Endpoints) == 0 {
				fmt.Fprintln(w, "  (no activity)")
			} else {
				// Align columns.
				maxMethod, maxPath := 0, 0
				for _, e := range ai.Endpoints {
					if len(e.Method) > maxMethod {
						maxMethod = len(e.Method)
					}
					if len(e.Path) > maxPath {
						maxPath = len(e.Path)
					}
				}
				for _, e := range ai.Endpoints {
					fmt.Fprintf(w, "  %-*s  %-*s  OK: %-4d ERR: %d\n",
						maxMethod, e.Method, maxPath, e.Path, e.OK, e.Err)
				}
			}
		}

		// ---- Sentinel section ----
		fmt.Fprintf(w, "\nSENTINEL  %s\n", label)
		if sentinelErr != nil {
			fmt.Fprintf(w, "  [ERROR] %v\n", sentinelErr)
		} else {
			si := insights.ParseSentinelLogs(sentinelLines)
			if len(si.Topics) == 0 {
				fmt.Fprintln(w, "  (no activity)")
			} else {
				maxTopic := 0
				for _, t := range si.Topics {
					if len(t.Topic) > maxTopic {
						maxTopic = len(t.Topic)
					}
				}
				for _, t := range si.Topics {
					fmt.Fprintf(w, "  %-*s  cycles: %-4d published: %-4d skipped: %d\n",
						maxTopic, t.Topic, t.Cycles, t.Published, t.Skipped)
				}
			}
		}

		// ---- Adapters section ----
		fmt.Fprintf(w, "\nADAPTERS  %s\n", label)
		if adapterErr != nil {
			fmt.Fprintf(w, "  [ERROR] %v\n", adapterErr)
		} else {
			adp := insights.ParseAdapterLogs(adapterLines)
			if len(adp.Adapters) == 0 {
				fmt.Fprintln(w, "  (no activity)")
			} else {
				maxName := 0
				for _, a := range adp.Adapters {
					if len(a.Name) > maxName {
						maxName = len(a.Name)
					}
				}
				// Sort adapters for stable output.
				sorted := make([]insights.AdapterStat, len(adp.Adapters))
				copy(sorted, adp.Adapters)
				sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
				for _, a := range sorted {
					var phaseParts []string
					for _, p := range a.Phases {
						phaseParts = append(phaseParts, fmt.Sprintf("%s(%d)", p.Name, p.Count))
					}
					phaseStr := strings.Join(phaseParts, " ")
					if phaseStr == "" {
						phaseStr = "(none)"
					}
					fmt.Fprintf(w, "  %-*s  executions: %-4d phases: %s\n",
						maxName, a.Name, a.Executions, phaseStr)
				}
			}
		}

		return nil
	},
}

func init() {
	logsAdapterCmd.Flags().StringVar(&logsClusterID, "cluster-id", "",
		"filter by cluster ID (default: active cluster from state)")
	logsInsightsCmd.Flags().StringVarP(&logsSince, "since", "s", "1m",
		"time window to look back (e.g. 30s, 5m, 1h)")
	logsCmd.AddCommand(logsAdapterCmd)
	logsCmd.AddCommand(logsInsightsCmd)
	rootCmd.AddCommand(logsCmd)
}
