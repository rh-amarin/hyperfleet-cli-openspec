## Packages

### internal/config

**File layout:**
- `config.go` — `Store` struct, `New()`, `Load()`, `Get()`, `Set()`, `ActiveEnvironment()`, `RequireActiveEnvironment()`, `ClusterID()`, `NodePoolID()`, deep-merge helpers
- `config_test.go` — unit tests using temp dirs

**Key decisions:**
- `Store` holds two YAML maps: one for `config.yaml`, one for `state.yaml`
- Precedence: CLI flags passed in at call site > `HF_*` env vars (checked at `Get()` time) > active env profile (deep-merged on load) > `config.yaml` defaults
- Environment profiles loaded from `~/.config/hf/environments/<name>.yaml` and deep-merged over base config when active
- Atomic writes: write to `<file>.tmp`, then `os.Rename` → `<file>`
- Config dir created with `os.MkdirAll` on first `Load()`
- Files written with mode `0600`

**Public API:**
```go
type Store struct { ... }
func New(configDir string) *Store
func (s *Store) Load() error
func (s *Store) Get(section, key string) string
func (s *Store) Set(section, key, value string) error
func (s *Store) GetState(key string) string
func (s *Store) SetState(key, value string) error
func (s *Store) ActiveEnvironment() string
func (s *Store) RequireActiveEnvironment() (string, error)
func (s *Store) ClusterID(explicit string) (string, error)
func (s *Store) NodePoolID(explicit string) (string, error)
func (s *Store) ConfigDir() string
```

### internal/api

**File layout:**
- `client.go` — `Client` struct, `NewClient()`, `Get[T]`, `Post[T]`, `Patch[T]`, `Delete[T]`, `ResourceHref()`
- `errors.go` — `APIError`, `ValidationError`, error parsing logic
- `client_test.go` — httptest-based tests

**Key decisions:**
- `Client` wraps `*http.Client` with 30s timeout
- All methods build URL as `baseURL + path`; baseURL = `{api-url}/api/hyperfleet/{api-version}/`
- Bearer token set only when non-empty
- Error parsing: check status code; if non-2xx, try JSON unmarshal into `APIError`; fall back to raw text
- HTML detection: body starts with `<!` or `<html`
- Verbose logging: `fmt.Fprintf(os.Stderr, "[DEBUG] %s %s → %d (%dms)\n", ...)`

**Public API:**
```go
type Client struct { ... }
func NewClient(baseURL, token string, verbose bool) *Client
func Get[T any](ctx context.Context, c *Client, path string) (T, error)
func Post[T any](ctx context.Context, c *Client, path string, body any) (T, error)
func Patch[T any](ctx context.Context, c *Client, path string, body any) (T, error)
func Delete[T any](ctx context.Context, c *Client, path string) (T, error)
func (c *Client) ResourceHref(resourcePath string) string
```

### internal/resource

**File layout:**
- `resource.go` — all type definitions

**Key decisions:**
- Pure type definitions; no logic
- All fields use `json:"snake_case"` tags
- `ListResponse[T any]` uses Go generics
- `AdapterCondition` vs `ResourceCondition` are distinct types per spec

### internal/output

**File layout:**
- `printer.go` — `Printer` struct, `NewPrinter()`, `Print()`, `PrintTable()`, `Warn()`, `Info()`, `Error()`
- `dots.go` — `StatusDot()` function and color helpers
- `columns.go` — `DynamicColumns()` function
- `printer_test.go` — format and dot renderer tests

**Key decisions:**
- Color detection: TTY check via `term.IsTerminal(int(os.Stdout.Fd()))` from `golang.org/x/term`; OR check `NO_COLOR` env var; OR check `noColor` flag
- JSON colorization: walk the raw JSON bytes and apply ANSI codes to keys/strings/numbers/booleans/null
- Table output: `text/tabwriter` with `\t` separators, flush after writing all rows
- YAML output: `gopkg.in/yaml.v3` encoder
- Dynamic columns: fixed cols + Available (if present) + sorted others + Reconciled (if present)

## Dependencies to Add

```
gopkg.in/yaml.v3
golang.org/x/term
```

## Testing Strategy

- `internal/config`: all tests use `t.TempDir()` to avoid touching real home dir
- `internal/api`: all tests use `httptest.NewServer`; no real network calls
- `internal/resource`: JSON marshal/unmarshal round-trip tests only
- `internal/output`: tests write to `bytes.Buffer`; TTY check bypassed by passing `noColor=true`
