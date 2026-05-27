## MODIFIED Requirements

### Requirement: Kube Command Group

The `hf kube` command group SHALL provide Kubernetes utility subcommands.

#### Scenario: Kubeconfig and context resolution

- **WHEN** any `hf kube` subcommand or auto-port-forward runs
- **THEN** kubeconfig loading MUST respect this precedence: `--kubeconfig` flag → `kubernetes.kubeconfig` config (including `HF_KUBECONFIG`) → `KUBECONFIG` env → `~/.kube/config`
- **AND** context selection MUST respect: `kubernetes.context` config (including `HF_CONTEXT`) → kubeconfig current-context
