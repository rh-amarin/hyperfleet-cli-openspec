// Package resource defines Go struct types for all HyperFleet API resources.
// All types conform to the canonical OpenAPI specification and use JSON snake_case tags.
// Explicit yaml tags are required on all multi-word fields because yaml.v3 lowercases
// Go field names directly rather than reading json struct tags.
package resource

// ObjectReference is a reference to a parent resource (e.g., cluster owning a nodepool).
type ObjectReference struct {
	ID   string `json:"id"   yaml:"id"`
	Kind string `json:"kind" yaml:"kind"`
	Href string `json:"href" yaml:"href"`
}

// ResourceCondition represents a status condition on a Cluster or NodePool.
// Status accepts only "True" or "False" (no "Unknown").
type ResourceCondition struct {
	Type               string `json:"type"                 yaml:"type"`
	Status             string `json:"status"               yaml:"status"`
	LastTransitionTime string `json:"last_transition_time" yaml:"last_transition_time"`
	ObservedGeneration int32  `json:"observed_generation"  yaml:"observed_generation"`
	CreatedTime        string `json:"created_time"         yaml:"created_time"`
	LastUpdatedTime    string `json:"last_updated_time"    yaml:"last_updated_time"`
	Reason             string `json:"reason,omitempty"     yaml:"reason,omitempty"`
	Message            string `json:"message,omitempty"    yaml:"message,omitempty"`
}

// AdapterCondition represents a status condition within an AdapterStatus report.
// Status accepts "True", "False", or "Unknown".
type AdapterCondition struct {
	Type               string `json:"type"                 yaml:"type"`
	Status             string `json:"status"               yaml:"status"`
	LastTransitionTime string `json:"last_transition_time" yaml:"last_transition_time"`
	Reason             string `json:"reason,omitempty"     yaml:"reason,omitempty"`
	Message            string `json:"message,omitempty"    yaml:"message,omitempty"`
}

// ClusterStatus is the server-computed status of a Cluster.
type ClusterStatus struct {
	Conditions []ResourceCondition `json:"conditions" yaml:"conditions"`
}

// NodePoolStatus is the server-computed status of a NodePool.
type NodePoolStatus struct {
	Conditions []ResourceCondition `json:"conditions" yaml:"conditions"`
}

// Cluster represents a HyperFleet cluster resource.
type Cluster struct {
	ID          string            `json:"id"                    yaml:"id"`
	Kind        string            `json:"kind"                  yaml:"kind"`
	Name        string            `json:"name"                  yaml:"name"`
	Generation  int32             `json:"generation"            yaml:"generation"`
	Labels      map[string]string `json:"labels"                yaml:"labels"`
	Spec        map[string]any    `json:"spec"                  yaml:"spec"`
	Status      ClusterStatus     `json:"status"                yaml:"status"`
	CreatedBy   string            `json:"created_by"            yaml:"created_by"`
	CreatedTime string            `json:"created_time"          yaml:"created_time"`
	UpdatedBy   string            `json:"updated_by"            yaml:"updated_by"`
	UpdatedTime string            `json:"updated_time"          yaml:"updated_time"`
	DeletedBy   string            `json:"deleted_by,omitempty"  yaml:"deleted_by,omitempty"`
	DeletedTime string            `json:"deleted_time,omitempty" yaml:"deleted_time,omitempty"`
	Href        string            `json:"href"                  yaml:"href"`
}

