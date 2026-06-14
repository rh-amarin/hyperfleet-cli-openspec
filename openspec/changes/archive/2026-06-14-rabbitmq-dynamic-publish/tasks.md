## 1. Generic Event Builder

- [x] 1.1 Add `AncestorID` struct and `BuildGenericReconcileEvent` to `internal/pubsub/events.go`
- [x] 1.2 Remove `BuildClusterEvent` and `BuildNodePoolEvent` from `internal/pubsub/events.go`
- [x] 1.3 Update `internal/pubsub/events_test.go`: replace per-type tests with generic builder tests (root resource, child resource with owner_references, missing ID error)

## 2. Dynamic Command

- [x] 2.1 Replace `rabbitmqPublishClusterCmd` and `rabbitmqPublishNodePoolCmd` in `cmd/rabbitmq.go` with a single dynamic `publish` handler that accepts `<resource-type> <exchange> [routing-key]`
- [x] 2.2 Update `cmd/rabbitmq_test.go` to cover: valid type dispatches correctly, unknown type returns error, missing state returns error

## 3. Verification

- [x] 3.1 Run `go build ./...` and `go vet ./...` — zero errors
- [x] 3.2 Run `go test ./...` — capture output to `verification_proof/unit_tests.txt`
- [x] 3.3 Live verify: `hf rabbitmq publish clusters <exchange>` against a real environment — save output to `verification_proof/live_publish.txt`
