package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:     "resource",
	Aliases: []string{"rs"},
	Short:   "Overview and manage config-defined HyperFleet API resources",
	Long: `Manage config-defined HyperFleet API resources.

Resource types and parent/child relationships are declared under resource-types
in the active environment file. Subcommands are registered dynamically per type.

Run hf rs (or hf resource) with no subcommand for a hierarchical overview of all
configured types, similar to hf table for clusters and nodepools.`,
	RunE: runResourceOverview,
}

var resourceTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List configured resource types and active state keys",
	Args:  cobra.NoArgs,
	RunE:  runResourceTypes,
}

var (
	registeredResourceTypeCmd = map[string]bool{}

	genericListSearch    string
	genericListWatch     bool
	genericListWatchSecs int
	genericCreateName    string
	genericCreateFile    string
	genericPatchFile     string
	genericInteractive   bool
	genericIDInteractive bool
)

// genericSel is swapped in tests.
var genericSel selector.Selector = selector.FuzzySelector{}

func registerResourceTypesFromStore(s *config.Store) error {
	types, err := s.ResourceTypes()
	if err != nil {
		return err
	}
	for _, def := range types {
		if registeredResourceTypeCmd[def.Name] {
			continue
		}
		resourceCmd.AddCommand(newResourceTypeCmd(def))
		registeredResourceTypeCmd[def.Name] = true
	}
	return nil
}

// preloadResourceCommands registers dynamic resource subcommands before Cobra parses args.
func preloadResourceCommands(args []string) error {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "resource", "rs":
	default:
		return nil
	}
	s := config.NewFromEnv()
	if err := s.Load(); err != nil {
		return err
	}
	if _, err := s.RequireActiveEnvironment(); err != nil {
		return nil
	}
	return registerResourceTypesFromStore(s)
}

func resetResourceRegistrationForTest() {
	registeredResourceTypeCmd = map[string]bool{}
	for _, c := range resourceCmd.Commands() {
		if c.Name() != "types" && c.Name() != "help" {
			resourceCmd.RemoveCommand(c)
		}
	}
}

