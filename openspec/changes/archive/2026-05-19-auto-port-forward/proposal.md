# Proposal: Auto Port-Forward for All API and Maestro Commands

## Problem

Every `hf` command that calls the HyperFleet REST API or Maestro requires the endpoints
to be directly reachable (public LoadBalancer or manually started port-forwards). In
environments where services are only reachable through Kubernetes port-forwarding, the
user must run `hf kube port-forward start` before any cluster/nodepool/maestro command.

## Proposed Solution

A config-driven auto-port-forward mode. When `hyperfleet.auto-port-forward: "true"` is
set in the active environment file, the CLI automatically establishes ephemeral in-process
port-forwards to the API and Maestro services before any command runs, routes all traffic
through them, and tears them down after the command completes.

## Activation

Config option: `hyperfleet.auto-port-forward: "true"` in the env file.

## Scope

Three services forwarded concurrently:
- HyperFleet API (pod `hyperfleet-api`, remote port 8000)
- Maestro HTTP (pod `maestro`, remote port from config, default 8000)
- Maestro gRPC (pod `maestro`, remote port from config, default 8090)
