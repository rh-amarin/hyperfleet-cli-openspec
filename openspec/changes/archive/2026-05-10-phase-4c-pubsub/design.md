## Design

### Package layout

```
internal/pubsub/
├── interfaces.go   — GCPPublisher, RabbitPublisher interfaces; TopicGroup struct
├── events.go       — BuildClusterEvent, BuildNodePoolEvent (shared by both cmd files)
├── gcp.go          — GCPClient: wraps cloud.google.com/go/pubsub; ADC auth
└── rabbitmq.go     — RabbitClient: HTTP Management API via net/http + BasicAuth
cmd/
├── pubsub.go       — hf pubsub list [filter], hf pubsub publish cluster|nodepool <topic>
└── rabbitmq.go     — hf rabbitmq publish cluster|nodepool <exchange> [routing-key]
```

### Interfaces

```go
// GCPPublisher abstracts GCP Pub/Sub operations for testability.
type GCPPublisher interface {
    Publish(ctx context.Context, projectID, topicID string, data []byte) (string, error)
    ListTopics(ctx context.Context, projectID string) ([]TopicGroup, error)
}

// RabbitPublisher abstracts RabbitMQ HTTP management API publishing.
type RabbitPublisher interface {
    Publish(ctx context.Context, baseURL, user, password, vhost, exchange, routingKey string, payload []byte) error
}

// TopicGroup pairs a Pub/Sub topic with its subscriptions.
type TopicGroup struct {
    Topic         string
    Subscriptions []string
}
```

### GCPClient

- Constructed via `NewGCPClient(ctx)` — returns `(*GCPClient, error)` or a sentinel `ErrNoCredentials` if ADC is not configured
- Credential check: `google.FindDefaultCredentials(ctx, pubsub.ScopePubSub)` — if it fails → `ErrNoCredentials`
- `ListTopics`: pages through `pubsub.Client.Topics(ctx)`, then for each topic calls `topic.Subscriptions(ctx)` to collect names → returns `[]TopicGroup`
- `Publish`: creates a `pubsub.Client`, gets or creates a topic reference, calls `topic.Publish(ctx, &pubsub.Message{Data: data}).Get(ctx)` → returns server-assigned message ID

### RabbitClient

- `Publish(ctx, baseURL, user, password, vhost, exchange, routingKey string, payload []byte) error`
  - vhost is URL-encoded: `/` → `%2F`
  - POST to `{baseURL}/api/exchanges/{vhost-encoded}/{exchange}/publish`
  - Body: `{"properties":{},"routing_key":"<routingKey>","payload":"<CloudEvent JSON string>","payload_encoding":"string"}`
  - BasicAuth with user/password
  - Checks HTTP response status code

### CloudEvent builders (`internal/pubsub/events.go`)

Both builders live in `internal/pubsub` to be importable by both `cmd/pubsub.go` and `cmd/rabbitmq.go`.

```go
func BuildClusterEvent(clusterID, apiURL, apiVersion string) ([]byte, error)
func BuildNodePoolEvent(clusterID, nodepoolID, apiURL, apiVersion string) ([]byte, error)
```

href pattern: `{apiURL}/api/hyperfleet/{apiVersion}/clusters/{clusterID}[/nodepools/{nodepoolID}]`

### Commands

**`hf pubsub list [filter]`**
1. `loadConfig()` + ADC check via `gcpFactory`
2. `s.Get("hyperfleet", "gcp-project")` for project ID
3. Calls `GCPPublisher.ListTopics(ctx, projectID)`
4. Prints `[INFO] Listing topics in project: <project>`; then for each TopicGroup prints topic name and subscriptions indented 4 spaces
5. Filter: if arg provided, skip topics/subscriptions whose name doesn't contain the term

**`hf pubsub publish cluster <topic>`**
1. `loadConfig()` → `s.GetState("cluster-id")` — fail with `[ERROR] No cluster-id set in state...` if empty
2. Build CloudEvent → print JSON
3. `GCPPublisher.Publish(ctx, projectID, topic, data)`
4. Print `[INFO] Published cluster <id> to topic <topic> (msg-id: <id>)` on success

**`hf pubsub publish nodepool <topic>`**
- Same as above but also requires `nodepool-id` from state

**`hf rabbitmq publish cluster <exchange> [routing-key]`**
- Build mgmt API base URL from `rabbitmq.host` + `rabbitmq.mgmt-port`
- Read cluster-id from state; build CloudEvent; POST via RabbitClient

**`hf rabbitmq publish nodepool <exchange> [routing-key]`**
- Same but requires both cluster-id and nodepool-id from state

### Factory pattern

Commands use package-level factory vars (injectable in tests):

```go
// cmd/pubsub.go
var gcpFactory func(ctx context.Context) (pubsub.GCPPublisher, error)

// cmd/rabbitmq.go
var rabbitFactory func() pubsub.RabbitPublisher
```

Defaults are set in `init()` to the real constructors.

### Testing strategy

- `internal/pubsub/events_test.go`: unit tests for CloudEvent JSON structure — no network
- `internal/pubsub/rabbitmq_test.go`: `httptest.NewServer` validates request path, method, auth, body
- `cmd/pubsub_test.go`: inject mock `GCPPublisher`; test error paths and success output
- `cmd/rabbitmq_test.go`: inject `httptest.NewServer`-backed `RabbitPublisher` mock; test error paths

### Key decisions

- **No AMQP library**: The spec uses the HTTP Management API only — no `github.com/rabbitmq/amqp091-go` needed
- **GCP auth**: `google.FindDefaultCredentials` returns a meaningful error when no credentials exist; we surface this as the spec's `[ERROR] GCP credentials not found...` message
- **Shared builders**: `BuildClusterEvent` / `BuildNodePoolEvent` are in `internal/pubsub` — `cmd/` files must not import each other