func newResourceTypeCmd(def config.ResourceTypeDef) *cobra.Command {
	typeName := def.Name
	profile := reconciledProfile(typeName)
	cmd := &cobra.Command{
		Use:   typeName,
		Short: fmt.Sprintf("Manage %s resources", typeName),
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s resources", typeName),
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile != "" {
				return runReconciledList(cmd, typeName)
			}
			return runGenericList(cmd, typeName)
		},
	}
	listCmd.Flags().StringVar(&genericListSearch, "search", "", "filter list by search query")
	listCmd.Flags().BoolVar(&genericListWatch, "watch", false, "continuously refresh the table (requires --output table)")
	listCmd.Flags().IntVarP(&genericListWatchSecs, "seconds", "s", 5, "refresh interval in seconds (used with --watch)")

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: fmt.Sprintf("Get a %s resource by ID", typeName),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile == "cluster" {
				return runReconciledClusterGet(cmd, args)
			}
			if profile == "nodepool" {
				return runReconciledNodepoolGet(cmd, args)
			}
			explicit := ""
			if len(args) > 0 {
				explicit = args[0]
			}
			return runGenericGet(cmd, typeName, explicit)
		},
	}
	getCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")

	createArgs := cobra.MaximumNArgs(1)
	createUse := "create [name]"
	if profile == "cluster" {
		createArgs = cobra.MaximumNArgs(3)
		createUse = "create [name] [region] [version]"
	}
	createCmd := &cobra.Command{
		Use:   createUse,
		Short: fmt.Sprintf("Create a %s resource", typeName),
		Args:  createArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile == "cluster" {
				return runReconciledClusterCreate(cmd, args)
			}
			if profile == "nodepool" {
				return runReconciledNodepoolCreate(cmd, args)
			}
			name := genericCreateName
			if len(args) > 0 {
				name = args[0]
			}
			return runGenericCreate(cmd, typeName, name)
		},
	}
	createCmd.Flags().StringVar(&genericCreateName, "name", "", "resource name (overrides template)")
	createCmd.Flags().StringVarP(&genericCreateFile, "file", "f", "", "JSON template file")
	if profile == "cluster" {
		createCmd.Flags().IntVar(&genericCreateReplicas, "replicas", 0, "number of replicas (overrides template)")
		createCmd.Flags().StringVar(&genericCreateNodepoolID, "nodepool-id", "", "nodepool ID")
	}
	if profile == "nodepool" {
		createCmd.Flags().StringVar(&genericCreatePlatformType, "type", "", "platform type (overrides template)")
		createCmd.Flags().IntVar(&genericCreateReplicas, "replicas", 0, "number of replicas (overrides template)")
	}

	searchCmd := &cobra.Command{
		Use:   "search [name]",
		Short: fmt.Sprintf("Search for a %s by name and set active context", typeName),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile == "cluster" {
				return runReconciledClusterSearch(cmd, args)
			}
			if profile == "nodepool" {
				return runReconciledNodepoolSearch(cmd, args)
			}
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runGenericSearch(cmd, typeName, name)
		},
	}

	patchCmd := &cobra.Command{
		Use:   "patch {spec|labels} [id]",
		Short: fmt.Sprintf("Patch a %s resource", typeName),
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || (args[0] != "spec" && args[0] != "labels") {
				return fmt.Errorf("usage: hf rs %s patch {spec|labels} [id]", typeName)
			}
			section := args[0]
			explicit := ""
			if len(args) == 2 {
				explicit = args[1]
			}
			if profile == "cluster" {
				return runReconciledClusterPatch(cmd, section, explicit)
			}
			if profile == "nodepool" {
				return runReconciledNodepoolPatch(cmd, section, explicit)
			}
			return runGenericPatch(cmd, typeName, section, explicit)
		},
	}
	patchCmd.Flags().StringVarP(&genericPatchFile, "file", "f", "", "JSON patch body file (omit to increment counter on reconciled types)")
	patchCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")

	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: fmt.Sprintf("Delete a %s resource", typeName),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := ""
			if len(args) > 0 {
				explicit = args[0]
			}
			if profile == "cluster" {
				return runReconciledClusterDelete(cmd, explicit)
			}
			if profile == "nodepool" {
				return runReconciledNodepoolDelete(cmd, explicit)
			}
			return runGenericDelete(cmd, typeName, explicit)
		},
	}
	deleteCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")
	if profile == "cluster" {
		deleteCmd.Flags().BoolVar(&genericReconciledDeleteForce, "force", false, "force-delete the cluster via /force-delete endpoint")
		deleteCmd.Flags().StringVar(&genericReconciledDeleteReason, "reason", "", "reason for force-deletion (required with --force)")
	}

	idCmd := &cobra.Command{
		Use:   "id",
		Short: fmt.Sprintf("Print or interactively set the active %s ID", typeName),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile == "cluster" {
				return runReconciledClusterID(cmd)
			}
			if profile == "nodepool" {
				return runReconciledNodepoolID(cmd)
			}
			return runGenericID(cmd, typeName)
		},
	}
	idCmd.Flags().BoolVarP(&genericIDInteractive, "interactive", "i", false, "interactively select and set the active resource")

	adapterReportCmd := &cobra.Command{
		Use:   "adapter-report <adapter_name> <True|False|Unknown> <generation> [id]",
		Short: fmt.Sprintf("Report adapter status for a %s resource (simulate adapter reporting)", typeName),
		Args:  cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenericAdapterReport(cmd, typeName, args)
		},
	}
	adapterReportCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")

	tableCmd := &cobra.Command{
		Use:   "table",
		Short: fmt.Sprintf("List %s resources in table format", typeName),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if profile != "" {
				return runReconciledTable(cmd, typeName)
			}
			prev := outputFmt
			outputFmt = "table"
			defer func() { outputFmt = prev }()
			return runGenericList(cmd, typeName)
		},
	}

	var subcmds []*cobra.Command
	subcmds = append(subcmds, listCmd, tableCmd, getCmd, createCmd, searchCmd, patchCmd, deleteCmd, idCmd, adapterReportCmd)

	if profile != "" {
		conditionsCmd := &cobra.Command{
			Use:   "conditions [id]",
			Short: fmt.Sprintf("Get %s status conditions", typeName),
			Args:  cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				explicit := ""
				if len(args) > 0 {
					explicit = args[0]
				}
				if profile == "cluster" {
					return runReconciledClusterConditions(cmd, explicit)
				}
				return runReconciledNodepoolConditions(cmd, explicit)
			},
		}
		conditionsCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")

		statusesCmd := &cobra.Command{
			Use:   "statuses [id]",
			Short: fmt.Sprintf("Get %s adapter statuses", typeName),
			Args:  cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				explicit := ""
				if len(args) > 0 {
					explicit = args[0]
				}
				if profile == "cluster" {
					return runReconciledClusterStatuses(cmd, explicit)
				}
				return runReconciledNodepoolStatuses(cmd, explicit)
			},
		}
		statusesCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")
		statusesCmd.Flags().BoolVar(&genericStatusesFilter, "filter", false, "open interactive status filter UI")

		subcmds = append(subcmds, conditionsCmd, statusesCmd)
	}

	if profile == "nodepool" {
		forceDeleteCmd := &cobra.Command{
			Use:   "force-delete [id]",
			Short: "Force-delete a nodepool",
			Args:  cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				explicit := ""
				if len(args) > 0 {
					explicit = args[0]
				}
				return runReconciledNodepoolForceDelete(cmd, explicit)
			},
		}
		forceDeleteCmd.Flags().StringVar(&genericReconciledForceReason, "reason", "", "reason for force-deleting the nodepool (required)")
		forceDeleteCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")
		subcmds = append(subcmds, forceDeleteCmd)
	}

	cmd.AddCommand(subcmds...)
	return cmd
}

