# Design: Phase 4b — Maestro

## Packages

### internal/maestro

`Client` struct holds the HTTP endpoint (from `maestro.http-endpoint`) and gRPC endpoint (from `maestro.grpc-endpoint`). HTTP calls go to the Maestro REST API at `/api/maestro/v1/`.

```go
type Client struct {
    httpEndpoint string
    http         *http.Client
}

func NewFromConfig(s *config.Store) *Client
func (c *Client) ListResourceBundles(ctx context.Context) (*ResourceBundleList, error)
func (c *Client) GetResourceBundle(ctx context.Context, name string) (*ResourceBundle, error)
func (c *Client) DeleteResourceBundle(ctx context.Context, id string) error
func (c *Client) ListConsumers(ctx context.Context) (*ConsumerList, error)
```

Resource types defined in the same package (not in `internal/resource`) because they are Maestro-specific and don't belong with HyperFleet API types.

### cmd/maestro.go

Extends the existing stub `maestroCmd` with five subcommands:

| Subcommand | API call |
|---|---|
| `hf maestro list` | GET `/api/maestro/v1/resource-bundles?consumer.name=<consumer>` |
| `hf maestro get <name>` | GET `/api/maestro/v1/resource-bundles` filtered by name |
| `hf maestro delete <name>` | GET to resolve name→ID, then DELETE `/api/maestro/v1/resource-bundles/<id>` |
| `hf maestro bundles` | GET `/api/maestro/v1/resource-bundles` (unfiltered) |
| `hf maestro consumers` | GET `/api/maestro/v1/consumers` |

## Key Decisions

- The `internal/maestro` client does NOT use `internal/api.Client` because Maestro has a different base path and auth model (no Bearer token by default).
- Default output is JSON (not table) per the spec.
- `hf maestro list` filters by `maestro.consumer` config key; `hf maestro bundles` lists all.
- Interactive selection for `get` / `delete` with no args prints a numbered menu and reads stdin — no external TUI library.
- The existing `cmd/maestro.go` stub's `init()` registers `maestroCmd`; we extend it by adding `AddCommand` calls in a new `init()` block in the same file.
