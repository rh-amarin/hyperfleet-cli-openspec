# Command Hierarchy and Architecture Specification

## Purpose

Define the Cobra command tree, Go module structure, shared library contracts, and developer tooling for the HyperFleet CLI. The architecture prioritizes a self-contained binary with no external tool dependencies for core operations.
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
  │   ├── config.go           # hf config [show|set]
  │   ├── env.go              # hf env [create|list|show|activate|delete]
  │   ├── db.go               # hf db [query|delete|config]
  │   ├── maestro.go          # hf maestro [list|get|delete|bundles|consumers]
  │   ├── pubsub.go           # hf pubsub [list|publish cluster|publish nodepool]
  │   ├── rabbitmq.go         # hf rabbitmq [publish]
  │   ├── kube.go             # hf kube [port-forward|curl|debug]
  │   ├── logs.go             # hf logs [<pattern>|adapter|insights]
  │   ├── repos.go            # hf repos
  │   └── resources.go        # hf resources / hf table (combined overview of clusters and nodepools)
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
- THEN the command tree MUST include:
  ```
  hf
  ├── resource | rs
  │   ├── types
  │   └── <type-name>          # dynamically registered from resource-types config
  │       ├── list
  │       ├── get
  │       ├── create
  │       ├── search
  │       ├── patch
  │       ├── delete
  │       └── id
  ```
- AND dynamic `<type-name>` groups MUST be registered after the active environment is loaded
- AND existing `hf cluster`, `hf nodepool`, and `hf resources` commands MUST remain unchanged

### Requirement: Version Package

The CLI SHALL expose a canonical version string through `internal/version`.

#### Scenario: Default version

- WHEN the binary is built without `-ldflags` injection
- THEN `internal/version.Version` MUST equal `"dev"`

#### Scenario: Injected version

- WHEN the binary is built with `-ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=<tag>"`
- THEN `internal/version.Version` MUST equal `<tag>`

#### Scenario: Version command output

- WHEN the user runs `hf version`
- THEN the CLI MUST print the version string to stdout and exit 0

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
  - RFC 7807 Problem Details error parsing with a structured error type implementing `error`
  - Request/response logging when `--verbose` is set (format: `[DEBUG] METHOD URL → STATUS (DURATIONms)`)
  - Default timeout of 30 seconds
  - Context propagation for cancellation

#### Scenario: API error handling

- GIVEN the API returns a non-2xx response
- WHEN the client parses the response
- THEN it MUST return a structured error type containing code, detail, status, title, trace_id, type, timestamp
- AND the error MUST implement Go's `error` interface with format `[{status}] {title}: {detail}`
- AND commands MUST be able to output the raw error JSON (exit 0) or propagate as a Go error
- AND non-JSON error responses MUST be wrapped with the raw body as `detail`

### Requirement: Shared Output Package (internal/output)

The CLI SHALL provide a shared output formatting package supporting multiple formats.

#### Scenario: Output format dispatch

- GIVEN the `--output` flag is set
- WHEN a command produces output
- THEN the output package MUST dispatch to the appropriate formatter:
  - `json`: pretty-printed JSON with 2-space indentation and trailing newline
  - `table`: formatted table with uppercase headers and aligned columns via `text/tabwriter`
  - `yaml`: YAML serialization via `gopkg.in/yaml.v3`

#### Scenario: Dynamic column table rendering

- GIVEN a table output is requested for cluster or nodepool resources
- WHEN conditions vary across resources
- THEN the table renderer MUST:
  - Collect all unique condition types across all items, excluding types ending in `Successful`
  - Order condition columns: `Available` first, then remaining conditions alphabetically, `Reconciled` last
  - Append adapter columns after condition columns, ordered by first occurrence
  - Render status values as colored dots: green `●`=True, red `●`=False, yellow `●`=Unknown, `-`=absent
  - Respect `--no-color` flag and `NO_COLOR` environment variable to disable ANSI colors

### Requirement: Shared Resource Types Package (internal/resource)

The CLI SHALL define shared Go types for all HyperFleet resources.

#### Scenario: Core resource types

- GIVEN the `internal/resource` package exists
- WHEN resource types are defined
- THEN the package MUST include:
  - `Cluster` struct: ID, Kind, Href, Name, Generation, Labels, Spec, Status, CreatedBy, CreatedTime, UpdatedBy, UpdatedTime, DeletedBy, DeletedTime
  - `NodePool` struct: ID, Kind, Href, Name, Generation, Labels, Spec, Status, OwnerReferences, CreatedBy, CreatedTime, UpdatedBy, UpdatedTime, DeletedBy, DeletedTime
  - `ResourceCondition` struct: Type, Status (True|False only), Reason, Message, LastTransitionTime, ObservedGeneration, CreatedTime, LastUpdatedTime
  - `AdapterCondition` struct: Type, Status (True|False|Unknown), Reason, Message, LastTransitionTime
  - `AdapterStatus` struct: Adapter, ObservedGeneration, Conditions, Metadata, Data, CreatedTime, LastReportTime
  - `AdapterStatusCreateRequest` struct: Adapter, ObservedGeneration, ObservedTime, Conditions, Metadata, Data
  - `CloudEvent` struct: SpecVersion, Type, Source, ID, Data
  - Generic `ListResponse[T]`: Items, Kind, Page, Size, Total
