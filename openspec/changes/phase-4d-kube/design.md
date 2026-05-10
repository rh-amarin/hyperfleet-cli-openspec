## Context

The `hf` CLI already has `cmd/kube.go` and `cmd/logs.go` as empty stubs (just the top-level cobra command with no subcommands). The `internal/kube/` directory does not yet exist. The spec (`openspec/specs/kubernetes/spec.md`) defines the full surface area.

## Goals / Non-Goals

**Goals:**
- Implement `internal/kube` with all functions named in the spec
- Wire `cmd/kube.go` with `port-forward`, `curl`, and `debug` subcommands
- Wire `cmd/logs.go` with default (pattern matching) and `adapter` subcommands
- Pass `go build ./...`, `go vet ./...`, `go test ./...` with zero failures

**Non-Goals:**
- GKE workload identity / auth plugin support (beyond `HF_KUBE_TOKEN` bypass)
- Websocket-based exec (SPDY is sufficient)
- Interactive shell passthrough for non-debug exec

## Decisions

### 1. Port-forward daemon model

Port-forward processes run as detached child processes writing PID files. The parent CLI returns immediately after `Start`. This matches the spec's `pf-<name>.pid` format.

**Why over in-process goroutines**: The CLI is short-lived; a background OS process survives the parent exit and can be killed independently by `hf kube port-forward stop`.

### 2. `internal/kube` interface for testability

All functions that need a Kubernetes clientset accept `kubernetes.Interface` (not a concrete `*kubernetes.Clientset`). This allows `fake.NewSimpleClientset()` in tests without a real cluster.

**Why**: The spec's `FindRunningPod` and `StreamLogs` need to list pods — fakeable with `fake` package from `k8s.io/client-go/kubernetes/fake`.

### 3. Log streaming fan-out

`StreamLogs` finds all pods matching pattern then launches one goroutine per pod, prefixing lines with `[pod-name]`. If `stern` is on PATH, `hf logs` delegates to it instead.

**Why**: `stern` is the de-facto standard; falling back to goroutine fan-out avoids a hard dependency.

### 4. Config source for kubeconfig

Resolution order: `--kubeconfig` flag → `KUBECONFIG` env → `~/.kube/config`. If not found, print `[ERROR] kubeconfig not found at <path>. Set KUBECONFIG or use --kubeconfig.` and exit 1.

## Risks / Trade-offs

- [Risk] SPDY exec requires the apiserver to support SPDY → Mitigation: GKE standard clusters support it; document as requirement
- [Risk] Port-forward daemon re-launch race (two CLIs starting simultaneously) → Mitigation: PID file written atomically before the child starts
- [Risk] `k8s.io/client-go` is a large dependency tree → Mitigation: acceptable given it's purpose-built for this; already listed in go.mod requirements

## Migration Plan

1. Add dependencies to `go.mod` via `go get`
2. Create `internal/kube/kube.go` with all exported functions
3. Create `internal/kube/kube_test.go` covering unit-testable functions
4. Expand `cmd/kube.go` with subcommands, keeping existing top-level command intact
5. Expand `cmd/logs.go` with subcommands, keeping existing top-level command intact
6. Run verification suite; save proof files

## Open Questions

_(none — spec is complete)_
