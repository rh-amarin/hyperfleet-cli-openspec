## Why

Table output for resources with long column names (e.g., `PROVISIONING_STATE`, `OBSERVED_GENERATION`, `CLUSTER_ID`) makes rows unnecessarily wide, forcing horizontal scrolling on standard 80- or 120-column terminals. This hurts readability when viewing cluster lists, nodepool statuses, or condition tables in the field.

The fix is to constrain the visual width of any header that exceeds 10 characters by wrapping it across multiple header lines instead of stretching the column to fit the full name.

## What Changes

`Printer.PrintTable` in `internal/output` gains header-wrapping behaviour:
- Any column header longer than 10 characters after uppercasing is split across multiple stacked lines.
- Splitting prefers underscore word boundaries; single tokens longer than 10 characters are hard-broken at 10.
- All header lines are emitted as separate tabwriter rows before the data rows, so column alignment is preserved.
- Headers of 10 characters or fewer are unaffected.

No command-level code changes. No new packages. No changes to JSON, YAML, or dot-rendering output.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `output-formatting`: The table output requirement gains a new scenario for multi-line header wrapping.
