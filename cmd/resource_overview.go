package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	Type     string                   `json:"type" yaml:"type"`
	Resource resource.GenericResource `json:"resource" yaml:"resource"`
	Children []dynamicOverviewNode    `json:"children,omitempty" yaml:"children,omitempty"`
}

// resourceOverviewJSON is the non-table combined overview payload when clusters and
// other configured types are both present.
type resourceOverviewJSON struct {
	Clusters  []clusterEntry        `json:"clusters,omitempty"`
	Resources []dynamicOverviewNode `json:"resources,omitempty"`
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
	useCombined := useReconciled && s != nil && hasNonReconciledOverviewRoots(s)

	if resourceOverviewWatch && effectiveFmt == "table" {
		if curlMode {
			return fetchAndRenderResourceOverview(cmd, effectiveFmt, 0, 0)
		}
		ctx, cancel := watchContext(context.Background())
		defer cancel()

		if useCombined {
			var (
				cachedEntries  []clusterEntry
				cachedAdCols   []string
				cachedForest   []dynamicOverviewNode
				cachedWarnings []string
				next           time.Time
			)
			return runWatchFast(ctx, cmd.OutOrStdout(), resourceOverviewWatchSecs, func(tick int, refresh bool) error {
				if refresh || cachedEntries == nil {
					entries, adCols, err := fetchResourceEntries(cmd)
					if err != nil {
						p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
						return handleAPIError(p, err)
					}
					fetch, err := fetchDynamicResourceOverview(cmd, "clusters")
					if err != nil {
						p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
						return handleAPIError(p, err)
					}
					cachedEntries, cachedAdCols = entries, adCols
					cachedForest, cachedWarnings = fetch.Forest, fetch.Warnings
					next = time.Now().Add(time.Duration(resourceOverviewWatchSecs) * time.Second)
				}
				treeOverview = true
				defer func() { treeOverview = false }()
				return renderCombinedResourceOverviewTable(cmd, cachedEntries, cachedAdCols, cachedForest, cachedWarnings, tick, resourceOverviewWatchSecs, secsUntil(next))
			})
		}

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
	if useReconciledOverview(s) && hasNonReconciledOverviewRoots(s) {
		return fetchAndRenderCombinedOverview(cmd, effectiveFmt, tick, frequencySecs)
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

func fetchAndRenderCombinedOverview(cmd *cobra.Command, effectiveFmt string, tick, frequencySecs int) error {
	entries, adCols, err := fetchResourceEntries(cmd)
	if err != nil {
		p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return handleAPIError(p, err)
	}

	fetch, err := fetchDynamicResourceOverview(cmd, "clusters")
	if err != nil {
		p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		return handleAPIError(p, err)
	}

	if effectiveFmt != "table" {
		p := output.NewPrinter(effectiveFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		printOverviewWarnings(cmd, fetch.Warnings)
		payload := resourceOverviewJSON{Clusters: entries, Resources: fetch.Forest}
		return p.Print(payload)
	}

	treeOverview = true
	defer func() { treeOverview = false }()
	return renderCombinedResourceOverviewTable(cmd, entries, adCols, fetch.Forest, fetch.Warnings, tick, frequencySecs, 0)
}

func fetchDynamicResourceOverview(cmd *cobra.Command, skipRootNames ...string) (dynamicOverviewFetch, error) {
	skip := make(map[string]bool, len(skipRootNames))
	for _, name := range skipRootNames {
		skip[name] = true
	}

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
		if skip[root.Name] {
			continue
		}
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
		if isOverviewListNotFound(err) {
			return nil, nil
		}
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

// isOverviewListNotFound reports whether a list GET failed because the collection
// is absent (404). Overview treats that as an empty list so optional child types
// (e.g. releases) do not spam warnings on every refresh.
func isOverviewListNotFound(err error) bool {
	var apiErr *api.APIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
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

func renderCombinedResourceOverviewTable(
	cmd *cobra.Command,
	entries []clusterEntry,
	adCols []string,
	forest []dynamicOverviewNode,
	warnings []string,
	tick, frequencySecs, secsLeft int,
) error {
	if frequencySecs > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", secsLeft, output.SpinnerFrame(tick))
		printWatchFooter(cmd, frequencySecs)
	}
	printOverviewWarnings(cmd, warnings)
	return renderUnifiedOverviewTable(cmd, entries, adCols, forest, tick, frequencySecs)
}

// unifiedOverviewHeaders returns the single overview table column list.
func unifiedOverviewHeaders(adapterCols []string) []string {
	headers := make([]string, 0, 4+len(fixedCondCols)+len(adapterCols))
	headers = append(headers, "ID", "NAME", "KIND", "GEN")
	headers = append(headers, fixedCondCols...)
	headers = append(headers, adapterCols...)
	return headers
}

// emptyReconciledCells fills condition and adapter columns for non-reconciled types.
func emptyReconciledCells(adapterCols []string) []string {
	out := make([]string, len(fixedCondCols)+len(adapterCols))
	for i := range out {
		out[i] = "-"
	}
	return out
}

// prependUnifiedKind inserts KIND before GEN in a reconciled row (ID, NAME, GEN, …).
func prependUnifiedKind(row []string, kind string) []string {
	if len(row) < 3 {
		return row
	}
	out := make([]string, 0, len(row)+1)
	out = append(out, row[0], row[1], kind, row[2])
	out = append(out, row[3:]...)
	return out
}

func renderUnifiedOverviewTable(
	cmd *cobra.Command,
	entries []clusterEntry,
	adapterCols []string,
	forest []dynamicOverviewNode,
	tick, frequencySecs int,
) error {
	p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	headers := unifiedOverviewHeaders(adapterCols)
	var rows [][]string
	appendReconciledOverviewRows(&rows, entries, adapterCols, tick, frequencySecs)
	for _, node := range forest {
		appendDynamicOverviewUnifiedRows(&rows, node, adapterCols, 0, false)
	}
	return p.PrintTable(headers, rows)
}

func appendReconciledOverviewRows(rows *[][]string, entries []clusterEntry, adapterCols []string, tick, frequencySecs int) {
	for _, e := range entries {
		clRow := buildClusterRow(e.cluster, e.adapterStatuses, adapterCols, tick, frequencySecs)
		*rows = append(*rows, prependUnifiedKind(clRow, e.cluster.Kind))
		for i, np := range e.nodepools {
			idPrefix := "  "
			namePrefix := "  "
			if treeOverview {
				idPrefix = treeLinePrefix(1, i == len(e.nodepools)-1)
				namePrefix = idPrefix
			}
			npRow := buildNodePoolRowPrefixed(np, e.npStatuses[np.ID], adapterCols, tick, frequencySecs, idPrefix, namePrefix)
			*rows = append(*rows, prependUnifiedKind(npRow, np.Kind))
		}
	}
}

func appendDynamicOverviewUnifiedRows(rows *[][]string, node dynamicOverviewNode, adapterCols []string, depth int, isLast bool) {
	prefix := treeLinePrefix(depth, isLast)
	gen := node.Resource.Generation()
	if node.Resource.DeletedTime() != "" {
		gen += " ❌"
	}
	pad := emptyReconciledCells(adapterCols)
	row := make([]string, 0, 4+len(pad))
	row = append(row, prefix+node.Resource.ID(), prefix+node.Resource.Name(), node.Resource.Kind(), gen)
	row = append(row, pad...)
	*rows = append(*rows, row)

	for i, child := range node.Children {
		appendDynamicOverviewUnifiedRows(rows, child, adapterCols, depth+1, i == len(node.Children)-1)
	}
}

func printWatchFooter(cmd *cobra.Command, frequencySecs int) {
	fmt.Fprintf(cmd.OutOrStdout(), "[watch] refreshing every %ds — press Ctrl-C to exit\n\n", frequencySecs)
}

func renderResourceOverviewTable(cmd *cobra.Command, forest []dynamicOverviewNode, warnings []string, tick, frequencySecs, secsLeft int) error {
	if frequencySecs > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "↻ %ds  %s\n", secsLeft, output.SpinnerFrame(tick))
	}
	printOverviewWarnings(cmd, warnings)
	p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
	headers := []string{"ID", "NAME", "KIND", "GEN"}
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
