# NodePool Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet nodepools. Nodepools are always scoped to a parent cluster, requiring a `cluster-id` to be set in config. All nodepool operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`.

## Prerequisites

**cluster-id required**: All nodepool commands require `cluster-id` to be set in state. If it is not set, the CLI MUST display:
```
[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.
```
AND exit with code 1 before making any API call.

**nodepool-id required for single-resource commands**: `hf nodepool patch`, `hf nodepool delete`, `hf nodepool conditions`, and `hf nodepool statuses` additionally require `nodepool-id` to be set in state (unless an explicit ID argument is provided). If cluster-id is set but nodepool-id is not, the CLI MUST display:
```
[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.
```
AND exit with code 1.

## Requirements

### Requirement: Create NodePool

`hf nodepool create` SHALL load the request body from a JSON template. The binary embeds a built-in default template (`cmd/assets/nodepool-template.json`). When no `--file` flag is given, the CLI MUST use the embedded default bytes directly in memory — it MUST NOT read from or write to `<config-dir>`. The `--name` flag overrides only the `name` field in the template.

#### Scenario: Create nodepool with default template

- GIVEN no `--file` flag is provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST use the built-in embedded template in memory
- AND MUST NOT write any file to `<config-dir>`
- AND MUST create the nodepool using the default payload (`kind=NodePool`, `name=my-nodepool`, default labels and spec)

#### Scenario: Create nodepool with `--name` override

- GIVEN a template (embedded or from `--file`) with `"name": "<template-name>"`
- WHEN the user runs `hf nodepool create --name <name>`
- THEN the CLI MUST set `name` to `<name>` in the request body, overriding the template value
- AND all other template fields MUST remain unchanged

#### Scenario: Create nodepool with `--file` override

- GIVEN a file at `<path>` containing a valid JSON nodepool payload
- WHEN the user runs `hf nodepool create --file <path>`
- THEN the CLI MUST use that file's content as the request body
- AND `--name` MAY still be used to override the name field from that file

#### Scenario: Create nodepool with no arguments

- GIVEN no flags and no arguments
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the embedded template payload

#### Scenario: Malformed template file

- GIVEN a file provided via `--file` containing invalid JSON
- WHEN the user runs `hf nodepool create --file <path>`
- THEN the CLI MUST exit with `[ERROR] loading template: <reason>` and code 1

### Requirement: List NodePools

The CLI SHALL list all nodepools in the current cluster.

#### Scenario: List nodepools

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf nodepool list`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools` using the cluster-id from state
- AND output the response as JSON with shape `{"kind": "NodePoolList", "items": [...], "page": N, "size": N, "total": N}`

### Requirement: Search NodePool

The CLI SHALL search for a nodepool by name within the current cluster and set it as the current context.

#### Scenario: Search with no arguments

- GIVEN a nodepool-id is set in config
- WHEN the user runs `hf nodepool search` with no arguments
- THEN the CLI MUST behave identically to `hf nodepool get` — fetching and returning the current nodepool from state

#### Scenario: Search with no arguments and no nodepool in state

- GIVEN no nodepool-id is set in state
- WHEN the user runs `hf nodepool search` with no arguments
- THEN the CLI MUST display `[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.`
- AND exit with code 1

#### Scenario: Search for existing nodepool

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST filter nodepools by name within the cluster
- AND output the matching nodepools as a JSON array of full NodePool objects
- AND persist the found nodepool's ID to active state via `config.SetNodePoolID`
- AND print `[INFO] NodePool context set to '<id>'` on stderr after persisting

#### Scenario: Search for non-existent nodepool

- GIVEN no nodepool matches the search name within the cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST display `[WARN] No nodepools found matching '<name>'`
- AND output an empty JSON array `[]`
- AND exit with code 0

#### Scenario: Multiple matches

- GIVEN multiple nodepools match the search name within the cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST display `[WARN] Multiple nodepools found matching '<name>', using first result`
- AND set nodepool-id to the first element in the returned `items` array
- AND persist that nodepool-id to active state

### Requirement: Get NodePool

The CLI SHALL retrieve and display full details of a specific nodepool.

