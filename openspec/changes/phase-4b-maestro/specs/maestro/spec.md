# Delta Spec: Maestro (Phase 4b)

This delta extends `openspec/specs/maestro/spec.md` with implementation details.

## internal/maestro package

The `internal/maestro` package provides a typed HTTP client for the Maestro REST API.

### Client

```go
type Client struct {
    httpEndpoint string
    http         *http.Client
}
```

- `NewFromConfig(s *config.Store) *Client` — reads `maestro.http-endpoint` from config
- All methods accept a `context.Context` as the first argument
- 30-second HTTP timeout

### Resource types

```go
type ManifestSummary struct { Kind, Name, Namespace string }
type BundleCondition struct { Type, Status, Reason string }
type ResourceBundle struct {
    ID            string
    Name          string
    ConsumerName  string
    Version       int
    ManifestCount int
    Manifests     []ManifestSummary
    Conditions    []BundleCondition
}
type ResourceBundleList struct { Items []ResourceBundle; Kind string; Total int }
type Consumer struct { ID, Kind, Name string }
type ConsumerList struct { Items []Consumer; Kind string; Total int }
```

### Methods

| Method | HTTP | Path |
|---|---|---|
| `ListResourceBundles(ctx)` | GET | `/api/maestro/v1/resource-bundles` |
| `ListResourceBundlesByConsumer(ctx, consumer)` | GET | `/api/maestro/v1/resource-bundles?search=consumer_name='<consumer>'` |
| `GetResourceBundle(ctx, name)` | GET | `/api/maestro/v1/resource-bundles` filtered by name client-side |
| `DeleteResourceBundle(ctx, id)` | DELETE | `/api/maestro/v1/resource-bundles/<id>` |
| `ListConsumers(ctx)` | GET | `/api/maestro/v1/consumers` |

## cmd/maestro.go subcommands

Extends the existing `maestroCmd` stub:

| Subcommand | Config used | Behavior |
|---|---|---|
| `hf maestro list` | `maestro.consumer` | Lists bundles filtered by consumer |
| `hf maestro get <name>` | — | Fetches all bundles, filters by name; interactive if no arg |
| `hf maestro delete <name>` | — | Resolves name to ID, then deletes; interactive if no arg |
| `hf maestro bundles` | — | Lists all resource bundles unfiltered |
| `hf maestro consumers` | — | Lists all consumers |

Default output format for all maestro commands: JSON.
