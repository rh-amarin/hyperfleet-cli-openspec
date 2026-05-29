// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
	"github.com/spf13/cobra"
)

// nodepoolCmd is the top-level group for all nodepool operations.
var nodepoolCmd = &cobra.Command{
	Use:   "nodepool",
	Short: "Manage HyperFleet nodepools",
	Long: `Manage HyperFleet nodepools.

Subcommands: create, get, list, search, patch, delete, conditions, statuses.`,
}

// ---- flag vars ----

var (
	nodepoolCreateName      string
	nodepoolCreateFile      string
	nodepoolCreateType      string
	nodepoolCreateReplicas  int
	nodepoolListWatch       bool
	nodepoolListWatchSecs   int
	nodepoolListSearch      string
	nodepoolStatusesFilter  bool
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

// npBase returns the cluster-scoped nodepool collection path.
func npBase(clusterID string) string {
	return "clusters/" + clusterID + "/nodepools"
}

// requireClusterID reads cluster-id from state and returns the spec-mandated error if absent.
func requireClusterID(s interface{ GetState(string) string }) (string, error) {
	id := s.GetState("cluster-id")
	if id == "" {
		return "", fmt.Errorf("[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
	}
	return id, nil
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
		if nodepoolListWatch && outputFmt == "table" {
			if curlMode {
				return fetchAndRenderNodepoolList(cmd, 0, 0)
			}
			ctx, cancel := watchContext(context.Background())
			defer cancel()
			return runWatch(ctx, cmd.OutOrStdout(), nodepoolListWatchSecs, func(tick int) error {
				return fetchAndRenderNodepoolList(cmd, tick, nodepoolListWatchSecs)
			})
		}
		return fetchAndRenderNodepoolList(cmd, 0, 0)
	},
}

func fetchAndRenderNodepoolList(cmd *cobra.Command, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	clusterID, err := requireClusterID(s)
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	npPath := npBase(clusterID)
	if nodepoolListSearch != "" {
		npPath = npBase(clusterID) + "?search=" + url.QueryEscape(nodepoolListSearch)
	}
	list, err := api.Get[resource.ListResponse[resource.NodePool]](context.Background(), client, npPath)
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
		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}
		if nodepoolInteractive && explicit == "" {
			explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		np, err := api.Get[resource.NodePool](context.Background(), client, npBase(clusterID)+"/"+id)
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

		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}

		// No name arg: behave like `hf nodepool get` using the state nodepool-id.
		if len(args) == 0 {
			id := s.GetState("nodepool-id")
			if id == "" {
				return fmt.Errorf("[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.")
			}
			client := newAPIClient(s)
			np, err := api.Get[resource.NodePool](context.Background(), client, npBase(clusterID)+"/"+id)
			if err != nil {
				return handleAPIError(p, err)
			}
			return p.Print(np)
		}

		name := args[0]
		client := newAPIClient(s)

		list, err := api.Get[resource.ListResponse[resource.NodePool]](
			context.Background(), client,
			npBase(clusterID)+"?search=name='"+name+"'",
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

		body, err := loadTemplate("nodepool", nodepoolCreateFile)
		if err != nil {
			return fmt.Errorf("[ERROR] %w", err)
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

		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		if !curlMode {
			existing, err := api.Get[resource.ListResponse[resource.NodePool]](
				context.Background(), client,
				npBase(clusterID)+"?search=name='"+name+"'",
			)
			if err == nil && len(existing.Items) > 0 {
				p.Warn(fmt.Sprintf("NodePool '%s' already exists, skipping creation", name))
				return nil
			}
		}

		np, err := api.Post[resource.NodePool](context.Background(), client, npBase(clusterID), body)
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
		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}
		if nodepoolInteractive && explicit == "" {
			explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		np, err := api.Get[resource.NodePool](context.Background(), client, npBase(clusterID)+"/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}

		oldVal, newVal := bumpCounter(section, np.Spec, np.Labels)

		p.Info(fmt.Sprintf("Incrementing %s.counter: %d -> %d", section, oldVal, newVal))

		body := patchCounterBody(section, newVal, np.Spec, np.Labels)

		_, err = api.Patch[resource.NodePool](context.Background(), client, npBase(clusterID)+"/"+id, body)
		if err != nil {
			return handleAPIError(p, err)
		}
		return nil
	},
}

// ---- nodepool delete ----

var (
	nodepoolDeleteForce  bool
	nodepoolDeleteReason string
)

var nodepoolDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a nodepool",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		explicit := ""
		if len(args) > 0 {
			explicit = args[0]
		}
		s, err := loadConfig()
		if err != nil {
			return err
		}
		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}
		if nodepoolInteractive && explicit == "" {
			explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		if nodepoolDeleteForce {
			_, err = api.Post[resource.NodePool](context.Background(), client,
				npBase(clusterID)+"/"+id+"/force-delete",
				map[string]string{"reason": nodepoolDeleteReason},
			)
			if err != nil {
				var apiErr *api.APIError
				if errors.As(err, &apiErr) && apiErr.Status == 404 {
					return fmt.Errorf("[ERROR] NodePool '%s' not found", id)
				}
				return handleAPIError(p, err)
			}
			p.Info(fmt.Sprintf("NodePool '%s' force-deleted", id))
			return nil
		}

		_, err = api.Delete[resource.NodePool](context.Background(), client, npBase(clusterID)+"/"+id)
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

// ---- nodepool force-delete ----

var nodepoolForceDeleteReason string

var nodepoolForceDeleteCmd = &cobra.Command{
	Use:   "force-delete [id]",
	Short: "Permanently remove a nodepool that is stuck in Finalizing state",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if nodepoolForceDeleteReason == "" {
			return fmt.Errorf("[ERROR] --reason is required")
		}
		id := ""
		if len(args) > 0 {
			id = args[0]
		}
		if id == "" && !nodepoolInteractive {
			return fmt.Errorf("[ERROR] nodepool ID required. Pass an explicit ID or use -i to select interactively.")
		}
		s, err := loadConfig()
		if err != nil {
			return err
		}
		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}
		if nodepoolInteractive && id == "" {
			id, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || id == "" {
				return err
			}
		}
		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		_, err = api.Post[resource.NodePool](context.Background(), client,
			npBase(clusterID)+"/"+id+"/force-delete",
			map[string]string{"reason": nodepoolForceDeleteReason},
		)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				return fmt.Errorf("[ERROR] NodePool '%s' not found", id)
			}
			return handleAPIError(p, err)
		}
		p.Info(fmt.Sprintf("NodePool '%s' force-deleted", id))
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
		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}
		if nodepoolInteractive && explicit == "" {
			explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		np, err := api.Get[resource.NodePool](context.Background(), client, npBase(clusterID)+"/"+id)
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
		clusterID, err := requireClusterID(s)
		if err != nil {
			return err
		}
		if nodepoolInteractive && explicit == "" {
			explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || explicit == "" {
				return err
			}
		}
		id, err := s.NodePoolID(explicit)
		if err != nil {
			return err
		}

		client := newAPIClient(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
			context.Background(), client, npBase(clusterID)+"/"+id+"/statuses",
		)
		if err != nil {
			return handleAPIError(p, err)
		}

		if nodepoolStatusesFilter {
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
		if nodepoolInteractive && explicit == "" {
			explicit, err = pickNodepoolInteractive(cmd, s, clusterID)
			if err != nil || explicit == "" {
				return err
			}
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

		result, err := api.Put[resource.AdapterStatus](context.Background(), client, "clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses", body)
		if err != nil {
			return handleAPIError(p, err)
		}

		if result.Adapter == "" {
			p.Info(fmt.Sprintf("Posted adapter status for %s on nodepool %s (no-op: status unchanged)", adapterName, nodepoolID))
			return nil
		}
		p.Info(fmt.Sprintf("Posted adapter status for %s on nodepool %s", adapterName, nodepoolID))
		return p.Print(result)
	},
}

// ---- nodepool table ----

var nodepoolTableCmd = &cobra.Command{
	Use:   "table",
	Short: "List all nodepools in table format (alias for: nodepool list --output table)",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFmt = "table"
		return fetchAndRenderNodepoolList(cmd, 0, 0)
	},
}

