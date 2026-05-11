## Why

The existing `ColorizeYAMLSections` only adds bold-cyan to top-level section headers
(`hyperfleet:`, `kubernetes:`, …). Every indented key and its value renders in the same
default terminal color, so scanning `hf config` output still requires reading each word
to distinguish structure from data. Adding distinct colors for keys, values, and special
sentinel values makes the config immediately scannable at a glance.

## What Changes

`ColorizeYAMLSections` in `internal/output/config_display.go` is replaced with a
line-by-line parser that applies three additional color rules on top of the existing
section-header rule:

| Line type | Example | Color |
|-----------|---------|-------|
| Top-level section header | `hyperfleet:` | bold cyan (unchanged) |
| Indented key with a value | `  api-url: http://…` | key part → bold white; `: ` → dim; value part → green |
| Indented key with no value | `  context:` | key part → bold white; `:` → dim; value blank |
| Sentinel values | `<not set>`, `<set>` | dim/italic |

All colors are suppressed when `noColor=true`, `--no-color` is set, or stdout is not a TTY
— identical guard as the existing implementation.

## Capabilities

### Modified Capabilities

- `config`: "Show Configuration" gains per-token colorization (key vs value vs sentinel).

## Testing Scope

- `ColorizeYAMLSections` unit tests extended to cover:
  - Indented key+value line: key segment contains bold-white ANSI, value segment contains green ANSI
  - Indented key-only line: key segment contains bold-white ANSI, no value ANSI
  - Sentinel `<not set>`: contains dim ANSI codes
  - `noColor=true`: output equals input for all line types
- Live `hf config` run to verify visual output.
