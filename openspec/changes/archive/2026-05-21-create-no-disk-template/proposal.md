## Why

`hf cluster create` and `hf nodepool create` currently write the embedded default template to `<config-dir>/cluster-template.json` / `nodepool-template.json` on first use. This causes a stale-file problem: if the embedded template changes (e.g. a new field like `shard` is added), existing users keep the old on-disk copy and never see the update. Since `--file` already gives users a clean path to supply their own template, there is no benefit to writing one to disk automatically.

## What Changes

- `loadTemplate` no longer writes anything to `<config-dir>` when no `--file` is given. It uses the embedded default bytes directly in memory.
- The `created` bool return value is removed from `loadTemplate`; callers no longer print the `[INFO] Created default template` message.
- `hf cluster create` and `hf nodepool create` are unchanged in every other respect: `--file` still works, flag overrides still apply, error handling is unchanged.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `cluster-lifecycle`: The "Create cluster with default template (no template file exists)" scenario changes — the CLI MUST NOT write a template to disk; it uses the embedded default in memory.
- `nodepool-lifecycle`: Same change as cluster-lifecycle for the nodepool path.

## Impact

- `cmd/templates.go`: `loadTemplate` drops the `os.Stat` / `os.WriteFile` branch and the `created` return value; always uses embedded bytes when `flagFile` is empty.
- `cmd/cluster.go`: Remove `created` variable and the `if created` info-print block.
- `cmd/nodepool.go`: Same removal.
- `cmd/templates_test.go`: Remove tests asserting the file is written to disk; add/update test asserting it is NOT written.
- No API, config schema, or flag interface changes.

## Testing Scope

| File | Changes |
|---|---|
| `cmd/templates_test.go` | Remove disk-write assertions; assert file is NOT created when no `--file` given |
| `cmd/cluster_test.go` | Remove any assertions on `[INFO] Created default cluster template` message |
| `cmd/nodepool_test.go` | Remove any assertions on `[INFO] Created default nodepool template` message |

Live cluster verification: not required — purely internal behaviour change with no API-visible effect.
