# Proposal: picker-header-and-config-set-preview

## Problem

Both `hf env` and `hf config set` use interactive fuzzy pickers, but:

1. `hf config set` still uses the simple single-panel picker with no preview — the user selects a key to edit without being able to see the full active configuration.
2. Neither picker shows any contextual help. First-time users have no indication of what the picker is for or how to operate it (keyboard shortcuts).

## Proposed Solution

1. Upgrade `hf config set` interactive mode to the same split-screen layout as `hf env`: left panel for filtering/selecting a config key, right panel showing the full active configuration (colorized YAML). The right panel is static — it always shows the current config regardless of which key is highlighted, because it shows what the user is about to edit.

2. Add a header message above the item list in both pickers using go-fuzzyfinder's WithHeader(). The header names the operation and lists the keyboard shortcuts.

## Changes Required

- internal/selector: add header string parameter to PreviewSelector.SelectWithPreview; use WithHeader() in FuzzyPreviewSelector
- cmd/config.go: change configSetSel to PreviewSelector; add renderConfigPreview helper; update configSetInteractive
- cmd/env.go: pass header constant to SelectWithPreview
- Tests: update mock signatures and configSetSel test double
