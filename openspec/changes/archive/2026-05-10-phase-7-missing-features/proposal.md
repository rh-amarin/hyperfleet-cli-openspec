## Why

A gap analysis against all archived specs revealed commands that were specified but never implemented in previous phases. The cluster `search` and `patch` commands are needed for the standard test workflow (create → patch → check conditions). The optional cluster-id on `delete`, positional args on `create`, the statuses FINALIZED column, and the config improvements align the CLI with its documented interface.

## What Changes

- `hf cluster search [name]` — search clusters by name, set as active context; no-args falls back to current get behavior
- `hf cluster patch {spec|labels} [id]` — increment counter field in cluster spec or labels, PATCH with new value
- `hf cluster delete [id]` — ID is now optional (falls back to configured cluster-id); outputs deleted cluster JSON
- `hf cluster create [name] [region] [version]` — adds positional arg support; `--name` flag still works
- `hf nodepool search [name]` — mirrors cluster search for nodepools
- `hf nodepool patch {spec|labels} [id]` — mirrors cluster patch for nodepools
- `hf cluster statuses --table` / `hf nodepool statuses --table` — adds FINALIZED column
- `hf config show [env-name]` — optional env-name argument displays a specific environment profile
- `hf config env new` — alias for `hf config env create`
