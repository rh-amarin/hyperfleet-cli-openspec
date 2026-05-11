## Why

`hf config` displays all configuration sections and runtime state as a single flat YAML block with no visual hierarchy. When scanning the output it is hard to tell at a glance where the live config ends and the ephemeral runtime state begins, and all YAML keys look identical. Adding color to section headers and a visual separator between the config and state sections makes the output immediately scannable.

## What Changes

- Section headers in `hf config` / `hf config show` output are colorized (bold cyan) when writing to an interactive terminal.
- A separator line is printed between the configuration sections (`hyperfleet`, `kubernetes`, etc.) and the `state:` block.
- Colors are suppressed automatically when `--no-color` is set, `NO_COLOR` is in the environment, or stdout is not a TTY — consistent with all other colored output in `hf`.

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `config`: The "Show Configuration" requirement gains colorized section headers and a separator between the config and state blocks.

## Testing Scope

- New helper functions in `internal/output/` need unit tests covering:
  - `ColorizeYAMLSections`: with `noColor=true` (no ANSI codes injected) and `noColor=false` (ANSI escape codes present on section-header lines).
  - `SectionSeparator`: returns a non-empty string in both modes.
- No live-cluster interaction is required to verify the color rendering; a manual `hf config` run against the real cluster is needed for the live screenshot.

## Impact

- `cmd/config.go` — `configShowCmd.RunE`: marshal config and state sections separately; inject separator and apply colorization.
- `internal/output/` — new file `config_display.go` with `ColorizeYAMLSections` and `SectionSeparator`.
- No API, wire format, or flag surface changes.
