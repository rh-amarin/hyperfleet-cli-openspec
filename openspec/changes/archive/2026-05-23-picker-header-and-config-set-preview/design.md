# Design: picker-header-and-config-set-preview

## selector.PreviewSelector interface change

Add a `header string` parameter to `SelectWithPreview`. When non-empty, the header is rendered above the item list using `fuzzyfinder.WithHeader()`.

```go
type PreviewSelector interface {
    SelectWithPreview(items []Item, preview func(i int) string, header string) (int, error)
}
```

`FuzzyPreviewSelector` builds a `[]fuzzyfinder.Option` slice and appends `WithHeader(header)` only when header != "".

## hf env picker header

A package-level constant in cmd/env.go:
```
hf env  —  select an environment to activate
type to filter  ·  ↑↓ navigate  ·  Enter to activate  ·  Esc to cancel
```

## hf config set preview + header

`configSetSel` changes from `selector.Selector` to `selector.PreviewSelector`.

`configSetInteractive` adds:
1. `renderConfigPreview(s *config.Store) string` — renders the active config sections as colorized YAML (reuses `resolvedSection`, `marshalYAMLOrdered`, `output.ColorizeYAMLSections` already in config.go). The preview is computed once and returned by a static closure `func(_ int) string { return preview }`.
2. A header constant:
```
hf config set  —  select a key to edit
type to filter  ·  ↑↓ navigate  ·  Enter to set value  ·  Esc to cancel
```

## Testing

- `mockPreviewSel.SelectWithPreview` gains `_ string` parameter.
- `TestConfigSet_Interactive` replaces `mockSel{idx:0}` with `mockPreviewSel{idx:0}` since `configSetSel` is now `PreviewSelector`.
- `selector_test.go` gets a compile-time check `var _ PreviewSelector = FuzzyPreviewSelector{}`.
