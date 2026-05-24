# Cluster Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet clusters, including creation, retrieval, listing, searching, patching, and deletion. All cluster operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters`.

## Requirements

### Requirement: Create Cluster

`hf cluster create` SHALL load the request body from a JSON template. The binary embeds a built-in default template (`cmd/assets/cluster-template.json`). When no `--file` flag is given, the CLI MUST use the embedded default bytes directly in memory — it MUST NOT read from or write to `<config-dir>`. The `--name` flag overrides only the `name` field in the template.

#### Scenario: Create cluster with default template

- GIVEN no `--file` flag is provided
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST use the built-in embedded template in memory
- AND MUST NOT write any file to `<config-dir>`
- AND MUST create the cluster using the default payload (`kind=Cluster`, `name=my-cluster`, default labels and spec)

#### Scenario: Create cluster with `--name` override

- GIVEN a template (embedded or from `--file`) with `"name": "<template-name>"`
- WHEN the user runs `hf cluster create --name <name>`
- THEN the CLI MUST set `name` to `<name>` in the request body, overriding the template value
- AND all other template fields MUST remain unchanged

#### Scenario: Create cluster with `--file` override

- GIVEN a file at `<path>` containing a valid JSON cluster payload
- WHEN the user runs `hf cluster create --file <path>`
- THEN the CLI MUST use that file's content as the request body
- AND `--name` MAY still be used to override the name field from that file

#### Scenario: Create cluster with no arguments

- GIVEN no flags and no arguments
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the embedded template payload

#### Scenario: Malformed template file

- GIVEN a file provided via `--file` containing invalid JSON
- WHEN the user runs `hf cluster create --file <path>`
- THEN the CLI MUST exit with `[ERROR] loading template: <reason>` and code 1

### Requirement: Search Cluster

The CLI SHALL search for clusters by name and set the found cluster as the current context.

#### Scenario: Search with no arguments

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster search` with no arguments
- THEN the CLI MUST behave identically to `hf cluster get` — fetching and returning the current cluster from state

#### Scenario: Search with no arguments and no cluster in state

- GIVEN no cluster-id is set in state
- WHEN the user runs `hf cluster search` with no arguments
- THEN the CLI MUST display `[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.`
- AND exit with code 1

#### Scenario: Search for existing cluster

- GIVEN clusters exist in the API
- WHEN the user runs `hf cluster search <name>`
- THEN the CLI MUST query the API filtering by name
- AND output the matching clusters as a JSON array of full Cluster objects
- AND persist the found cluster's ID to active state via `config.SetClusterID`
- AND print `[INFO] Cluster context set to '<id>'` on stderr after persisting

#### Scenario: Search for non-existent cluster

- GIVEN no cluster matches the search name
- WHEN the user runs `hf cluster search <name>`
- THEN the CLI MUST display `[WARN] No clusters found matching '<name>'`
- AND output an empty JSON array `[]`
- AND exit with code 0

#### Scenario: Multiple matches

- GIVEN multiple clusters match the search name
- WHEN the user runs `hf cluster search <name>`
- THEN the CLI MUST display `[WARN] Multiple clusters found matching '<name>', using first result`
- AND set cluster-id to the first element in the returned `items` array
- AND persist that cluster-id to active state

### Requirement: Get Cluster

The CLI SHALL retrieve and display full details of a specific cluster.

#### Scenario: Get current cluster

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster get`
- THEN the CLI MUST send a GET request to `/api/hyperfleet/v1/clusters/{cluster_id}`
- AND output the full cluster JSON including: id, kind, name, generation, labels, spec, status.conditions, created_by, created_time, updated_by, updated_time, href

#### Scenario: Get cluster by explicit ID

- GIVEN a valid cluster ID is provided
- WHEN the user runs `hf cluster get <cluster_id>`
- THEN the CLI MUST use the provided ID instead of the configured cluster-id

#### Scenario: Get non-existent cluster

- GIVEN an invalid or non-existent cluster ID is used
- WHEN the user runs `hf cluster get <invalid_id>`
- THEN the CLI MUST output the API error response (RFC 7807 format)
- AND the error MUST contain code `HYPERFLEET-NTF-001`, status 404, title `Resource Not Found`
- AND the CLI MUST exit with code 0
- NOTE: Exit code 0 for API errors is intentional to maintain backwards compatibility with the original shell scripts. All API errors exit 0 and output the error JSON. See `errors-and-usage/spec.md` and `command-hierarchy/spec.md` Error Handling Strategy.

**Example** — `hf cluster get 00000000-0000-0000-0000-000000000000`:
```json
{
  "code": "HYPERFLEET-NTF-001",
  "detail": "Cluster with id='00000000-0000-0000-0000-000000000000' not found",
  "status": 404,
  "title": "Resource Not Found",
  "type": "https://api.hyperfleet.io/errors/not-found"
}
```

### Requirement: Interactive Cluster View

The CLI SHALL support an interactive split-screen viewer for a cluster via `hf cluster get -i`. This is a **read-only viewer** — there is no text input or filter. It replaces the old "pick a cluster" behavior that `-i` had on other subcommands.

#### Scenario: Open interactive viewer

- GIVEN a cluster-id is set in config (or an explicit ID is provided)
- WHEN the user runs `hf cluster get -i` (or `hf cluster get <id> -i`)
- THEN the CLI MUST fetch the cluster and open an interactive split-screen view
- AND the left panel MUST display a compact summary of the cluster's key fields: `name`, `id`, `generation`, and a one-line status derived from `status.conditions` (e.g. `Available: False | Reconciled: False`)
- AND the right panel MUST display the complete cluster JSON with syntax coloring, scrollable with arrow keys
- AND there is NO text input field — the view is read-only; the user navigates the right panel directly
- AND pressing `q` or Escape MUST exit the view cleanly with code 0

#### Scenario: Interactive viewer — no cluster in state

- GIVEN no cluster-id is set in state and no explicit ID is provided
- WHEN the user runs `hf cluster get -i`
- THEN the CLI MUST display the standard missing-cluster-id error and exit 1

### Requirement: Patch Cluster

The CLI SHALL increment a counter field in the cluster's spec or labels section, triggering a generation bump.

#### Scenario: Patch spec counter

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster patch spec`
- THEN the CLI MUST fetch the current cluster
- AND read the current `spec.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}` with the full existing `spec` map, updating only `counter` to the incremented string value (e.g., `"2"`)
- AND MUST NOT omit other existing keys in the patched section — the API replaces the entire `spec`/`labels` object on PATCH
- AND display `[INFO] Incrementing spec.counter: <old> -> <new>` where `<old>` and `<new>` are integer strings (e.g., `1 -> 2`; first increment displays `0 -> 1`)
- AND the cluster's generation MUST increment

**Example** — `hf cluster patch spec` when current `spec.counter` is `"1"`:
```
[INFO] Incrementing spec.counter: 1 -> 2
```

#### Scenario: Patch labels counter

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster patch labels`
- THEN the CLI MUST fetch the current cluster
- AND read the current `labels.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}` with the full existing `labels` map, updating only `counter` to the incremented string value
- AND MUST NOT omit other existing keys in the patched section — the API replaces the entire `spec`/`labels` object on PATCH
- AND display `[INFO] Incrementing labels.counter: <old> -> <new>`
- AND the cluster's generation MUST increment

