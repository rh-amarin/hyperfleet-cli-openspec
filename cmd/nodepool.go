// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

// nodepoolCmd is the top-level group for all nodepool operations.
var nodepoolCmd = &cobra.Command{
	Use:   "nodepool",
	Short: "Manage HyperFleet nodepools",
	Long: `Manage HyperFleet nodepools.

Subcommands: create, get, list, search, update, patch, delete, conditions, statuses.`,
}

// ---- flag vars ----

var (
	nodepoolCreateName     string
	nodepoolCreateFile     string
	nodepoolCreateType     string
	nodepoolCreateReplicas int
	nodepoolUpdateName     string
	nodepoolUpdateReplicas int
)

// ---- helpers ----

// nodepoolType extracts the type from a nodepool's spec map.
// Checks spec["platform"]["type"] first, then spec["type"] as fallback.
func nodepoolType(np resource.NodePool) string {
	if platform, ok := np.Spec["platform"].(map[string]any); ok {
		if t, ok := platform["type"].(string); ok {
			return t
		}
	}
	if t, ok := np.Spec["type"].(string); ok {
		return t
	}
	return ""
}

// nodepoolReplicas extracts replica count from a nodepool's spec map.
func nodepoolReplicas(np resource.NodePool) string {
	if r, ok := np.Spec["replicas"]; ok {
		switch v := r.(type) {
		case float64:
			return strconv.Itoa(int(v))
		case int:
			return strconv.Itoa(v)
		case int32:
			return strconv.Itoa(int(v))
		case string:
			return v
		}
	}
	return ""
}

// nodepoolOverallStatus derives a table status from nodepool conditions.
func nodepoolOverallStatus(np resource.NodePool, nc bool) string {
	available, reconciled := "", ""
	for _, cond := range np.Status.Conditions {
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

// nodepoolConditionsView is the JSON shape emitted by `hf nodepool conditions`.
type nodepoolConditionsView struct {
	Generation int32                        `json:"generation"`
	Status     nodepoolConditionsViewStatus `json:"status"`
}

type nodepoolConditionsViewStatus struct {
	Conditions []resource.ResourceCondition `json:"conditions"`
}

// ---- nodepool list ----

var nodepoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodepools",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := api.Get[resource.ListResponse[resource.NodePool]](context.Background(), client, "nodepools")
		if err != nil {
			return handleAPIError(p, err)
		}

		if outputFmt == "table" {
			headers := []string{"ID", "NAME", "TYPE", "GEN", "REPLICAS", "STATUS"}
			rows := make([][]string, 0, len(list.Items))
			for _, np := range list.Items {
				rows = append(rows, []string{
					np.ID,
					np.Name,
					nodepoolType(np),
					strconv.Itoa(int(np.Generation)),
					nodepoolReplicas(np),
					nodepoolOverallStatus(np, noColor),
				})
			}
			return p.PrintTable(headers, rows)
		}
		return p.Print(list)
	},
}

// ---- nodepool get ----

var nodepoolGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a nodepool by ID",
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
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		np, err := api.Get[resource.NodePool](context.Background(), client, "nodepools/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		if outputFmt == "table" {
			headers := []string{"ID", "NAME", "TYPE", "GEN", "REPLICAS", "STATUS"}
			rows := [][]string{{
				np.ID,
				np.Name,
				nodepoolType(np),
				strconv.Itoa(int(np.Generation)),
				nodepoolReplicas(np),
				nodepoolOverallStatus(np, noColor),
			}}
			return p.PrintTable(headers, rows)
		}
		return p.Print(np)
	},
}

// ---- nodepool search ----

var nodepoolSearchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "Search for a nodepool by name and set as active context",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		// No name arg: behave like `hf nodepool get` using the state nodepool-id.
		if len(args) == 0 {
			id := s.GetState("nodepool-id")
			if id == "" {
				return fmt.Errorf("[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.")
			}
			client := newAPIClient(s)
			np, err := api.Get[resource.NodePool](context.Background(), client, "nodepools/"+id)
			if err != nil {
				return handleAPIError(p, err)
			}
			return p.Print(np)
		}

		name := args[0]
		client := newAPIClient(s)

		list, err := api.Get[resource.ListResponse[resource.NodePool]](
			context.Background(), client,
			"nodepools?search=name='"+name+"'",
		)
		if err != nil {
			return handleAPIError(p, err)
		}

		if len(list.Items) == 0 {
			p.Warn(fmt.Sprintf("No nodepools found matching '%s'", name))
			return p.Print([]resource.NodePool{})
		}

		if len(list.Items) > 1 {
			p.Warn(fmt.Sprintf("Multiple nodepools found matching '%s', using first result", name))
		}

		first := list.Items[0]
		if setErr := s.SetState("nodepool-id", first.ID); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist nodepool-id: %v\n", setErr)
		} else {
			p.Info(fmt.Sprintf("NodePool context set to '%s'", first.ID))
		}

		return p.Print(list.Items)
	},
}

// ---- nodepool create ----

var nodepoolCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new nodepool",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}

		body, created, err := loadTemplate(s.ConfigDir(), "nodepool", nodepoolCreateFile)
		if err != nil {
			return fmt.Errorf("[ERROR] %w", err)
		}
		if created {
			p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
			p.Info(fmt.Sprintf("Created default nodepool template at %s/nodepool-template.json", s.ConfigDir()))
		}

		// Name: positional arg > --name flag > template value.
		if nodepoolCreateName != "" {
			body["name"] = nodepoolCreateName
		}
		if len(args) >= 1 {
			body["name"] = args[0]
		}

		// Flag overrides into spec.
		spec := ensureSpecMap(body)
		if nodepoolCreateType != "" {
			platform, ok := spec["platform"].(map[string]any)
			if !ok {
				platform = map[string]any{}
				spec["platform"] = platform
			}
			platform["type"] = nodepoolCreateType
		}
		if nodepoolCreateReplicas > 0 {
			spec["replicas"] = nodepoolCreateReplicas
		}

		name, _ := body["name"].(string)

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		// Duplicate check.
		existing, err := api.Get[resource.ListResponse[resource.NodePool]](
			context.Background(), client,
			"nodepools?search=name='"+name+"'",
		)
		if err == nil && len(existing.Items) > 0 {
			p.Warn(fmt.Sprintf("NodePool '%s' already exists, skipping creation", name))
			return nil
		}

		np, err := api.Post[resource.NodePool](context.Background(), client, "nodepools", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		if setErr := s.SetState("nodepool-id", np.ID); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist nodepool-id: %v\n", setErr)
		} else {
			p.Info(fmt.Sprintf("NodePool context set to '%s'", np.ID))
		}

		return p.Print(np)
	},
}

// ---- nodepool update ----

var nodepoolUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a nodepool",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		s, err := loadConfig()
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		body := map[string]any{}
		if nodepoolUpdateName != "" {
			body["name"] = nodepoolUpdateName
		}
		if nodepoolUpdateReplicas > 0 {
			if _, ok := body["spec"]; !ok {
				body["spec"] = map[string]any{}
			}
			body["spec"].(map[string]any)["replicas"] = nodepoolUpdateReplicas
		}

		np, err := api.Patch[resource.NodePool](context.Background(), client, "nodepools/"+id, body)
		if err != nil {
			return handleAPIError(p, err)
		}
		return p.Print(np)
	},
}

// ---- nodepool patch ----

var nodepoolPatchCmd = &cobra.Command{
	Use:   "patch {spec|labels} [nodepool_id]",
	Short: "Increment a counter field in nodepool spec or labels",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != "spec" && args[0] != "labels") {
			fmt.Fprintln(cmd.OutOrStdout(), "Usage: hf nodepool patch {spec|labels} [nodepool_id]")
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "Arguments:")
			fmt.Fprintln(cmd.OutOrStdout(), "  spec|labels   Which section to increment the counter field in (required)")
			fmt.Fprintln(cmd.OutOrStdout(), "  nodepool_id   NodePool ID (default: current nodepool)")
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
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		np, err := api.Get[resource.NodePool](context.Background(), client, "nodepools/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		var oldVal int
		if section == "spec" {
			if v, ok := np.Spec["counter"].(string); ok {
				oldVal, _ = strconv.Atoi(v)
			}
		} else {
			if v, ok := np.Labels["counter"]; ok {
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

		_, err = api.Patch[resource.NodePool](context.Background(), client, "nodepools/"+id, body)
		if err != nil {
			return handleAPIError(p, err)
		}
		return nil
	},
}

// ---- nodepool delete ----

var nodepoolDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a nodepool",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		s, err := loadConfig()
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		_, err = api.Delete[resource.NodePool](context.Background(), client, "nodepools/"+id)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				return fmt.Errorf("[ERROR] NodePool '%s' not found", id)
			}
			return handleAPIError(p, err)
		}
		return nil
	},
}

