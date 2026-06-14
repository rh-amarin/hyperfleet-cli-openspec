## Packages

| Package | Change |
|---|---|
| `internal/pubsub` | Add `BuildGenericReconcileEvent`; remove `BuildClusterEvent`, `BuildNodePoolEvent` |
| `cmd` | Replace `rabbitmqPublishClusterCmd` + `rabbitmqPublishNodePoolCmd` with one dynamic `rabbitmqPublishCmd` handler |

## Key Decisions

### Event type string uses the type name verbatim (plural)

The CloudEvent `type` field is derived directly from the resource type name as defined in `resource-types` config, without singularization:

```
com.redhat.hyperfleet.<typeName>.reconcile.v1
```

e.g., `clusters` → `com.redhat.hyperfleet.clusters.reconcile.v1`

This matches the environment's own naming convention and requires no extra config field.

### Generic event builder signature

```go
func BuildGenericReconcileEvent(
    typeName string,       // resource-types key (e.g. "nodepools")
    resourceID string,     // active ID from state
    ancestorIDs []AncestorID, // ordered root→immediate-parent
    apiURL, apiVersion string,
) ([]byte, error)
```

`AncestorID` is a small struct `{TypeName, ID string}` — enough to build each ancestor's `href`.

The caller (cmd) resolves state IDs by walking `config.resourceAncestorChain` and calling `s.GetState(def.StateKey)` for each ancestor, then passes them in order.

### owner_references: immediate parent only

Mirrors the existing NodePool behaviour. Deeply nested resources (e.g., `versions` under `channels`) include only the direct parent in `owner_references`. No recursive chain in the payload.

### href construction

Uses the same pattern as the existing pubsub spec:
```
{api-url}/api/hyperfleet/{api-version}/{resolved-path}/{resource-id}
```

The `resolved-path` for child resources is built inline from the ancestor chain (not via `config.ResolveResourcePath`, which requires a full `Store`) — the cmd layer resolves paths via `Store`, passes only the final resolved path string to the builder.

### Command arg validation

The dynamic command validates `args[0]` against the active environment's `resource-types` via `s.ResourceType(typeName)`. An unknown type returns `[ERROR] unknown resource type "<name>"`.

## File Map

```
internal/pubsub/events.go        — add BuildGenericReconcileEvent, AncestorID; remove BuildClusterEvent, BuildNodePoolEvent
internal/pubsub/events_test.go   — replace cluster/nodepool tests with generic builder tests
cmd/rabbitmq.go                  — single dynamic publish command
cmd/rabbitmq_test.go             — update to new command shape
openspec/changes/rabbitmq-dynamic-publish/specs/pubsub/spec.md  — delta spec
```