**Example** — `hf cluster patch labels` when current `labels.counter` is `"1"`:
```
[INFO] Incrementing labels.counter: 1 -> 2
```

#### Scenario: Patch with no arguments

- GIVEN the user provides no arguments
- WHEN the user runs `hf cluster patch`
- THEN the CLI MUST display usage: `Usage: hf cluster patch {spec|labels} [cluster_id]`
- AND exit with code 1

### Requirement: Delete Cluster

`hf cluster delete` SHALL accept an optional cluster ID, falling back to the configured cluster-id.

#### Scenario: Delete cluster

- WHEN the user runs `hf cluster delete [cluster_id]`
- THEN the CLI MUST use the provided ID, or the configured cluster-id if none is provided
- AND the CLI MUST output the deleted cluster JSON subject to the `--output` flag

### Requirement: Get Cluster Conditions

The CLI SHALL display the generation and status conditions of a cluster.

#### Scenario: Get conditions

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster conditions`
- THEN the CLI MUST fetch the cluster from `/api/hyperfleet/v1/clusters/{cluster_id}`
- AND extract only `generation` and `status.conditions`
- AND output them as JSON

**Example** — `hf cluster conditions` immediately after creation (no adapters yet):
```json
{
  "generation": 1,
  "status": {
    "conditions": [
      {
        "type": "Available",
        "status": "False",
        "reason": "AdaptersNotAtSameGeneration",
        "message": "Required adapters do not report a consistent Available state",
        "observed_generation": 1
      },
      {
        "type": "Reconciled",
        "status": "False",
        "reason": "MissingRequiredAdapters",
        "message": "Required adapters not reporting Available=True: [cl-deployment, cl-invalid-resource, cl-job, cl-maestro, cl-namespace, cl-precondition-error]. Currently reporting: []",
        "observed_generation": 1
      }
    ]
  }
}
```

### Requirement: Get Cluster Conditions Table

The CLI SHALL display cluster conditions in a formatted table via `--output table`.

#### Scenario: Display conditions table

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster conditions --output table`
- THEN the CLI MUST output a table with columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE
- AND status values MUST be color-coded: True=green, False=red, Unknown=yellow

