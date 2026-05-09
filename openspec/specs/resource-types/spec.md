# Resource Types Specification

## Purpose

Define Go struct types for all HyperFleet API resources, providing a type-safe data
model layer with JSON tags that match the API field names exactly. All types MUST
conform to the canonical OpenAPI specification at
[openshift-hyperfleet/hyperfleet-api-spec](https://github.com/openshift-hyperfleet/hyperfleet-api-spec/blob/main/schemas/core/openapi.yaml).

## Requirements

### Requirement: API Spec Conformance

The resource package SHALL define Go types that match the HyperFleet OpenAPI 3.0 specification exactly.

#### Scenario: Type alignment with OpenAPI schema

- GIVEN the canonical API spec at `openshift-hyperfleet/hyperfleet-api-spec/schemas/core/openapi.yaml`
- WHEN Go resource types are defined
- THEN every struct field, JSON tag, and type MUST correspond to the matching `components/schemas/*` definition in the OpenAPI spec
- AND any divergence from the OpenAPI spec MUST be explicitly documented with rationale

### Requirement: Cluster Resource Type

The resource package SHALL define a Cluster struct matching the `Cluster` schema in the OpenAPI spec.

#### Scenario: Cluster struct fields

- GIVEN the `resource` package is imported
- WHEN a Cluster is defined
- THEN it MUST include fields: `ID` (string), `Kind` (string), `Name` (string), `Generation` (int32), `Labels` (map[string]string), `Spec` (map[string]any), `Status` (ClusterStatus), `CreatedBy` (string), `CreatedTime` (string), `UpdatedBy` (string), `UpdatedTime` (string), `DeletedBy` (string, omitempty), `DeletedTime` (string, omitempty), `Href` (string)
- AND `Labels` MUST be `map[string]string` (not `map[string]any`) per the OpenAPI `additionalProperties: type: string`
- AND `Spec` MUST be `map[string]any` per the OpenAPI `ClusterSpec` open object
- AND all fields MUST have JSON struct tags matching the API field names (snake_case)

#### Scenario: Cluster JSON round-trip

- GIVEN a JSON blob representing a cluster from the HyperFleet API
- WHEN the JSON is unmarshaled into a Cluster struct and re-marshaled
- THEN the output JSON MUST preserve all fields without data loss
- AND `spec` MUST preserve arbitrary nested keys
- AND `labels` MUST preserve all string key-value pairs

### Requirement: NodePool Resource Type

The resource package SHALL define a NodePool struct matching the `NodePool` schema in the OpenAPI spec.

#### Scenario: NodePool struct fields

- GIVEN the `resource` package is imported
- WHEN a NodePool is defined
- THEN it MUST include fields: `ID` (string), `Kind` (string), `Name` (string), `Generation` (int32), `Labels` (map[string]string), `Spec` (map[string]any), `Status` (NodePoolStatus), `OwnerReferences` (ObjectReference — single object, not array), `CreatedBy` (string), `CreatedTime` (string), `UpdatedBy` (string), `UpdatedTime` (string), `DeletedBy` (string, omitempty), `DeletedTime` (string, omitempty), `Href` (string)
- AND `OwnerReferences` MUST be a single `ObjectReference` struct (not a slice), matching the OpenAPI `$ref: ObjectReference`
- AND `ObjectReference` MUST include `ID` (string), `Kind` (string), `Href` (string) fields

#### Scenario: NodePool spec extensibility

- GIVEN a nodepool JSON with provider-specific spec fields (e.g., `platform.type`, `replicas`)
- WHEN the JSON is unmarshaled into a NodePool struct
- THEN the `Spec` map MUST preserve all nested fields including `platform.type`
- AND the `Labels` map MUST preserve all string key-value pairs

### Requirement: ClusterStatus Type

The resource package SHALL define a ClusterStatus struct matching the `ClusterStatus` schema in the OpenAPI spec.

#### Scenario: ClusterStatus struct fields

- GIVEN the `ClusterStatus` schema at `openshift-hyperfleet/hyperfleet-api-spec`
- WHEN ClusterStatus is defined
- THEN it MUST include: `Conditions` ([]ResourceCondition, required, minimum 2)
- AND the mandatory condition types MUST be: `Reconciled` (replaces deprecated `Ready`) and `Available`
- AND this object is server-computed and MUST NOT be modified directly by the CLI

### Requirement: NodePoolStatus Type

The resource package SHALL define a NodePoolStatus struct matching the `NodePoolStatus` schema in the OpenAPI spec.

#### Scenario: NodePoolStatus struct fields

- GIVEN the `NodePoolStatus` schema at `openshift-hyperfleet/hyperfleet-api-spec`
- WHEN NodePoolStatus is defined
- THEN it MUST include: `Conditions` ([]ResourceCondition, required, minimum 2)
- AND the mandatory condition types MUST be: `Reconciled` (replaces deprecated `Ready`) and `Available`
- AND this object is server-computed and MUST NOT be modified directly by the CLI

### Requirement: ResourceCondition Type

The resource package SHALL define a ResourceCondition struct for cluster/nodepool status conditions, matching the `ResourceCondition` schema.

#### Scenario: ResourceCondition struct fields

- GIVEN a status condition from a cluster or nodepool
- WHEN it is represented as a ResourceCondition struct
- THEN it MUST include all required fields: `Type` (string), `Status` (string), `LastTransitionTime` (string), `ObservedGeneration` (int32), `CreatedTime` (string), `LastUpdatedTime` (string)
- AND optional fields: `Reason` (string), `Message` (string)
- AND the `Status` field MUST accept only: `True`, `False` (per `ResourceConditionStatus` enum — no `Unknown`)

#### Scenario: ResourceCondition vs AdapterCondition distinction

- GIVEN the API defines two distinct condition types
- WHEN conditions appear on cluster/nodepool status objects
- THEN `ResourceCondition` MUST be used (includes `observed_generation`, `created_time`, `last_updated_time`)
- AND when conditions appear inside `AdapterStatus` objects, `AdapterCondition` MUST be used (no `observed_generation`, `created_time`, `last_updated_time`)

### Requirement: AdapterCondition Type

The resource package SHALL define an AdapterCondition struct for conditions within adapter status reports, matching the `AdapterCondition` schema.

#### Scenario: AdapterCondition struct fields

- GIVEN a condition inside an adapter status report
- WHEN it is represented as an AdapterCondition struct
- THEN it MUST include required fields: `Type` (string), `Status` (string), `LastTransitionTime` (string)
- AND optional fields: `Reason` (string), `Message` (string)
- AND the `Status` field MUST accept values: `True`, `False`, `Unknown` (per `AdapterConditionStatus` enum)
- AND `AdapterCondition` MUST NOT include `observed_generation`, `created_time`, or `last_updated_time` (these are at the AdapterStatus level)

### Requirement: AdapterStatus Resource Type

The resource package SHALL define an AdapterStatus struct matching the `AdapterStatus` schema.

#### Scenario: AdapterStatus struct fields

- GIVEN an adapter status report from the API
- WHEN it is represented as an AdapterStatus struct
- THEN it MUST include required fields: `Adapter` (string), `ObservedGeneration` (int32), `Conditions` ([]AdapterCondition), `CreatedTime` (string), `LastReportTime` (string)
- AND optional fields: `Metadata` (AdapterStatusMetadata), `Data` (map[string]any)
- AND `Conditions` MUST use `AdapterCondition` (not `ResourceCondition`)

#### Scenario: AdapterStatus metadata fields

- GIVEN an adapter status with job execution metadata
- WHEN the metadata is represented
- THEN `AdapterStatusMetadata` MUST include optional fields: `JobName` (string), `JobNamespace` (string), `Attempt` (int32), `StartedTime` (string), `CompletedTime` (string), `Duration` (string)

### Requirement: AdapterStatusCreateRequest Type

The resource package SHALL define a request type for posting adapter statuses, matching the `AdapterStatusCreateRequest` schema.

#### Scenario: AdapterStatusCreateRequest struct fields

- GIVEN an adapter is posting a status report
- WHEN the request is constructed
- THEN it MUST include required fields: `Adapter` (string), `ObservedGeneration` (int32), `ObservedTime` (string), `Conditions` ([]ConditionRequest)
- AND optional fields: `Metadata` (AdapterStatusMetadata), `Data` (map[string]any)
- AND `ConditionRequest` MUST include required fields: `Type` (string), `Status` (string)
- AND `ConditionRequest` optional fields: `Reason` (string), `Message` (string)
- NOTE: `last_transition_time` is NOT part of `ConditionRequest` per the OpenAPI schema — it is absent from the request payload entirely

### Requirement: CloudEvent Type

The resource package SHALL define a CloudEvent struct for event publishing.

#### Scenario: CloudEvent struct fields

- GIVEN a CloudEvents 1.0 message
- WHEN it is represented as a CloudEvent struct
- THEN it MUST include fields: `SpecVersion` (string), `Type` (string), `Source` (string), `ID` (string), `Data` (any)
- AND `SpecVersion` MUST default to `"1.0"`

### Requirement: Generic List Response

The resource package SHALL define a generic ListResponse type for paginated API responses.

#### Scenario: ListResponse generic wrapper

- GIVEN a paginated API response (e.g., ClusterList, NodePoolList, AdapterStatusList)
- WHEN it is represented as a ListResponse[T]
- THEN it MUST include fields: `Items` ([]T), `Kind` (string), `Page` (int32), `Size` (int32), `Total` (int32)
- AND the `Kind` field MUST reflect the list type (e.g., `ClusterList`, `NodePoolList`, `AdapterStatusList`)

#### Scenario: Empty list response

- GIVEN no items match the query
- WHEN the API returns an empty list
- THEN `Items` MUST be an empty slice (not nil)
- AND `Size` MUST be 0
- AND `Total` MUST be 0

### Requirement: ObjectReference Type

The resource package SHALL define an ObjectReference struct matching the `ObjectReference` schema.

#### Scenario: ObjectReference struct fields

- GIVEN a reference to a parent resource (e.g., cluster owning a nodepool)
- WHEN it is represented as an ObjectReference struct
- THEN it MUST include fields: `ID` (string), `Kind` (string), `Href` (string)
- AND this type MUST be used as the `OwnerReferences` field on NodePool (single object, not array)

### Requirement: ValidationError Type

The resource package SHALL define a ValidationError struct matching the `ValidationError` schema.

#### Scenario: ValidationError struct fields

- GIVEN the API returns a validation error with field-level details
- WHEN the error is parsed
- THEN `ValidationError` MUST include required fields: `Field` (string), `Message` (string)
- AND optional fields: `Value` (any), `Constraint` (string)
- AND the `Constraint` field MUST accept enum values: `required`, `min`, `max`, `min_length`, `max_length`, `pattern`, `enum`, `format`, `unique`
