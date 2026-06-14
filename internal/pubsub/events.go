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

type resourceData struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"`
	Href            string    `json:"href"`
	OwnerReferences *ownerRef `json:"owner_references,omitempty"`
}

type ownerRef struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

// AncestorID holds the resolved identity of one ancestor resource, ordered root→immediate-parent.
type AncestorID struct {
	TypeName string // resource-types key (e.g. "clusters")
	ID       string // active ID from state
	Path     string // resolved API path for this ancestor (e.g. "clusters")
}

func apiBase(apiURL, apiVersion string) string {
	return strings.TrimRight(apiURL, "/") + "/api/hyperfleet/" + apiVersion
}

// BuildGenericReconcileEvent returns a marshalled CloudEvent 1.0 for any resource type.
// ancestors must be ordered root→immediate-parent; may be empty for root resources.
// resourcePath is the resolved API path for the resource (e.g. "clusters/id/nodepools").
func BuildGenericReconcileEvent(typeName, resourceID string, ancestors []AncestorID, resourcePath, apiURL, apiVersion string) ([]byte, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("resourceID is required")
	}
	base := apiBase(apiURL, apiVersion)
	href := fmt.Sprintf("%s/%s/%s", base, resourcePath, resourceID)

	rd := resourceData{
		ID:   resourceID,
		Kind: typeName,
		Href: href,
	}

	if len(ancestors) > 0 {
		parent := ancestors[len(ancestors)-1]
		parentHref := fmt.Sprintf("%s/%s/%s", base, parent.Path, parent.ID)
		rd.OwnerReferences = &ownerRef{
			ID:   parent.ID,
			Kind: parent.TypeName,
			Href: parentHref,
		}
	}

	data, err := json.Marshal(rd)
	if err != nil {
		return nil, err
	}

	ev := cloudEvent{
		SpecVersion:     "1.0",
		Type:            fmt.Sprintf("com.redhat.hyperfleet.%s.reconcile.v1", typeName),
		Source:          "/hyperfleet/service/sentinel",
		ID:              resourceID,
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data:            data,
	}
	return json.MarshalIndent(ev, "", "  ")
}
