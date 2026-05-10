// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

// clusterCmd is the top-level group for all cluster operations.
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage HyperFleet clusters",
	Long: `Manage HyperFleet clusters.

Subcommands: create, get, list, search, update, patch, delete, conditions, statuses.`,
}

// ---- flag vars ----

var (
	clusterCreateName     string
	clusterCreateReplicas int
	clusterCreateNPID     string
	clusterUpdateName     string
	clusterUpdateReplicas int
)

// ---- helpers ----

// newAPIClient builds an API client from the active config store.
// baseURL is constructed as: <api-url>/api/hyperfleet/<api-version>/
func newAPIClient(s *config.Store) *api.Client {
	apiURLVal := s.Get("hyperfleet", "api-url")
	apiVersion := s.Get("hyperfleet", "api-version")
	if apiURLVal == "" {
		apiURLVal = "http://localhost:8000"
	}
	if apiVersion == "" {
		apiVersion = "v1"
	}
	baseURL := strings.TrimRight(apiURLVal, "/") + "/api/hyperfleet/" + apiVersion + "/"
	token := s.Get("hyperfleet", "token")
	return api.NewClient(baseURL, token, verbose)
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
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		_ = p.Print(apiErr)
		return nil
	}
	return err
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
		s, err := loadConfig()
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, "clusters")
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
	},
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

		// No name arg: behave like `hf cluster get` using the state cluster-id.
		if len(args) == 0 {
			id := s.GetState("cluster-id")
			if id == "" {
				return fmt.Errorf("[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
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
		if setErr := s.SetState("cluster-id", first.ID); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist cluster-id: %v\n", setErr)
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

		name := clusterCreateName
		region := "us-east-1"
		version := "4.15.0"

		if len(args) >= 1 {
			name = args[0]
		}
		if len(args) >= 2 {
			region = args[1]
		}
		if len(args) >= 3 {
			version = args[2]
		}
		if name == "" {
			name = "my-cluster"
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		// Duplicate check.
		existing, err := api.Get[resource.ListResponse[resource.Cluster]](
			context.Background(), client,
			"clusters?search=name='"+name+"'",
		)
		if err == nil && len(existing.Items) > 0 {
			p.Warn(fmt.Sprintf("Cluster '%s' already exists, skipping creation", name))
			return nil
		}

		spec := map[string]any{
			"counter": "1",
			"region":  region,
			"version": version,
		}
		if clusterCreateReplicas > 0 {
			spec["replicas"] = strconv.Itoa(clusterCreateReplicas)
		}

		body := map[string]any{
			"kind": "Cluster",
			"name": name,
			"labels": map[string]string{
				"counter":     "1",
				"environment": "development",
				"shard":       "1",
				"team":        "core",
			},
			"spec": spec,
		}
		if clusterCreateNPID != "" {
			body["nodepool_id"] = clusterCreateNPID
		}

		cluster, err := api.Post[resource.Cluster](context.Background(), client, "clusters", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		if setErr := s.SetState("cluster-id", cluster.ID); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist cluster-id: %v\n", setErr)
		} else {
			p.Info(fmt.Sprintf("Cluster context set to '%s'", cluster.ID))
		}

		return p.Print(cluster)
	},
}

// ---- cluster update ----

var clusterUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		s, err := loadConfig()
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		body := map[string]any{}
		if clusterUpdateName != "" {
			body["name"] = clusterUpdateName
		}
		if clusterUpdateReplicas > 0 {
			if _, ok := body["spec"]; !ok {
				body["spec"] = map[string]any{}
			}
			body["spec"].(map[string]any)["replicas"] = strconv.Itoa(clusterUpdateReplicas)
		}

		cluster, err := api.Patch[resource.Cluster](context.Background(), client, "clusters/"+id, body)
		if err != nil {
			return handleAPIError(p, err)
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

		var oldVal int
		if section == "spec" {
			if v, ok := cluster.Spec["counter"].(string); ok {
				oldVal, _ = strconv.Atoi(v)
			}
		} else {
			if v, ok := cluster.Labels["counter"]; ok {
				oldVal, _ = strconv.Atoi(v)
			}
		}
		newVal := oldVal + 1

		p.Info(fmt.Sprintf("Incrementing %s.counter: %d -> %d", section, oldVal, newVal))

		var body map[string]any
		if section == "spec" {
			body = map[string]any{
				"spec": map[string]any{"counter": strconv.Itoa(newVal)},
			}
		} else {
			body = map[string]any{
				"labels": map[string]any{"counter": strconv.Itoa(newVal)},
			}
		}

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
		s, err := loadConfig()
		if err != nil {
			return err
		}
		explicit := ""
		if len(args) > 0 {
			explicit = args[0]
		}
		id, err := s.ClusterID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

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
	Args:  cobra.ExactArgs(3),
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
		clusterID, err := s.ClusterID("")
		if err != nil {
			return err
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

		result, err := api.Post[resource.AdapterStatus](context.Background(), client, "clusters/"+clusterID+"/statuses", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		p.Info(fmt.Sprintf("Posted adapter status for %s on cluster %s", adapterName, clusterID))
		return p.Print(result)
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterGetCmd)
	clusterCmd.AddCommand(clusterSearchCmd)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterUpdateCmd)
	clusterCmd.AddCommand(clusterPatchCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterCmd.AddCommand(clusterConditionsCmd)
	clusterCmd.AddCommand(clusterStatusesCmd)
	clusterCmd.AddCommand(clusterAdapterCmd)
	clusterAdapterCmd.AddCommand(clusterAdapterPostStatusCmd)

	clusterCreateCmd.Flags().StringVar(&clusterCreateName, "name", "", "cluster name (default: my-cluster)")
	clusterCreateCmd.Flags().IntVar(&clusterCreateReplicas, "replicas", 0, "number of replicas")
	clusterCreateCmd.Flags().StringVar(&clusterCreateNPID, "nodepool-id", "", "nodepool ID")

	clusterUpdateCmd.Flags().StringVar(&clusterUpdateName, "name", "", "new cluster name")
	clusterUpdateCmd.Flags().IntVar(&clusterUpdateReplicas, "replicas", 0, "new number of replicas")
}
