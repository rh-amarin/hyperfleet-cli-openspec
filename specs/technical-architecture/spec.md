# Technical Architecture Specification

## Purpose

Define the modular Go architecture for the HyperFleet CLI, including package structure, shared libraries, the Cobra command tree, and dependency bundling strategy. The architecture prioritizes maintainability, extensibility, and a self-contained binary with minimal external tool requirements.

## Requirements

### Requirement: Go Module Structure

The CLI SHALL be organized as a single Go module with internal packages following domain-driven boundaries.

#### Scenario: Top-level module layout

- GIVEN the CLI is built as a Go project
- WHEN the repository is initialized
- THEN the module MUST follow this package structure:
  ```
  hf/
  ├── cmd/                    # Cobra command definitions (one file per command group)
  │   ├── root.go             # Root command, global flags
  │   ├── cluster.go          # hf cluster [create|get|list|search|patch|delete|conditions|statuses|id]
  │   ├── nodepool.go         # hf nodepool [create|get|list|search|patch|delete|conditions|statuses|id]
  │   ├── adapter.go          # hf cluster adapter post-status, hf nodepool adapter post-status
  │   ├── config.go           # hf config [show|set|env]
  │   ├── db.go               # hf db [query|delete|config]
  │   ├── maestro.go          # hf maestro [list|get|delete|bundles|consumers]
  │   ├── pubsub.go           # hf pubsub [list|publish cluster|publish nodepool]
  │   ├── rabbitmq.go         # hf rabbitmq [publish]
  │   ├── kube.go             # hf kube [port-forward|curl|debug]
  │   ├── logs.go             # hf logs [<pattern>|adapter]
  │   ├── repos.go            # hf repos
  │   └── resources.go        # hf resources (combined overview of clusters and nodepools)
  ├── internal/
  │   ├── api/                # HyperFleet API client
  │   ├── config/             # Configuration management (split YAML model)
  │   ├── output/             # Output formatting (JSON, table, YAML, colored dots)
  │   ├── resource/           # Shared resource types and data structures
  │   ├── kube/               # Kubernetes operations (client-go wrapper)
  │   ├── maestro/            # Maestro HTTP API client
  │   ├── pubsub/             # Pub/Sub and RabbitMQ event publishing
  │   ├── db/                 # PostgreSQL database operations
  │   └── version/            # Build version info
  ├── main.go                 # Entry point
  ├── go.mod
  └── go.sum
  ```
- AND each `cmd/` file MUST register its commands with the Cobra root command
- AND all business logic MUST reside in `internal/` packages, not in `cmd/` files

### Requirement: Cobra Command Tree

