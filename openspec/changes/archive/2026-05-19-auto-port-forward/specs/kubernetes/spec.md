# Spec: Auto Port-Forward Feature

## Activation

Set `hyperfleet.auto-port-forward: "true"` in the active environment file.
Defaults to `"false"` (backward-compatible).

## Behavior

When enabled, before every command (except bypass commands):

1. Concurrently start ephemeral in-process SPDY port-forwards:
   - `hyperfleet-api` pod → `HF_API_URL=http://127.0.0.1:<random-port>`
   - `maestro` pod → `HF_MAESTRO_HTTP=http://127.0.0.1:<random-port>`
   - `maestro` pod → `HF_MAESTRO_GRPC=127.0.0.1:<random-port>`

2. Each successful forward prints:
   ```
   [INFO] auto port-forward: <service> (<namespace>) → localhost:<port>
   ```

3. Each failed forward prints a warning and the command proceeds without it:
   ```
   [WARN] auto port-forward: <service>: <reason>
   ```

4. After the command completes, all port-forwards are torn down.

## Port allocation

`net.Listen("tcp", "127.0.0.1:0")` selects a random free port; the listener
is closed before port-forwarding begins (brief TOCTOU window, acceptable for CLI use).

## Config keys used

| Key | Section | Default |
|-----|---------|---------|
| `auto-port-forward` | `hyperfleet` | `"false"` |
| `namespace` | `hyperfleet` | `"hyperfleet"` |
| `namespace` | `maestro` | `"maestro"` |
| `maestro-http-remote-port` | `port-forward` | `"8000"` |
| `maestro-grpc-remote-port` | `port-forward` | `"8090"` |

## Env vars set

| Env var | Format |
|---------|--------|
| `HF_API_URL` | `http://127.0.0.1:<port>` |
| `HF_MAESTRO_HTTP` | `http://127.0.0.1:<port>` |
| `HF_MAESTRO_GRPC` | `127.0.0.1:<port>` |
