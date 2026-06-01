package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

// dynamicOverviewNode is one resource and its nested descendants for overview output.
type dynamicOverviewNode struct {
	Type     string                    `json:"type" yaml:"type"`
	Resource resource.GenericResource  `json:"resource" yaml:"resource"`
	Children []dynamicOverviewNode     `json:"children,omitempty" yaml:"children,omitempty"`
}

var (
	resourceOverviewWatch     bool
	resourceOverviewWatchSecs int
)

// dynamicOverviewFetch holds a partial resource tree and non-fatal load errors.
type dynamicOverviewFetch struct {
	Forest   []dynamicOverviewNode
	Warnings []string
}

func runResourceOverview(cmd *cobra.Command, _ []string) error {
	effectiveFmt := outputFmt
	if !rootCmd.PersistentFlags().Changed("output") {
		effectiveFmt = "table"
	}

	s, _ := loadConfig()
	useReconciled := s != nil && useReconciledOverview(s)

	if resourceOverviewWatch && effectiveFmt == "table" {
		if curlMode {
			return fetchAndRenderResourceOverview(cmd, effectiveFmt, 0, 0)
		}
		ctx, cancel := watchContext(context.Background())
		defer cancel()

		if useReconciled {
			var (
				cachedEntries []clusterEntry
				cachedAdCols  []string
				next          time.Time
			)
			return runWatchFast(ctx, cmd.OutOrStdout(), resourceOverviewWatchSecs, func(tick int, refresh bool) error {
				if refresh || cachedEntries == nil {
					entries, adCols, err := fetchResourceEntries(cmd)
					if err != nil {
						p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
						return handleAPIError(p, err)
					}
					cachedEntries, cachedAdCols = entries, adCols
					next = time.Now().Add(time.Duration(resourceOverviewWatchSecs) * time.Second)
				}
				treeOverview = true
				defer func() { treeOverview = false }()
				return renderResourcesTable(cmd, cachedEntries, cachedAdCols, tick, resourceOverviewWatchSecs, secsUntil(next))
			})
		}

		var (
			cached         []dynamicOverviewNode
			cachedWarnings []string
			next           time.Time
		)
		return runWatch(ctx, cmd.OutOrStdout(), resourceOverviewWatchSecs, func(tick int) error {
			if tick == 0 || time.Now().After(next) {
				fetch, err := fetchDynamicResourceOverview(cmd)
				if err != nil {
					p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
					return handleAPIError(p, err)
				}
				cached = fetch.Forest
				cachedWarnings = fetch.Warnings
				next = time.Now().Add(time.Duration(resourceOverviewWatchSecs) * time.Second)
			}
			return renderResourceOverviewTable(cmd, cached, cachedWarnings, tick, resourceOverviewWatchSecs, secsUntil(next))
		})
	}

	return fetchAndRenderResourceOverview(cmd, effectiveFmt, 0, 0)
}

func fetchAndRenderResourceOverview(cmd *cobra.Command, effectiveFmt string, tick, frequencySecs int) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	if useReconciledOverview(s) {
		return fetchAndRenderReconciledOverview(cmd, effectiveFmt, tick, frequencySecs)
	}

	fetch, err := fetchDynamicResourceOverview(cmd)
	if err != nil {
		p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return handleAPIError(p, err)
	}

	if effectiveFmt != "table" {
		p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		if len(fetch.Forest) == 0 && len(fetch.Warnings) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No resource types configured in the active environment.")
			return nil
		}
		printOverviewWarnings(cmd, fetch.Warnings)
		return p.Print(fetch.Forest)
	}

	if len(fetch.Forest) == 0 && len(fetch.Warnings) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No resource types configured in the active environment.")
		return nil
	}
	return renderResourceOverviewTable(cmd, fetch.Forest, fetch.Warnings, tick, frequencySecs, 0)
}

