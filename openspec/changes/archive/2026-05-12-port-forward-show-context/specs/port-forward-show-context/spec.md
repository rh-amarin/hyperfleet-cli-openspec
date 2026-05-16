## ADDED Requirements

### Requirement: Display resolved Kubernetes context

`hf kube port-forward start` and `hf kube port-forward status` SHALL print the resolved Kubernetes context name as the first line of their output, before any service lines or status table.

#### Scenario: Context header on start

- **WHEN** the user runs `hf kube port-forward start`
- **THEN** the first line of output MUST be `[INFO] Kubernetes context: <contextName>`
- **AND** `<contextName>` MUST be the context that will actually be used (the `kubernetes.context` config override if set, otherwise the kubeconfig's current-context)
- **AND** the per-service `[INFO] Started …` lines and status table MUST follow as normal

#### Scenario: Context header on status

- **WHEN** the user runs `hf kube port-forward status`
- **THEN** the first line of output MUST be `[INFO] Kubernetes context: <contextName>`
- **AND** the port-forward bullet table (or "No port-forwards tracked.") MUST follow

#### Scenario: Context resolved from config override

- **GIVEN** `kubernetes.context` is set to a non-empty value in the active config (or via `HF_CONTEXT`)
- **WHEN** the user runs `hf kube port-forward start` or `hf kube port-forward status`
- **THEN** the context header MUST show the configured override value
- **AND** all Kubernetes API calls MUST use that context

#### Scenario: Context resolved from kubeconfig current-context

- **GIVEN** `kubernetes.context` is empty (default)
- **WHEN** the user runs `hf kube port-forward start` or `hf kube port-forward status`
- **THEN** the context header MUST show the name of the kubeconfig's current-context

#### Scenario: Context resolution failure

- **GIVEN** the kubeconfig file is missing or the named context does not exist
- **WHEN** the CLI attempts to resolve the context name
- **THEN** the CLI MUST print `[WARN] Could not resolve kubernetes context: <reason>` and continue