The CLI SHALL use [spf13/cobra](https://github.com/spf13/cobra) for command routing, flag parsing, and help generation.

#### Scenario: Command hierarchy

- GIVEN Cobra is the CLI framework
- WHEN commands are registered
- THEN the command tree MUST follow this structure:
  ```
  hf
  ├── cluster
  │   ├── create    <name> [region] [version]
  │   ├── get       [cluster_id]
  │   ├── list      [--table]
  │   ├── search    [name]
  │   ├── patch     {spec|labels} [cluster_id]
  │   ├── delete    [cluster_id]
  │   ├── id
  │   ├── conditions      [--table] [cluster_id]
  │   ├── statuses        [cluster_id]
  │   └── adapter
  │       └── post-status <adapter> <status> <generation>
  ├── nodepool
  │   ├── create    <name> [count] [instance-type]
  │   ├── get       [nodepool_id]
  │   ├── list      [--table]
  │   ├── search    [name]
  │   ├── patch     {spec|labels} [nodepool_id]
  │   ├── delete    [nodepool_id]
  │   ├── id
  │   ├── conditions      [--table] [nodepool_id]
  │   ├── statuses        [nodepool_id]
  │   └── adapter
  │       └── post-status <adapter> <status> <generation> [nodepool_id]
  ├── config
  │   ├── show      [env-name]
  │   ├── set       <key> <value>
  │   └── env
  │       ├── new      [name]
  │       ├── list
  │       ├── show     <name>
  │       └── activate <name>
  ├── resources
  ├── db
  │   ├── query     <sql> | -f <file>
  │   ├── delete    <clusters|nodepools|adapter_statuses|ALL>
  │   └── config
  ├── maestro
  │   ├── list
  │   ├── get       [name]
  │   ├── delete    [name]
  │   ├── bundles
  │   └── consumers
  ├── pubsub
  │   ├── list      [filter]
  │   └── publish
  │       ├── cluster  <topic>
  │       └── nodepool <topic>
  ├── rabbitmq
  │   └── publish
  │       ├── cluster  <exchange> [routing-key]
  │       └── nodepool <exchange> [routing-key]
  ├── kube
  │   ├── port-forward  start|stop|status
  │   ├── curl       [options] <url>
  │   └── debug      <deployment> [namespace]
  ├── logs           <pattern> [flags]
  │   └── adapter    <pattern> [flags]
  ├── repos
  ├── version
  └── completion     bash|zsh|fish|powershell
  ```
NOTE: `hf resources` is a standalone command that displays all clusters and their nodepools in a combined table. `--table` is a flag available only on `list` and `conditions` subcommands — it is NOT supported on `hf resources`.

#### Scenario: Global flags

- GIVEN the root command is defined
- WHEN global flags are registered
- THEN the following persistent flags MUST be available on every command:
  - `--config <path>`: override config file location
  - `--output <format>`: output format (`json`, `table`, `yaml`); default varies per command
  - `--no-color`: disable colored output
  - `--verbose` / `-v`: enable verbose/debug logging
  - `--api-url <url>`: override API URL for this invocation
  - `--api-token <token>`: override API token for this invocation
- NOTE: There is no `--force-color` flag. Color is enabled when stdout is a TTY, disabled otherwise. Use `--no-color` to explicitly disable.

### Requirement: Shared API Client Package (internal/api)

The CLI SHALL provide a shared HTTP client for all HyperFleet API operations.

#### Scenario: API client capabilities

- GIVEN the `internal/api` package exists
- WHEN any command needs to call the HyperFleet API
- THEN the client MUST provide:
  - Base URL construction from config (`{api-url}/api/hyperfleet/{api-version}/`)
  - Generic typed methods: `Get[T]`, `Post[T]`, `Patch[T]`, `Delete[T]` using Go type parameters
  - Authentication via Bearer token from config (omitted when token is empty)
  - Automatic JSON marshaling/unmarshaling with `encoding/json`
  - RFC 7807 Problem Details error parsing with structured `APIError` type implementing `error`
  - Request/response logging when `--verbose` is set (format: `[DEBUG] METHOD URL → STATUS (DURATIONms)`)
  - Default timeout of 30 seconds via `http.Client.Timeout`
  - Context propagation for cancellation via `http.NewRequestWithContext`

#### Scenario: API error handling

- GIVEN the API returns a non-2xx response
- WHEN the client parses the response
- THEN it MUST return a structured `APIError` type containing code, detail, status, title, trace_id, type, timestamp
- AND the error MUST implement Go's `error` interface with format `[{status}] {title}: {detail}`
- AND commands MUST be able to output the raw error JSON (exit 0) or propagate as a Go error
- AND non-JSON error responses MUST be wrapped in an `APIError` with the raw body as `detail`

### Requirement: Shared Output Package (internal/output)

The CLI SHALL provide a shared output formatting package supporting multiple formats.

#### Scenario: Output format dispatch

- GIVEN the `--output` flag is set
- WHEN a command produces output
- THEN the output package MUST dispatch to the appropriate formatter:
  - `json`: pretty-printed JSON with 2-space indentation and trailing newline
  - `table`: formatted table with uppercase headers and aligned columns via `text/tabwriter`
  - `yaml`: YAML serialization via `gopkg.in/yaml.v3`
- AND the default format MUST be determined per command (table for list views, json for get views)

#### Scenario: Dynamic column table rendering

- GIVEN a table output is requested for cluster or nodepool resources
- WHEN conditions vary across resources
- THEN the table renderer MUST:
  - Collect all unique condition types across all items
  - Order columns: fixed columns first, then `Available`, then alphabetical adapter conditions, then `Reconciled` last
  - Render status values as colored dots: green `●`=True, red `●`=False, yellow `●`=Unknown, `-`=absent
  - Respect `--no-color` flag and `NO_COLOR` environment variable to disable ANSI colors
  - In no-color mode, render status as plain text: `True`, `False`, `Unknown`, `-`

Status dot rendering follows the spec defined in `output-formatting/spec.md` Requirement: Colored Dot Rendering.

### Requirement: Shared Utility Functions

The CLI SHALL expose well-defined shared helper functions used across command implementations.

#### Scenario: API lookup helpers

- GIVEN the `internal/api` package exists
- WHEN commands need to find resources by name
- THEN the package MUST provide:
  - `FindClusterByName(ctx context.Context, client *Client, name string) (*resource.Cluster, error)` — queries the clusters list endpoint filtering by exact name match; returns the first match or nil if not found
  - `FindNodePoolByName(ctx context.Context, client *Client, clusterID, name string) (*resource.NodePool, error)` — queries the nodepools list endpoint for the given cluster, filtering by exact name match

#### Scenario: Config state helpers

- GIVEN the `internal/config` package manages active state
- WHEN commands need to read or write the active cluster or nodepool
- THEN the package MUST provide:
  - `SetClusterID(id string) error` — writes `cluster-id` to `state.yaml`
  - `GetClusterID() (string, error)` — reads `cluster-id` from `state.yaml`; returns error if not set
  - `SetNodePoolID(id string) error` — writes `nodepool-id` to `state.yaml`
  - `GetNodePoolID() (string, error)` — reads `nodepool-id` from `state.yaml`; returns error if not set

### Requirement: Shared Resource Types Package (internal/resource)

The CLI SHALL define shared Go types for all HyperFleet resources.

#### Scenario: Core resource types

- GIVEN the `internal/resource` package exists
- WHEN resource types are defined
- THEN the package MUST include:
  - `Cluster` struct with fields: ID, Kind, Href, Name, Generation (int32), Labels (map[string]string), Spec (map[string]any), Status (ClusterStatus), CreatedBy, CreatedTime, UpdatedBy, UpdatedTime, DeletedBy, DeletedTime
  - `NodePool` struct with fields: ID, Kind, Href, Name, Generation (int32), Labels (map[string]string), Spec (map[string]any), Status (NodePoolStatus), OwnerReferences (ObjectReference — single object), CreatedBy, CreatedTime, UpdatedBy, UpdatedTime, DeletedBy, DeletedTime
  - `ResourceCondition` struct for cluster/nodepool conditions: Type, Status (True|False only), Reason, Message, LastTransitionTime, ObservedGeneration, CreatedTime, LastUpdatedTime
  - `AdapterCondition` struct for adapter status conditions: Type, Status (True|False|Unknown), Reason, Message, LastTransitionTime
  - `AdapterStatus` struct: Adapter, ObservedGeneration, Conditions ([]AdapterCondition), Metadata (AdapterStatusMetadata), Data, CreatedTime, LastReportTime
  - `AdapterStatusMetadata` struct: JobName, JobNamespace, Attempt, StartedTime, CompletedTime, Duration
  - `AdapterStatusCreateRequest` struct: Adapter, ObservedGeneration, ObservedTime, Conditions ([]ConditionRequest), Metadata, Data
  - `ObjectReference` struct: ID, Kind, Href
  - `CloudEvent` struct: SpecVersion, Type, Source, ID, Data
  - `ValidationError` struct: Field, Message, Value, Constraint
  - Generic `ListResponse[T]` with fields: Items, Kind, Page (int32), Size (int32), Total (int32)
- AND all types MUST conform to the canonical OpenAPI spec at `openshift-hyperfleet/hyperfleet-api-spec`
- AND all types MUST have JSON struct tags matching the API field names (snake_case)
- AND `Spec` MUST use `map[string]any` and `Labels` MUST use `map[string]string`
- AND `ListResponse[T]` MUST use Go type parameters for type-safe list operations

### Requirement: Kubernetes Operations Package (internal/kube)

The CLI SHALL bundle `client-go` for all Kubernetes operations without requiring an external `kubectl` binary.

#### Scenario: Bundled client-go capabilities

- GIVEN `client-go` is bundled
- WHEN the kube package is used
- THEN it MUST provide:
  - Kubeconfig loading (respecting `--kubeconfig` flag, `KUBECONFIG` env, and default `~/.kube/config`)
  - Port-forward lifecycle management (start, stop, status with PID tracking)
  - Pod log streaming with label/name filtering and multi-pod fan-out (replacing stern)
  - Pod exec for in-cluster curl and debug operations
- AND the binary MUST NOT require `kubectl` to be installed for any core operation

### Requirement: Dependency Bundling Strategy

The CLI SHALL bundle Go libraries to replace external tool dependencies, producing a self-contained binary.

#### Scenario: Bundled dependencies

- GIVEN the CLI is compiled
- WHEN external tool equivalents are needed
- THEN the following MUST be bundled as Go libraries:
  | Former Tool | Go Replacement | Library |
  |-------------|---------------|---------|
  | jq | encoding/json | stdlib |
  | curl | net/http | stdlib |
  | awk/sed | text/tabwriter + string processing | stdlib |
  | lsof/ss | net.Listen / os.FindProcess | stdlib |
  | psql | database/sql + pgx | jackc/pgx/v5 |
  | kubectl | client-go | k8s.io/client-go |
  | gcloud pubsub | cloud.google.com/go/pubsub | Google Cloud Go SDK |
  | gh | go-github | google/go-github/v60 |
  | stern | client-go pod log streaming | k8s.io/client-go |

#### Scenario: Maestro HTTP API

- GIVEN maestro-http-endpoint is configured
- WHEN maestro commands are invoked
- THEN `hf maestro list`, `hf maestro get`, `hf maestro delete`, `hf maestro bundles`, and `hf maestro consumers` MUST use the Maestro HTTP API directly via `net/http`
- AND the CLI MUST NOT require any external `maestro-cli` tool for any maestro command

#### Scenario: Zero external dependencies for core operations

- GIVEN the CLI binary is installed on a clean system
- WHEN the user runs cluster, nodepool, adapter-status, config, resources, or output commands
- THEN the CLI MUST NOT require any external tools to be installed
- AND only GCP credentials (for Pub/Sub) MAY be required for their respective specialized commands

### Requirement: Error Handling Strategy

The CLI SHALL follow a consistent error handling pattern across all commands.

#### Scenario: Error propagation

- GIVEN an error occurs during command execution
- WHEN the error is an API error (RFC 7807)
- THEN the CLI MUST output the structured error in the current output format (json/table/yaml)
- AND exit with code 0 to maintain backwards compatibility with the shell scripts

#### Scenario: CLI-level errors

- GIVEN an error occurs that is not an API error (network failure, config missing, etc.)
- WHEN the error is reported
- THEN the CLI MUST print the error to stderr with a `[ERROR]` prefix
- AND exit with a non-zero exit code

#### Scenario: Warning and info messages

- GIVEN a non-fatal condition occurs (duplicate creation, empty results)
- WHEN the condition is reported
- THEN the CLI MUST print to stderr with `[WARN]` or `[INFO]` prefix
- AND continue execution or exit with code 0

### Requirement: Logging and Verbosity

The CLI SHALL support structured logging with configurable verbosity.

#### Scenario: Verbose mode

- GIVEN `--verbose` or `-v` is passed
- WHEN the CLI executes
- THEN debug-level messages MUST be printed to stderr
- AND HTTP request/response details (method, URL, status code, duration) MUST be logged
- AND config resolution steps MUST be logged

#### Scenario: Default mode

- GIVEN no verbosity flag is set
- WHEN the CLI executes
- THEN only warnings, errors, and command output MUST be displayed
- AND no debug information MUST appear