func fetchDynamicResourceOverview(cmd *cobra.Command) (dynamicOverviewFetch, error) {
	s, err := loadConfig()
	if err != nil {
		return dynamicOverviewFetch{}, err
	}
	types, err := s.ResourceTypes()
	if err != nil {
		return dynamicOverviewFetch{}, err
	}
	if len(types) == 0 {
		return dynamicOverviewFetch{}, nil
	}

	client := newAPIClient(s)
	ctx := context.Background()
	var forest []dynamicOverviewNode
	var warnings []string
	for _, root := range config.RootResourceTypes(types) {
		nodes, warns := fetchResourceSubtree(ctx, client, s, types, root, nil)
		warnings = append(warnings, warns...)
		forest = append(forest, nodes...)
	}
	return dynamicOverviewFetch{Forest: forest, Warnings: warnings}, nil
}

func fetchResourceSubtree(
	ctx context.Context,
	client *api.Client,
	s *config.Store,
	allTypes []config.ResourceTypeDef,
	def config.ResourceTypeDef,
	ancestorIDs map[string]string,
) ([]dynamicOverviewNode, []string) {
	path, err := s.ResolveListPath(def.Name, ancestorIDs)
	if err != nil {
		return nil, []string{overviewWarning(def.Name, path, err)}
	}
	list, err := api.Get[resource.ListResponse[resource.GenericResource]](ctx, client, path)
	if err != nil {
		return nil, []string{overviewWarning(def.Name, path, err)}
	}

	nodes := make([]dynamicOverviewNode, 0, len(list.Items))
	childTypes := config.ChildResourceTypes(allTypes, def.Name)
	var warnings []string
	for _, item := range list.Items {
		ids := cloneAncestorIDs(ancestorIDs)
		ids[def.StateKey] = item.ID()

		var children []dynamicOverviewNode
		for _, childDef := range childTypes {
			sub, warns := fetchResourceSubtree(ctx, client, s, allTypes, childDef, ids)
			warnings = append(warnings, warns...)
			children = append(children, sub...)
		}
		nodes = append(nodes, dynamicOverviewNode{
			Type:     def.Name,
			Resource: item,
			Children: children,
		})
	}
	return nodes, warnings
}

func overviewWarning(typeName, path string, err error) string {
	if path == "" {
		return fmt.Sprintf("%s: %v", typeName, err)
	}
	return fmt.Sprintf("%s (%s): %v", typeName, path, err)
}

func printOverviewWarnings(cmd *cobra.Command, warnings []string) {
	if len(warnings) == 0 {
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[WARN] %d error(s) while loading resources:\n", len(warnings))
	for _, w := range warnings {
		fmt.Fprintf(cmd.OutOrStdout(), "  • %s\n", w)
	}
	fmt.Fprintln(cmd.OutOrStdout())
}

func cloneAncestorIDs(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func renderResourceOverviewTable(cmd *cobra.Command, forest []dynamicOverviewNode, warnings []string, tick, frequencySecs, secsLeft int) error {
	if frequencySecs > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", secsLeft, output.SpinnerFrame(tick))
	}
	printOverviewWarnings(cmd, warnings)
	p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	headers := []string{"TYPE", "ID", "NAME", "KIND", "GEN"}
	var rows [][]string
	for _, node := range forest {
		appendOverviewRows(&rows, node, 0, false)
	}
	return p.PrintTable(headers, rows)
}

func appendOverviewRows(rows *[][]string, node dynamicOverviewNode, depth int, isLast bool) {
	prefix := treeLinePrefix(depth, isLast)
	id := prefix + node.Resource.ID()
	name := prefix + node.Resource.Name()
	gen := node.Resource.Generation()
	if node.Resource.DeletedTime() != "" {
		gen += " ❌"
	}
	*rows = append(*rows, []string{
		node.Type,
		id,
		name,
		node.Resource.Kind(),
		gen,
	})

	for i, child := range node.Children {
		appendOverviewRows(rows, child, depth+1, i == len(node.Children)-1)
	}
}

// treeLinePrefix builds an ASCII tree prefix for hierarchical table rows.
// depth 0 is a root row (no prefix). depth 1+ use space indent and ├─ / └─.
func treeLinePrefix(depth int, isLast bool) string {
	if depth == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		b.WriteString("   ")
	}
	if isLast {
		b.WriteString("└─ ")
	} else {
		b.WriteString("├─ ")
	}
	return b.String()
}
