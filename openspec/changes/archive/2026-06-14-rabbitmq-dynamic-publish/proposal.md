## Why

`hf rabbitmq publish` has two hardcoded subcommands (`cluster`, `nodepool`) that cannot publish events for other resource types defined in the environment's `resource-types` config (e.g., `channels`, `versions`). Any new resource type requires a new hardcoded command and a new hand-written event builder — the dynamic config system is completely bypassed.

## What Changes

- **BREAKING**: Remove `hf rabbitmq publish cluster` and `hf rabbitmq publish nodepool` subcommands
- Add single `hf rabbitmq publish <resource-type> <exchange> [routing-key]` command that works for any type defined in `resource-types`
- Add `pubsub.BuildGenericReconcileEvent()` that derives event shape from the resource type's ancestor chain — no per-type code needed
- Remove `pubsub.BuildClusterEvent()` and `pubsub.BuildNodePoolEvent()` (replaced by generic builder)
- CloudEvent `type` uses the config key name verbatim, plural: `com.redhat.hyperfleet.<typeName>.reconcile.v1`
- `owner_references` is automatically populated from the immediate parent in the ancestor chain

## Capabilities

### New Capabilities
- `rabbitmq-dynamic-publish`: Dynamic `hf rabbitmq publish <resource-type>` command driven by `resource-types` config; generic CloudEvent builder with auto-derived event type and owner references

### Modified Capabilities
- `pubsub`: Replace hardcoded cluster/nodepool RabbitMQ publish requirements with the new dynamic resource-type-driven publish requirement

## Impact

- `cmd/rabbitmq.go`: replace two `cobra.Command` vars with one dynamic command
- `internal/pubsub/events.go`: add `BuildGenericReconcileEvent`, remove `BuildClusterEvent` / `BuildNodePoolEvent`
- `internal/pubsub/events_test.go`: replace per-type test cases with generic builder tests
- `cmd/rabbitmq_test.go`: update to use the new command signature
- `openspec/specs/pubsub/spec.md`: update RabbitMQ publish requirements

## Testing Scope

- `internal/pubsub`: `BuildGenericReconcileEvent` — root resource (no owner_references), child resource (owner_references set), deeply-nested resource (two ancestors), missing state ID error
- `cmd/rabbitmq_test.go`: valid resource type dispatches correctly, unknown resource type returns error, missing state returns error

Live cluster verification required for the publish path (RabbitMQ management API call).
