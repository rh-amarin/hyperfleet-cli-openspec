// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
	"github.com/spf13/cobra"
)

// clusterCmd is the top-level group for all cluster operations.
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage HyperFleet clusters",
	Long: `Manage HyperFleet clusters.

Subcommands: create, get, list, search, patch, delete, conditions, statuses.`,
}

// ---- flag vars ----

var (
	clusterCreateName      string
	clusterCreateFile      string
	clusterCreateReplicas  int
	clusterCreateNPID      string
	clusterListWatch       bool
	clusterListWatchSecs   int
	clusterListSearch      string
	clusterDeleteForce     bool
	clusterDeleteReason    string
	clusterStatusesFilter  bool
)

// ---- helpers ----

// newAPIClient builds an API client from the active config store.
// baseURL is constructed as: <api-url>/api/hyperfleet/<api-version>/
func newAPIClient(s *config.Store) *api.Client {
	apiURLVal := s.Get("hyperfleet", "api-url")
	apiVersion := s.Get("hyperfleet", "api-version")
	baseURL := strings.TrimRight(apiURLVal, "/") + "/api/hyperfleet/" + apiVersion + "/"
	token := s.Get("hyperfleet", "token")
	return api.NewClient(baseURL, token, verbose, curlMode)
}

// loadConfig loads config and returns a store ready for use.
func loadConfig() (*config.Store, error) {
	s := config.NewFromEnv()
	if err := s.Load(); err != nil {
		return nil, fmt.Errorf("[ERROR] loading config: %w", err)
	}
	return s, nil
}

// handleAPIError prints RFC 7807 errors as JSON (exit 0) and returns nil.
// Non-API errors are returned as-is (exit 1).
func handleAPIError(p *output.Printer, err error) error {
	if errors.Is(err, api.ErrDryRun) {
		return nil
	}
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		_ = p.Print(apiErr)
		return nil
	}
	return err
}

func errCurlInteractive(interactive bool) error {
	if curlMode && interactive {
		return fmt.Errorf("[ERROR] --curl cannot be used with interactive mode")
	}
	return nil
}

// clusterOverallStatus derives a table status from cluster conditions.
// Returns a colored dot (or plain text) representing Available AND Reconciled.
func clusterOverallStatus(c resource.Cluster, nc bool) string {
	available, reconciled := "", ""
	for _, cond := range c.Status.Conditions {
		switch cond.Type {
		case "Available":
			available = cond.Status
		case "Reconciled":
			reconciled = cond.Status
		}
	}
	if available == "True" && reconciled == "True" {
		return output.StatusDot("True", nc)
	}
	return output.StatusDot("False", nc)
}

// conditionsView is the JSON shape emitted by `hf cluster conditions`.
type conditionsView struct {
	Generation int32                `json:"generation"`
	Status     conditionsViewStatus `json:"status"`
}

type conditionsViewStatus struct {
	Conditions []resource.ResourceCondition `json:"conditions"`
}

// ---- cluster list ----

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		if clusterListWatch && outputFmt == "table" {
			if curlMode {
				return fetchAndRenderClusterList(cmd, 0, 0)
			}
			ctx, cancel := watchContext(context.Background())
			defer cancel()
			return runWatch(ctx, cmd.OutOrStdout(), clusterListWatchSecs, func(tick int) error {
				return fetchAndRenderClusterList(cmd, tick, clusterListWatchSecs)
			})
		}
		return fetchAndRenderClusterList(cmd, 0, 0)
	},
}

func fetchAndRenderClusterList(cmd *cobra.Command, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	path := "clusters"
	if clusterListSearch != "" {
		path = "clusters?search=" + url.QueryEscape(clusterListSearch)
	}
	list, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, path)
	if err != nil {
		return handleAPIError(p, err)
	}

	if outputFmt == "table" {
		headers := []string{"ID", "NAME", "GEN", "STATUS"}
		rows := make([][]string, 0, len(list.Items))
		for _, c := range list.Items {
			rows = append(rows, []string{
				c.ID,
				c.Name,
				strconv.Itoa(int(c.Generation)),
				clusterOverallStatus(c, noColor),
			})
		}
		return p.PrintTable(headers, rows)
	}
	return p.Print(list)
}

