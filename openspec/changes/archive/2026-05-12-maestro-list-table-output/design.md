# Design: maestro-list-table-output

## Affected packages

- `cmd/maestro.go` — only file that changes

No new packages. No changes to `internal/maestro/` or `internal/output/`.

## Approach

In `maestroListCmd.RunE`, after `list` is fetched, dispatch on `outputFmt`:

```go
if outputFmt == "table" {
    return printMaestroListTable(cmd.OutOrStdout(), list, noColor)
}
return p.Print(list)
```

Add a private function `printMaestroListTable(w io.Writer, list *maestro.ResourceBundleList, noColor bool)` in `cmd/maestro.go` that iterates over `list.Items` and writes the hierarchical output directly.

## Output format

```
<id>  <name>  v<version>
  <kind>/<name>  <namespace>
  <kind>/<name>  <namespace>
```

- Bundle line: `id`, two spaces, `name`, two spaces, `v<version>` (prefix "v" added)
- Manifest lines: two-space indent, `<kind>/<name>`, two spaces, `<namespace>`
- Empty manifest list: no child lines (bundle line is still shown)
- Empty item list: print `No resource bundles.` and return

## Color

- Respect the existing `noColor` global flag; no new color conventions introduced

## Test

- Unit test in `cmd/maestro_test.go` using `httptest.NewServer` to simulate the Maestro API
- Two scenarios: bundle with manifests and empty list
- Assert exact output lines (strip trailing whitespace to avoid fragility)