- AND all types MUST have JSON struct tags matching the API field names (snake_case)
- AND `Spec` MUST use `map[string]any` and `Labels` MUST use `map[string]string`

### Requirement: Kubernetes Operations Package (internal/kube)

The CLI SHALL bundle `client-go` for all Kubernetes operations without requiring an external `kubectl` binary.

#### Scenario: Bundled client-go capabilities

- GIVEN `client-go` is bundled
- WHEN the kube package is used
- THEN it MUST provide:
  - Kubeconfig loading (respecting `--kubeconfig` flag, `KUBECONFIG` env, and default `~/.kube/config`)
  - Port-forward lifecycle management (start, stop, status with PID tracking)
  - Pod log streaming with label/name filtering and multi-pod fan-out
  - Pod exec for in-cluster curl and debug operations
- AND the binary MUST NOT require `kubectl` to be installed for any core operation

### Requirement: Dependency Bundling Strategy

The CLI SHALL bundle Go libraries to replace external tool dependencies, producing a self-contained binary.

#### Scenario: Bundled dependencies

- GIVEN the CLI is compiled
- WHEN external tool equivalents are needed
- THEN the following MUST be bundled as Go libraries:
  | Former Tool | Go Replacement | Library |
  |-------------|----------------|---------|
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
- AND the CLI MUST NOT require any external `maestro-cli` tool

#### Scenario: Zero external dependencies for core operations

- GIVEN the CLI binary is installed on a clean system
- WHEN the user runs cluster, nodepool, adapter-status, config, resources, or output commands
- THEN the CLI MUST NOT require any external tools to be installed
- AND only GCP credentials (for Pub/Sub) MAY be required for their respective specialized commands

### Requirement: Makefile Build Targets

The repository SHALL provide a `Makefile` with standard developer targets.

#### Scenario: Build target

- WHEN the developer runs `make build`
- THEN the binary `bin/hf` MUST be produced with no errors

#### Scenario: Test target

- WHEN the developer runs `make test`
- THEN `go test ./...` MUST run and exit 0 when all tests pass

#### Scenario: Vet target

- WHEN the developer runs `make vet`
- THEN `go vet ./...` MUST run and exit 0 when no issues are found

### Requirement: Error Handling Strategy

The CLI SHALL follow a consistent error handling pattern across all commands.

#### Scenario: API error propagation

- GIVEN an error occurs during command execution
- WHEN the error is an API error (RFC 7807)
- THEN the CLI MUST output the structured error in the current output format (json/table/yaml)
- AND exit with code 0

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

The CLI SHALL support configurable verbosity.

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

### Requirement: Generic Resource Command Group

The CLI SHALL provide a top-level command group `hf resource` with alias `hf rs` for config-defined HyperFleet API resource types.

#### Scenario: Command group registration

- GIVEN the root command is initialized
- WHEN the user runs `hf resource --help` or `hf rs --help`
- THEN the CLI MUST show the `resource` command group
- AND `hf rs` MUST be a registered alias for `hf resource`
- AND this group MUST NOT be named `resources` (that name is reserved for the cluster+nodepool overview)

#### Scenario: Distinction from hf resources

- GIVEN `hf resources` displays combined cluster and nodepool overview
- WHEN a user runs `hf resource <type> list`
- THEN the CLI MUST operate on config-defined API resource types
- AND MUST NOT invoke the cluster+nodepool overview command

#### Scenario: Dynamic type subcommands

- GIVEN the active environment defines resource types under `resource-types`
- WHEN the user runs `hf resource --help` after config load
- THEN the CLI MUST register one subcommand group per configured type name (e.g. `channels`, `versions`)
- AND each type group MUST expose subcommands: `list`, `get`, `create`, `search`, `patch`, `delete`, `id`

#### Scenario: No configured types

- GIVEN the active environment has no `resource-types` section or an empty map
- WHEN the user runs `hf resource --help`
- THEN the CLI MUST show at minimum the `types` subcommand
- AND MUST NOT fail with a parse error

### Requirement: Resource Types Subcommand

The CLI SHALL provide `hf resource types` to display configured resource types and their relationships.

#### Scenario: List configured types

- GIVEN resource types are defined in the active environment
- WHEN the user runs `hf resource types`
- THEN the CLI MUST print each type name, its API path template, its parent (if any), and its state-key
- AND MUST indicate whether each state-key is currently set in `state.yaml`

#### Scenario: Parent chain display

- GIVEN a child type with `parent: channels`
- WHEN the user runs `hf resource types`
- THEN the output MUST show the parent relationship
- AND MUST indicate which parent state-key is required before child commands can run

