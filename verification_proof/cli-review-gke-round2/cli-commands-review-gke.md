# HyperFleet CLI — Live Command Review (GKE)

Generated: 2026-06-05 06:40 UTC

> **Round 2 update:** Environment includes `clusters` and `nodepools` in `resource-types`.
> Destructive commands executed live. Interactive commands tested via **ttyd** + browser.

## Test Environment

| Setting | Value |
|---------|-------|
| Active environment | `gke` |
| Kubernetes context | `gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1` |
| Namespace | `hyperfleet-e2e-gke1` |
| Resource types | `clusters`, `nodepools`, `channels`, `versions` |
| Port-forwards | API `:8000`, PostgreSQL `:5432`, Maestro HTTP `:8100`, gRPC `:8090` |
| Interactive testing | ttyd ports 7683–7685, browser-driven |

## Results Summary

| ✅ Success / expected | 0 |
| ❌ Error / missing prereq | 0 |

### Key Findings

- **Clusters/nodepools** — full CRUD against live GKE API
- **Patch** — `hf rs clusters patch spec` increments counter (not key=value syntax)
- **Force-delete** — 409 when not in Finalizing state
- **db delete** — requires typing `yes`; deleted 13 adapter_statuses rows live
- **RabbitMQ** — connection refused (no :15672 port-forward)
- **TUI/env/`-i`** — verified via ttyd + browser

---

## Meta & Help

### ⚠️ `00-version.txt` — capture missing

### ⚠️ `01-help.txt` — capture missing

## Shell Completion

### ⚠️ `02-completion-bash.txt` — capture missing

### ⚠️ `03-completion-zsh.txt` — capture missing

### ⚠️ `04-completion-fish.txt` — capture missing

### ⚠️ `05-completion-powershell.txt` — capture missing

## Environment (hf env)

### ⚠️ `10-env-list.txt` — capture missing

### ⚠️ `11-env-show.txt` — capture missing

### ⚠️ `12-env-show-gke.txt` — capture missing

### ⚠️ `401-tty-env-picker.txt` — capture missing

## Database (hf db)

### ⚠️ `20-db-config.txt` — capture missing

### ⚠️ `21-db-query-count-channels.txt` — capture missing

### ⚠️ `23-db-query-clusters.txt` — capture missing

### ⚠️ `25-db-query-resources.txt` — capture missing

### ⚠️ `26-db-query-resources-sample.txt` — capture missing

### ⚠️ `262-db-exec-noop.txt` — capture missing

### ⚠️ `27-db-exec-dryrun.txt` — capture missing

### ⚠️ `263-db-delete-clusters-abort.txt` — capture missing

### ⚠️ `264-db-delete-adapter-statuses.txt` — capture missing

## Kubernetes (hf kube)

### ⚠️ `30-kube-pf-status.txt` — capture missing

### ⚠️ `31-kube-curl-api-health.txt` — capture missing

### ⚠️ `32-kube-curl-api-clusters.txt` — capture missing

### ⚠️ `280-kube-pf-stop.txt` — capture missing

### ⚠️ `281-kube-pf-start.txt` — capture missing

### ⚠️ `282-kube-pf-status.txt` — capture missing

## Maestro (hf maestro)

### ⚠️ `40-maestro-list.txt` — capture missing

### ⚠️ `41-maestro-bundles.txt` — capture missing

### ⚠️ `42-maestro-consumers.txt` — capture missing

## Pub/Sub (hf pubsub)

### ⚠️ `50-pubsub-list.txt` — capture missing

### ⚠️ `250-pubsub-publish-cluster.txt` — capture missing

### ⚠️ `251-pubsub-publish-nodepool.txt` — capture missing

## RabbitMQ (hf rabbitmq)

### ⚠️ `252-rabbitmq-publish-cluster.txt` — capture missing

### ⚠️ `253-rabbitmq-publish-nodepool.txt` — capture missing

## Repos (hf repos)

### ⚠️ `70-repos-table.txt` — capture missing

## Resource Overview (hf rs)

### ⚠️ `240-rs-overview-table.txt` — capture missing

