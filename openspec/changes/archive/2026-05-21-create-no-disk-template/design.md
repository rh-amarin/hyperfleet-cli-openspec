## Context

`cmd/templates.go:loadTemplate` currently has two paths when no `--file` flag is given:
1. File exists on disk → read and parse it
2. File does not exist → write embedded default to disk, then parse it

Path 2 is the footgun: once written, the on-disk file is never updated, so users miss changes to the embedded template. Path 1 compounds the problem by silently preferring a potentially stale file over the known-good embedded default.

The fix is to collapse both paths into one: when no `--file` is given, always parse the embedded bytes directly.

## Goals / Non-Goals

**Goals:**
- `loadTemplate` never touches `<config-dir>` unless `--file` points there explicitly
- Users always get the current embedded template when not using `--file`
- `--file` path is entirely unchanged

**Non-Goals:**
- No change to what fields are in the embedded templates
- No change to flag names, API payload structure, or output format
- No migration needed (existing on-disk templates are simply ignored, not deleted)

## Decisions

### Drop `created` return value entirely

The `created` bool existed solely to trigger the `[INFO] Created default … template` message. With no disk write, it is always false and serves no purpose. Removing it simplifies the function signature and removes two dead-code blocks in the callers.

Alternative considered: keep `created` and always return `false`. Rejected — dead return values are noise and make the intent less clear.

### Do not delete existing on-disk template files

Users who already have `~/.config/hf/nodepool-template.json` keep the file untouched. The CLI simply stops reading from it (when `--file` is not given). This is non-destructive and requires no migration.

Alternative considered: delete on-disk files at startup. Rejected — destructive, surprising, and unnecessary.

### No `--use-saved-template` escape hatch

A flag to opt back into the old disk-read behaviour would add complexity with no real benefit — users who want a saved template already have `--file`.

## Risks / Trade-offs

- **Users who customised `<config-dir>/nodepool-template.json`** will silently lose that customisation unless they pass `--file`. Mitigation: this is documented; the `--file` flag covers the use case cleanly.
- **Simpler code** — the `os.Stat` / `os.WriteFile` branch and `MkdirAll` call are removed, making `loadTemplate` a pure parse function for the default path.

## Migration Plan

No migration steps required. `loadTemplate` signature change (`created` bool removed) is internal to the `cmd` package — no external callers.
