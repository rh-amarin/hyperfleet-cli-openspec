# HyperFleet CLI Requirements Specification

## Purpose

Complete requirements specification for the HyperFleet CLI tool (`hf`), reverse-engineered from the shell scripts at [rh-amarin/hyperfleet-cli](https://github.com/rh-amarin/hyperfleet-cli) and extended with architecture, configuration model, and non-functional requirements for the Go reimplementation.

## Specification Index

### Functional Requirements

| # | Domain | Spec | Description | Scripts Covered |
|---|--------|------|-------------|-----------------|
| 01 | [Configuration](config/spec.md) | Config management, `hf env` profiles, diagnostics | hf.config.sh, hf.cluster.id.sh, hf.nodepool.id.sh |
| 02 | [Cluster Lifecycle](cluster-lifecycle/spec.md) | Cluster operations via `hf rs clusters` (legacy `hf cluster` removed) | hf.cluster.*.sh → `hf rs clusters` |
| 03 | [NodePool Lifecycle](nodepool-lifecycle/spec.md) | NodePool operations via `hf rs nodepools` (legacy `hf nodepool` removed) | hf.nodepool.*.sh → `hf rs nodepools` |
| 04 | [Adapter Status](adapter-status/spec.md) | Adapter reporting via `hf rs <entity> adapter-report` | hf.cluster.adapter.post.status.sh |
| 05 | [Tables and Lists](tables-and-lists/spec.md) | Combined and per-entity tables via `hf rs`, watch mode, spinner | hf.resources.sh → `hf rs` |
| 13 | [RS Entity Commands](rs-entity-commands/spec.md) | Canonical `hf rs <entity>` command tree for clusters, nodepools, and config-defined types | — |
| 06 | [Database](database/spec.md) | Direct PostgreSQL operations (query, delete, config) | hf.db.{query,delete,config}.sh |
| 07 | [Maestro](maestro/spec.md) | Maestro resource management via HTTP API | hf.maestro.{list,bundles,consumers,get,delete}.sh |
| 08 | [Pub/Sub & Messaging](pubsub/spec.md) | Event publishing to GCP Pub/Sub and RabbitMQ | hf.pubsub.{list,publish.*}.sh, hf.rabbitmq.publish.*.sh |
| 09 | [Kubernetes](kubernetes/spec.md) | Port-forwarding, debug pods, log tailing, log insights | hf.kube.{port.forward,curl,debug.pod}.sh, hf.logs.{sh,adapter}.sh |
| 10 | [Repos](repos/spec.md) | GitHub repository status overview | hf.repos.sh |
| 11 | [Errors & Usage](errors-and-usage/spec.md) | Error handling, usage messages, edge cases | Cross-cutting across all commands |
| 12 | ~~Config Registry~~ | **ARCHIVED** — superseded by [Config Model](config-model/spec.md) | — |

### Technical & Non-Functional Requirements

| # | Domain | Spec | Description |
|---|--------|------|-------------|
| T1 | [Command Hierarchy](command-hierarchy/spec.md) | Go module structure, Cobra command tree, shared packages, dependency bundling |
| T2 | [Configuration Model](config-model/spec.md) | Self-contained environment files, precedence chain, state.yaml, secret handling |
| T3 | [Non-Functional](non-functional/spec.md) | Shell completions, output format flag, cross-compilation, CI/CD pipelines, testing, security |
| T4 | [Output Formatting](output-formatting/spec.md) | Multi-format output dispatch, colored dot rendering, dynamic column ordering, JSON colorization |
| T5 | [Config Template](config-template/spec.md) | Bundled environment template embedding and consistency requirements |
| T6 | [Resource Types](resource-types/spec.md) | Canonical resource type definitions for clusters, nodepools, adapter statuses |
| T7 | [API Client](api-client/spec.md) | HTTP client contract, RFC 7807 error parsing, request/response conventions |
| T8 | [Generic Resource Lifecycle](generic-resource-lifecycle/spec.md) | Config-driven `hf rs` overview and CRUD for arbitrary API types (channels, versions, etc.) |
| T9 | [RS Entity Commands](rs-entity-commands/spec.md) | Full `hf rs <entity>` subcommand contract (table, conditions, statuses, adapter-report, force-delete) |

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | **Go** | Single binary, strong k8s ecosystem, cross-platform |
| CLI Framework | **Cobra** | Industry standard (kubectl, gh, docker), subcommand trees, auto-completions |
| Config Format | **Environment files + state** | Self-contained `environments/<name>.yaml` seeded from bundled template; `state.yaml` for active state |
| K8s Client | **client-go (bundled)** | Self-contained binary, no kubectl dependency |
| DB Driver | **pgx** | Native Go PostgreSQL driver, no psql needed |
| GCP Pub/Sub | **Cloud Go SDK** | Official library, no gcloud dependency |
| GitHub | **go-github** | REST client, no gh dependency |
| Build/Release | **GoReleaser** | Cross-compile linux/mac/windows amd64/arm64 |
| Extensibility | **None** | Core binary only; no plugin system |

## Dependency Bundling

| Former External Tool | Go Replacement | Status |
|---------------------|----------------|--------|
| jq | encoding/json (stdlib) | Bundled |
| curl | net/http (stdlib) | Bundled |
| awk/sed | text/tabwriter + strings | Bundled |
| lsof/ss | net.Listen / os.FindProcess | Bundled |
| psql | jackc/pgx/v5 | Bundled |
| kubectl | k8s.io/client-go | Bundled |
| gcloud | cloud.google.com/go/pubsub | Bundled |
| gh | google/go-github/v60 | Bundled |
| stern | client-go log streaming | Bundled |
| maestro-cli | net/http (Maestro HTTP API) | Bundled |

## Key Design Patterns

1. **Environment files**: Self-contained `environments/<name>.yaml` (seeded from bundled template) for all settings; `state.yaml` for active cluster/nodepool/environment
2. **`hf env` as top-level command group**: `hf env create|list|show|activate|delete` — not nested under `hf config`
3. **Template-based creation**: `hf rs clusters create` and `hf rs nodepools create` use templates under `~/.config/hf/templates/`; `--name` overrides the name, `--file` uses a custom template
4. **`--output` flag everywhere**: `--output json|table|yaml` on every data-producing command; no `--table` flag
5. **`hf rs` overview**: Default table output for combined cluster+nodepool view when `clusters` and `nodepools` are configured; `--watch` for live refresh
6. **Defaults over usage**: Create commands with no args use embedded defaults, not a usage message
7. **Generation tracking**: Resources track generation; adapters report observed_generation
8. **Convergence logic**: Reconciled becomes True when ALL required adapters report Available=True at current generation
9. **Zero external deps for core**: Only GCP credentials needed for Pub/Sub; all other commands are fully self-contained
10. **RFC 7807 errors**: API errors follow Problem Details format; CLI exits 0 and outputs the error body
11. **Config precedence**: flags > env vars > active environment file > built-in defaults

## API Base Path

All HyperFleet API calls use: `/api/hyperfleet/v1/`

## Environment Context

From the recording environment:

```json
{
  "api_url": "http://localhost:8000",
  "api_version": "v1",
  "context": "kind-kind",
  "namespace": "hyperfleet",
  "db_host": "localhost",
  "db_port": 5432,
  "db_name": "hyperfleet",
  "maestro_http": "http://localhost:8100",
  "maestro_grpc": "localhost:8090",
  "gcp_project": "hcm-hyperfleet"
}
```