func runResourceTypes(cmd *cobra.Command, args []string) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	types, err := s.ResourceTypes()
	if err != nil {
		return err
	}
	if len(types) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No resource types configured in the active environment.")
		return nil
	}

	sortResourceTypes(types)
	for _, def := range types {
		stateVal := s.GetState(def.StateKey)
		stateDisplay := "<not set>"
		if stateVal != "" {
			stateDisplay = stateVal
		}
		indent := ""
		if def.Parent != "" {
			indent = "  └─ "
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s  path: %s  state: %s=%s\n", indent, def.Name, def.Path, def.StateKey, stateDisplay)
		if def.Parent != "" {
			parent, _ := s.ResourceType(def.Parent)
			fmt.Fprintf(cmd.OutOrStdout(), "     requires: %s\n", parent.StateKey)
		}
	}
	return nil
}

func sortResourceTypes(types []config.ResourceTypeDef) {
	for i := 0; i < len(types); i++ {
		for j := i + 1; j < len(types); j++ {
			if types[j].Name < types[i].Name {
				types[i], types[j] = types[j], types[i]
			}
		}
	}
}

func runGenericList(cmd *cobra.Command, typeName string) error {
	if genericListWatch && outputFmt == "table" {
		if curlMode {
			return fetchAndRenderGenericList(cmd, typeName, 0, 0)
		}
		ctx, cancel := watchContext(context.Background())
		defer cancel()
		return runWatch(ctx, cmd.OutOrStdout(), genericListWatchSecs, func(tick int) error {
			return fetchAndRenderGenericList(cmd, typeName, tick, genericListWatchSecs)
		})
	}
	return fetchAndRenderGenericList(cmd, typeName, 0, 0)
}

func fetchAndRenderGenericList(cmd *cobra.Command, typeName string, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	path := basePath
	if genericListSearch != "" {
		path = basePath + "?search=" + url.QueryEscape(genericListSearch)
	}
	list, err := api.Get[resource.ListResponse[resource.GenericResource]](context.Background(), client, path)
	if err != nil {
		return handleAPIError(p, err)
	}

	if outputFmt == "table" {
		headers := []string{"ID", "NAME", "KIND", "GEN"}
		rows := make([][]string, 0, len(list.Items))
		for _, item := range list.Items {
			rows = append(rows, []string{
				item.ID(),
				item.Name(),
				item.Kind(),
				item.Generation(),
			})
		}
		return p.PrintTable(headers, rows)
	}
	return p.Print(list)
}