#### Scenario: Get current nodepool

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool get`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
- AND output the full nodepool JSON

#### Scenario: Get nodepool by explicit ID

- GIVEN a valid nodepool ID is provided
- WHEN the user runs `hf nodepool get <nodepool_id>`
- THEN the CLI MUST use the provided ID instead of the configured nodepool-id

### Requirement: Interactive NodePool View

The CLI SHALL support an interactive split-screen viewer for a nodepool via `hf nodepool get -i`. This is a **read-only viewer** — there is no text input or filter. It replaces the old "pick a cluster" behavior that `-i` had on other subcommands.

#### Scenario: Open interactive viewer

- GIVEN cluster-id and nodepool-id are set in config (or an explicit nodepool ID is provided)
- WHEN the user runs `hf nodepool get -i` (or `hf nodepool get <id> -i`)
- THEN the CLI MUST fetch the nodepool and open an interactive split-screen view
- AND the left panel MUST display a compact summary of the nodepool's key fields: `name`, `id`, `generation`, `spec.replicas`, `spec.platform.type`, and a one-line status derived from `status.conditions`
- AND the right panel MUST display the complete nodepool JSON with syntax coloring, scrollable with arrow keys
- AND there is NO text input field — the view is read-only; the user navigates the right panel directly
- AND pressing `q` or Escape MUST exit the view cleanly with code 0

#### Scenario: Interactive viewer — no nodepool in state

- GIVEN no nodepool-id is set in state and no explicit ID is provided
- WHEN the user runs `hf nodepool get -i`
- THEN the CLI MUST display the standard missing-nodepool-id error and exit 1

### Requirement: Patch NodePool

The CLI SHALL increment a counter field in the nodepool's spec or labels section, triggering a generation bump.

#### Scenario: Patch with no arguments

- GIVEN the user provides no arguments
- WHEN the user runs `hf nodepool patch`
- THEN the CLI MUST display usage: `Usage: hf nodepool patch {spec|labels} [nodepool_id]`
- AND exit with code 1

#### Scenario: Patch spec counter

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool patch spec`
- THEN the CLI MUST fetch the current nodepool
- AND read the current `spec.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}` with the incremented counter as a string
- AND display `[INFO] Incrementing spec.counter: <old> -> <new>` where `<old>` and `<new>` are integer strings (e.g., `1 -> 2`; first increment displays `0 -> 1`)
- AND the nodepool's generation MUST increment

#### Scenario: Patch labels counter

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool patch labels`
- THEN the CLI MUST fetch the current nodepool
- AND read the current `labels.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}` with the incremented counter as a string
- AND display `[INFO] Incrementing labels.counter: <old> -> <new>`
- AND the nodepool's generation MUST increment

### Requirement: Delete NodePool

The CLI SHALL delete a nodepool by ID.

#### Scenario: Delete nodepool

- GIVEN a nodepool exists
- WHEN the user runs `hf nodepool delete [nodepool_id]`
- THEN the CLI MUST send DELETE to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
- AND the response MUST include the full nodepool object with `deleted_by`, `deleted_time`, and incremented `generation`
- AND the CLI MUST output the deleted nodepool object subject to the `--output` flag (default: JSON)

#### Scenario: Delete current nodepool

- GIVEN nodepool-id is set in config and no explicit ID is provided
- WHEN the user runs `hf nodepool delete`
- THEN the CLI MUST use the configured nodepool-id

### Requirement: Get NodePool Conditions

The CLI SHALL display the generation and status conditions of a nodepool.

#### Scenario: Get conditions

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool conditions`
- THEN the CLI MUST fetch the nodepool and extract `generation` and `status.conditions` as JSON

**Example** — `hf nodepool conditions` after one patch (generation 2, no adapters yet):
```json
{
  "generation": 2,
  "status": {
    "conditions": [
      {
        "type": "Available",
        "status": "False",
        "reason": "AdaptersNotAtSameGeneration",
        "message": "Required adapters do not report a consistent Available state",
        "observed_generation": 2
      },
      {
        "type": "Reconciled",
        "status": "False",
        "reason": "MissingRequiredAdapters",
        "message": "Required adapters not reporting Available=True: [np-configmap]. Currently reporting: []",
        "observed_generation": 2
      }
    ]
  }
}
```

### Requirement: Get NodePool Conditions Table

The CLI SHALL display nodepool conditions in a formatted table via `--output table`.

#### Scenario: Display conditions table before adapters report

- GIVEN a nodepool exists with no adapter statuses
- WHEN the user runs `hf nodepool conditions --output table`
- THEN the CLI MUST output a table with columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE
- AND Reconciled and Available MUST show `False`

**Example** — `hf nodepool conditions --output table` before any adapters report:
```
TYPE        STATUS  LAST TRANSITION      REASON                       MESSAGE
---         ---     ---                  ---                          ---
Available   False   2026-04-24T16:05:00Z AdaptersNotAtSameGeneration  Required adapters do not report a consistent Available state
Reconciled  False   2026-04-24T16:05:00Z MissingRequiredAdapters      Required adapters not reporting Available=True: [np-configmap]. Currently reporting: []
```

#### Scenario: Display conditions table after all adapters report