**Example** — `hf cluster conditions --output table` before any adapters report:
```
TYPE        STATUS  LAST TRANSITION      REASON                       MESSAGE
---         ---     ---                  ---                          ---
Available   False   2026-04-24T16:00:00Z AdaptersNotAtSameGeneration  Required adapters do not report a consistent Available state
Reconciled  False   2026-04-24T16:00:00Z MissingRequiredAdapters      Required adapters not reporting Available=True: [cl-deployment, ...]. Currently reporting: []
```

**Example** — `hf cluster conditions --output table` after some adapters report (partial convergence):
```
TYPE                    STATUS  LAST TRANSITION      REASON               MESSAGE
---                     ---     ---                  ---                  ---
Available               False   2026-04-24T16:01:00Z AdaptersNotAtSame... ...
Reconciled              False   2026-04-24T16:01:00Z MissingRequired...   ...
ClDeploymentSuccessful  True    2026-04-24T16:01:00Z ManualStatusPost      Status posted via hf adapter post-status
ClJobSuccessful         False   2026-04-24T16:01:00Z ManualStatusPost      Status posted via hf adapter post-status
ClNamespaceSuccessful   True    2026-04-24T16:01:00Z ManualStatusPost      Status posted via hf adapter post-status
```

### Requirement: Get Cluster Adapter Statuses

The statuses table SHALL include a FINALIZED column in addition to AVAILABLE.

#### Scenario: Get statuses table

- WHEN the user runs `hf cluster statuses --output table`
- THEN the CLI MUST output columns: ADAPTER, GEN, AVAILABLE, FINALIZED
- AND AVAILABLE and FINALIZED columns MUST be color-coded dots: green=True, red=False, `-`=not present

### Requirement: Interactive Cluster Status Filter

The CLI SHALL provide an interactive split-screen filter for adapter statuses via `hf cluster statuses --filter`. This is a **typed filter** — the primary interaction is typing a query string, not navigating a list. It is entirely different from `hf cluster get -i` (which is a read-only viewer with no input). The `-i` flag on `hf cluster statuses` retains its original behavior of interactively selecting the active cluster.

#### Scenario: Split-screen layout and filter input

- GIVEN a cluster-id is set in config and adapter statuses exist
- WHEN the user runs `hf cluster statuses --filter`
- THEN the CLI MUST fetch all adapter statuses and open a split-screen view
- AND the left panel MUST contain:
  - A **text input field at the top** where the user types their filter query
  - A **reference section below** showing two labelled groups:
    - **Adapters** — the unique adapter names (from the `adapter` field of each status item)
    - **Conditions** — the unique condition types (from `conditions[].type` across all items); these always start with an uppercase letter
  - The reference groups are for visual reference only — they are not selectable or navigable; the user interacts only via the text input
- AND the right panel MUST show the filtered results, updating as the user types
- AND pressing `q` or Escape MUST exit the view cleanly with code 0
- AND pressing Enter (selecting the highlighted item) MUST also exit cleanly with code 0

#### Scenario: Filter by adapter name (lowercase-prefixed query)

- GIVEN the user is in the interactive status filter view
- WHEN the user types a query string whose first character is **lowercase** (e.g. `cl-job`)
- THEN the filter targets the `adapter` field of each status item
- AND the right panel MUST show the full status item(s) where `adapter` contains the query as a substring
- AND all conditions of the matching adapter(s) MUST be shown

**Example** — query `cl-job` shows the entire `cl-job` adapter status including all its conditions.

#### Scenario: Filter by condition type (uppercase-prefixed query)

- GIVEN the user is in the interactive status filter view
- WHEN the user types a query string whose first character is **uppercase** (e.g. `Health`)
- THEN the filter targets `conditions[].type` across all adapter status items
- AND the right panel MUST show only the individual conditions where `type` contains the query as a substring, collected from all adapters
- AND each displayed condition MUST be labelled with its parent adapter name (e.g. `cl-job › Health: False`) to differentiate same-type conditions from different adapters

**Example** — query `Health` shows one row per adapter that has a `Health` condition, each labelled with the adapter name.

#### Scenario: Empty query — show all

- GIVEN the user is in the interactive status filter view
- WHEN the filter input is empty
- THEN the right panel MUST display all adapter statuses unfiltered (same as `hf cluster statuses` default JSON output)

#### Scenario: No matching results

- GIVEN the user types a query that matches no adapter name and no condition type
- THEN the right panel MUST display `(no results)`

### Requirement: List Clusters

The CLI SHALL list all clusters via GET /clusters.

#### Scenario: List clusters as JSON

- GIVEN an active environment is configured
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters`
- AND output the full `ListResponse[Cluster]` as pretty-printed JSON

#### Scenario: List clusters as table

- GIVEN an active environment is configured
- WHEN the user runs `hf cluster list --output table`
- THEN the CLI MUST output a table with columns: ID, NAME, GEN, STATUS
- AND STATUS MUST be derived from conditions: green dot if Available=True AND Reconciled=True, otherwise red dot (or plain text in no-color mode)
