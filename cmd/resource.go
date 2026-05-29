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
	Short:   "Manage config-defined HyperFleet API resources",
	Long: `Manage config-defined HyperFleet API resources.

Resource types and parent/child relationships are declared under resource-types
in the active environment file. Subcommands are registered dynamically per type.`,
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
	cmd := &cobra.Command{
		Use:   typeName,
		Short: fmt.Sprintf("Manage %s resources", typeName),
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s resources", typeName),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			explicit := ""
			if len(args) > 0 {
				explicit = args[0]
			}
			return runGenericGet(cmd, typeName, explicit)
		},
	}
	getCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")

	createCmd := &cobra.Command{
		Use:   "create [name]",
		Short: fmt.Sprintf("Create a %s resource", typeName),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := genericCreateName
			if len(args) > 0 {
				name = args[0]
			}
			return runGenericCreate(cmd, typeName, name)
		},
	}
	createCmd.Flags().StringVar(&genericCreateName, "name", "", "resource name (overrides template)")
	createCmd.Flags().StringVarP(&genericCreateFile, "file", "f", "", "JSON template file")

	searchCmd := &cobra.Command{
		Use:   "search [name]",
		Short: fmt.Sprintf("Search for a %s by name and set active context", typeName),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				return fmt.Errorf("usage: hf resource %s patch {spec|labels} [id]", typeName)
			}
			section := args[0]
			explicit := ""
			if len(args) == 2 {
				explicit = args[1]
			}
			return runGenericPatch(cmd, typeName, section, explicit)
		},
	}
	patchCmd.Flags().StringVarP(&genericPatchFile, "file", "f", "", "JSON patch body file (required)")
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
			return runGenericDelete(cmd, typeName, explicit)
		},
	}
	deleteCmd.Flags().BoolVarP(&genericInteractive, "interactive", "i", false, "interactively select a resource")

	idCmd := &cobra.Command{
		Use:   "id",
		Short: fmt.Sprintf("Print or interactively set the active %s ID", typeName),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenericID(cmd, typeName)
		},
	}
	idCmd.Flags().BoolVarP(&genericIDInteractive, "interactive", "i", false, "interactively select and set the active resource")

	cmd.AddCommand(listCmd, getCmd, createCmd, searchCmd, patchCmd, deleteCmd, idCmd)
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
	genericListSearch = ""
	genericListWatch = false
	genericListWatchSecs = 5
	genericCreateName = ""
	genericCreateFile = ""
	genericPatchFile = ""
	genericInteractive = false
	genericIDInteractive = false
}

func init() {
	rootCmd.AddCommand(resourceCmd)
	resourceCmd.AddCommand(resourceTypesCmd)
}
