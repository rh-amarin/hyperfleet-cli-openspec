## ADDED Requirements

### Requirement: CollectLogs

Package `internal/kube` SHALL provide a `CollectLogs` function that fetches pod logs
for a bounded time window and returns all lines as a flat slice.

```go
func CollectLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, sinceSeconds int64) ([]string, error)
```

#### Scenario: Collect logs from matching pods

- GIVEN pods exist in the namespace whose names contain `podPattern`
- WHEN `CollectLogs` is called with a positive `sinceSeconds` value
- THEN it MUST return all log lines from all matching pods as `[]string`
- AND lines from different pods MUST be combined into a single slice

#### Scenario: No matching pods

- GIVEN no pods in the namespace match `podPattern`
- WHEN `CollectLogs` is called
- THEN it MUST return an empty slice with no error

#### Scenario: Pod list error

- GIVEN the Kubernetes API returns an error listing pods
- WHEN `CollectLogs` is called
- THEN it MUST return the error immediately

### Requirement: ParseLogfmt exported

Package `internal/kube` SHALL export `ParseLogfmt(line string) map[string]string`
so that other packages can reuse logfmt parsing without duplication.

#### Scenario: Parse logfmt line

- WHEN `ParseLogfmt` is called with a valid logfmt line
- THEN it MUST return a map of all key-value pairs including quoted values

### Requirement: insights package

Package `internal/insights` SHALL provide three pure log-parsing functions that
operate on `[]string` log lines and return structured summary types.

#### Scenario: ParseAPILogs extracts completed requests

- WHEN `ParseAPILogs` is called with API pod log lines
- THEN it MUST parse only lines where `message == "HTTP request completed"`
- AND group counts by `method + path` with UUIDs normalised to `:id`
- AND track `OK` count (status_code < 400) and `Err` count (status_code >= 400) per group

#### Scenario: ParseSentinelLogs extracts cycle summaries

- WHEN `ParseSentinelLogs` is called with sentinel pod log lines
- THEN it MUST parse only lines where `message` starts with `"Trigger cycle completed"`
- AND accumulate per-topic cycle count, published count, and skipped count

#### Scenario: ParseAdapterLogs extracts adapter activity

- WHEN `ParseAdapterLogs` is called with adapter pod log lines
- THEN it MUST count `"Processing event"` messages per `component` as executions
- AND count `"Phase <name>: RUNNING"` messages per `component` and phase name

### Requirement: Log Insights Command

The CLI SHALL provide `hf logs insights [-s <duration>]` that fetches logs from
running pods and displays a human-readable summary of recent system activity.

#### Scenario: Run log insights with default window

- WHEN the user runs `hf logs insights`
- THEN the CLI MUST fetch logs from the last 1 minute from pods matching `api`, `sentinel`, and `adapter`
- AND display an API section with request counts grouped by `METHOD /normalised/path` and OK/error split
- AND display a Sentinel section with cycle and published-message counts per topic
- AND display an Adapter section with execution counts and phase activity per adapter component

#### Scenario: Run log insights with custom window

- WHEN the user runs `hf logs insights -s 5m`
- THEN the CLI MUST fetch logs from the last 5 minutes
- AND all output sections reflect the extended window

#### Scenario: Invalid duration

- WHEN the user runs `hf logs insights -s notaduration`
- THEN the CLI MUST display `[ERROR] invalid --since value "notaduration": ...`
- AND exit with code 1

#### Scenario: No active environment

- GIVEN no environment is activated
- WHEN the user runs `hf logs insights`
- THEN the CLI MUST fail with `[ERROR]` and exit 1

#### Scenario: No activity in window

- GIVEN pods exist but emitted no relevant log lines in the time window
- WHEN the user runs `hf logs insights`
- THEN the CLI MUST display `(no activity)` for that section
