package tui

import (
	"context"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// NewAPIFetcher returns an EntryFetcher backed by the HyperFleet API client.
func NewAPIFetcher(client *api.Client) EntryFetcher {
	return func() ([]ClusterEntry, error) {
		clusterList, err := api.Get[resource.ListResponse[resource.Cluster]](context.Background(), client, "clusters")
		if err != nil {
			return nil, err
		}

		entries := make([]ClusterEntry, 0, len(clusterList.Items))
		for _, cl := range clusterList.Items {
			adStatuses, _ := api.Get[resource.ListResponse[resource.AdapterStatus]](
				context.Background(), client, "clusters/"+cl.ID+"/statuses",
			)
			npList, _ := api.Get[resource.ListResponse[resource.NodePool]](
				context.Background(), client, "clusters/"+cl.ID+"/nodepools",
			)
			npStatuses := make(map[string][]resource.AdapterStatus)
			for _, np := range npList.Items {
				npAdStatus, _ := api.Get[resource.ListResponse[resource.AdapterStatus]](
					context.Background(), client,
					"clusters/"+cl.ID+"/nodepools/"+np.ID+"/statuses",
				)
				npStatuses[np.ID] = npAdStatus.Items
			}
			entries = append(entries, ClusterEntry{
				Cluster:         cl,
				AdapterStatuses: adStatuses.Items,
				Nodepools:       npList.Items,
				NPStatuses:      npStatuses,
			})
		}
		return entries, nil
	}
}
