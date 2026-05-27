## Why

Each environment profile already pins `kubernetes.context`, but kubeconfig path resolution ignores the profile entirely — it only uses the `--kubeconfig` flag, `KUBECONFIG`, or `~/.kube/config`. Users running multiple clusters (kind, GKE, etc.) must export `KUBECONFIG` manually when switching environments instead of storing the path alongside the context in the environment file.

## What Changes

- Add `kubernetes.kubeconfig` config key (empty default = fall through to existing resolution)
- Add `HF_KUBECONFIG` environment variable override (consistent with `HF_CONTEXT`)
- Update kubeconfig resolution precedence: `--kubeconfig` flag → `HF_KUBECONFIG` / env file → `KUBECONFIG` → `~/.kube/config`
- Expose the key in `hf config set` interactive picker and config template
- Update error message to mention `kubernetes.kubeconfig` config key

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `config-model`: add `kubernetes.kubeconfig` key and `HF_KUBECONFIG` env var mapping
- `kubernetes`: document kubeconfig resolution precedence including config file
- `command-hierarchy`: update kubeconfig loading description

## Impact

- `internal/config/assets/config-template.yaml` — new default key
- `internal/config/config.go` — `HF_KUBECONFIG` in `envVarMap`
- `cmd/kube.go` — `resolvedKubeconfig()` reads config store
- `cmd/config.go` — `knownKeysForSection("kubernetes")`
- `internal/kube/kube.go` — error message update (optional)
- Tests in `cmd/kube_test.go`, `internal/config/config_test.go`

## Testing Scope

- `internal/config`: `Get()` returns `kubernetes.kubeconfig` from env file; `HF_KUBECONFIG` overrides file
- `cmd`: `resolvedKubeconfig()` precedence unit test (flag > config > KUBECONFIG > default)

## Live Verification

- Set `kubernetes.kubeconfig` in an environment profile and run `hf kube port-forward status` — context header should reflect the configured kubeconfig/context pair
