# Tables and Lists — delta spec for watch-countdown-display

Only the scenarios below change. All other scenarios from `openspec/specs/tables-and-lists/spec.md` remain in force.

## MODIFIED Requirements

### Requirement: Watch Mode for Table Commands

`hf table --watch`, `hf resources --watch`, `hf cluster list --watch`, and `hf nodepool list --watch` MUST continuously refresh the table output at the configured interval. The refresh interval MUST default to 5 seconds and MAY be changed with `--seconds / -s`.

For commands that use fast-tick rendering (500 ms spinner interval decoupled from the data-fetch interval), the CLI MUST additionally display a live countdown line above the table headers on every render tick, showing the number of seconds remaining until the next data fetch and an animated braille spinner.

#### Scenario: Combined table watch mode — countdown line shown

- GIVEN `hf table --watch` is running with a refresh interval of N seconds
- WHEN the table is rendered (every 500 ms)
- THEN the CLI MUST print a line of the form `↻ Xs  <spinner>` above the table headers
- AND `X` MUST be the ceiling of the number of seconds remaining until the next data fetch (range: 1 to N)
- AND `<spinner>` MUST be the current braille spinner frame, advancing every 500 ms
- AND the line MUST appear flush left, directly above the `ID` column header

#### Scenario: Combined table watch mode — countdown resets after data refresh

- GIVEN `hf table --watch` is running with a refresh interval of N seconds
- WHEN a data fetch completes successfully
- THEN the countdown MUST reset to N on the next render tick

#### Scenario: No countdown line in non-watch mode

- GIVEN the user runs `hf table` without `--watch`
- WHEN the table is rendered once and exits
- THEN the CLI MUST NOT print any `↻` countdown line
- AND the output MUST be byte-for-byte identical to the output before this change