// ---- cluster get ----

var clusterGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a cluster by ID",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		explicit := ""
		if len(args) > 0 {
			explicit = args[0]
		}
		if clusterInteractive && explicit == "" {
			explicit, err = pickClusterInteractive(cmd, s)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.ClusterID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		cluster, err := api.Get[resource.Cluster](context.Background(), client, "clusters/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		if outputFmt == "table" {
			headers := []string{"ID", "NAME", "GEN", "STATUS"}
			rows := [][]string{{
				cluster.ID,
				cluster.Name,
				strconv.Itoa(int(cluster.Generation)),
				clusterOverallStatus(cluster, noColor),
			}}
			return p.PrintTable(headers, rows)
		}
		return p.Print(cluster)
	},
}

// ---- cluster search ----

var clusterSearchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "Search for a cluster by name and set as active context",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		// No name arg: behave like `hf cluster get` using the state clusters.
		if len(args) == 0 {
			id := s.GetState("clusters")
			if id == "" {
				return fmt.Errorf("[ERROR] No clusters set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
			}
			client := newAPIClient(s)
			cluster, err := api.Get[resource.Cluster](context.Background(), client, "clusters/"+id)
			if err != nil {
				return handleAPIError(p, err)
			}
			return p.Print(cluster)
		}

		name := args[0]
		client := newAPIClient(s)

		list, err := api.Get[resource.ListResponse[resource.Cluster]](
			context.Background(), client,
			"clusters?search=name='"+name+"'",
		)
		if err != nil {
			return handleAPIError(p, err)
		}

		if len(list.Items) == 0 {
			p.Warn(fmt.Sprintf("No clusters found matching '%s'", name))
			return p.Print([]resource.Cluster{})
		}

		if len(list.Items) > 1 {
			p.Warn(fmt.Sprintf("Multiple clusters found matching '%s', using first result", name))
		}

		first := list.Items[0]
		if setErr := s.SetState("clusters", first.ID); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist clusters: %v\n", setErr)
		} else {
			p.Info(fmt.Sprintf("Cluster context set to '%s'", first.ID))
		}

		return p.Print(list.Items)
	},
}

// ---- cluster create ----

var clusterCreateCmd = &cobra.Command{
	Use:   "create [name] [region] [version]",
	Short: "Create a new cluster",
	Args:  cobra.MaximumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}

		body, err := loadTemplate("cluster", clusterCreateFile)
		if err != nil {
			return fmt.Errorf("[ERROR] %w", err)
		}

		// Name: positional arg > --name flag > template value.
		if clusterCreateName != "" {
			body["name"] = clusterCreateName
		}
		if len(args) >= 1 {
			body["name"] = args[0]
		}

		// Region / version positional overrides into spec.
		spec := ensureSpecMap(body)
		if len(args) >= 2 {
			spec["region"] = args[1]
		}
		if len(args) >= 3 {
			spec["version"] = args[2]
		}

		// Flag overrides.
		if clusterCreateReplicas > 0 {
			spec["replicas"] = strconv.Itoa(clusterCreateReplicas)
		}
		if clusterCreateNPID != "" {
			body["nodepool_id"] = clusterCreateNPID
		}

		name, _ := body["name"].(string)

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		if !curlMode {
			existing, err := api.Get[resource.ListResponse[resource.Cluster]](
				context.Background(), client,
				"clusters?search=name='"+name+"'",
			)
			if err == nil && len(existing.Items) > 0 {
				p.Warn(fmt.Sprintf("Cluster '%s' already exists, skipping creation", name))
				return nil
			}
		}

		cluster, err := api.Post[resource.Cluster](context.Background(), client, "clusters", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		if setErr := s.SetState("clusters", cluster.ID); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist clusters: %v\n", setErr)
		} else {
			p.Info(fmt.Sprintf("Cluster context set to '%s'", cluster.ID))
		}

		return p.Print(cluster)
	},
}

// ---- cluster patch ----