### ⚠️ `241-rs-overview-json.txt` — capture missing

### ⚠️ `242-rs-types.txt` — capture missing

### ⚠️ `243-table-deprecated.txt` — capture missing

## Clusters (hf rs clusters)

### ⚠️ `200-clusters-create.txt` — capture missing

### ⚠️ `201-clusters-list.txt` — capture missing

### ⚠️ `203-clusters-get.txt` — capture missing

### ⚠️ `205-clusters-search.txt` — capture missing

### ⚠️ `206-clusters-conditions.txt` — capture missing

### ⚠️ `207-clusters-statuses.txt` — capture missing

### ⚠️ `208-clusters-adapter-report.txt` — capture missing

### ⚠️ `209-clusters-patch.txt` — capture missing

### ⚠️ `304-clusters-patch-spec.txt` — capture missing

### ⚠️ `291-clusters-delete.txt` — capture missing

### ⚠️ `302-clusters-force-delete.txt` — capture missing

### ⚠️ `402-tty-clusters-id-i.txt` — capture missing

## Nodepools (hf rs nodepools)

### ⚠️ `220b-nodepools-create-short.txt` — capture missing

### ⚠️ `221b-nodepools-list.txt` — capture missing

### ⚠️ `223-nodepools-get.txt` — capture missing

### ⚠️ `228-nodepools-adapter-report.txt` — capture missing

### ⚠️ `229-nodepools-patch.txt` — capture missing

### ⚠️ `305-nodepools-patch-spec.txt` — capture missing

### ⚠️ `290-nodepools-delete.txt` — capture missing

### ⚠️ `306-nodepools-force-delete.txt` — capture missing

## Channels (hf rs channels)

### ⚠️ `90-channels-list.txt` — capture missing

### ⚠️ `92-channels-get.txt` — capture missing

### ⚠️ `94-channels-search.txt` — capture missing

### ⚠️ `96-channels-create-curl.txt` — capture missing

### ⚠️ `97-channels-patch-curl.txt` — capture missing

### ⚠️ `98-channels-delete-curl.txt` — capture missing

### ⚠️ `99-channels-adapter-report-curl.txt` — capture missing

## Versions (hf rs versions)

### ⚠️ `100-versions-list.txt` — capture missing

### ⚠️ `101-versions-get.txt` — capture missing

### ⚠️ `103-versions-create-curl.txt` — capture missing

## Logs (hf logs)

### ⚠️ `270-logs-sentinel.txt` — capture missing

### ⚠️ `271-logs-adapter.txt` — capture missing

### ⚠️ `111-logs-insights.txt` — capture missing

## UI & TUI

### ⚠️ `120-ui-start.txt` — capture missing

### ⚠️ `400-tty-tui.txt` — capture missing

### ⚠️ `121-tui.txt` — capture missing

## Destructive Commands Tested

| Command | Result |
|---------|--------|
| `hf rs clusters create` | Created test clusters on GKE |
| `hf rs nodepools create` | Created `np-cli-test` (name ≤15 chars) |
| `hf rs clusters/nodepools patch spec` | Counter 1→2 |
| `hf rs nodepools delete` | Soft-delete nodepool |
| `hf rs clusters delete` | Soft-delete cluster |
| `hf rs * force-delete` | 409 conflict (not Finalizing) |
| `echo yes \| hf db delete adapter_statuses` | Deleted 13 rows |
| `echo no \| hf db delete clusters` | Aborted |
| `hf pubsub publish cluster/nodepool` | Published to GCP |
| `hf kube port-forward stop/start` | Cycled 4 forwards |

## Interactive (ttyd + browser)

| Command | Port | Result |
|---------|------|--------|
| `hf tui` | 7683 | Cluster table + keybindings; `q` quits |
| `hf env` | 7684 | 9-env fuzzy picker + YAML preview |
| `hf rs clusters id -i` | 7685 | Selected cluster from fuzzy list |

## Raw Captures

- `verification_proof/cli-review-gke/` — round 1 (meta, channels, kube, maestro, …)
- `verification_proof/cli-review-gke-round2/` — round 2 (clusters, nodepools, destructive, ttyd)
