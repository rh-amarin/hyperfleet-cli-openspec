package cmd

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
)

func tuiPatchResource(target tui.PatchTarget, section string) (string, error) {
	if section != "spec" && section != "labels" {
		return "", fmt.Errorf("invalid patch section %q", section)
	}

	s, err := loadConfig()
	if err != nil {
		return "", err
	}
	client := newAPIClient(s)

	if target.IsNodePool {
		return patchNodePoolCounter(context.Background(), client, target.ClusterID, target.NodePoolID, section)
	}
	return patchClusterCounter(context.Background(), client, target.ClusterID, section)
}

func patchClusterCounter(ctx context.Context, client *api.Client, id, section string) (string, error) {
	cluster, err := api.Get[resource.Cluster](ctx, client, "clusters/"+id)
	if err != nil {
		if errors.Is(err, api.ErrDryRun) {
			return "", nil
		}
		return "", err
	}
	oldVal, newVal := bumpCounter(section, cluster.Spec, cluster.Labels)
	body := patchCounterBody(section, newVal, cluster.Spec, cluster.Labels)
	if _, err := api.Patch[resource.Cluster](ctx, client, "clusters/"+id, body); err != nil {
		return "", err
	}
	return fmt.Sprintf("[INFO] Incrementing %s.counter: %d -> %d", section, oldVal, newVal), nil
}

func patchNodePoolCounter(ctx context.Context, client *api.Client, clusterID, npID, section string) (string, error) {
	np, err := api.Get[resource.NodePool](ctx, client, npBase(clusterID)+"/"+npID)
	if err != nil {
		if errors.Is(err, api.ErrDryRun) {
			return "", nil
		}
		return "", err
	}
	oldVal, newVal := bumpCounter(section, np.Spec, np.Labels)
	body := patchCounterBody(section, newVal, np.Spec, np.Labels)
	if _, err := api.Patch[resource.NodePool](ctx, client, npBase(clusterID)+"/"+npID, body); err != nil {
		return "", err
	}
	return fmt.Sprintf("[INFO] Incrementing %s.counter: %d -> %d", section, oldVal, newVal), nil
}

func bumpCounter(section string, spec map[string]any, labels map[string]string) (oldVal, newVal int) {
	if section == "spec" {
		if v, ok := spec["counter"].(string); ok {
			oldVal, _ = strconv.Atoi(v)
		}
	} else if v, ok := labels["counter"]; ok {
		oldVal, _ = strconv.Atoi(v)
	}
	newVal = oldVal + 1
	return oldVal, newVal
}

func patchCounterBody(section string, newVal int, spec map[string]any, labels map[string]string) map[string]any {
	counter := strconv.Itoa(newVal)
	if section == "spec" {
		merged := maps.Clone(spec)
		if merged == nil {
			merged = map[string]any{}
		}
		merged["counter"] = counter
		return map[string]any{"spec": merged}
	}
	merged := make(map[string]any, len(labels)+1)
	for k, v := range labels {
		merged[k] = v
	}
	merged["counter"] = counter
	return map[string]any{"labels": merged}
}
