# Delta: kubernetes — Port-forward start with stray-process cleanup

## Modified Scenario: Start port forwards (all)

Replaces the stop requirement in the existing "Start port forwards" scenario:

**Before:**
- AND the CLI MUST stop all tracked port-forwards before starting new ones
- AND print `[INFO] Stopped <name>` for each stopped forward

**After:**
- AND the CLI MUST stop all tracked port-forwards before starting new ones
- AND print `[INFO] Stopped <name>` for each tracked forward that is stopped
- AND the CLI MUST additionally kill any process still listening on each service's local port, even if no PID file exists for that service
- AND print `[INFO] Killed stray process <pid> on port <port>` when such a process is found and terminated

## Modified Scenario: Start port forward — single service

Replaces the stop requirement in the "Start port forward — single service" scenario:

**Before:**
- AND the CLI MUST stop the tracked port-forward for `<name>` when one exists

**After:**
- AND the CLI MUST stop the tracked port-forward for `<name>` when one exists
- AND the CLI MUST additionally kill any process still listening on the service's local port, even if no PID file exists for that service
- AND print `[INFO] Killed stray process <pid> on port <port>` when such a process is found and terminated
