# Design: Phase 3b NodePool Lifecycle

## Package Structure

No new packages. All code lives in `cmd/` alongside `cluster.go`.

## cmd/nodepool.go

### Helpers reused from cluster.go (same package)
- `newAPIClient(s *config.Store) *api.Client`
- `loadConfig() (*config.Store, error)`
- `handleAPIError(p *output.Printer, err error) error`

### Flag variables
```
nodepoolCreateName     string
nodepoolCreateType     string
nodepoolCreateReplicas int
nodepoolUpdateName     string
nodepoolUpdateReplicas int
```

### Subcommands
| Command | Method | Path |
|---|---|---|
| list | GET | nodepools |
| get [id] | GET | nodepools/<id> |
| create | POST | nodepools |
| update <id> | PATCH | nodepools/<id> |
| delete <id> | DELETE | nodepools/<id> |
| conditions [id] | GET | nodepools/<id> |
| statuses [id] | GET | nodepools/<id>/statuses |

### Table columns
- **list**: ID, NAME, TYPE, GEN, REPLICAS, STATUS
- **conditions**: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE
- **statuses**: ADAPTER, GEN, AVAILABLE

### NodePool type/replicas extraction
`Spec` is `map[string]any`. Type is at `spec["platform"]["type"]` (or `spec["type"]` as fallback). Replicas is at `spec["replicas"]`.

### State persistence
On successful create: `s.SetState("nodepool-id", np.ID)` and print `[INFO] NodePool context set to '<id>'`.

### Duplicate guard on create
Same pattern as cluster: GET list with name search; if found, warn and skip POST.

### Delete behavior
Silent success. On 404: return `fmt.Errorf("[ERROR] NodePool '%s' not found", id)` (exit 1).

## cmd/nodepool_test.go

Mirror `cluster_test.go` structure exactly. Add `resetNodepoolFlags()` helper. Use shared `makeEnv`, `setActiveEnv`, `runCmd` helpers from `config_test.go`. Add `setNodepoolIDInState` parallel to `setClusterIDInState`.