// ---- nodepool conditions ----

var nodepoolConditionsCmd = &cobra.Command{
	Use:   "conditions [id]",
	Short: "Get nodepool status conditions",
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
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		np, err := api.Get[resource.NodePool](context.Background(), client, "nodepools/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		if outputFmt == "table" {
			headers := []string{"TYPE", "STATUS", "LAST TRANSITION", "REASON", "MESSAGE"}
			rows := make([][]string, 0, len(np.Status.Conditions))
			for _, cond := range np.Status.Conditions {
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

		out := nodepoolConditionsView{
			Generation: np.Generation,
			Status:     nodepoolConditionsViewStatus{Conditions: np.Status.Conditions},
		}
		return p.Print(out)
	},
}

// ---- nodepool statuses ----

var nodepoolStatusesCmd = &cobra.Command{
	Use:   "statuses [id]",
	Short: "Get nodepool adapter statuses",
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
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
			context.Background(), client, "nodepools/"+id+"/statuses",
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

// ---- nodepool adapter ----

var nodepoolAdapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Adapter operations for a nodepool",
}

var nodepoolAdapterPostStatusCmd = &cobra.Command{
	Use:   "post-status <adapter_name> <True|False|Unknown> <generation> [nodepool_id]",
	Short: "Post adapter status conditions for the current nodepool",
	Args:  cobra.RangeArgs(3, 4),
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

		explicit := ""
		if len(args) == 4 {
			explicit = args[3]
		}
		nodepoolID, err := s.NodePoolID(explicit)
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

		result, err := api.Post[resource.AdapterStatus](context.Background(), client, "clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		p.Info(fmt.Sprintf("Posted adapter status for %s on nodepool %s", adapterName, nodepoolID))
		return p.Print(result)
	},
}

func init() {
	rootCmd.AddCommand(nodepoolCmd)

	nodepoolCmd.AddCommand(nodepoolListCmd)
	nodepoolCmd.AddCommand(nodepoolGetCmd)
	nodepoolCmd.AddCommand(nodepoolSearchCmd)
	nodepoolCmd.AddCommand(nodepoolCreateCmd)
	nodepoolCmd.AddCommand(nodepoolUpdateCmd)
	nodepoolCmd.AddCommand(nodepoolPatchCmd)
	nodepoolCmd.AddCommand(nodepoolDeleteCmd)
	nodepoolCmd.AddCommand(nodepoolConditionsCmd)
	nodepoolCmd.AddCommand(nodepoolStatusesCmd)
	nodepoolCmd.AddCommand(nodepoolAdapterCmd)
	nodepoolAdapterCmd.AddCommand(nodepoolAdapterPostStatusCmd)

	nodepoolCreateCmd.Flags().StringVar(&nodepoolCreateName, "name", "", "nodepool name (overrides template)")
	nodepoolCreateCmd.Flags().StringVarP(&nodepoolCreateFile, "file", "f", "", "JSON template file (default: <config-dir>/nodepool-template.json)")
	nodepoolCreateCmd.Flags().StringVar(&nodepoolCreateType, "type", "", "instance type (overrides template)")
	nodepoolCreateCmd.Flags().IntVar(&nodepoolCreateReplicas, "replicas", 0, "number of replicas (overrides template)")

	nodepoolUpdateCmd.Flags().StringVar(&nodepoolUpdateName, "name", "", "new nodepool name")
	nodepoolUpdateCmd.Flags().IntVar(&nodepoolUpdateReplicas, "replicas", 0, "new number of replicas")
}
