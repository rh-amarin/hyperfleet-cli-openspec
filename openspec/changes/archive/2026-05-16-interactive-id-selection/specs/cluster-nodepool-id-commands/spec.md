# Delta: cluster-nodepool-id-commands — interactive-id-selection

Extends the existing `cluster-nodepool-id-commands` spec with interactive selection behavior.

---

## Requirement: `hf cluster id --interactive` (ADDED)

The CLI SHALL support an `--interactive` / `-i` flag on `hf cluster id` that fetches all available clusters and presents an interactive fuzzy-searchable picker. Selecting a cluster saves it as the active cluster context.

### Scenario: Interactive selection — cluster selected

- **GIVEN** the user runs `hf cluster id --interactive` (or `-i`)
- **AND** at least one cluster is available from the API
- **WHEN** the user selects a cluster in the picker and presses Enter
- **THEN** the CLI MUST save the selected cluster's ID via `s.SetState("cluster-id", <id>)` (persisted to `state.yaml`)
- **AND** print `Active cluster set to: <name> (<id>)` to stdout
- **AND** exit 0

### Scenario: Interactive selection — user aborts

- **GIVEN** the user runs `hf cluster id -i`
- **WHEN** the user presses Esc or Ctrl+C
- **THEN** the CLI MUST NOT modify state
- **AND** produce no output
- **AND** exit 0

### Scenario: Interactive selection — no clusters available

- **GIVEN** the API returns an empty cluster list
- **WHEN** the user runs `hf cluster id -i`
- **THEN** the CLI MUST return `[ERROR] no clusters available` and exit 1 without opening the picker

### Scenario: Non-interactive path unchanged

- **GIVEN** `--interactive` is not passed
- **WHEN** the user runs `hf cluster id`
- **THEN** behaviour is identical to the pre-existing spec (print active ID or error if none set)

---

## Requirement: `hf nodepool id --interactive` (ADDED)

The CLI SHALL support an `--interactive` / `-i` flag on `hf nodepool id` that fetches all nodepools for the active cluster and presents an interactive fuzzy-searchable picker. Selecting a nodepool saves it as the active nodepool context.

**Prerequisite**: `cluster-id` MUST be set in state. If it is not, the CLI MUST print the standard missing-cluster-id error and exit 1 before opening the picker or making any API call.

### Scenario: Interactive selection — nodepool selected

- **GIVEN** the user runs `hf nodepool id --interactive` (or `-i`)
- **AND** `cluster-id` is set in state
- **AND** at least one nodepool exists for the active cluster
- **WHEN** the user selects a nodepool in the picker and presses Enter
- **THEN** the CLI MUST save the selected nodepool's ID via `s.SetState("nodepool-id", <id>)`
- **AND** print `Active nodepool set to: <name> (<id>)` to stdout
- **AND** exit 0

### Scenario: Interactive selection — user aborts

- **GIVEN** the user runs `hf nodepool id -i`
- **WHEN** the user presses Esc or Ctrl+C
- **THEN** the CLI MUST NOT modify state, produce no output, and exit 0

### Scenario: Interactive selection — no nodepools available

- **GIVEN** the API returns an empty nodepool list for the active cluster
- **WHEN** the user runs `hf nodepool id -i`
- **THEN** the CLI MUST return `[ERROR] no nodepools available for cluster <cluster-id>` and exit 1

### Scenario: Interactive selection — no cluster-id in state

- **GIVEN** `cluster-id` is NOT set in state
- **WHEN** the user runs `hf nodepool id -i`
- **THEN** the CLI MUST print the standard missing-cluster-id error and exit 1 without making any API call

### Scenario: Non-interactive path unchanged

- **GIVEN** `--interactive` is not passed
- **WHEN** the user runs `hf nodepool id`
- **THEN** behaviour is identical to the pre-existing spec

---

## Picker UX Contract

- The picker displays each row as `<name padded to 40 chars>  <id>`
- The user can type to fuzzy-filter, use ↑↓ to navigate, Enter to confirm, Esc to cancel
- Implemented via `github.com/ktr0731/go-fuzzyfinder` (fzf-style)
- The `Selector` interface in `internal/selector` allows test injection without a real terminal