func runGenericGet(cmd *cobra.Command, typeName, explicit string) error {
	if err := errCurlInteractive(genericInteractive); err != nil {
		return err
	}
	s, err := loadConfig()
	if err != nil {
		return err
	}
	if genericInteractive && explicit == "" {
		explicit, err = pickGenericInteractive(cmd, s, typeName)
		if err != nil || explicit == "" {
			return err
		}
	}
	id, err := s.ResourceID(typeName, explicit)
	if err != nil {
		return err
	}
	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return err
	}

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	item, err := api.Get[resource.GenericResource](context.Background(), client, basePath+"/"+id)
	if err != nil {
		return handleAPIError(p, err)
	}

	if outputFmt == "table" {
		headers := []string{"ID", "NAME", "KIND", "GEN"}
		rows := [][]string{{
			item.ID(),
			item.Name(),
			item.Kind(),
			item.Generation(),
		}}
		return p.PrintTable(headers, rows)
	}
	return p.Print(item)
}

func runGenericSearch(cmd *cobra.Command, typeName, name string) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	def, err := s.ResourceType(typeName)
	if err != nil {
		return err
	}

	if name == "" {
		id := s.GetState(def.StateKey)
		if id == "" {
			return fmt.Errorf("[ERROR] No %s set in state. Run 'hf resource %s search <name>' first.", def.StateKey, typeName)
		}
		basePath, err := s.ResolveResourcePath(typeName)
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		item, err := api.Get[resource.GenericResource](context.Background(), client, basePath+"/"+id)
		if err != nil {
			return handleAPIError(p, err)
		}
		return p.Print(item)
	}

	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return err
	}
	client := newAPIClient(s)
	list, err := api.Get[resource.ListResponse[resource.GenericResource]](
		context.Background(), client,
		basePath+"?search=name='"+name+"'",
	)
	if err != nil {
		return handleAPIError(p, err)
	}
	if len(list.Items) == 0 {
		p.Warn(fmt.Sprintf("No %s found matching '%s'", typeName, name))
		return p.Print([]resource.GenericResource{})
	}
	if len(list.Items) > 1 {
		p.Warn(fmt.Sprintf("Multiple %s found matching '%s', using first result", typeName, name))
	}
	first := list.Items[0]
	if setErr := s.SetState(def.StateKey, first.ID()); setErr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist %s: %v\n", def.StateKey, setErr)
	} else {
		p.Info(fmt.Sprintf("%s context set to '%s'", typeName, first.ID()))
	}
	return p.Print(list.Items)
}

func runGenericCreate(cmd *cobra.Command, typeName, name string) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	def, err := s.ResourceType(typeName)
	if err != nil {
		return err
	}
	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return err
	}

	body, err := loadResourceTemplate(s, def.CreateTemplate, genericCreateFile)
	if err != nil {
		return fmt.Errorf("[ERROR] %w", err)
	}
	if name != "" {
		body["name"] = name
	}

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	created, err := api.Post[resource.GenericResource](context.Background(), client, basePath, body)
	if err != nil {
		return handleAPIError(p, err)
	}
	if created.ID() != "" {
		if setErr := s.SetState(def.StateKey, created.ID()); setErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist %s: %v\n", def.StateKey, setErr)
		} else {
			p.Info(fmt.Sprintf("%s context set to '%s'", typeName, created.ID()))
		}
	}
	return p.Print(created)
}

func runGenericPatch(cmd *cobra.Command, typeName, section, explicit string) error {
	if genericPatchFile == "" {
		return fmt.Errorf("[ERROR] --file is required for patch")
	}
	if err := errCurlInteractive(genericInteractive); err != nil {
		return err
	}
	raw, err := os.ReadFile(genericPatchFile)
	if err != nil {
		return fmt.Errorf("[ERROR] loading patch file: %w", err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return fmt.Errorf("[ERROR] loading patch file: %w", err)
	}

	s, err := loadConfig()
	if err != nil {
		return err
	}
	if genericInteractive && explicit == "" {
		explicit, err = pickGenericInteractive(cmd, s, typeName)
		if err != nil || explicit == "" {
			return err
		}
	}
	id, err := s.ResourceID(typeName, explicit)
	if err != nil {
		return err
	}
	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return err
	}

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	_, err = api.Patch[resource.GenericResource](context.Background(), client, basePath+"/"+id+"/"+section, body)
	if err != nil {
		return handleAPIError(p, err)
	}
	return nil
}

