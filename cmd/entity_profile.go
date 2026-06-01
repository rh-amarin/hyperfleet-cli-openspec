package cmd

import (
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
)

// reconciledProfile returns "cluster", "nodepool", or "" for bundled reconciled entities.
func reconciledProfile(typeName string) string {
	switch typeName {
	case "clusters":
		return "cluster"
	case "nodepools":
		return "nodepool"
	default:
		return ""
	}
}

// useReconciledOverview reports whether hf rs should render the adapter-rich cluster+nodepool table.
func useReconciledOverview(s *config.Store) bool {
	types, err := s.ResourceTypes()
	if err != nil {
		return false
	}
	var hasClusters, hasNodepools bool
	for _, d := range types {
		if d.Name == "clusters" && d.Path == "clusters" {
			hasClusters = true
		}
		if d.Name == "nodepools" && d.Parent == "clusters" {
			hasNodepools = true
		}
	}
	return hasClusters && hasNodepools
}
