# HyperFleet CLI — `hf`

A self-contained Go CLI for managing HyperFleet clusters. `hf` replaces a suite of bash scripts with a single binary — no `kubectl`, `psql`, `gcloud`, or `gh` required.

---

## Quickstart

### 1. Build

Prerequisites: Go 1.22+.

```bash
git clone https://github.com/rh-amarin/hyperfleet-cli
cd hyperfleet-cli
make build          # produces ./bin/hf
./bin/hf version
```

Move the binary to your `PATH`:

```bash
sudo mv bin/hf /usr/local/bin/hf
```

### 2. Create and configure an environment

`hf` uses *environment profiles* — named YAML files under `~/.config/hf/environments/` — to store per-cluster settings. All commands require an active environment.

```bash
# Create a new environment and activate it immediately
hf env create eu

# The new profile is written to ~/.config/hf/environments/eu.yaml
# Edit it to point at your cluster:
```

Open the generated file and fill in the required fields:

```yaml
hyperfleet:
  api-url: "http://localhost:8000"    # set after port-forwarding (step 3)
  api-version: "v1"
  token: ""                           # bearer token if your API requires auth
  namespace: "hyperfleet"

kubernetes:
  context: "gke_my-project_europe-west1-b_my-cluster"   # kubectl context name

maestro:
  http-endpoint: "http://localhost:8100"
  grpc-endpoint: "localhost:8090"
  namespace: "maestro"

database:
  host: "localhost"
  port: "5432"
  name: "hyperfleet"
  user: "hyperfleet"
  password: "your-db-password"

# Optional: declare extra HyperFleet API types (see "Config-defined resources")
resource-types: {}
```

Inspect the active configuration (secrets are redacted):

```bash
hf env show
```

```
hyperfleet:
  api-url: http://localhost:8000
  api-version: v1
  namespace: hyperfleet
  token: <not set>
kubernetes:
  context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-eu1
maestro:
  grpc-endpoint: localhost:8090
  http-endpoint: http://localhost:8100
  namespace: maestro
port-forward:
  api-port: 8000
  pg-port: 5432
  ...
database:
  host: localhost
  name: hyperfleet
  password: <set>
  port: 5432
  user: hyperfleet
────────────────────────────────────────
state:
  active-environment: eu
────────────────────────────────────────
Environment file: /home/user/.config/hf/environments/eu.yaml [active]
State file:       /home/user/.config/hf/state.yaml

Edit these files to change configuration and runtime state.
```

### 3. Set up port-forwards

`hf` uses `client-go` to port-forward in-cluster services directly — no `kubectl` needed. Place your kubeconfig at the path referenced by `KUBECONFIG` (or `~/.kube/config`) and run:

```bash
hf kube port-forward start
```

This starts four background daemons in parallel:

```
[INFO] Kubernetes context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-eu1
[INFO] Started hyperfleet-api (hyperfleet/hyperfleet-api): localhost:8000 → 8000 (pid 42811)
[INFO] Started postgresql    (hyperfleet/postgresql):     localhost:5432 → 5432 (pid 42812)
[INFO] Started maestro-http  (maestro/maestro):           localhost:8100 → 8000 (pid 42813)
[INFO] Started maestro-grpc  (maestro/maestro):           localhost:8090 → 8090 (pid 42814)
  ✓ hyperfleet-api - localhost:8000 (PID: 42811)
  ✓ postgresql     - localhost:5432 (PID: 42812)
  ✓ maestro-http   - localhost:8100 (PID: 42813)
  ✓ maestro-grpc   - localhost:8090 (PID: 42814)
```

Check status or stop at any time:

```bash
hf kube port-forward status
hf kube port-forward stop          # stop all
hf kube port-forward stop maestro-http   # stop one
```

Start a single named forward with a custom port mapping:

```bash
hf kube port-forward start hyperfleet-api 9000:8000
```

> **Tip:** Predefined services (`hyperfleet-api`, `postgresql`, `maestro-http`, `maestro-grpc`, `rabbitmq`) are resolved from your environment profile. For any other pod pass `<pod-pattern> <localPort>:<remotePort>`.

### 4. Create a cluster and node pool

Create templates live under `~/.config/hf/templates/` (`clusters.json`, `nodepools.json` are seeded in the config template). Embedded defaults apply when a file is missing.

```bash
# Create with positional args (name, region, version)
hf rs clusters create prod-eu-west1 europe-west1 4.16.0
```

