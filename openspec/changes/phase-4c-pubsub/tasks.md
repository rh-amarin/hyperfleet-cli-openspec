# Tasks: phase-4c-pubsub

- [x] 1. Add `cloud.google.com/go/pubsub` dependency via `go get`
- [x] 2. Create `internal/pubsub/interfaces.go` — `GCPPublisher`, `RabbitPublisher` interfaces and `TopicGroup` struct
- [x] 3. Create `internal/pubsub/events.go` — `BuildClusterEvent` and `BuildNodePoolEvent` builders
- [x] 4. Create `internal/pubsub/events_test.go` — unit tests for both builders
- [x] 5. Create `internal/pubsub/gcp.go` — `GCPClient` implementing `GCPPublisher` with ADC auth
- [x] 6. Create `internal/pubsub/rabbitmq.go` — `RabbitClient` implementing `RabbitPublisher` via HTTP management API
- [x] 7. Create `internal/pubsub/rabbitmq_test.go` — test `RabbitClient.Publish` via `httptest.NewServer`
- [x] 8. Extend `cmd/pubsub.go` — implement `hf pubsub list [filter]` and `hf pubsub publish cluster|nodepool <topic>`
- [x] 9. Create `cmd/pubsub_test.go` — test pubsub commands with mock `GCPPublisher`
- [x] 10. Extend `cmd/rabbitmq.go` — implement `hf rabbitmq publish cluster|nodepool <exchange> [routing-key]`
- [x] 11. Create `cmd/rabbitmq_test.go` — test rabbitmq commands via mock `RabbitPublisher`
- [x] 12. Run `go build ./...`, `go vet ./...`, `go test ./...` — capture output to `verification_proof/`
- [x] 13. Commit `verification_proof/` and all implementation files
