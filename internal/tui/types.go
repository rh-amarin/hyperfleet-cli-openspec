package tui

import "github.com/rh-amarin/hyperfleet-cli/internal/resource"

// RowKind identifies whether a table row is a cluster or nodepool.
type RowKind int

const (
	RowCluster RowKind = iota
	RowNodePool
)

// RowMeta maps a displayed table row back to its underlying resource.
type RowMeta struct {
	Kind        RowKind
	ClusterIdx  int
	NodePoolIdx int // -1 for cluster rows
}

// ClusterEntry holds a cluster with adapter statuses and nodepools.
type ClusterEntry struct {
	Cluster         resource.Cluster
	AdapterStatuses []resource.AdapterStatus
	Nodepools       []resource.NodePool
	NPStatuses      map[string][]resource.AdapterStatus
}

// Snapshot is the full dataset for one render/refresh cycle.
type Snapshot struct {
	Headers []string
	Rows    [][]string
	Meta    []RowMeta
	Entries []ClusterEntry
}

// DetailFormat is the resource detail display mode.
type DetailFormat int

const (
	DetailJSON DetailFormat = iota
	DetailYAML
	DetailOverview
)

// EntryFetcher loads cluster/nodepool entries from the API.
type EntryFetcher func() ([]ClusterEntry, error)

// PatchTarget identifies the selected resource for a counter patch.
type PatchTarget struct {
	IsNodePool bool
	ClusterID  string
	NodePoolID string
}

// Deleter deletes the selected cluster or nodepool. When force is true, uses force-delete.
type Deleter func(target PatchTarget, force bool) (info string, err error)

// Patcher increments spec or labels counter on the target resource.
type Patcher func(target PatchTarget, section string) (info string, err error)
