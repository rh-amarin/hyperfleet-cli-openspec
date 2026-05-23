## Context

`hf config env` subcommands (create, activate, list, delete, show) live in `cmd/config.go` as `configEnvCmd` and 5 child vars. They will move verbatim to `cmd/env.go` under `envCmd`.

The interactive picker uses `go-fuzzyfinder v0.9.0` which already supports `WithPreviewWindow(func(i, width, height int) string)`. The preview panel is always rendered on the right half of the terminal (50/50 split, hardcoded by the library) — exactly the desired layout.

## PreviewSelector interface

Added to `internal/selector/selector.go` alongside the existing `Selector` interface:

```go
type PreviewSelector interface {
    SelectWithPreview(items []Item, preview func(i int) string) (int, error)
}

type FuzzyPreviewSelector struct{}

func (FuzzyPreviewSelector) SelectWithPreview(items []Item, preview func(i int) string) (int, error) {
    idx, err := fuzzyfinder.Find(
        items,
        func(i int) string { return items[i].Name },
        fuzzyfinder.WithPreviewWindow(func(i, _, _ int) string {
            if i == -1 { return "" }
            return preview(i)
        }),
    )
    if err == fuzzyfinder.ErrAbort { return -1, nil }
    return idx, err
}
```

The `width` and `height` parameters are ignored — `ColorizeYAMLSections` does not wrap and go-fuzzyfinder handles rendering.

## envCmd bare RunE

```
s = config.NewFromEnv(); s.Load()
names, _ = s.ListEnvironments()

if len(names) == 0:
    cmd.Help()
    fmt.Fprintln("No environments found. Run 'hf env create <name>' to create one.")
    return nil

active = s.ActiveEnvironment()
items[i].Name = names[i] + " ✓" if active else names[i]
previewFn(i) = ColorizeYAMLSections(ReadFile(s.EnvFilePath(names[i])), noColor)

idx = envSel.SelectWithPreview(items, previewFn)
if idx == -1: return nil  // aborted

s.ActivateEnvironment(names[idx])
fmt.Fprintf("[INFO] Activated '<name>'.\n\n")
configShowCmd.RunE(cmd, nil)
```

## Injectable selector

```go
var envSel selector.PreviewSelector = selector.FuzzyPreviewSelector{}
```

Test doubles implement `PreviewSelector` with a fixed index.

## isBypassCommand

Change `strings.Contains(path, "config env")` to `strings.HasPrefix(path, "hf env")`.
`hf env` and all its subcommands must bypass the active-environment guard.

## No backward compat for `hf config env`

`configEnvCmd` and its 5 subcommand vars are deleted. `configCmd.init()` no longer adds them.
The `showEnvProfile` helper moves to `cmd/env.go`.

## Selector display format in left panel

Each row in the left panel: `items[i].Name` (the env name, optionally with ` ✓`). No ID column is needed — environment names are unique and human-readable.

## Preview content

`os.ReadFile(s.EnvFilePath(names[i]))` then `output.ColorizeYAMLSections(string(raw), noColor)`.
If the file is unreadable, the preview shows the error string — never panics.
