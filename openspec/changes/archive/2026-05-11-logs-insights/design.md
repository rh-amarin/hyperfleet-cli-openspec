# Design: hf logs insights

## Packages

### `internal/kube` — new function `CollectLogs`

```go
func CollectLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, sinceSeconds int64) ([]string, error)
```

- Lists all pods in `namespace` whose name contains `podPattern`.
- For each matching pod, fetches logs with `corev1.PodLogOptions{SinceSeconds: &sinceSeconds}` (non-streaming).
- Collects all lines across pods and returns them as `[]string`.
- Runs pod fetches sequentially (simplicity over micro-optimisation — the window is small).

### `internal/insights` — new package

Three independent parsers, each operating on `[]string` (raw log lines):

#### `ParseAPILogs(lines []string) APIInsights`

API pods emit **JSON** log lines. Each completed request is logged as:
```json
{"message":"HTTP request completed","method":"GET","path":"/api/hyperfleet/v1/clusters","status_code":200,...}
```

- Filter lines where `message == "HTTP request completed"`.
- Normalize `path` by replacing UUID-shaped segments with `:id`
  (regex: `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}` → `:id`).
- Group by `METHOD + normalized_path`, count `ok` (status < 400) vs `err` (status >= 400).

```go
type APIEndpointStat struct {
    Method string
    Path   string
    OK     int
    Err    int
}
type APIInsights struct {
    Endpoints []APIEndpointStat
}
```

#### `ParseSentinelLogs(lines []string) SentinelInsights`

Sentinel pods emit **JSON** log lines. Cycle summaries look like:
```json
{"message":"Trigger cycle completed total=1 published=1 skipped=0 duration=0.017s","topic":"amarin-e2e-clusters",...}
```

- Filter lines where `message` starts with `"Trigger cycle completed"`.
- Parse `total=N`, `published=N`, `skipped=N` from the message string.
- Group by `topic` field, accumulate cycle count, published count, skipped count.

```go
type SentinelTopicStat struct {
    Topic     string
    Cycles    int
    Published int
    Skipped   int
}
type SentinelInsights struct {
    Topics []SentinelTopicStat
}
```

#### `ParseAdapterLogs(lines []string) AdapterInsights`

Adapter pods emit **logfmt** lines. Key events:
```
time=... level=INFO msg="Processing event" component=cl-deployment cluster_id=... event_id=...
time=... level=INFO msg="Phase param_extraction: RUNNING" component=cl-deployment ...
time=... level=INFO msg="Phase resources: SUCCESS - 1 processed" component=cl-deployment ...
time=... level=INFO msg="Phase preconditions: SKIPPED" component=cl-deployment ...
```

- Parse logfmt (reuse existing `parseLogfmt` — will be exported as `ParseLogfmt`).
- Track per `component`:
  - Count of `"Processing event"` messages → executions
  - For each `"Phase <name>: RUNNING"` msg → increment phase run count
  - (SUCCESS/SKIPPED appear in later msgs but `RUNNING` is the definitive marker that a phase executed)

```go
type AdapterPhaseStat struct {
    Name  string
    Count int
}
type AdapterStat struct {
    Name       string
    Executions int
    Phases     []AdapterPhaseStat
}
type AdapterInsights struct {
    Adapters []AdapterStat
}
```

### `cmd/logs.go` — new `logsInsightsCmd`

```
hf logs insights [-s <duration>]
```

- Flag: `--since` / `-s` `string`, default `"1m"`.
- Parse duration with `time.ParseDuration`; convert to `int64` seconds.
- Determine namespace from config (default `"my-namespace"`).
- Build `kube.NewClientset` from resolved kubeconfig.
- Collect logs in parallel for three pod groups: `"api"`, `"sentinel"`, `"adapter"`.
- Call the three parsers.
- Print human-readable output to `cmd.OutOrStdout()`.

## Key decisions

| Decision | Choice | Reason |
|---|---|---|
| Log collection | Pull (non-streaming) with `sinceSeconds` | Bounded window; streaming would require a timeout dance |
| Parallelism | Three goroutines for the three pod groups | api/sentinel/adapter are independent; reduces latency |
| `parseLogfmt` | Export existing private function as `ParseLogfmt` from `internal/kube` | Reuse; tested there |
| `--output` flag | Not supported | Command is inherently a human-readable summary |
| Phase counting | Count `RUNNING` msgs | `RUNNING` is always emitted first; `SUCCESS`/`SKIPPED` are outcome variants |
| Path normalisation | UUID regex only | Sufficient for current API paths; no over-engineering |
| No `--output json` | Intentional | Proposal specifies "human understandable" |
| Test approach | `httptest`-free — pure string slice input | Parser logic has no HTTP dependency; unit tests feed raw lines |

## Files changed

| File | Change |
|---|---|
| `internal/kube/kube.go` | Add `CollectLogs`; export `ParseLogfmt` (rename private → public, update call sites) |
| `internal/kube/kube_test.go` | Add `CollectLogs` test |
| `internal/insights/insights.go` | New — `ParseAPILogs`, `ParseSentinelLogs`, `ParseAdapterLogs` |
| `internal/insights/insights_test.go` | New — unit tests for all three parsers |
| `cmd/logs.go` | Add `logsInsightsCmd`, registered under `logsCmd` |
| `openspec/changes/logs-insights/specs/kubernetes/spec.md` | Delta spec for new kube function and new command |