var clusterPatchCmd = &cobra.Command{
	Use:   "patch {spec|labels} [cluster_id]",
	Short: "Increment a counter field in cluster spec or labels",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != "spec" && args[0] != "labels") {
			fmt.Fprintln(cmd.OutOrStdout(), "Usage: hf cluster patch {spec|labels} [cluster_id]")
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "Arguments:")
			fmt.Fprintln(cmd.OutOrStdout(), "  spec|labels   Which section to increment the counter field in (required)")
			fmt.Fprintln(cmd.OutOrStdout(), "  cluster_id    Cluster ID (default: current cluster)")
			return fmt.Errorf("invalid arguments")
		}

		section := args[0]
		explicit := ""
		if len(args) == 2 {
			explicit = args[1]
		}

		s, err := loadConfig()
		if err != nil {
			return err
		}
		if clusterInteractive && explicit == "" {
			explicit, err = pickClusterInteractive(cmd, s)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.ClusterID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		cluster, err := api.Get[resource.Cluster](context.Background(), client, "clusters/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		oldVal, newVal := bumpCounter(section, cluster.Spec, cluster.Labels)

		p.Info(fmt.Sprintf("Incrementing %s.counter: %d -> %d", section, oldVal, newVal))

		body := patchCounterBody(section, newVal, cluster.Spec, cluster.Labels)

		_, err = api.Patch[resource.Cluster](context.Background(), client, "clusters/"+id, body)
		if err != nil {
			return handleAPIError(p, err)
		}
		return nil
	},
}

// ---- cluster delete ----

var clusterDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a cluster",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if clusterDeleteForce && clusterDeleteReason == "" {
			return fmt.Errorf("[ERROR] --reason is required when using --force")
		}
		s, err := loadConfig()
		if err != nil {
			return err
		}
		explicit := ""
		if len(args) > 0 {
			explicit = args[0]
		}
		if clusterInteractive && explicit == "" {
			explicit, err = pickClusterInteractive(cmd, s)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.ClusterID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		if clusterDeleteForce {
			_, err = api.Post[resource.Cluster](context.Background(), client,
				"clusters/"+id+"/force-delete",
				map[string]string{"reason": clusterDeleteReason},
			)
			if err != nil {
				var apiErr *api.APIError
				if errors.As(err, &apiErr) && apiErr.Status == 404 {
					return fmt.Errorf("[ERROR] Cluster '%s' not found", id)
				}
				return handleAPIError(p, err)
			}
			p.Info(fmt.Sprintf("Cluster '%s' force-deleted", id))
			return nil
		}

		deleted, err := api.Delete[resource.Cluster](context.Background(), client, "clusters/"+id)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				return fmt.Errorf("[ERROR] Cluster '%s' not found", id)
			}
			return handleAPIError(p, err)
		}
		return p.Print(deleted)
	},
}

// ---- cluster conditions ----

var clusterConditionsCmd = &cobra.Command{
	Use:   "conditions [id]",
	Short: "Get cluster status conditions",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		explicit := ""
		if len(args) > 0 {
			explicit = args[0]
		}
		if clusterInteractive && explicit == "" {
			explicit, err = pickClusterInteractive(cmd, s)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.ClusterID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		cluster, err := api.Get[resource.Cluster](context.Background(), client, "clusters/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		if outputFmt == "table" {
			headers := []string{"TYPE", "STATUS", "LAST TRANSITION", "REASON", "MESSAGE"}
			rows := make([][]string, 0, len(cluster.Status.Conditions))
			for _, cond := range cluster.Status.Conditions {
				rows = append(rows, []string{
					cond.Type,
					output.StatusDot(cond.Status, noColor),
					cond.LastTransitionTime,
					cond.Reason,
					cond.Message,
				})
			}
			return p.PrintTable(headers, rows)
		}

		out := conditionsView{
			Generation: cluster.Generation,
			Status:     conditionsViewStatus{Conditions: cluster.Status.Conditions},
		}
		return p.Print(out)
	},
}

// ---- cluster statuses ----

