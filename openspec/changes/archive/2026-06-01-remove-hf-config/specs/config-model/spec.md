## REMOVED Requirements

### Requirement: Set Configuration Value

**Reason**: No CLI setter; users edit environment YAML files directly.

**Migration**: Edit `~/.config/hf/environments/<active>.yaml` or use `HF_*` environment variables.

## ADDED Requirements

### Requirement: State File Path

The config Store MUST expose the absolute path to `state.yaml`.

#### Scenario: StateFilePath returns state.yaml location

- **WHEN** `Store.StateFilePath()` is called
- **THEN** it MUST return `<config-dir>/state.yaml`
