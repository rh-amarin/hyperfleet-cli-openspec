## Context

`configShowCmd.RunE` in `cmd/config.go` currently marshals all config sections and the state block into a single YAML string via `marshalYAMLOrdered`, then writes it to stdout in one call. There is no TTY awareness in this path — the output is always plain text.

The existing `internal/output/` package already owns all color logic (dots, JSON colorization) and already imports `golang.org/x/term` for TTY detection. The ANSI constants (`ansiGreen`, `ansiRed`, `ansiYellow`, `ansiReset`) are unexported; adding new exported helpers follows the same pattern as `StatusDot`.

## Goals / Non-Goals

**Goals:**
- Colorize top-level YAML section headers in bold cyan when stdout is a TTY.
- Print a separator line (`────────────────────────────────────────`) between the last config section and the `state:` block.
- Respect `--no-color`, `NO_COLOR`, and non-TTY suppression identically to existing colored output.

**Non-Goals:**
- Colorizing values (only section header keys).
- Changing any other subcommand output (`hf config env show`, `hf config get`, etc.).
- Introducing a new flag or changing the output format when `--output json|yaml` is used (colorization applies to the default YAML-style display only).

## Decisions

### Post-process the YAML string
`marshalYAMLOrdered` already produces a well-formed YAML string. Rather than restructuring the marshal pipeline, a `ColorizeYAMLSections(s string, noColor bool) string` helper scans line by line and prefixes ANSI codes on lines matching `^(\w[\w-]*):\s*$` (top-level YAML keys with no inline value). This keeps the marshal logic untouched and the colorization self-contained.

**Alternative considered:** Marshal config and state sections into separate YAML strings and skip post-processing. Rejected because it duplicates the YAML-ordering logic and breaks if new top-level keys are added.

### New file in `internal/output/`
`config_display.go` is added to `internal/output/` alongside `dots.go`. Both export small, stateless color helper functions. Putting it in `cmd/` would bury general-purpose helpers behind the command layer.

### Separator style
Unicode `────────────────────────────────────────` (40 × U+2500 BOX DRAWINGS LIGHT HORIZONTAL). Visually distinct, no ambiguity with YAML `---` document separator, and consistent with terminal UI conventions. In `noColor` mode the same character is used (it carries no ANSI codes).

### Color choice for section headers
`\033[1;36m` (bold cyan) — cyan is already used for JSON object keys, so it is a consistent "structural element" color. Bold weight makes headers visually heavier than their child keys.

### TTY detection in the command
`configShowCmd.RunE` must detect whether its writer is a TTY. It uses `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))` when `cmd.OutOrStdout()` is `os.Stdout`. The `--no-color` flag already exists on the root command and is accessible via the persistent pre-run.

## Risks / Trade-offs

- **Test isolation**: `ColorizeYAMLSections` with `noColor=false` injects ANSI escape codes. Tests MUST pass `noColor=true` or strip codes when asserting plain-text content. → Tests use `noColor=true` and a separate ANSI-presence test with `noColor=false`.
- **Windows terminals**: older Windows consoles do not render ANSI codes. The `NO_COLOR` / non-TTY suppression already handles this in the existing output package. No additional work needed.
- **Future YAML keys with inline values**: the regex `^(\w[\w-]*):\s*$` only matches keys with no value on the same line (i.e., section headers). Keys like `api-url: http://...` are not matched and will not be colorized. This is the desired behavior.

## Open Questions

- None — scope is fully bounded by the config show command.
