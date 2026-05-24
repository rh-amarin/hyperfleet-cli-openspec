package tui

import (
	"encoding/json"

	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"gopkg.in/yaml.v3"
)

// SelectedResource returns the cluster or nodepool for a row index.
func (s Snapshot) SelectedResource(row int) (cluster *resource.Cluster, nodepool *resource.NodePool, statuses []resource.AdapterStatus, ok bool) {
	if row < 0 || row >= len(s.Meta) {
		return nil, nil, nil, false
	}
	m := s.Meta[row]
	e := s.Entries[m.ClusterIdx]
	switch m.Kind {
	case RowCluster:
		c := e.Cluster
		return &c, nil, e.AdapterStatuses, true
	case RowNodePool:
		if m.NodePoolIdx < 0 || m.NodePoolIdx >= len(e.Nodepools) {
			return nil, nil, nil, false
		}
		np := e.Nodepools[m.NodePoolIdx]
		return &e.Cluster, &np, e.NPStatuses[np.ID], true
	default:
		return nil, nil, nil, false
	}
}

// RenderDetail returns formatted detail panel content.
func RenderDetail(cluster *resource.Cluster, nodepool *resource.NodePool, statuses []resource.AdapterStatus, format DetailFormat, noColor bool) string {
	if format == DetailOverview {
		return renderDetailOverview(cluster, nodepool, statuses, noColor)
	}
	if nodepool != nil {
		return renderResourceDetail(nodepool, format, noColor)
	}
	if cluster != nil {
		return renderResourceDetail(cluster, format, noColor)
	}
	return ""
}

func renderResourceDetail(v any, format DetailFormat, noColor bool) string {
	switch format {
	case DetailYAML:
		b, err := yaml.Marshal(v)
		if err != nil {
			return err.Error()
		}
		return output.ColorizeYAMLSections(string(b), noColor)
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err.Error()
		}
		return string(output.ColorizeJSON(b, noColor))
	}
}
