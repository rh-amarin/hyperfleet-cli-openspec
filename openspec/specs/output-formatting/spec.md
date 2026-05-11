# Output Formatting Specification

## Purpose

Provide a shared output formatting package that dispatches on the `--output` flag
(json, table, yaml), renders colored status dots for condition tables, and computes
dynamic column ordering for resource tables.
## Requirements
### Requirement: Multi-Format Output Dispatch

The output package SHALL dispatch rendering based on the `--output` flag value.

#### Scenario: JSON output

- GIVEN `--output json` is set (or is the default for the command)
- WHEN a command calls `Printer.Print(v)` with any Go value
- THEN the output MUST be pretty-printed JSON with 2-space indentation
- AND a trailing newline MUST be appended

#### Scenario: Table output

- GIVEN `--output table` is set
- WHEN a command calls `Printer.PrintTable(headers, rows)`
- THEN the output MUST be rendered using aligned columns (tab-separated via tabwriter)
- AND headers MUST be displayed in uppercase

#### Scenario: YAML output

- GIVEN `--output yaml` is set
- WHEN a command calls `Printer.Print(v)` with any Go value
- THEN the output MUST be rendered as YAML
- AND struct field names MUST be serialized as `snake_case` keys matching the API field names (e.g., `created_time`, `observed_generation`) — the same JSON tags are reused for YAML via `gopkg.in/yaml.v3`

### Requirement: Colored Dot Rendering

The output package SHALL render condition status values as colored dot characters.

#### Scenario: True status renders green

- GIVEN a condition status is `True`
- WHEN the dot renderer is called
- THEN it MUST return a green-colored dot character (`●`)

#### Scenario: False status renders red

- GIVEN a condition status is `False`
- WHEN the dot renderer is called
- THEN it MUST return a red-colored dot character (`●`)

#### Scenario: Unknown status renders yellow

- GIVEN a condition status is `Unknown` (only valid for `AdapterCondition`, not `ResourceCondition`)
- WHEN the dot renderer is called
- THEN it MUST return a yellow-colored dot character (`●`)

#### Scenario: Absent condition renders dash

- GIVEN a condition is not present for a resource
- WHEN the dot renderer is called
- THEN it MUST return a dash character (`-`)

#### Scenario: No-color mode

- GIVEN `--no-color` is set or the `NO_COLOR` environment variable is set
- WHEN the dot renderer is called
- THEN it MUST return the dot character without ANSI color codes
- AND True MUST render as `True`, False as `False`, Unknown as `Unknown`

#### Scenario: Non-TTY auto color disabling

- GIVEN stdout is not a TTY (e.g., output is piped to a file or another command)
- WHEN any output with color is produced
- THEN ANSI color codes MUST be disabled automatically (equivalent to `--no-color`)
- AND there is no flag to override this — use `--no-color` to explicitly disable color on a TTY

### Requirement: Dynamic Column Ordering

The output package SHALL compute column order for condition-based resource tables.

#### Scenario: Column ordering algorithm

- GIVEN a set of resources each with varying conditions
- WHEN the dynamic column builder processes the conditions
- THEN columns MUST be ordered as:
  1. Fixed columns first (e.g., ID, NAME, GEN — provided by the caller)
  2. `Available` column (if present in any resource's conditions)
  3. All other condition types sorted alphabetically
  4. `Reconciled` column last (if present in any resource's conditions)

#### Scenario: Collect unique conditions across resources

- GIVEN multiple resources with different sets of conditions
- WHEN the column builder collects condition types
- THEN the resulting column list MUST include every unique condition type across all resources
- AND no duplicates MUST appear

#### Scenario: No conditions present

- GIVEN resources have no conditions (e.g., newly created with empty status)
- WHEN the column builder processes the conditions
- THEN only the fixed columns MUST appear (no dynamic condition columns)

### Requirement: Colored JSON Output

The CLI SHALL colorize JSON output when writing to an interactive terminal. JSON colorization respects the same `--no-color` flag, `NO_COLOR` environment variable, and non-TTY auto-disable as dot rendering — all three mechanisms suppress JSON colors.

#### Scenario: Color enabled (default)

- GIVEN the writer is an interactive TTY
- AND `--no-color` is not set
- AND the `NO_COLOR` environment variable is not set
- WHEN `hf` prints JSON output
- THEN object keys MUST be rendered in cyan
- AND string values MUST be rendered in green
- AND numeric values MUST be rendered in yellow
- AND `true` MUST be rendered in green, `false` in red
- AND `null` MUST be rendered in dim/faint style
- AND delimiters (`{`, `}`, `[`, `]`) MUST be uncolored

#### Scenario: Color suppressed — --no-color flag

- GIVEN the `--no-color` flag is set
- WHEN `hf` prints JSON output
- THEN the output MUST be plain text with no ANSI escape codes

#### Scenario: Color suppressed — NO_COLOR env var

- GIVEN the `NO_COLOR` environment variable is set to any non-empty value
- WHEN `hf` prints JSON output
- THEN the output MUST be plain text with no ANSI escape codes

#### Scenario: Color suppressed — non-TTY writer

- GIVEN the output is piped to a file or another process (not a TTY)
- WHEN `hf` prints JSON output
- THEN the output MUST be plain text with no ANSI escape codes

### Requirement: Printer Initialization

The output printer SHALL be initialized with format and color settings.

#### Scenario: Create printer with defaults

- GIVEN `--output` is not set and `--no-color` is not set
- WHEN `NewPrinter` is called with format="" and noColor=false
- THEN the printer MUST default to writing to stdout
- AND the format MUST be determined by the caller (each command sets its own default)

#### Scenario: Create printer for stderr messages

- GIVEN a warning or info message needs to be written
- WHEN message helpers are used
- THEN `[WARN]`, `[INFO]`, and `[ERROR]` prefixed messages MUST be written to stderr
- AND these MUST NOT be affected by the `--output` flag

### Requirement: Braille Spinner Helper

The `internal/output` package SHALL expose two pure functions supporting the activity indicator feature:

`SpinnerFrame(tick int) string` — returns a single braille spinner character from the fixed sequence `["⠋","⠙","⠹","⠸","⠼","⠴","⠦","⠧","⠇","⠏"]` at index `tick % 10`.

`IsActive(lastReportTime string, frequencySecs int) bool` — parses `lastReportTime` as RFC 3339 and returns `true` when `time.Since(t) < time.Duration(2*frequencySecs) * time.Second`. Returns `false` if the string is empty or unparseable.

These functions are pure (no side effects, no I/O) and are the sole location for activity-indicator logic.

#### Scenario: SpinnerFrame cycles correctly

- **WHEN** `SpinnerFrame` is called with ticks 0 through 9
- **THEN** it MUST return each of the 10 braille frames in order
- **AND** tick 10 MUST return the same frame as tick 0

#### Scenario: IsActive — recent report

- **WHEN** `lastReportTime` is an RFC 3339 timestamp less than `2 × frequencySecs` seconds ago
- **THEN** `IsActive` MUST return `true`

#### Scenario: IsActive — stale report

- **WHEN** `lastReportTime` is an RFC 3339 timestamp older than `2 × frequencySecs` seconds
- **THEN** `IsActive` MUST return `false`

#### Scenario: IsActive — empty or malformed string

- **WHEN** `lastReportTime` is empty or not a valid RFC 3339 timestamp
- **THEN** `IsActive` MUST return `false`

