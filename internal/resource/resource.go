// Package resource defines Go struct types for all HyperFleet API resources.
// All types conform to the canonical OpenAPI specification and use JSON snake_case tags.
package resource

// ObjectReference is a reference to a parent resource (e.g., cluster owning a nodepool).
type ObjectReference struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

// ResourceCondition represents a status condition on a Cluster or NodePool.
// Status accepts only "True" or "False" (no "Unknown").
type ResourceCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastTransitionTime string `json:"last_transition_time"`
	ObservedGeneration int32  `json:"observed_generation"`
	CreatedTime        string `json:"created_time"`
	LastUpdatedTime    string `json:"last_updated_time"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
}

// AdapterCondition represents a status condition within an AdapterStatus report.
// Status accepts "True", "False", or "Unknown".
type AdapterCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastTransitionTime string `json:"last_transition_time"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
}

// ClusterStatus is the server-computed status of a Cluster.
type ClusterStatus struct {
	Conditions []ResourceCondition `json:"conditions"`
}

// NodePoolStatus is the server-computed status of a NodePool.
type NodePoolStatus struct {
	Conditions []ResourceCondition `json:"conditions"`
}

// Cluster represents a HyperFleet cluster resource.
type Cluster struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Name        string            `json:"name"`
	Generation  int32             `json:"generation"`
	Labels      map[string]string `json:"labels"`
	Spec        map[string]any    `json:"spec"`
	Status      ClusterStatus     `json:"status"`
	CreatedBy   string            `json:"created_by"`
	CreatedTime string            `json:"created_time"`
	UpdatedBy   string            `json:"updated_by"`
	UpdatedTime string            `json:"updated_time"`
	DeletedBy   string            `json:"deleted_by,omitempty"`
	DeletedTime string            `json:"deleted_time,omitempty"`
	Href        string            `json:"href"`
}

// NodePool represents a HyperFleet node pool resource.
type NodePool struct {
	ID              string          `json:"id"`
	Kind            string          `json:"kind"`
	Name            string          `json:"name"`
	Generation      int32           `json:"generation"`
	Labels          map[string]string `json:"labels"`
	Spec            map[string]any  `json:"spec"`
	Status          NodePoolStatus  `json:"status"`
	OwnerReferences ObjectReference `json:"owner_references"`
	CreatedBy       string          `json:"created_by"`
	CreatedTime     string          `json:"created_time"`
	UpdatedBy       string          `json:"updated_by"`
	UpdatedTime     string          `json:"updated_time"`
	DeletedBy       string          `json:"deleted_by,omitempty"`
	DeletedTime     string          `json:"deleted_time,omitempty"`
	Href            string          `json:"href"`
}

// AdapterStatusMetadata holds optional job execution metadata for an AdapterStatus.
type AdapterStatusMetadata struct {
	JobName       string `json:"job_name,omitempty"`
	JobNamespace  string `json:"job_namespace,omitempty"`
	Attempt       int32  `json:"attempt,omitempty"`
	StartedTime   string `json:"started_time,omitempty"`
	CompletedTime string `json:"completed_time,omitempty"`
	Duration      string `json:"duration,omitempty"`
}

// AdapterStatus represents a status report from an adapter.
type AdapterStatus struct {
	Adapter            string                `json:"adapter"`
	ObservedGeneration int32                 `json:"observed_generation"`
	Conditions         []AdapterCondition    `json:"conditions"`
	CreatedTime        string                `json:"created_time"`
	LastReportTime     string                `json:"last_report_time"`
	Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"`
	Data               map[string]any        `json:"data,omitempty"`
}

// ConditionRequest is a condition entry in an AdapterStatusCreateRequest.
// Note: last_transition_time is NOT included — it is absent from the request payload.
type ConditionRequest struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// AdapterStatusCreateRequest is the request body for posting an adapter status report.
type AdapterStatusCreateRequest struct {
	Adapter            string                `json:"adapter"`
	ObservedGeneration int32                 `json:"observed_generation"`
	ObservedTime       string                `json:"observed_time"`
	Conditions         []ConditionRequest    `json:"conditions"`
	Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"`
	Data               map[string]any        `json:"data,omitempty"`
}

// CloudEvent represents a CloudEvents 1.0 message.
type CloudEvent struct {
	SpecVersion string `json:"specversion"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	ID          string `json:"id"`
	Data        any    `json:"data"`
}

// ListResponse is a generic paginated API list response.
type ListResponse[T any] struct {
	Items []T    `json:"items"`
	Kind  string `json:"kind"`
	Page  int32  `json:"page"`
	Size  int32  `json:"size"`
	Total int32  `json:"total"`
}

// ValidationError represents a field-level validation failure from the API.
type ValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Value      any    `json:"value,omitempty"`
	Constraint string `json:"constraint,omitempty"`
}