// ---- nodepool id ----

var nodepoolInteractive bool
var nodepoolIDInteractive bool

// nodepoolIDSel is the selector used by hf nodepool id -i and pickNodepoolInteractive; swapped in tests.
var nodepoolIDSel selector.Selector = selector.FuzzySelector{}

func pickNodepoolInteractive(cmd *cobra.Command, s *config.Store, clusterID string) (string, error) {
	if err := errCurlInteractive(true); err != nil {
		return "", err
	}
	client := newAPIClient(s)
	list, err := api.Get[resource.ListResponse[resource.NodePool]](context.Background(), client, npBase(clusterID))
	if err != nil {
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return "", handleAPIError(p, err)
	}
	if len(list.Items) == 0 {
		return "", fmt.Errorf("[ERROR] no nodepools available for cluster %s", clusterID)
	}
	items := make([]selector.Item, len(list.Items))
	for i, np := range list.Items {
		items[i] = selector.Item{ID: np.ID, Name: np.Name}
	}
	idx, err := nodepoolIDSel.Select(items)
	if err != nil {
		return "", err
	}
	if idx < 0 {
		return "", nil
	}
	if err := s.SetState("nodepool-id", items[idx].ID); err != nil {
		return "", err
	}
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	p.Info(fmt.Sprintf("nodepool context set to: %s (%s)", items[idx].Name, items[idx].ID))
	return items[idx].ID, nil
}

