## Why

HyperFleet operators need to manually trigger adapter reconciliation by publishing CloudEvents to GCP Pub/Sub topics and RabbitMQ exchanges. Today this requires `gcloud pubsub` or manual HTTP calls; the `hf` CLI should provide this without external tools.

## What Changes

- New `internal/pubsub` package: `GCPPublisher` / `RabbitPublisher` interfaces, `GCPClient` (wraps `cloud.google.com/go/pubsub`), `RabbitClient` (HTTP Management API), and shared `BuildClusterEvent` / `BuildNodePoolEvent` builders
- Extend `cmd/pubsub.go`: add `hf pubsub list [filter]` and `hf pubsub publish cluster|nodepool <topic>` subcommands
- Extend `cmd/rabbitmq.go`: add `hf rabbitmq publish cluster|nodepool <exchange> [routing-key]` subcommands
- Config keys already present in `internal/config`: `hyperfleet.gcp-project`, `rabbitmq.*`

## Capabilities

### New Capabilities

- `pubsub`: GCP Pub/Sub topic listing and CloudEvent publishing; RabbitMQ exchange publishing via HTTP Management API

### Modified Capabilities

_(none — no existing spec-level requirements change)_

## Impact

- New dependency: `cloud.google.com/go/pubsub` (GCP client library)
- Auth: Application Default Credentials (`GOOGLE_APPLICATION_CREDENTIALS` → gcloud ADC → GCE metadata)
- All hrefs constructed from `hyperfleet.api-url` + `hyperfleet.api-version` — never hardcoded
- RabbitMQ calls go to the HTTP Management API (port 15672); no AMQP library needed
- No breaking changes to existing commands

## Testing Scope

| Package | Test cases |
|---|---|
| `internal/pubsub` | `BuildClusterEvent` / `BuildNodePoolEvent` struct validation; `RabbitClient.Publish` via `httptest.NewServer`; `GCPClient` tested via interface mocks |
| `cmd/pubsub.go` | missing-credentials error path; missing-cluster-id error; publish success via mock GCPPublisher |
| `cmd/rabbitmq.go` | missing-cluster-id error; publish success via `httptest.NewServer` for mgmt API |

## Verification

- Live verification requires: GCP credentials (`GOOGLE_APPLICATION_CREDENTIALS` or `gcloud auth application-default login`) and a reachable RabbitMQ management port
- `go build ./...`, `go vet ./...`, `go test ./...` can run without cluster access
