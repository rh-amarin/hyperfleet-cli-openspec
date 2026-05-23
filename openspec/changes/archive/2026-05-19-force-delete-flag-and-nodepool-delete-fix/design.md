## Context

`cmd/cluster.go` implements `clusterDeleteCmd` using `api.Delete[resource.Cluster]`. There is no `/force-delete` path for clusters. `cmd/nodepool.go` already has a standalone `nodepoolForceDeleteCmd` subcommand that POSTs to the force-delete endpoint; we are surfacing the same capability via a `--force` flag on the regular delete command and fixing the ID resolution bug in `nodepoolDeleteCmd`.

## Goals / Non-Goals

**Goals:**
- Add `--force` (bool) and `--reason` (string, optional) flags to `hf cluster delete`.
- Add `--force` (bool) and `--reason` (string, optional) flags to `hf nodepool delete`.
- When `--force` is set, POST `{"reason": <reason>}` to `<resource>/force-delete` instead of sending DELETE.
- Fix `nodepoolDeleteCmd` to resolve the active nodepool-id from state via `s.NodePoolID(explicit)` when no explicit ID is provided, consistent with all other single-resource nodepool commands.

**Non-Goals:**
- Removing or changing the existing `hf nodepool force-delete` subcommand (kept for backward compatibility).
- Adding `--force` to `hf cluster delete` on the existing `nodepoolForceDeleteCmd` (already has `--reason`).

## Decisions

**`--reason` is optional (not required):** The existing `nodepoolForceDeleteCmd` requires `--reason`. For the `--force` flag on delete, `--reason` is optional — an empty string is allowed. This keeps the flag simple for the common case while still permitting an audit trail. The API can decide how to handle an absent reason.

**POST to `/force-delete` sub-path:** The force-delete endpoint is a separate action endpoint (`POST .../force-delete`) rather than a flag on the DELETE verb. This matches the existing nodepool force-delete implementation and the server contract.

**Cluster force-delete body:** Same shape as nodepool: `{"reason": "<text>"}`. If reason is empty, send an empty object `{}` or omit the field — we'll omit by sending `map[string]string{"reason": reason}` always (API can handle empty string).

**nodepool-id fallback fix:** Replace the early-return error (`"nodepool ID required"`) with the standard pattern: call `s.NodePoolID(explicit)` after loading config, matching `nodepoolGetCmd`, `nodepoolPatchCmd`, `nodepoolConditionsCmd`, and `nodepoolStatusesCmd`. This makes the command consistent with the spec's "Delete current nodepool" scenario.

## Risks / Trade-offs

- [Risk] Force-deleting a cluster/nodepool is destructive and irreversible → Mitigation: no additional prompt is added (per the user requirement); the `--force` flag is the explicit acknowledgment.
- [Risk] Empty `--reason` may be rejected by the API → Mitigation: the API already handles empty reason in the existing `force-delete` subcommand flow; no change needed here.
