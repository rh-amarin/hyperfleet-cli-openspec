## ADDED Requirements

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
