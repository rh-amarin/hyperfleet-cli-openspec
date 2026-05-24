## ADDED Requirements

### Requirement: hf tui command

The CLI SHALL provide a top-level `hf tui` command that launches a full-screen terminal user interface for monitoring HyperFleet clusters and nodepools. The command MUST require an active environment (same as other data commands). A `-s / --seconds` flag MUST control the data refresh interval with default `5` and minimum `1`.

#### Scenario: Launch TUI with active environment

- **WHEN** the user runs `hf tui` with an active environment configured
- **THEN** the CLI MUST enter the alternate screen buffer and display the combined cluster+nodepool table
- **AND** the table MUST auto-refresh at the configured interval

#### Scenario: Launch TUI without active environment

- **WHEN** the user runs `hf tui` without an active environment
- **THEN** the CLI MUST exit with code 1 and print an `[ERROR]` message

#### Scenario: Custom refresh interval

- **WHEN** the user runs `hf tui -s 10`
- **THEN** the CLI MUST refresh table data every 10 seconds

### Requirement: Main table view

The main panel MUST display the same combined cluster+nodepool table as `hf table --watch`, including fixed condition columns (`Reconciled`, `LastKnownReconciled`), dynamic adapter columns, nodepool row indentation, deletion markers, status dots, and adapter activity spinners during refresh cycles.

#### Scenario: Table matches hf table --watch

- **WHEN** the TUI main view is displayed
- **THEN** row content and column ordering MUST match `hf table --watch` for the same cluster data

### Requirement: Header bar

The TUI MUST display a fixed header section above the main panel with three lines: (1) the current view name and refresh countdown, (2) active environment name, API URL, and Kubernetes context, and (3) port-forward connectivity for configured services (`hyperfleet-api`, `postgresql`, `maestro-http`, `maestro-grpc`). Port-forward status MUST refresh on the same interval as table data.

#### Scenario: Header shows list view context

- **WHEN** the TUI is in list mode
- **THEN** line 1 MUST show `View: Clusters/Nodepools` and the refresh countdown
- **AND** line 2 MUST show the active environment name and API URL

#### Scenario: Header shows describe view context

- **WHEN** the describe view is open
- **THEN** line 1 MUST show `View: Describe` with the selected resource and format

#### Scenario: Port-forward status in header

- **WHEN** port-forward services are configured
- **THEN** line 3 MUST show each service with a connected/disconnected indicator and local port

### Requirement: Single-panel layout

The TUI MUST use a single main panel at all times. The default view MUST be a full-width cluster+nodepool table. Pressing Enter or `d` MUST switch the main panel to a full-width describe view for the selected row. Pressing `Esc` MUST return to the table list. The CLI MUST NOT render a persistent split-screen layout with table and detail side by side.

#### Scenario: List view uses full terminal width

- **WHEN** the TUI is in list mode
- **THEN** the table MUST occupy the full width of the main panel

#### Scenario: Describe replaces main panel

- **WHEN** the user selects a row and presses Enter or `d`
- **THEN** the main panel MUST switch to full-width describe content
- **AND** the table MUST NOT remain visible alongside the describe view

#### Scenario: Return to list from describe

- **WHEN** the describe view is open and the user presses `Esc`
- **THEN** the main panel MUST return to the full-width table list

### Requirement: Detail panel

When a row is selected, pressing Enter or `d` MUST open a full-width describe view in the main panel. The view MUST be vertically scrollable. Pressing `Esc` MUST close the describe view and return to the table list.

#### Scenario: Open describe view for cluster

- **WHEN** the user selects a cluster row and presses Enter or `d`
- **THEN** the main panel MUST show the cluster resource
- **AND** the default view MUST be syntax-highlighted JSON

#### Scenario: Open describe view for nodepool

- **WHEN** the user selects a nodepool row and presses Enter or `d`
- **THEN** the main panel MUST show the nodepool resource

#### Scenario: Close describe view