```
[INFO] Cluster context set to 'e3f2a1b0-cafe-4dea-beef-000000000001'
{
  "id": "e3f2a1b0-cafe-4dea-beef-000000000001",
  "kind": "Cluster",
  "name": "prod-eu-west1",
  "generation": 1,
  "spec": {
    "region": "europe-west1",
    "version": "4.16.0"
  },
  "status": {
    "conditions": [
      { "type": "Available",  "status": "False", "observed_generation": 1 },
      { "type": "Reconciled", "status": "False", "observed_generation": 1 }
    ]
  },
  "created_time": "2026-05-15T10:00:00Z"
}
```

The newly created cluster ID is stored in state and used automatically by subsequent commands. Create a node pool under it:

```bash
hf rs nodepools create workers-n2 --type n2-standard-4 --replicas 3
```

```
[INFO] NodePool context set to 'e3f2a1b0-cafe-4dea-0001-000000000001'
{
  "id": "e3f2a1b0-cafe-4dea-0001-000000000001",
  "kind": "NodePool",
  "name": "workers-n2",
  "generation": 1,
  "spec": {
    "platform": { "type": "n2-standard-4" },
    "replicas": 3
  },
  "status": {
    "conditions": [
      { "type": "Available",  "status": "False", "observed_generation": 1 },
      { "type": "Reconciled", "status": "False", "observed_generation": 1 }
    ]
  }
}
```

> **Tip:** Pass `--file path/to/template.json` to use a custom body. Any positional args or flags override template values.

### 5. Watch cluster state with `hf rs --watch`

`hf rs` renders a live combined view of all configured resource types in one table. Rows include `ID`, `NAME`, `KIND`, and `GEN`. Clusters and nodepools also show reconciled condition dots and per-adapter columns; other types (e.g. `Channel`, `Version`) use `-` in those columns. Child rows use tree prefixes (`├─` / `└─`).

```bash
hf rs                           # one-shot snapshot (table default)
hf rs --watch                   # refresh every 5 s (default)
hf rs --watch --seconds 10      # refresh every 10 s
```

Screenshot of `hf rs --watch` against a live EU cluster:

```
ID                                    NAME               GEN  AVAILABLE  RECONCILED  CL-CONFIGM  CL-DEPLOYM  NP-CONFIGM
                                                                                     AP          ENT         AP
a1b2c3d4-0001-0001-0001-000000000001  prod-eu-west1      7    ● 7        ● 7         ⠸ ● 7       ⠸ ● 7       -
  np000001-0001-0001-0001-000000000001  workers-n2       5    ● 5        ● 5         -           -           ⠸ ● 5
a1b2c3d4-0002-0002-0002-000000000002  staging-eu-west1   2    ● 2        ● 2         -           ⠸ ● 2       -

[watch] refreshing every 5s — press Ctrl-C to exit
```

> Columns are discovered dynamically from the cluster's adapter status responses, so the table expands automatically as new adapters register. The **⠸** spinner indicates adapters that reported within the last two refresh intervals; the green **●** dot indicates `Available: True`.

Other output formats are available when not in watch mode:

```bash
hf rs -o json
hf rs -o yaml
```

---

## Environment management

```bash
hf env list              # list all environments
hf env create staging    # create + activate
hf env activate prod     # switch active environment
hf env delete staging    # remove a profile
hf env show              # inspect the active profile
hf env show eu           # inspect a named profile
```

Config precedence (highest to lowest):

```
CLI flags  >  HF_* env vars  >  environment profile  >  defaults
```

---

## Config-defined resources (`resource-types`)

Besides built-in `cluster` and `nodepool` commands, `hf` can manage arbitrary HyperFleet API resource types declared in the active environment file. Each type becomes a subcommand under `hf resource` (alias `hf rs`).

Add a `resource-types` map to `~/.config/hf/environments/<name>.yaml`. New environments created with `hf env create` include an empty `resource-types: {}` stub you can fill in.

### Fields

| Field | Required | Description |
|---|---|---|
| `path` | yes | API path relative to `{api-url}/api/hyperfleet/{api-version}/` (e.g. `channels` or `channels/{channel_id}/versions`) |
| `parent` | no | Name of the immediate parent type; required for nested paths |
| `path-param` | no | Placeholder name this type's ID fills in child paths (defaults from entity name: `clusters` → `{cluster_id}`) |
| `create-template` | no | JSON filename under `~/.config/hf/templates/` used by `hf rs <type> create` |

The map key (entity name) is also the key in `state.yaml` when a resource is selected (e.g. type `channels` → state key `channels`).

Validation rules:

- `parent` must reference another defined type; cycles are rejected
- Root types (no `parent`) must not contain unresolved `{placeholders}` in `path`

### Example: root and nested types

```yaml
resource-types:
  channels:
    path: channels
    create-template: channels.json
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
  releases:
    parent: versions
    path: "channels/{channel_id}/versions/{version_id}/releases"
```

Place create payloads next to other templates:

```bash
mkdir -p ~/.config/hf/templates
# ~/.config/hf/templates/channels.json
```