// NodePool represents a HyperFleet node pool resource.
type NodePool struct {
	ID              string            `json:"id"                     yaml:"id"`
	Kind            string            `json:"kind"                   yaml:"kind"`
	Name            string            `json:"name"                   yaml:"name"`
	Generation      int32             `json:"generation"             yaml:"generation"`
	Labels          map[string]string `json:"labels"                 yaml:"labels"`
	Spec            map[string]any    `json:"spec"                   yaml:"spec"`
	Status          NodePoolStatus    `json:"status"                 yaml:"status"`
	OwnerReferences ObjectReference   `json:"owner_references"       yaml:"owner_references"`
	CreatedBy       string            `json:"created_by"             yaml:"created_by"`
	CreatedTime     string            `json:"created_time"           yaml:"created_time"`
	UpdatedBy       string            `json:"updated_by"             yaml:"updated_by"`
	UpdatedTime     string            `json:"updated_time"           yaml:"updated_time"`
	DeletedBy       string            `json:"deleted_by,omitempty"   yaml:"deleted_by,omitempty"`
	DeletedTime     string            `json:"deleted_time,omitempty" yaml:"deleted_time,omitempty"`
	Href            string            `json:"href"                   yaml:"href"`
}

// AdapterStatusMetadata holds optional job execution metadata for an AdapterStatus.
type AdapterStatusMetadata struct {
	JobName       string `json:"job_name,omitempty"       yaml:"job_name,omitempty"`
	JobNamespace  string `json:"job_namespace,omitempty"  yaml:"job_namespace,omitempty"`
	Attempt       int32  `json:"attempt,omitempty"        yaml:"attempt,omitempty"`
	StartedTime   string `json:"started_time,omitempty"   yaml:"started_time,omitempty"`
	CompletedTime string `json:"completed_time,omitempty" yaml:"completed_time,omitempty"`
	Duration      string `json:"duration,omitempty"       yaml:"duration,omitempty"`
}

// AdapterStatus represents a status report from an adapter.
type AdapterStatus struct {
	Adapter            string                 `json:"adapter"              yaml:"adapter"`
	ObservedGeneration int32                  `json:"observed_generation"  yaml:"observed_generation"`
	Conditions         []AdapterCondition     `json:"conditions"           yaml:"conditions"`
	CreatedTime        string                 `json:"created_time"         yaml:"created_time"`
	LastReportTime     string                 `json:"last_report_time"     yaml:"last_report_time"`
	Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"   yaml:"metadata,omitempty"`
	Data               map[string]any         `json:"data,omitempty"       yaml:"data,omitempty"`
}

// ConditionRequest is a condition entry in an AdapterStatusCreateRequest.
// Note: last_transition_time is NOT included — it is absent from the request payload.
type ConditionRequest struct {
	Type    string `json:"type"             yaml:"type"`
	Status  string `json:"status"           yaml:"status"`
	Reason  string `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// AdapterStatusCreateRequest is the request body for posting an adapter status report.
type AdapterStatusCreateRequest struct {
	Adapter            string                 `json:"adapter"              yaml:"adapter"`
	ObservedGeneration int32                  `json:"observed_generation"  yaml:"observed_generation"`
	ObservedTime       string                 `json:"observed_time"        yaml:"observed_time"`
	Conditions         []ConditionRequest     `json:"conditions"           yaml:"conditions"`
	Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"   yaml:"metadata,omitempty"`
	Data               map[string]any         `json:"data,omitempty"       yaml:"data,omitempty"`
}

// CloudEvent represents a CloudEvents 1.0 message.
type CloudEvent struct {
	SpecVersion string `json:"specversion" yaml:"specversion"`
	Type        string `json:"type"        yaml:"type"`
	Source      string `json:"source"      yaml:"source"`
	ID          string `json:"id"          yaml:"id"`
	Data        any    `json:"data"        yaml:"data"`
}

// ListResponse is a generic paginated API list response.
type ListResponse[T any] struct {
	Items []T    `json:"items" yaml:"items"`
	Kind  string `json:"kind"  yaml:"kind"`
	Page  int32  `json:"page"  yaml:"page"`
	Size  int32  `json:"size"  yaml:"size"`
	Total int32  `json:"total" yaml:"total"`
}

// ValidationError represents a field-level validation failure from the API.
type ValidationError struct {
	Field      string `json:"field"                yaml:"field"`
	Message    string `json:"message"              yaml:"message"`
	Value      any    `json:"value,omitempty"      yaml:"value,omitempty"`
	Constraint string `json:"constraint,omitempty" yaml:"constraint,omitempty"`
}