- **WHEN** the describe view is open and the user presses `Esc`
- **THEN** the describe view MUST close
- **AND** the main panel MUST show the table list at full width

#### Scenario: Scroll describe view

- **WHEN** the describe view is open and content exceeds panel height
- **THEN** the user MUST be able to scroll the content with ↑/↓ or PgUp/PgDn

#### Scenario: Row selection with cursor keys

- **WHEN** the user presses ↑ or ↓ in list mode
- **THEN** the highlighted row MUST move among cluster and nodepool rows
- **AND** pressing Enter or `d` MUST open the describe view for the selected row

### Requirement: Detail format cycling

Pressing `V` while the describe view is open MUST cycle the view among JSON, YAML, and overview modes. Syntax highlighting MUST be applied for JSON and YAML when color is enabled.

#### Scenario: Cycle JSON to YAML to overview

- **WHEN** the describe view is open and the user presses `V` repeatedly
- **THEN** the display MUST cycle: JSON → YAML → overview → JSON

### Requirement: Statuses view

Pressing `S` while the describe view is open MUST switch to adapter statuses for the selected cluster or nodepool. Pressing `S` again MUST return to the resource describe view (preserving the current format mode).

#### Scenario: Toggle statuses view for cluster

- **WHEN** a cluster is selected and the user presses `S`
- **THEN** the describe view MUST show adapter statuses for that cluster

#### Scenario: Toggle statuses view for nodepool

- **WHEN** a nodepool is selected and the user presses `S`
- **THEN** the describe view MUST show adapter statuses for that nodepool

### Requirement: Smart search in statuses view

Pressing `/` while the describe view is in statuses mode MUST open a search text field. Filter behavior MUST depend on the case of the first typed character:

- If the first character is lowercase, the filter MUST match adapter names partially (case-insensitive) against `items[].adapter` and display only matching adapter statuses.
- If the first character is uppercase, the filter MUST match condition `type` values partially and display matching conditions with the adapter name of each condition.

#### Scenario: Lowercase filter by adapter name

- **WHEN** the user is in statuses view, presses `/`, and types `sent` (lowercase start)
- **THEN** only adapter statuses whose `adapter` field contains `sent` (case-insensitive) MUST be shown

#### Scenario: Uppercase filter by condition type

- **WHEN** the user is in statuses view, presses `/`, and types `Hea` (uppercase start)
- **THEN** only conditions whose `type` contains `Hea` MUST be shown
- **AND** each displayed condition MUST include the adapter name

#### Scenario: Cancel search

- **WHEN** the search field is open and the user presses `Esc`
- **THEN** the search field MUST close without changing the current filter

### Requirement: Patch selected resource

Pressing `c` MUST open a patch prompt for the currently selected cluster or nodepool row. The user MUST choose `s` to increment `spec.counter` or `l` to increment `labels.counter`, matching `hf cluster patch` / `hf nodepool patch` behavior. Pressing `Esc` or `c` again MUST cancel without patching. On success the TUI MUST refresh table data and display an `[INFO]` message with the old and new counter values.

#### Scenario: Patch cluster spec from TUI

- **WHEN** a cluster row is selected and the user presses `c` then `s`
- **THEN** the CLI MUST PATCH the cluster with an incremented `spec.counter`
- **AND** display `[INFO] Incrementing spec.counter: <old> -> <new>`

#### Scenario: Patch nodepool labels from TUI

- **WHEN** a nodepool row is selected and the user presses `c` then `l`
- **THEN** the CLI MUST PATCH the nodepool with an incremented `labels.counter`

#### Scenario: Cancel patch prompt

- **WHEN** the patch prompt is open and the user presses `Esc`
- **THEN** no PATCH request MUST be sent

### Requirement: Exit

Pressing `q` or `Ctrl+C` MUST exit the TUI cleanly and restore the primary screen buffer with exit code 0.

#### Scenario: Quit with q

- **WHEN** the user presses `q`
- **THEN** the TUI MUST exit with code 0 and restore the terminal
