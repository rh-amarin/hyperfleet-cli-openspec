package pubsub

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type cloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	Type            string          `json:"type"`
	Source          string          `json:"source"`
	ID              string          `json:"id"`
	Time            string          `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

type clusterData struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

type nodepoolData struct {
	ID              string          `json:"id"`
	Kind            string          `json:"kind"`
	Href            string          `json:"href"`
	OwnerReferences ownerRef        `json:"owner_references"`
}

type ownerRef struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

func baseURL(apiURL, apiVersion string) string {
	return strings.TrimRight(apiURL, "/") + "/api/hyperfleet/" + apiVersion
}

// BuildClusterEvent returns a marshalled CloudEvent 1.0 for cluster reconciliation.
func BuildClusterEvent(clusterID, apiURL, apiVersion string) ([]byte, error) {
	base := baseURL(apiURL, apiVersion)
	href := fmt.Sprintf("%s/clusters/%s", base, clusterID)

	data, err := json.Marshal(clusterData{ID: clusterID, Kind: "Cluster", Href: href})
	if err != nil {
		return nil, err
	}

	ev := cloudEvent{
		SpecVersion:     "1.0",
		Type:            "com.redhat.hyperfleet.cluster.reconcile.v1",
		Source:          "/hyperfleet/service/sentinel",
		ID:              clusterID,
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data:            data,
	}
	return json.MarshalIndent(ev, "", "  ")
}

// BuildNodePoolEvent returns a marshalled CloudEvent 1.0 for nodepool reconciliation.
func BuildNodePoolEvent(clusterID, nodepoolID, apiURL, apiVersion string) ([]byte, error) {
	base := baseURL(apiURL, apiVersion)
	clusterHref := fmt.Sprintf("%s/clusters/%s", base, clusterID)
	npHref := fmt.Sprintf("%s/clusters/%s/nodepools/%s", base, clusterID, nodepoolID)

	data, err := json.Marshal(nodepoolData{
		ID:   nodepoolID,
		Kind: "NodePool",
		Href: npHref,
		OwnerReferences: ownerRef{
			ID:   clusterID,
			Kind: "Cluster",
			Href: clusterHref,
		},
	})
	if err != nil {
		return nil, err
	}

	ev := cloudEvent{
		SpecVersion:     "1.0",
		Type:            "com.redhat.hyperfleet.nodepool.reconcile.v1",
		Source:          "/hyperfleet/service/sentinel",
		ID:              nodepoolID,
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data:            data,
	}
	return json.MarshalIndent(ev, "", "  ")
}