```json
{
  "kind": "Channel",
  "name": "example"
}
```

Child types resolve ancestor IDs from `state.yaml`. If `channel-id` is missing, commands such as `hf rs versions list` fail with a hint to run `hf rs channels search <name>` first.

### State

`hf rs <type> search <name>` writes the matched ID to `state.yaml` under the entity name. Generic keys coexist with `clusters`, `nodepools`, and `active-environment`:

```yaml
# ~/.config/hf/state.yaml
active-environment: eu
clusters: e3f2a1b0-...
cluster-name: my-cluster
channels: abc-123
versions: ver-456
```

Inspect configured types and current state:

```bash
hf env show    # lists resource-types and state keys from the active environment
```

### Commands

| Command | Description |
|---|---|
| `hf rs` | Hierarchical overview of all configured types (table by default; `--watch` / `-s N`) |
| `hf rs <type> list` / `table` | List resources (`--search`, `--watch`, `-o json\|table\|yaml`) |
| `hf rs <type> get [id]` | Get one resource (`-i` to pick interactively) |
| `hf rs <type> search [name]` | Find by name and set active context |
| `hf rs <type> create [name]` | Create from template (`--file`, `--name`; clusters/nodepools have extra flags) |
| `hf rs <type> patch spec\|labels [id]` | Counter increment (no `--file`) or file-based patch (`--file`) |
| `hf rs <type> delete [id]` | Delete resource (`clusters delete --force --reason` for force-delete) |
| `hf rs <type> conditions [id]` | Show status conditions (clusters, nodepools) |
| `hf rs <type> statuses [id]` | Show adapter statuses (`--filter` for interactive filter) |
| `hf rs nodepools force-delete [id] --reason` | Force-delete a stuck nodepool |
| `hf rs <type> id` | Print or set (`-i`) active ID |
| `hf rs <type> adapter-report <adapter> <True\|False\|Unknown> <gen> [id]` | Simulate adapter status reporting |

Overview table example (one table; child rows use tree prefixes; `GEN` shows `❌` when `deleted_time` is set):

```
ID              NAME    KIND      GEN   RECONCILED  …
cl-1            prod    Cluster   2     ● 2         …
├─ np-1         workers NodePool  1     ● 1         …
abc-123         alpha   Channel   3     -           …
└─ ver-1        v1      Version   1     -           …
```

If some lists fail to load (missing parent state, API errors), `hf rs` still prints whatever it could fetch and shows `[WARN]` lines at the top describing each failure.

> When `clusters` and `nodepools` are configured under `resource-types`, `hf rs` (no subcommand) renders the adapter-rich overview (formerly `hf table` / `hf resources`) in a single table together with any other configured types (e.g. `channels`, `versions`).

---

## Command reference

| Command | Description |
|---|---|
| `hf rs` / `hf resource` | Combined overview (clusters + nodepools when configured) or generic type tree |
| `hf rs clusters …` / `hf rs nodepools …` | Full cluster and nodepool lifecycle (see [resource-types](#config-defined-resources-resource-types)) |
| `hf rs <type> …` | CRUD, conditions, statuses, adapter-report for any configured type |
| `hf kube port-forward start [name]` | Start port-forward(s) to in-cluster services |
| `hf kube port-forward stop [name]` | Stop port-forward(s) |
| `hf kube port-forward status` | Show running port-forward status |
| `hf kube curl -- <flags> <url>` | Run curl from inside the cluster |
| `hf kube debug <deployment>` | Exec into a debug pod |
| `hf db query <sql>` | Run a SQL query against the HyperFleet database |
| `hf maestro list` | List Maestro resource bundles |
| `hf env show [name]` | Display an environment profile (active when omitted) |
| `hf env create\|list\|activate\|delete` | Manage environment profiles |
| `hf version` | Print the binary version |
| `hf completion bash\|zsh\|fish` | Generate shell completion scripts |

Global flags available on every command:

```
--output, -o   json | table | yaml   (default: json; table commands default to table)
--no-color                           disable ANSI colour codes
--verbose, -v                        enable debug logging
--api-url                            override hyperfleet.api-url for one invocation
--api-token                          override bearer token for one invocation
--kubeconfig                         path to kubeconfig (hf kube commands only)
```

---

## Development

```bash
make build      # compile → ./bin/hf
make test       # go test ./...
make vet        # go vet ./...
make completions # generate shell completions into ./completions/
```

Integration tests require a live cluster and are gated with `//go:build integration`. Run them with:

```bash
go test -tags integration ./...
```

External binaries named `hf-<name>` on `PATH` are automatically delegated as plugins:

```bash
hf my-plugin arg1 arg2   # delegates to hf-my-plugin
```
