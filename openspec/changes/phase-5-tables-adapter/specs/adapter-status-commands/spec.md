# Adapter Status Commands — Delta Spec

## Delta against: openspec/specs/adapter-status/spec.md

This delta adds the Go command interface for adapter status posting as implemented in phase-5.

## Go Command Interface

### hf cluster adapter post-status

```
hf cluster adapter post-status <adapter_name> <True|False|Unknown> <generation>
```

- Requires `cluster-id` in state (`~/.config/hf/state.yaml`)
- Validates `<status>` is one of `True`, `False`, `Unknown`; exits 1 with `[ERROR]` on invalid value
- Parses `<generation>` as integer via `strconv.Atoi`
- Sends POST to `clusters/{cluster_id}/statuses`
- Prints `[INFO] Posted adapter status for <adapter_name> on cluster <cluster_id>` on success
- Outputs the returned `AdapterStatus` JSON (or `{}` on HTTP 204)

### hf nodepool adapter post-status

```
hf nodepool adapter post-status <adapter_name> <True|False|Unknown> <generation> [nodepool_id]
```

- Requires `cluster-id` and `nodepool-id` in state (or explicit 4th arg overrides nodepool-id)
- Same validation as cluster variant
- Sends POST to `clusters/{cluster_id}/nodepools/{nodepool_id}/statuses`
- Prints `[INFO] Posted adapter status for <adapter_name> on nodepool <nodepool_id>`

## Request Payload (confirmed implementation)

```json
{
  "adapter": "<adapter_name>",
  "observed_generation": <int>,
  "observed_time": "<ISO8601 UTC>",
  "conditions": [
    {"type": "Available",  "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Applied",    "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Health",     "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"},
    {"type": "Finalized",  "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status"}
  ]
}
```
