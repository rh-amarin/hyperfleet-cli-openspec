# Proposal: hf logs insights

## What

Add a new `hf logs insights` subcommand that fetches logs from running pods (api,
sentinel, and adapter pods) for a configurable time window and produces a
human-readable summary of what the HyperFleet system has been doing.

## Why

Operators and developers currently have no fast way to understand system activity
at a glance. Tailing raw logs via `hf logs` or `stern` requires manual parsing.
`hf logs insights` gives a single command that answers "what happened in the last
N minutes?" in plain language.

## Command signature

```
hf logs insights [-s <duration>]
```

- `-s` / `--since` — time window to look back (Go duration string: `30s`, `5m`, `1h`).
  Default: `1m`.

## Output sections

### API

Requests performed, grouped by `METHOD /normalized/path`, with success (2xx) and
error (4xx/5xx) counts.

```
API  (last 1m)
  GET  /api/hyperfleet/v1/clusters            OK: 12  ERR: 0
  GET  /api/hyperfleet/v1/clusters/:id        OK: 48  ERR: 1
  POST /api/hyperfleet/v1/clusters/:id/statuses  OK: 24  ERR: 0
```

### Sentinel

Number of trigger cycles and published messages per topic, across all sentinel pods.

```
SENTINEL  (last 1m)
  amarin-e2e-clusters    cycles: 12  published: 12  skipped: 0
  amarin-e2e-nodepools   cycles: 12  published: 0   skipped: 12
```

### Adapters

Number of executions per adapter (component name), and which phases they ran,
showing outcome counts (RUNNING→SUCCESS / SKIPPED).

```
ADAPTERS  (last 1m)
  cl-deployment     executions: 12  phases: param_extraction(12) preconditions(12) resources(12) post_actions(12)
  cl-job            executions: 12  phases: param_extraction(12) preconditions(12) resources(12) post_actions(12)
  cl-maestro        executions: 12  phases: param_extraction(12) preconditions(12) resources(0 skipped)
  np-configmap      executions: 0
```

## What Changes

- New `hf logs insights [-s <duration>]` command in `cmd/logs.go`.
- New `internal/insights` package with three log parsers.
- New `kube.CollectLogs` function and exported `kube.ParseLogfmt`.

## Scope

- Read-only — fetches pod logs via `client-go`, no writes.
- No `--output json|table|yaml` — the format is inherently human-readable summary.
- Requires an active environment (kubeconfig + namespace).
- Works with the existing `internal/kube` Kubernetes client; adds `CollectLogs`.
- New `internal/insights` package handles all parsing — fully unit-testable.