- GIVEN all required adapters have reported `Available=True` at the current generation
- WHEN the user runs `hf nodepool conditions --output table`
- THEN Reconciled and Available MUST show `True` (green)
- AND per-adapter conditions (e.g., `NpConfigmapSuccessful`) MUST appear as additional rows

**Example** — `hf nodepool conditions --output table` after `np-configmap` reports `Available=True` at generation 2:
```
TYPE                   STATUS  LAST TRANSITION      REASON           MESSAGE
---                    ---     ---                  ---              ---
Available              True    2026-04-24T16:06:00Z AllAdapters...   All required adapters reported Available=True at generation 2
Reconciled             True    2026-04-24T16:06:00Z AllAdapters...   All required adapters report Available=True at generation 2
NpConfigmapSuccessful  True    2026-04-24T16:06:00Z ManualStatusPost  Status posted via hf adapter post-status
```

### Requirement: Get NodePool Adapter Statuses

The statuses table SHALL include a FINALIZED column in addition to AVAILABLE.

#### Scenario: Get statuses table

- WHEN the user runs `hf nodepool statuses --output table`
- THEN the CLI MUST output columns: ADAPTER, GEN, AVAILABLE, FINALIZED
- AND AVAILABLE and FINALIZED columns MUST be color-coded dots: green=True, red=False, `-`=not present

### Requirement: Interactive NodePool Status Filter

The CLI SHALL provide an interactive split-screen filter for nodepool adapter statuses via `hf nodepool statuses --filter`. This is a **typed filter** — the primary interaction is typing a query string, not navigating a list. It is entirely different from `hf nodepool get -i` (which is a read-only viewer with no input). The `-i` flag on `hf nodepool statuses` retains its original behavior of interactively selecting the active nodepool.

#### Scenario: Split-screen layout and filter input

- GIVEN cluster-id and nodepool-id are set in config and adapter statuses exist
- WHEN the user runs `hf nodepool statuses --filter`
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
- WHEN the user types a query string whose first character is **lowercase** (e.g. `np-configmap`)
- THEN the filter targets the `adapter` field of each status item
- AND the right panel MUST show the full status item(s) where `adapter` contains the query as a substring
- AND all conditions of the matching adapter(s) MUST be shown

**Example** — query `np-configmap` shows the entire `np-configmap` adapter status including all its conditions.

#### Scenario: Filter by condition type (uppercase-prefixed query)

- GIVEN the user is in the interactive status filter view
- WHEN the user types a query string whose first character is **uppercase** (e.g. `Available`)
- THEN the filter targets `conditions[].type` across all adapter status items
- AND the right panel MUST show only the individual conditions where `type` contains the query as a substring, collected from all adapters
- AND each displayed condition MUST be labelled with its parent adapter name (e.g. `np-configmap › Available: True`) to differentiate same-type conditions from different adapters

**Example** — query `Available` shows one row per adapter that has an `Available` condition, each labelled with the adapter name.

#### Scenario: Empty query — show all

- GIVEN the user is in the interactive status filter view
- WHEN the filter input is empty
- THEN the right panel MUST display all adapter statuses unfiltered (same as `hf nodepool statuses` default JSON output)

#### Scenario: No matching results

- GIVEN the user types a query that matches no adapter name and no condition type
- THEN the right panel MUST display `(no results)`

### Requirement: Display NodePool Table

The CLI SHALL display nodepools in the current cluster as a formatted table when `--output table` is passed to `hf nodepool list`.

#### Scenario: Display nodepool table

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool list --output table`
- THEN the CLI MUST fetch adapter statuses for each nodepool and output a table with:
  - Fixed columns: `ID`, `NAME`, `REPLICAS`, `TYPE`, `GEN`
  - Dynamic condition columns (excluding `*Successful` types)
  - Dynamic adapter columns (one per unique adapter name)
- AND status values MUST be displayed as colored dots with inline generation: `● N`
- AND the deletion marker (`❌`) MUST appear on GEN for nodepools with `deleted_time` set

**Example** — `hf nodepool list --output table` with two nodepools: `workers-1` (gen 2, converged) and `workers-2` (gen 1, not yet converged). Colors shown in parentheses, `● N` = dot + generation number:
```
ID                                    NAME      REPLICAS  TYPE           GEN  Available  Reconciled  np-configmap
---                                   ---       ---       ---            ---  ---        ---         ---
019dc049-e79e-72a9-94f8-0056a11193cd  workers-2  1        n2-standard-4  1    ● 1(red)   ● 1(red)    -
019dc049-e76c-7be1-b201-0db50e2c8ecb  workers-1  1        n2-standard-4  2    ● 2(green) ● 2(green)  ● 2(green)
```
