# hf — User Guide

`hf` is a self-contained CLI for managing HyperFleet clusters. It replaces a suite of bash scripts with a single binary — no `kubectl`, `psql`, `gcloud`, or other external tools required.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Configuration](#2-configuration)
3. [Shell Completion](#3-shell-completion)
4. [Cluster Management](#4-cluster-management)
5. [NodePool Management](#5-nodepool-management)
6. [Resources Overview](#6-resources-overview)
7. [Kubernetes Operations](#7-kubernetes-operations)
8. [Maestro Operations](#8-maestro-operations)
9. [Database Operations](#9-database-operations)
10. [Log Streaming](#10-log-streaming)
11. [Output Formats](#11-output-formats)
12. [Environment Variables](#12-environment-variables)
13. [Config Key Reference](#13-config-key-reference)

---

## 1. Prerequisites

- `hf` binary in your `PATH`
- A kubeconfig file (for `hf kube` commands) — set via `KUBECONFIG` env var or `--kubeconfig` flag
- A GCP access token (for GKE clusters without `gke-gcloud-auth-plugin`) — set via `HF_KUBE_TOKEN`

```bash
$ hf version
3d4e772
```

---

## 2. Configuration

`hf` stores configuration in `~/.config/hf/`. Each named **environment** is a self-contained YAML file in `~/.config/hf/environments/`. One environment is active at a time.

### 2.1 Create an environment

```bash
$ hf config env create prod
Environment 'prod' created and activated.
Edit your configuration: ~/.config/hf/environments/prod.yaml
```

### 2.2 List environments

```bash
$ hf config env list
NAME     ACTIVE
prod     ✓
staging
```

### 2.3 Switch environment

```bash
$ hf config env activate staging
Active environment set to 'staging'.
```

### 2.4 Show active configuration

`hf config show` prints the fully resolved configuration with secrets masked.

```bash
$ hf config show
~/.config/hf/environments/prod.yaml
hyperfleet:
  api-url: http://34.175.77.98:8000
  api-version: v1
  gcp-project: hcm-hyperfleet
  namespace: hyperfleet-e2e-amarin
  token: <not set>
kubernetes:
  context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
maestro:
  consumer: cluster1
  grpc-endpoint: localhost:8090
  http-endpoint: http://localhost:8100
  namespace: maestro
port-forward:
  api-port: 8000
  maestro-grpc-port: 8090
  maestro-grpc-remote-port: 8090
  maestro-http-port: 8100
  maestro-http-remote-port: 8000
  pg-port: 5432
database:
  host: localhost
  name: hyperfleet
  password: <set>
  port: 5432
  user: hyperfleet
rabbitmq:
  host: localhost
  mgmt-port: 15672
  password: <set>
  user: guest
  vhost: /
registry:
  name:
────────────────────────────────────────
state:
  active-environment: prod
```

### 2.5 Get a single value

```bash
$ hf config get hyperfleet.api-url
http://34.175.77.98:8000

$ hf config get hyperfleet.namespace
hyperfleet-e2e-amarin
```

### 2.6 Set a value

```bash
$ hf config set hyperfleet.api-url http://34.175.77.98:8000
$ hf config set hyperfleet.namespace hyperfleet-e2e-amarin
$ hf config set maestro.consumer cluster1
$ hf config set kubernetes.context gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
```

Key format is always `section.key`. Sections: `hyperfleet`, `kubernetes`, `maestro`, `port-forward`, `database`, `rabbitmq`, `registry`.

---

## 3. Shell Completion

### Load in the current session

```bash
# Bash
source <(hf completion bash)

# Zsh
source <(hf completion zsh)

# Fish
hf completion fish | source

# PowerShell
hf completion powershell | Out-String | Invoke-Expression
```

### Install permanently

```bash
# Bash
hf completion bash > /etc/bash_completion.d/hf

# Zsh
hf completion zsh > "${fpath[1]}/_hf"

# Fish
hf completion fish > ~/.config/fish/completions/hf.fish
```

After loading, `hf <TAB>` completes commands, and `hf cluster <TAB>` completes subcommands.

---

## 4. Cluster Management

All cluster commands talk to the HyperFleet REST API (`hyperfleet.api-url`).

### 4.1 List clusters

```bash
$ hf cluster list --output table
NAME          ID                                    REGION          VERSION  STATUS
my-cluster    019dbf43-65c5-7562-9077-e0a2331a1070  eu-west-1       1.30     Available

$ hf cluster list --output json
{
  "items": [
    {
      "id": "019dbf43-65c5-7562-9077-e0a2331a1070",
      "name": "my-cluster",
      ...
    }
  ],
  "total": 1
}
```

Use `--watch` to continuously refresh:

```bash
$ hf cluster list --output table --watch --seconds 10
```

### 4.2 Get a cluster

```bash
$ hf cluster get 019dbf43-65c5-7562-9077-e0a2331a1070

# Search by name and set as active context
$ hf cluster search my-cluster
```

`search` stores the matched cluster ID and name in state so subsequent commands targeting "the active cluster" work without passing the ID every time.

### 4.3 Create a cluster

```bash
# Using defaults from template
$ hf cluster create my-cluster eu-west-1 1.30

# Using a custom JSON template
$ hf cluster create --file /path/to/cluster.json
```

### 4.4 Update and patch

```bash
# Full update from a template file
$ hf cluster update --file cluster.json

# Increment a counter field
$ hf cluster patch spec.desiredCount
$ hf cluster patch labels.retryCount
```

### 4.5 Delete

```bash
$ hf cluster delete 019dbf43-65c5-7562-9077-e0a2331a1070
```

### 4.6 Status conditions and adapter statuses

```bash
$ hf cluster conditions
$ hf cluster statuses
$ hf cluster statuses --output table
```

---

## 5. NodePool Management

NodePool commands mirror cluster commands and also talk to the HyperFleet API.

### 5.1 List nodepools

```bash
$ hf nodepool list --output table
NAME           ID                                    CLUSTER-ID   VERSION  REPLICAS
my-nodepool    019dbf43-7199-7ea6-b786-d617fc793c28  019dbf43...  1.30     3
```

### 5.2 Create a nodepool

```bash
$ hf nodepool create my-nodepool eu-west-1 1.30

# Attach to a specific cluster and set replicas
$ hf nodepool create --name my-nodepool --nodepool-id <id> --replicas 3
```

### 5.3 Search, update, patch, delete

```bash
$ hf nodepool search my-nodepool
$ hf nodepool update --file nodepool.json
$ hf nodepool patch spec.desiredCount
$ hf nodepool delete 019dbf43-7199-7ea6-b786-d617fc793c28
```

### 5.4 Status

```bash
$ hf nodepool conditions
$ hf nodepool statuses --output table
```

---

## 6. Resources Overview

`hf resources` (alias: `hf table`) shows a combined cluster + nodepool table in one view.

```bash
$ hf resources
$ hf resources --watch           # auto-refresh every 5s
$ hf resources --watch --seconds 10
```

---

## 7. Kubernetes Operations

`hf kube` provides Kubernetes operations without requiring `kubectl`. Set `KUBECONFIG` and, for GKE clusters, set `HF_KUBE_TOKEN` to a valid GCP access token to bypass the `gke-gcloud-auth-plugin`.

```bash
export KUBECONFIG=~/.kube/config
export HF_KUBE_TOKEN=$(gcloud auth print-access-token)
```

### 7.1 Port-forwards

`hf kube port-forward` manages background port-forward processes to the four predefined in-cluster services. PID files are stored at `~/.config/hf/pf-<name>.pid`.

| Service | Pod pattern | Namespace config | Local port | Remote port |
|---|---|---|---|---|
| `hyperfleet-api` | `hyperfleet-api` | `hyperfleet.namespace` | 8000 | 8000 |
| `postgresql` | `postgresql` | `hyperfleet.namespace` | 5432 | 5432 |
| `maestro-http` | `maestro` | `maestro.namespace` | 8100 | 8000 |
| `maestro-grpc` | `maestro` | `maestro.namespace` | 8090 | 8090 |

#### Start all port-forwards

```bash
$ hf kube port-forward start
[INFO] Kubernetes context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
[INFO] Started hyperfleet-api (hyperfleet-e2e-amarin/api-hyperfleet-api-dc948665f-kk296): localhost:8000 → 8000 (pid 21095)
[INFO] Started postgresql (hyperfleet-e2e-amarin/api-hyperfleet-api-postgresql-78487b6d54-hw5r7): localhost:5432 → 5432 (pid 21101)
[INFO] Started maestro-http (maestro/maestro-64b5f757f9-gbsps): localhost:8100 → 8000 (pid 21107)
[INFO] Started maestro-grpc (maestro/maestro-64b5f757f9-gbsps): localhost:8090 → 8090 (pid 21113)
  ✓ hyperfleet-api - localhost:8000 (PID: 21095)
  ✓ maestro-grpc   - localhost:8090 (PID: 21113)
  ✓ maestro-http   - localhost:8100 (PID: 21107)
  ✓ postgresql     - localhost:5432 (PID: 21101)
```

After starting, `hf` waits one second then shows connectivity status using protocol-aware checks (HTTP, pgx Ping, gRPC Health).

#### Start a single service

```bash
$ hf kube port-forward start maestro-http
[INFO] Kubernetes context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
[INFO] Started maestro-http (maestro/maestro-64b5f757f9-dw74p): localhost:8100 → 8000 (pid 6398)
  ✓ maestro-http - localhost:8100 (PID: 6398)
```

#### Check status

```bash
$ hf kube port-forward status
[INFO] Kubernetes context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
  ✓ hyperfleet-api - localhost:8000 (PID: 21095)
  ✓ maestro-grpc   - localhost:8090 (PID: 21113)
  ✓ maestro-http   - localhost:8100 (PID: 21107)
  ✓ postgresql     - localhost:5432 (PID: 21101)
```

A red ✗ means the service is not reachable. Each service is checked with its native protocol: HTTP GET for API services, pgx Ping for PostgreSQL, gRPC Health/Check for maestro-grpc.

#### Stop

```bash
$ hf kube port-forward stop           # stop all
$ hf kube port-forward stop maestro-http  # stop one

[INFO] Stopped maestro-grpc
[INFO] Stopped maestro-http
[INFO] Stopped postgresql
```

### 7.2 In-cluster curl

Run a `curl` command from inside the cluster (useful for testing internal service endpoints):

```bash
$ hf kube curl http://hyperfleet-api:8000/api/hyperfleet/v1/clusters
```

### 7.3 Debug pod

Exec into a debug container spun up from the deployment template:

```bash
$ hf kube debug
```

---

## 8. Maestro Operations

`hf maestro` talks to the Maestro HTTP API (configured at `maestro.http-endpoint`). Port-forward `maestro-http` first.

### 8.1 List resource bundles for the configured consumer

```bash
$ hf maestro list
{
  "items": [
    {
      "id": "e2f1474b-648f-58bf-9e0e-9ea0bc0360c2",
      "name": "e2f1474b-648f-58bf-9e0e-9ea0bc0360c2",
      "consumer_name": "cluster1",
      "version": 1,
      "manifest_count": 0,
      "manifests": [
        { "kind": "Namespace", "name": "", "namespace": "" },
        { "kind": "ConfigMap", "name": "", "namespace": "" }
      ],
      "conditions": null
    }
  ],
  "kind": "ResourceBundleList",
  "total": 8
}
```

#### Table view — hierarchical tree

```bash
$ hf maestro list --output table
e2f1474b-648f-58bf-9e0e-9ea0bc0360c2  e2f1474b-648f-58bf-9e0e-9ea0bc0360c2  v1
  Namespace/my-ns  
  ConfigMap/my-cfg  default
28bcbc9b-b89a-522f-b826-61f3eaec7f89  28bcbc9b-b89a-522f-b826-61f3eaec7f89  v1
  Namespace/other-ns  
  ConfigMap/other-cfg  default
```

Each bundle line shows `id  name  v<version>`. Child lines show `Kind/name  namespace` indented by two spaces.

#### YAML view

```bash
$ hf maestro list --output yaml
items:
  - id: e2f1474b-648f-58bf-9e0e-9ea0bc0360c2
    name: e2f1474b-648f-58bf-9e0e-9ea0bc0360c2
    consumername: cluster1
    version: 1
    manifestcount: 0
    manifests:
      - kind: Namespace
        name: ""
        namespace: ""
      - kind: ConfigMap
        name: ""
        namespace: ""
    conditions: []
```

### 8.2 Get a specific bundle

```bash
$ hf maestro get e2f1474b-648f-58bf-9e0e-9ea0bc0360c2
{
  "id": "e2f1474b-648f-58bf-9e0e-9ea0bc0360c2",
  "name": "e2f1474b-648f-58bf-9e0e-9ea0bc0360c2",
  "consumer_name": "cluster1",
  "version": 1,
  "manifest_count": 0,
  "manifests": [
    { "kind": "Namespace", "name": "", "namespace": "" },
    { "kind": "ConfigMap", "name": "", "namespace": "" }
  ],
  "conditions": null
}
```

Running `hf maestro get` without an argument shows an interactive selection menu.

### 8.3 List all bundles (no consumer filter)

```bash
$ hf maestro bundles
```

### 8.4 List consumers

```bash
$ hf maestro consumers
{
  "items": [
    {
      "id": "5c9892e1-7ca7-4d70-be29-134af4430aa8",
      "kind": "Consumer",
      "name": "cluster1"
    }
  ],
  "kind": "ConsumerList",
  "total": 1
}
```

### 8.5 Delete a bundle

```bash
$ hf maestro delete e2f1474b-648f-58bf-9e0e-9ea0bc0360c2

# Delete all (prompts for confirmation)
$ hf maestro delete --all
3 resource bundle(s) will be deleted:
  - e2f1474b-... (e2f1474b-...)
  - 28bcbc9b-... (28bcbc9b-...)
  - 2da5e54e-... (2da5e54e-...)
Type 'yes' to confirm deletion: yes
```

---

## 9. Database Operations

`hf db` runs SQL queries and DML directly against the PostgreSQL database. Port-forward `postgresql` first (`hf kube port-forward start postgresql`).

### 9.1 Show connection parameters

```bash
$ hf db config
host=localhost port=5432 dbname=hyperfleet user=hyperfleet sslmode=disable
```

### 9.2 Run a query

```bash
$ hf db query --file query.sql --output table

# Inline SQL
$ hf db query "SELECT id, name FROM clusters LIMIT 5" --output table
ID                                    NAME
019dbf43-65c5-7562-9077-e0a2331a1070  my-cluster
```

### 9.3 Execute DML

```bash
$ hf db exec "UPDATE clusters SET spec = spec WHERE id = '019dbf43-...'"
Rows affected: 1
```

### 9.4 Delete records

```bash
$ hf db delete clusters            # deletes all rows from clusters table
$ hf db delete clusters --all      # same, skips confirmation prompt
$ hf db delete adapter_statuses
```

---

## 10. Log Streaming

### 10.1 Stream pod logs by pattern

```bash
$ hf logs hyperfleet-api
```

If `stern` is on your `PATH`, it is used; otherwise `hf` fans out goroutines to collect logs from all matching pods.

### 10.2 Stream adapter logs for the active cluster

```bash
$ hf logs adapter
```

### 10.3 Summarise recent activity (AI insights)

```bash
$ hf logs insights
```

---

## 11. Output Formats

Every data-producing command supports `--output` / `-o`:

| Flag | Description |
|---|---|
| `-o json` | Pretty-printed JSON with colour (default) |
| `-o yaml` | YAML with 2-space indent |
| `-o table` | Human-readable table (or hierarchical tree for `maestro list`) |

```bash
$ hf maestro consumers -o json      # default
$ hf maestro consumers -o yaml
$ hf cluster list -o table
$ hf cluster list -o table --watch  # live-refresh table
```

Colour is auto-disabled for non-TTY output (pipes, redirects) and when `NO_COLOR` is set.

---

## 12. Environment Variables

| Variable | Effect |
|---|---|
| `HF_CONFIG_DIR` | Override the config directory (default `~/.config/hf`) |
| `HF_API_URL` | Override `hyperfleet.api-url` for this invocation |
| `HF_API_VERSION` | Override `hyperfleet.api-version` |
| `HF_TOKEN` | Override `hyperfleet.token` (API bearer token) |
| `HF_CONTEXT` | Override `kubernetes.context` |
| `HF_NAMESPACE` | Override `hyperfleet.namespace` |
| `HF_KUBE_TOKEN` | GCP access token injected as the kubeconfig bearer token — bypasses `gke-gcloud-auth-plugin` |
| `KUBECONFIG` | Path to kubeconfig file |
| `NO_COLOR` | Disable all ANSI colour output |

Environment variables take precedence over the active environment file.

**GKE example** (no `gke-gcloud-auth-plugin` required):

```bash
export HF_KUBE_TOKEN=$(gcloud auth print-access-token)
export KUBECONFIG=~/.kube/gke-eu.yaml
hf kube port-forward start
```

---

## 13. Config Key Reference

| Section | Key | Default | Description |
|---|---|---|---|
| `hyperfleet` | `api-url` | `http://localhost:8000` | HyperFleet REST API base URL |
| `hyperfleet` | `api-version` | `v1` | API version path segment |
| `hyperfleet` | `token` | _(empty)_ | Bearer token for the API |
| `hyperfleet` | `gcp-project` | `hcm-hyperfleet` | GCP project ID |
| `hyperfleet` | `namespace` | `hyperfleet` | Kubernetes namespace for HyperFleet app pods |
| `kubernetes` | `context` | _(empty)_ | kubeconfig context name |
| `maestro` | `http-endpoint` | `http://localhost:8100` | Maestro HTTP API base URL |
| `maestro` | `grpc-endpoint` | `localhost:8090` | Maestro gRPC endpoint |
| `maestro` | `consumer` | `cluster1` | Consumer name filter for `hf maestro list` |
| `maestro` | `namespace` | `maestro` | Kubernetes namespace for Maestro pods |
| `port-forward` | `api-port` | `8000` | Local port for `hyperfleet-api` |
| `port-forward` | `pg-port` | `5432` | Local port for `postgresql` |
| `port-forward` | `maestro-http-port` | `8100` | Local port for `maestro-http` |
| `port-forward` | `maestro-grpc-port` | `8090` | Local port for `maestro-grpc` |
| `database` | `host` | `localhost` | PostgreSQL host (use `localhost` with port-forward) |
| `database` | `port` | `5432` | PostgreSQL port |
| `database` | `name` | `hyperfleet` | Database name |
| `database` | `user` | `hyperfleet` | Database user |
| `database` | `password` | _(set via config)_ | Database password — never hardcode |
