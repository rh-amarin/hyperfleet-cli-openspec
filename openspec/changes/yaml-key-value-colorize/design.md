## Key Decisions

**D1 — Replace, not wrap, `ColorizeYAMLSections`.**
The function signature stays the same so all callers (`cmd/config.go`) are untouched.
The implementation is rewritten in place; no new exported symbol is needed.

**D2 — Line-by-line regex-free parsing.**
Each line is classified by its leading whitespace and colon position using
`strings.IndexByte` and `strings.TrimLeft` — no `regexp` import, keeping
the output package dependency-free.

**D3 — ANSI palette.**
| Constant | Code | Usage |
|----------|------|-------|
| `ansiBoldCyan` | `\033[1;36m` | Section headers (existing) |
| `ansiBoldWhite` | `\033[1;37m` | Indented keys |
| `ansiGreen` | `\033[32m` | Values |
| `ansiDim` | `\033[2m` | `: ` separator token and sentinel values (`<…>`) |
| `ansiReset` | `\033[0m` | Reset after every colored segment |

**D4 — Sentinel detection.**
A value is a sentinel when it starts with `<` and ends with `>` (e.g. `<not set>`, `<set>`).
Sentinels are rendered dim instead of green.

**D5 — No changes outside `internal/output/config_display.go` and its test file.**
`cmd/config.go` is not modified; the improved colorization is transparent to callers.
