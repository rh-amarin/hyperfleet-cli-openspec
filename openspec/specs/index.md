# HyperFleet CLI Requirements Specification

## Purpose

Complete requirements specification for the HyperFleet CLI tool (`hf`), reverse-engineered from the shell scripts at [rh-amarin/hyperfleet-cli](https://github.com/rh-amarin/hyperfleet-cli) and extended with technical architecture, configuration model, and non-functional requirements for the Go reimplementation.

## Specification Index

### Functional Requirements

Organized to match the [output index](https://github.com/rh-amarin/hyperfleet-cli/blob/main/scripts/output/00-index.json):

| # | Domain | Spec | Req | Scenarios | Scripts Covered |
|---|--------|------|-----|-----------|-----------------|
| 01 | [Configuration](config/spec.md) | Config management, env profiles, diagnostics | 6 | 17 | hf.config.sh, hf.cluster.id.sh, hf.nodepool.id.sh |
| 02 | [Cluster Lifecycle](cluster-lifecycle/spec.md) | Cluster CRUD operations | 8 | 21 | hf.cluster.{create,search,get,patch,delete,conditions,conditions.table,statuses}.sh |
| 03 | [NodePool Lifecycle](nodepool-lifecycle/spec.md) | NodePool CRUD operations | 10 | 23 | hf.nodepool.{create,list,search,get,patch,delete,conditions,conditions.table,statuses,table}.sh |
| 04 | [Adapter Status](adapter-status/spec.md) | Adapter status posting and convergence model | 3 | 11 | hf.cluster.adapter.post.status.sh, hf.nodepool.adapter.post.status.sh |
| 05 | [Tables and Lists](tables-and-lists/spec.md) | Aggregated views and formatted tables | 4 | 6 | hf.cluster.{list,table}.sh, hf.nodepool.table.sh, hf.resources.sh |
| 06 | [Database](database/spec.md) | Direct PostgreSQL operations | 3 | 12 | hf.db.{query,delete,delete.all,config}.sh |
| 07 | [Maestro](maestro/spec.md) | Maestro resource management via HTTP API | 5 | 8 | hf.maestro.{list,bundles,consumers,get,delete}.sh |
| 08 | [Pub/Sub & Messaging](pubsub/spec.md) | Event publishing to GCP Pub/Sub and RabbitMQ | 5 | 10 | hf.pubsub.{list,publish.*}.sh, hf.rabbitmq.publish.*.sh |
| 09 | [Kubernetes](kubernetes/spec.md) | Port-forwarding, debugging, log tailing | 5 | 10 | hf.kube.{port.forward,context,curl,debug.pod}.sh, hf.logs.{sh,adapter}.sh |
| 10 | [Repos](repos/spec.md) | GitHub repository status overview | 1 | 3 | hf.repos.sh |
| 11 | [Errors & Usage](errors-and-usage/spec.md) | Error handling, usage messages, edge cases | 6 | 11 | Cross-cutting across all commands |
| 12 | [Config Registry](config-registry/spec.md) | Configuration property registry and storage model | 2 | 4 | hf.lib.sh (shared library) |

### Technical & Non-Functional Requirements

| # | Domain | Spec | Req | Scenarios |
|---|--------|------|-----|-----------|
| T1 | [Technical Architecture](technical-architecture/spec.md) | Go module structure, Cobra command tree, shared packages, dependency bundling | 10 | 19 |
| T2 | [Configuration Model](config-model/spec.md) | Self-contained environment files, precedence chain, state.yaml, secret handling | 8 | 18 |
| T5 | [Config Template](config-template/spec.md) | Bundled environment template embedding and consistency requirements | 1 | 2 |
| T3 | [Non-Functional](non-functional/spec.md) | Shell completions, output format flag, cross-compilation, testing, security | 9 | 26 |
| T4 | [Output Formatting](output-formatting/spec.md) | Multi-format output dispatch, colored dot rendering, dynamic column ordering, JSON colorization | 5 | 18 |

### Summary

| Category | Requirements | Scenarios |
|----------|-------------|-----------|
| Functional (01–12) | 58 | 136 |
| Technical & NFR (T1–T4) | 31 | 79 |
| **Total** | **89** | **215** |

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | **Go** | Single binary, strong k8s ecosystem, cross-platform |
| CLI Framework | **Cobra** | Industry standard (kubectl, gh, docker), subcommand trees, auto-completions |
| Config Format | **Environment files + state** | Self-contained `environments/<name>.yaml` seeded from bundled template; `state.yaml` for active state (top-level keys) |
| K8s Client | **client-go (bundled)** | Self-contained binary, no kubectl dependency |
| DB Driver | **pgx** | Native Go PostgreSQL driver, no psql needed |
| GCP Pub/Sub | **Cloud Go SDK** | Official library, no gcloud dependency |
| GitHub | **go-github** | REST client, no gh dependency |
| Build/Release | **GoReleaser** | Cross-compile linux/mac/windows amd64/arm64 |
| Extensibility | **None** | Core binary only; no plugin system |

## Dependency Bundling

| Former External Tool | Go Replacement | Status |
|---------------------|---------------|--------|
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
2. **Shared internal functions**: Commands reuse `internal/` packages (e.g., `api.FindClusterByName`, `config.SetClusterID`) rather than invoking each other as subprocesses
3. **Defaults over usage**: Create commands with no args use defaults, not usage display
4. **Generation tracking**: Resources track generation; adapters report observed_generation
5. **Convergence logic**: Reconciled becomes True when ALL required adapters report Available=True at current generation
6. **Multi-format output**: `--output json|table|yaml` on every data-producing command
7. **Zero external deps for core**: Only GCP credentials needed for Pub/Sub commands; all other commands are fully self-contained
9. **RFC 7807 errors**: API errors follow Problem Details format
10. **Config precedence**: flags > env vars > active environment file > built-in defaults

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

## API Base Path

All HyperFleet API calls use: `/api/hyperfleet/v1/`
