package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
)

const tuiForceDeleteReason = "force-deleted via hf tui"

func tuiDeleteResource(target tui.PatchTarget, force bool) (string, error) {
	s, err := loadConfig()
	if err != nil {
		return "", err
	}
	client := newAPIClient(s)
	ctx := context.Background()

	if target.IsNodePool {
		path := npBase(target.ClusterID) + "/" + target.NodePoolID
		if force {
			_, err := api.Post[resource.NodePool](ctx, client, path+"/force-delete",
				map[string]string{"reason": tuiForceDeleteReason})
			if errors.Is(err, api.ErrDryRun) {
				return "", nil
			}
			if err != nil {
				var apiErr *api.APIError
				if errors.As(err, &apiErr) && apiErr.Status == 404 {
					return "", fmt.Errorf("NodePool '%s' not found", target.NodePoolID)
				}
				return "", err
			}
			return fmt.Sprintf("[INFO] NodePool '%s' force-deleted", target.NodePoolID), nil
		}
		_, err := api.Delete[resource.NodePool](ctx, client, path)
		if errors.Is(err, api.ErrDryRun) {
			return "", nil
		}
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				return "", fmt.Errorf("NodePool '%s' not found", target.NodePoolID)
			}
			return "", err
		}
		return fmt.Sprintf("[INFO] NodePool '%s' deleted", target.NodePoolID), nil
	}

	if force {
		_, err := api.Post[resource.Cluster](ctx, client, "clusters/"+target.ClusterID+"/force-delete",
			map[string]string{"reason": tuiForceDeleteReason})
		if errors.Is(err, api.ErrDryRun) {
			return "", nil
		}
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				return "", fmt.Errorf("Cluster '%s' not found", target.ClusterID)
			}
			return "", err
		}
		return fmt.Sprintf("[INFO] Cluster '%s' force-deleted", target.ClusterID), nil
	}

	_, err = api.Delete[resource.Cluster](ctx, client, "clusters/"+target.ClusterID)
	if errors.Is(err, api.ErrDryRun) {
		return "", nil
	}
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.Status == 404 {
			return "", fmt.Errorf("Cluster '%s' not found", target.ClusterID)
		}
		return "", err
	}
	return fmt.Sprintf("[INFO] Cluster '%s' deleted", target.ClusterID), nil
}
