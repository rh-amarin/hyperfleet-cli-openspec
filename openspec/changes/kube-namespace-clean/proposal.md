# Proposal: hf kube namespace-clean

## Problem

After cluster lifecycle operations (create, test, delete), Kubernetes namespaces that were
created for HyperFleet workloads are often left behind. These namespaces can be identified
by labels whose key or value contains the string "hyperfleet". There is no automated way to
bulk-clean them from the CLI today.

## Value

Operators can clean up stale HyperFleet namespaces in one command without writing kubectl
one-liners. The command is safe: it shows the full list and count before asking for
confirmation, and it never deletes silently.

## Scope

- New subcommand `hf kube namespace-clean`
- Lists namespaces matching the label criterion, shows count, prompts y/N, then deletes in
  parallel with per-namespace progress output
- New helper functions in `internal/kube/`: `ListHyperfleetNamespaces`, `DeleteNamespace`
- Unit tests in `cmd/kube_test.go`
