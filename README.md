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
hf config env create eu

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
```

Or set individual keys from the CLI:

```bash
hf config set kubernetes.context gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-eu1
hf config set database.password my-secret
```

Verify the resolved configuration:

```bash
hf config show
```

```
/home/user/.config/hf/environments/eu.yaml
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

> **Tip:** Predefined services (`hyperfleet-api`, `postgresql`, `maestro-http`, `maestro-grpc`) are resolved from your environment profile. For any other pod pass `<pod-pattern> <localPort>:<remotePort>`.

### 4. Create a cluster and node pool

On first run, `hf cluster create` writes a default JSON template to `~/.config/hf/cluster-template.json`. Edit it once, then reuse it for every new cluster.

```bash
# Create with positional args (name, region, version)
hf cluster create prod-eu-west1 europe-west1 4.16.0
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
hf nodepool create workers-n2 --type n2-standard-4 --replicas 3
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

### 5. Watch cluster state with `hf table --watch`

`hf table` (alias for `hf resources`) renders a live combined view of all clusters and their node pools, with one adapter column per registered adapter. Node pool rows are indented under their parent cluster.

```bash
hf table                        # one-shot snapshot
hf table --watch                # refresh every 5 s (default)
hf table --watch --seconds 10   # refresh every 10 s
```

Screenshot of `hf table --watch` against a live EU cluster:

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
hf table -o json
hf table -o yaml
```

---

## Environment management

```bash
hf config env list              # list all environments
hf config env create staging    # create + activate
hf config env activate prod     # switch active environment
hf config env delete staging    # remove a profile
hf config env show eu           # inspect a named profile
```

Config precedence (highest to lowest):

```
CLI flags  >  HF_* env vars  >  environment profile  >  config.yaml  >  defaults
```

---

## Command reference

| Command | Description |
|---|---|
| `hf cluster create [name] [region] [version]` | Create a cluster (uses template + flag overrides) |
| `hf cluster list` | List all clusters |
| `hf cluster get [id]` | Get a cluster by ID (defaults to active) |
| `hf cluster search <name>` | Find by name and set as active context |
| `hf cluster delete [id]` | Delete a cluster |
| `hf cluster conditions [id]` | Show status conditions |
| `hf cluster statuses [id]` | Show per-adapter reconciliation status |
| `hf cluster patch spec\|labels` | Increment the `counter` field |
| `hf nodepool create [name]` | Create a node pool under the active cluster |
| `hf nodepool list` | List node pools for the active cluster |
| `hf nodepool get [id]` | Get a node pool by ID |
| `hf nodepool search <name>` | Find by name and set as active context |
| `hf nodepool delete <id>` | Delete a node pool |
| `hf table [--watch] [-s N]` | Combined cluster + node pool overview |
| `hf resources [--watch] [-s N]` | Alias for `hf table` |
| `hf kube port-forward start [name]` | Start port-forward(s) to in-cluster services |
| `hf kube port-forward stop [name]` | Stop port-forward(s) |
| `hf kube port-forward status` | Show running port-forward status |
| `hf kube curl -- <flags> <url>` | Run curl from inside the cluster |
| `hf kube debug <deployment>` | Exec into a debug pod |
| `hf db query <sql>` | Run a SQL query against the HyperFleet database |
| `hf maestro list` | List Maestro resource bundles |
| `hf config show` | Display the resolved active configuration |
| `hf config set <section.key> <value>` | Set a config value |
| `hf config get <section.key>` | Read a single config value |
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
