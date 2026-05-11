## Why

`hf config` shows the active environment name but not the file path, so users can't easily locate or inspect the environment file directly. Adding the path removes the need to know the `~/.config/hf/environments/` convention.

## What Changes

- `hf config` and `hf config show` (no argument) will display the absolute path of the active environment file below the environment name header.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `config`: add a requirement that the active environment file path is printed when displaying the active configuration.

## Impact

- `cmd/config.go`: `configShowCmd` prints one extra line with the env file path.
- `internal/config`: `EnvFilePath(name string) string` (or equivalent) may be added if the path is not already surfaced.
- No new dependencies.

## Testing Scope

- **`cmd/`**: existing `configShowCmd` unit tests extended to assert the file path line appears in output.
- Live verification: run `hf config` against the active environment and confirm the path is printed.