func runGenericDelete(cmd *cobra.Command, typeName, explicit string) error {
	if err := errCurlInteractive(genericInteractive); err != nil {
		return err
	}
	s, err := loadConfig()
	if err != nil {
		return err
	}
	if genericInteractive && explicit == "" {
		explicit, err = pickGenericInteractive(cmd, s, typeName)
		if err != nil || explicit == "" {
			return err
		}
	}
	id, err := s.ResourceID(typeName, explicit)
	if err != nil {
		return err
	}
	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return err
	}

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	_, err = api.Delete[resource.GenericResource](context.Background(), client, basePath+"/"+id)
	if err != nil {
		return handleAPIError(p, err)
	}
	p.Info(fmt.Sprintf("%s '%s' deleted", typeName, id))
	return nil
}

func runGenericID(cmd *cobra.Command, typeName string) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	def, err := s.ResourceType(typeName)
	if err != nil {
		return err
	}
	if genericIDInteractive {
		if err := errCurlInteractive(true); err != nil {
			return err
		}
		basePath, err := s.ResolveResourcePath(typeName)
		if err != nil {
			return err
		}
		client := newAPIClient(s)
		list, err := api.Get[resource.ListResponse[resource.GenericResource]](context.Background(), client, basePath)
		if err != nil {
			p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
			return handleAPIError(p, err)
		}
		if len(list.Items) == 0 {
			return fmt.Errorf("[ERROR] no %s resources available", typeName)
		}
		items := make([]selector.Item, len(list.Items))
		for i, item := range list.Items {
			items[i] = selector.Item{ID: item.ID(), Name: item.Name()}
		}
		idx, err := genericSel.Select(items)
		if err != nil {
			return err
		}
		if idx < 0 {
			return nil
		}
		if err := s.SetState(def.StateKey, items[idx].ID); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Active %s set to: %s (%s)\n", typeName, items[idx].Name, items[idx].ID)
		return nil
	}
	id := s.GetState(def.StateKey)
	if id == "" {
		return fmt.Errorf("[ERROR] No %s set in state. Run 'hf resource %s search <name>' first.", def.StateKey, typeName)
	}
	fmt.Fprintln(cmd.OutOrStdout(), id)
	return nil
}

func pickGenericInteractive(cmd *cobra.Command, s *config.Store, typeName string) (string, error) {
	basePath, err := s.ResolveResourcePath(typeName)
	if err != nil {
		return "", err
	}
	client := newAPIClient(s)
	list, err := api.Get[resource.ListResponse[resource.GenericResource]](context.Background(), client, basePath)
	if err != nil {
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return "", handleAPIError(p, err)
	}
	if len(list.Items) == 0 {
		return "", fmt.Errorf("[ERROR] no %s resources available", typeName)
	}
	items := make([]selector.Item, len(list.Items))
	for i, item := range list.Items {
		items[i] = selector.Item{ID: item.ID(), Name: item.Name()}
	}
	idx, err := genericSel.Select(items)
	if err != nil {
		return "", err
	}
	if idx < 0 {
		return "", nil
	}
	def, _ := s.ResourceType(typeName)
	if err := s.SetState(def.StateKey, items[idx].ID); err != nil {
		return "", err
	}
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	p.Info(fmt.Sprintf("%s context set to: %s (%s)", typeName, items[idx].Name, items[idx].ID))
	return items[idx].ID, nil
}

func resetGenericFlags() {
	resourceOverviewWatch = false
	resourceOverviewWatchSecs = 5
	genericListSearch = ""
	genericListWatch = false
	genericListWatchSecs = 5
	genericCreateName = ""
	genericCreateFile = ""
	genericPatchFile = ""
	genericInteractive = false
	genericIDInteractive = false
	genericReconciledDeleteForce = false
	genericReconciledDeleteReason = ""
	genericReconciledForceReason = ""
	genericCreateReplicas = 0
	genericCreateNodepoolID = ""
	genericCreatePlatformType = ""
	genericStatusesFilter = false
}

func init() {
	rootCmd.AddCommand(resourceCmd)
	resourceCmd.AddCommand(resourceTypesCmd)

	resourceCmd.Flags().BoolVar(&resourceOverviewWatch, "watch", false, "continuously refresh the overview table")
	resourceCmd.Flags().IntVarP(&resourceOverviewWatchSecs, "seconds", "s", 5, "refresh interval in seconds (used with --watch)")
}