var clusterStatusesCmd = &cobra.Command{
	Use:   "statuses [id]",
	Short: "Get cluster adapter statuses",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		explicit := ""
		if len(args) > 0 {
			explicit = args[0]
		}
		if clusterInteractive && explicit == "" {
			explicit, err = pickClusterInteractive(cmd, s)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.ClusterID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
			context.Background(), client, "clusters/"+id+"/statuses",
		)
		if err != nil {
			return handleAPIError(p, err)
		}

		if clusterStatusesFilter {
			return runStatusFilterUI(list.Items, noColor)
		}

		if outputFmt == "table" {
			headers := []string{"ADAPTER", "GEN", "AVAILABLE", "FINALIZED"}
			rows := make([][]string, 0, len(list.Items))
			for _, as := range list.Items {
				avail, final := "-", "-"
				for _, cond := range as.Conditions {
					switch cond.Type {
					case "Available":
						avail = output.StatusDot(cond.Status, noColor)
					case "Finalized":
						final = output.StatusDot(cond.Status, noColor)
					}
				}
				rows = append(rows, []string{
					as.Adapter,
					strconv.Itoa(int(as.ObservedGeneration)),
					avail,
					final,
				})
			}
			return p.PrintTable(headers, rows)
		}
		return p.Print(list)
	},
}

// ---- cluster adapter ----

var clusterAdapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Adapter operations for a cluster",
}

var clusterAdapterPostStatusCmd = &cobra.Command{
	Use:   "post-status <adapter_name> <True|False|Unknown> <generation>",
	Short: "Post adapter status conditions for the current cluster",
	Args:  helpOnNoArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		adapterName := args[0]
		status := args[1]
		genStr := args[2]

		if status != "True" && status != "False" && status != "Unknown" {
			return fmt.Errorf("[ERROR] Invalid status value '%s'. Must be one of: True, False, Unknown.", status)
		}
		gen, err := strconv.Atoi(genStr)
		if err != nil {
			return fmt.Errorf("[ERROR] Invalid generation '%s': must be an integer", genStr)
		}

		s, err := loadConfig()
		if err != nil {
			return err
		}
		var clusterID string
		if clusterInteractive {
			clusterID, err = pickClusterInteractive(cmd, s)
			if err != nil || clusterID == "" {
				return err
			}
		} else {
			clusterID, err = s.ClusterID("")
			if err != nil {
				return err
			}
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		body := resource.AdapterStatusCreateRequest{
			Adapter:            adapterName,
			ObservedGeneration: int32(gen),
			ObservedTime:       time.Now().UTC().Format(time.RFC3339),
			Conditions: []resource.ConditionRequest{
				{Type: "Available", Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
				{Type: "Applied", Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
				{Type: "Health", Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
				{Type: "Finalized", Status: status, Reason: "ManualStatusPost", Message: "Status posted via hf adapter post-status"},
			},
		}

		result, err := api.Put[resource.AdapterStatus](context.Background(), client, "clusters/"+clusterID+"/statuses", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		if result.Adapter == "" {
			p.Info(fmt.Sprintf("Posted adapter status for %s on cluster %s (no-op: status unchanged)", adapterName, clusterID))
			return nil
		}
		p.Info(fmt.Sprintf("Posted adapter status for %s on cluster %s", adapterName, clusterID))
		return p.Print(result)
	},
}

// ---- cluster table ----

var clusterTableCmd = &cobra.Command{
	Use:   "table",
	Short: "List all clusters in table format (alias for: cluster list --output table)",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFmt = "table"
		return fetchAndRenderClusterList(cmd, 0, 0)
	},
}

// ---- cluster id ----

var clusterInteractive bool
var clusterIDInteractive bool

// clusterIDSel is the selector used by hf cluster id -i and pickClusterInteractive; swapped in tests.
var clusterIDSel selector.Selector = selector.FuzzySelector{}

func pickClusterInteractive(cmd *cobra.Command, s *config.Store) (string, error) {
	if err := errCurlInteractive(true); err != nil {
		return "", err
	}
	client := newAPIClient(s)
	list, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, "clusters")
	if err != nil {
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return "", handleAPIError(p, err)
	}
	if len(list.Items) == 0 {
		return "", fmt.Errorf("[ERROR] no clusters available")
	}
	items := make([]selector.Item, len(list.Items))
	for i, c := range list.Items {
		items[i] = selector.Item{ID: c.ID, Name: c.Name}
	}
	idx, err := clusterIDSel.Select(items)
	if err != nil {
		return "", err
	}
	if idx < 0 {
		return "", nil
	}
	if err := s.SetState("clusters", items[idx].ID); err != nil {
		return "", err
	}
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	p.Info(fmt.Sprintf("cluster context set to: %s (%s)", items[idx].Name, items[idx].ID))
	return items[idx].ID, nil
}

var clusterIDCmd = &cobra.Command{
	Use:   "id",
	Short: "Print the active cluster ID",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		if clusterIDInteractive {
			return runClusterIDInteractive(cmd, s, clusterIDSel)
		}
		id := s.GetState("clusters")
		if id == "" {
			return fmt.Errorf("[ERROR] No clusters set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
		}
		fmt.Fprintln(cmd.OutOrStdout(), id)
		return nil
	},
}

func runClusterIDInteractive(cmd *cobra.Command, s *config.Store, sel selector.Selector) error {
	if err := errCurlInteractive(true); err != nil {
		return err
	}
	client := newAPIClient(s)
	list, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, "clusters")
	if err != nil {
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return handleAPIError(p, err)
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("[ERROR] no clusters available")
	}
	items := make([]selector.Item, len(list.Items))
	for i, c := range list.Items {
		items[i] = selector.Item{ID: c.ID, Name: c.Name}
	}
	idx, err := sel.Select(items)
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}
	if err := s.SetState("clusters", items[idx].ID); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Active cluster set to: %s (%s)\n", items[idx].Name, items[idx].ID)
	return nil
}

