## Why

Operators frequently need to verify which cluster or nodepool is currently active in their local state before running lifecycle commands. Today the only way to check is `hf config show` or inspecting the state file directly, which surfaces the full config rather than a targeted answer.

`hf cluster id` and `hf nodepool id` provide a fast, scriptable way to read the active IDs — useful in shell scripts, CI pipelines, and day-to-day debugging.

Also fixes a correctness bug: all `hf nodepool` subcommands were posting to the flat `/nodepools` endpoint rather than the cluster-scoped `/clusters/{cluster_id}/nodepools` endpoint specified in the nodepool lifecycle spec, causing the server to return 405 Method Not Allowed on create and other write operations.

## What Changes

- `cmd/cluster.go`: add `hf cluster id` subcommand — prints the active `cluster-id` from state, or exits 1 with a clear error if none is set.
- `cmd/nodepool.go`: add `hf nodepool id` subcommand — prints the active `nodepool-id` from state, or exits 1 with a clear error if none is set. Also fixes all nodepool API paths to use the cluster-scoped route `clusters/{cluster_id}/nodepools/...` as required by the spec.
- No new packages, no new config keys, no API changes.
