## ADDED Requirements

### Requirement: UI server command
The system SHALL provide a `hf ui` subcommand that starts an HTTP server serving a browser-based dashboard. The server SHALL listen on a configurable port (default 8088) and SHALL require an active environment to be set. The server SHALL print the listening address to stdout before accepting connections.

#### Scenario: Server starts with active environment
- **WHEN** user runs `hf ui` with an active environment configured
- **THEN** the server starts on port 8088 and prints `Serving HyperFleet UI at http://localhost:8088`

#### Scenario: Server starts on custom port
- **WHEN** user runs `hf ui --port 9090`
- **THEN** the server listens on port 9090 and prints the correct URL

#### Scenario: No active environment
- **WHEN** user runs `hf ui` with no active environment
- **THEN** the command exits with code 1 and prints `[ERROR] no active environment`

### Requirement: Dashboard HTML asset serving
The system SHALL serve a single self-contained HTML page at `GET /` with all CSS and JavaScript inline. The asset SHALL be embedded in the binary at build time via `go:embed`. No external CDN or runtime file dependencies SHALL be required.

#### Scenario: Browser requests root path
- **WHEN** browser sends `GET /` to the server
- **THEN** server responds with HTTP 200 and `Content-Type: text/html`

#### Scenario: Binary works without source files
- **WHEN** the built binary is copied to a machine without the source tree
- **THEN** `hf ui` serves the dashboard correctly

### Requirement: API proxy routes
The server SHALL expose read-only JSON proxy routes that forward requests to the configured HyperFleet API using the active environment's API client. All routes SHALL return `Content-Type: application/json`. Upstream errors SHALL be forwarded to the browser with the same HTTP status code.

Proxy route table:

| Browser path | Upstream path |
|---|---|
| `GET /api/clusters` | `clusters` (with statuses merged per cluster) |
| `GET /api/clusters/{id}` | `clusters/{id}` |
| `GET /api/clusters/{id}/statuses` | `clusters/{id}/statuses` |
| `GET /api/clusters/{id}/nodepools` | `clusters/{id}/nodepools` |
| `GET /api/clusters/{id}/nodepools/{npid}` | `clusters/{id}/nodepools/{npid}` |
| `GET /api/clusters/{id}/nodepools/{npid}/statuses` | `clusters/{id}/nodepools/{npid}/statuses` |

#### Scenario: Cluster list proxied successfully
- **WHEN** browser sends `GET /api/clusters`
- **THEN** server fetches clusters from upstream, merges per-cluster adapter statuses, and returns the combined JSON with HTTP 200

#### Scenario: Upstream returns error
- **WHEN** upstream API returns HTTP 404 for a cluster ID
- **THEN** server forwards the error response to the browser with HTTP 404 and JSON body

#### Scenario: Unknown route
- **WHEN** browser sends a request to an unrecognized path
- **THEN** server responds with HTTP 404

### Requirement: Dashboard cluster table
The dashboard SHALL display a table of all clusters with columns: Name, Age, and one column per unique condition type observed across all clusters. Condition columns SHALL be ordered: Available first, Reconciled last, all others alphabetically between. Each condition cell SHALL display a colored dot representing the condition status.

#### Scenario: Conditions rendered as colored dots
- **WHEN** a cluster has condition `Available=True`
- **THEN** the Available column shows a green dot (●) for that cluster's row

#### Scenario: Dynamic condition columns
- **WHEN** clusters have different condition type sets
- **THEN** the table shows the union of all condition types in the correct sort order

#### Scenario: Missing condition
- **WHEN** a cluster does not have a particular condition type
- **THEN** the corresponding cell shows a gray dot

### Requirement: Condition dot color semantics
Condition status SHALL be rendered as colored dots matching the CLI's `output.StatusDot()` semantics:
- `True` → green (`#3fb950`)
- `False` → red (`#f85149`)
- `Unknown` → yellow (`#d29922`)
- absent/empty → gray (`#484f58`)

#### Scenario: True status
- **WHEN** a condition has `Status: "True"`
- **THEN** the dot is rendered in green

#### Scenario: False status
- **WHEN** a condition has `Status: "False"`
- **THEN** the dot is rendered in red

#### Scenario: Unknown status
- **WHEN** a condition has `Status: "Unknown"`
- **THEN** the dot is rendered in yellow

### Requirement: Hover tooltips on condition dots
Each condition dot SHALL show a tooltip on hover containing: condition type, status, reason (if present), message (if present), and last transition time.

#### Scenario: Tooltip shown on hover
- **WHEN** user hovers over a condition dot
- **THEN** a tooltip appears showing condition type, status, reason, message, and last transition time

### Requirement: Auto-polling with countdown
The dashboard SHALL automatically re-fetch cluster data every 5 seconds. The header SHALL display a countdown showing seconds until the next refresh. Table rows SHALL update in place without a full page re-render (preserving scroll position and selected row state).

#### Scenario: Auto-refresh fires
- **WHEN** 5 seconds elapse since last fetch
- **THEN** dashboard fetches `/api/clusters`, updates changed rows in the table

#### Scenario: Countdown displayed
- **WHEN** dashboard is loaded
- **THEN** header shows a countdown timer decrementing from 5 to 0

### Requirement: Detail side panel for clusters
Clicking a cluster row SHALL open a side panel that slides in from the right. The panel SHALL remain open while the table is visible. The panel SHALL display: cluster name, ID, labels, created/updated timestamps, full list of ResourceConditions (with all fields), adapter statuses (grouped by adapter name, each with its AdapterConditions and job metadata), and an expandable list of nodepools.

#### Scenario: Row click opens panel
- **WHEN** user clicks a cluster row
- **THEN** the detail panel slides in showing the cluster's full conditions and adapter statuses

#### Scenario: Panel closes on dismiss
- **WHEN** user clicks the close (×) button
- **THEN** the detail panel slides out and no cluster is selected

#### Scenario: Panel re-fetches on poll
- **WHEN** the auto-poll fires while a cluster is selected
- **THEN** the detail panel data is also refreshed

### Requirement: Expandable nodepools in detail panel
The cluster detail panel SHALL list all nodepools for the selected cluster. Each nodepool SHALL be collapsible. Expanding a nodepool SHALL fetch and display its ResourceConditions and adapter statuses inline.

#### Scenario: Nodepool expand
- **WHEN** user clicks a nodepool entry in the detail panel
- **THEN** nodepool conditions and adapter statuses are fetched and displayed

#### Scenario: Nodepool collapse
- **WHEN** user clicks an expanded nodepool
- **THEN** the nodepool detail collapses