func init() {
	// clusterCmd is not registered on root; use hf rs clusters instead.

	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterTableCmd)
	clusterCmd.AddCommand(clusterGetCmd)
	clusterCmd.AddCommand(clusterSearchCmd)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterPatchCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterCmd.AddCommand(clusterConditionsCmd)
	clusterCmd.AddCommand(clusterStatusesCmd)
	clusterCmd.AddCommand(clusterAdapterCmd)
	clusterCmd.AddCommand(clusterIDCmd)
	clusterAdapterCmd.AddCommand(clusterAdapterPostStatusCmd)

	clusterCreateCmd.Flags().StringVar(&clusterCreateName, "name", "", "cluster name (overrides template)")
	clusterCreateCmd.Flags().StringVarP(&clusterCreateFile, "file", "f", "", "JSON template file (default: <config-dir>/cluster-template.json)")
	clusterCreateCmd.Flags().IntVar(&clusterCreateReplicas, "replicas", 0, "number of replicas (overrides template)")
	clusterCreateCmd.Flags().StringVar(&clusterCreateNPID, "nodepool-id", "", "nodepool ID")

	clusterListCmd.Flags().BoolVar(&clusterListWatch, "watch", false, "continuously refresh the table (requires --output table)")
	clusterListCmd.Flags().IntVarP(&clusterListWatchSecs, "seconds", "s", 5, "refresh interval in seconds (used with --watch)")
	clusterListCmd.Flags().StringVar(&clusterListSearch, "search", "", "TSL filter expression (e.g. \"labels.environment='prod'\")")

	clusterIDCmd.Flags().BoolVarP(&clusterIDInteractive, "interactive", "i", false, "interactively select and set the active cluster")
	clusterDeleteCmd.Flags().BoolVar(&clusterDeleteForce, "force", false, "force-delete the cluster via /force-delete endpoint")
	clusterDeleteCmd.Flags().StringVar(&clusterDeleteReason, "reason", "", "reason for force-deletion (required with --force)")

	clusterStatusesCmd.Flags().BoolVar(&clusterStatusesFilter, "filter", false, "open interactive split-screen filter for adapter statuses")

	for _, c := range []*cobra.Command{
		clusterGetCmd, clusterPatchCmd, clusterDeleteCmd,
		clusterConditionsCmd, clusterStatusesCmd, clusterAdapterPostStatusCmd,
	} {
		c.Flags().BoolVarP(&clusterInteractive, "interactive", "i", false,
			"interactively select the active cluster before running this command")
	}
}