var nodepoolIDCmd = &cobra.Command{
	Use:   "id",
	Short: "Print the active nodepool ID",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		if nodepoolIDInteractive {
			return runNodepoolIDInteractive(cmd, s, nodepoolIDSel)
		}
		id := s.GetState("nodepool-id")
		if id == "" {
			return fmt.Errorf("[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.")
		}
		fmt.Fprintln(cmd.OutOrStdout(), id)
		return nil
	},
}

func runNodepoolIDInteractive(cmd *cobra.Command, s *config.Store, sel selector.Selector) error {
	if err := errCurlInteractive(true); err != nil {
		return err
	}
	clusterID, err := requireClusterID(s)
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	list, err := api.Get[resource.ListResponse[resource.NodePool]](context.Background(), client, npBase(clusterID))
	if err != nil {
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return handleAPIError(p, err)
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("[ERROR] no nodepools available for cluster %s", clusterID)
	}
	items := make([]selector.Item, len(list.Items))
	for i, np := range list.Items {
		items[i] = selector.Item{ID: np.ID, Name: np.Name}
	}
	idx, err := sel.Select(items)
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}
	if err := s.SetState("nodepool-id", items[idx].ID); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Active nodepool set to: %s (%s)\n", items[idx].Name, items[idx].ID)
	return nil
}

func init() {
	rootCmd.AddCommand(nodepoolCmd)

	nodepoolCmd.AddCommand(nodepoolListCmd)
	nodepoolCmd.AddCommand(nodepoolTableCmd)
	nodepoolCmd.AddCommand(nodepoolGetCmd)
	nodepoolCmd.AddCommand(nodepoolSearchCmd)
	nodepoolCmd.AddCommand(nodepoolCreateCmd)
	nodepoolCmd.AddCommand(nodepoolPatchCmd)
	nodepoolCmd.AddCommand(nodepoolDeleteCmd)
	nodepoolCmd.AddCommand(nodepoolForceDeleteCmd)
	nodepoolCmd.AddCommand(nodepoolConditionsCmd)
	nodepoolCmd.AddCommand(nodepoolStatusesCmd)
	nodepoolCmd.AddCommand(nodepoolAdapterCmd)
	nodepoolCmd.AddCommand(nodepoolIDCmd)
	nodepoolAdapterCmd.AddCommand(nodepoolAdapterPostStatusCmd)

	nodepoolCreateCmd.Flags().StringVar(&nodepoolCreateName, "name", "", "nodepool name (overrides template)")
	nodepoolCreateCmd.Flags().StringVarP(&nodepoolCreateFile, "file", "f", "", "JSON template file (default: <config-dir>/nodepool-template.json)")
	nodepoolCreateCmd.Flags().StringVar(&nodepoolCreateType, "type", "", "instance type (overrides template)")
	nodepoolCreateCmd.Flags().IntVar(&nodepoolCreateReplicas, "replicas", 0, "number of replicas (overrides template)")

	nodepoolListCmd.Flags().BoolVar(&nodepoolListWatch, "watch", false, "continuously refresh the table (requires --output table)")
	nodepoolListCmd.Flags().IntVarP(&nodepoolListWatchSecs, "seconds", "s", 5, "refresh interval in seconds (used with --watch)")
	nodepoolListCmd.Flags().StringVar(&nodepoolListSearch, "search", "", "TSL filter expression (e.g. \"labels.role='worker'\")")

	nodepoolIDCmd.Flags().BoolVarP(&nodepoolIDInteractive, "interactive", "i", false, "interactively select and set the active nodepool")

	nodepoolForceDeleteCmd.Flags().StringVar(&nodepoolForceDeleteReason, "reason", "", "reason for force-deleting the nodepool (required)")

	nodepoolDeleteCmd.Flags().BoolVar(&nodepoolDeleteForce, "force", false, "force-delete the nodepool via POST .../force-delete")
	nodepoolDeleteCmd.Flags().StringVar(&nodepoolDeleteReason, "reason", "", "reason for force-deleting (used with --force)")

	nodepoolStatusesCmd.Flags().BoolVar(&nodepoolStatusesFilter, "filter", false, "open interactive split-screen filter for adapter statuses")

	for _, c := range []*cobra.Command{
		nodepoolGetCmd, nodepoolPatchCmd, nodepoolDeleteCmd, nodepoolForceDeleteCmd,
		nodepoolConditionsCmd, nodepoolStatusesCmd, nodepoolAdapterPostStatusCmd,
	} {
		c.Flags().BoolVarP(&nodepoolInteractive, "interactive", "i", false,
			"interactively select the active nodepool before running this command")
	}
}
