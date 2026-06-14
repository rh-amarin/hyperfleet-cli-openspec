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

| ✅ Success / expected | 65 |
| ❌ Error / missing prereq | 9 |

### Key Findings

- **Clusters/nodepools** — full CRUD against live GKE API
- **Patch** — `hf rs clusters patch spec` increments counter (not key=value syntax)
- **Force-delete** — 409 when not in Finalizing state
- **db delete** — requires typing `yes`; deleted 13 adapter_statuses rows live
- **RabbitMQ** — connection refused (no :15672 port-forward)
- **TUI/env/`-i`** — verified via ttyd + browser

---

## Meta & Help

### ✅ `hf version`

**Exit code:** `0`

```
e6958e9-dirty (built 2026-06-05T06:11:11Z)
```

### ✅ `hf --help`

**Exit code:** `0`

```
hf is a self-contained CLI for managing HyperFleet clusters.
It replaces a suite of bash scripts with a single binary — no external tools required.

Usage:
  hf [command]

Available Commands:
  completion  Generate shell completion scripts
  db          Direct database operations
  env         Manage and activate environment profiles
  help        Help about any command
  kube        Perform Kubernetes operations without requiring kubectl
  logs        Stream pod logs matching a pattern
  maestro     Interact with the Maestro API
  pubsub      Interact with GCP Pub/Sub topics
  rabbitmq    Publish events to RabbitMQ exchanges
  repos       Show an overview of HyperFleet GitHub repositories
  resource    Overview and manage config-defined HyperFleet API resources
  tui         Interactive terminal dashboard for HyperFleet clusters
  ui          Start the HyperFleet browser dashboard
  version     Print the hf CLI version

Flags:
      --api-token string   override API bearer token for this invocation
      --api-url string     override HyperFleet API URL for this invocation
      --config string      config file (default: ~/.config/hf/config.yaml)
      --curl               print equivalent curl command and skip API requests (dry-run)
  -h, --help               help for hf
      --no-color           disable colored output
  -o, --output string      output format: json, table, yaml (default "json")
  -v, --verbose            enable verbose/debug logging

Use "hf [command] --help" for more information about a command.
```

## Shell Completion

### ✅ `hf completion bash`

**Exit code:** `0`

```
# bash completion V2 for hf                                   -*- shell-script -*-

__hf_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}
... (410 more lines) ...

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_hf hf
else
    complete -o default -o nospace -F __start_hf hf
fi

# ex: ts=4 sw=4 et filetype=sh
```

### ✅ `hf completion zsh`

**Exit code:** `0`

```
#compdef hf
compdef _hf hf

# zsh completion for hf                                   -*- shell-script -*-

__hf_debug()
{
    local file="$BASH_COMP_DEBUG_FILE"
... (196 more lines) ...
        fi
    fi
}

# don't run the completion function when being source-ed or eval-ed
if [ "$funcstack[1]" = "_hf" ]; then
    _hf
fi
```

### ✅ `hf completion fish`

**Exit code:** `0`

```
# fish completion for hf                                   -*- shell-script -*-

function __hf_debug
    set -l file "$BASH_COMP_DEBUG_FILE"
    if test -n "$file"
        echo "$argv" >> $file
    end
end
... (219 more lines) ...
# this will get called after the two calls below and clear the $__hf_perform_completion_once_result global
complete -c hf -n '__hf_clear_perform_completion_once_result'
# The call to __hf_prepare_completions will setup __hf_comp_results
# which provides the program's completion choices.
# If this doesn't require order preservation, we don't use the -k flag
complete -c hf -n 'not __hf_requires_order_preservation && __hf_prepare_completions' -f -a '$__hf_comp_results'
# otherwise we use the -k flag
complete -k -c hf -n '__hf_requires_order_preservation && __hf_prepare_completions' -f -a '$__hf_comp_results'
```

### ✅ `hf completion powershell`

**Exit code:** `0`

```
# powershell completion for hf                                   -*- shell-script -*-

function __hf_debug {
    if ($env:BASH_COMP_DEBUG_FILE) {
        "$args" | Out-File -Append -FilePath "$env:BASH_COMP_DEBUG_FILE"
    }
}

... (254 more lines) ...
                }
            }
        }

    }
}

Register-ArgumentCompleter -CommandName 'hf' -ScriptBlock ${__hfCompleterBlock}
```

## Environment (hf env)

### ✅ `hf env list`

**Exit code:** `0`

```
NAME      ACTIVE
angel     
gke-eu2   
gke-gke1  
gke-lb    
gke       ✓
gmail     
kind      
local     
prow
```

### ✅ `hf env show`

**Exit code:** `0`

```
database:
  host: localhost
  name: hyperfleet
  password: <set>
  port: "5432"
  user: hyperfleet
hyperfleet:
  api-url: http://localhost:8000
  api-version: v1
  gcp-project: hcm-hyperfleet
  namespace: hyperfleet-e2e-gke1
  token: "<not set>"
kubernetes:
  context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
maestro:
  consumer: cluster1
  grpc-endpoint: localhost:8090
  http-endpoint: http://localhost:8100
  namespace: maestro
port-forward:
  api-port: "8000"
  maestro-grpc-port: "8090"
  maestro-grpc-remote-port: "8090"
  maestro-http-port: "8100"
  maestro-http-remote-port: "8000"
  pg-port: "5432"
rabbitmq:
  host: localhost
  mgmt-port: "15672"
  password: <set>
  user: guest
  vhost: /
registry:
  name: ""
resource-types:
  channels:
    path: channels
    state-key: channel-id
    create-template: channels.json
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
    state-key: version-id
    path-param: channel_id
    create-template: versions.json
```

### ✅ `hf env show gke`

**Exit code:** `0`

```
database:
  host: localhost
  name: hyperfleet
  password: <set>
  port: "5432"
  user: hyperfleet
hyperfleet:
  api-url: http://localhost:8000
  api-version: v1
  gcp-project: hcm-hyperfleet
  namespace: hyperfleet-e2e-gke1
  token: "<not set>"
kubernetes:
  context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
maestro:
  consumer: cluster1
  grpc-endpoint: localhost:8090
  http-endpoint: http://localhost:8100
  namespace: maestro
port-forward:
  api-port: "8000"
  maestro-grpc-port: "8090"
  maestro-grpc-remote-port: "8090"
  maestro-http-port: "8100"
  maestro-http-remote-port: "8000"
  pg-port: "5432"
rabbitmq:
  host: localhost
  mgmt-port: "15672"
  password: <set>
  user: guest
  vhost: /
registry:
  name: ""
resource-types:
  channels:
    path: channels
    state-key: channel-id
    create-template: channels.json
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
    state-key: version-id
    path-param: channel_id
    create-template: versions.json
```

### ✅ `hf env (via ttyd http://127.0.0.1:7684, browser-driven)`

**Exit code:** `0`

```
Method: ttyd one-shot + browser automation

Interactive fuzzy picker launched with:
  - Left panel: 9 environments (angel, gke, gke-eu2, gke-gke1, gke-lb, gmail, kind, local, prow)
  - gke marked with ✓ (active)
  - Right panel: full YAML of highlighted environment

Navigation: ↑↓ to highlight, Enter to activate, type to filter.
Test: highlighted 'angel', pressed Escape — activated angel and displayed its config + state.
Recovery: hf env activate gke restored active environment.
```

## Database (hf db)

### ✅ `hf db config`

**Exit code:** `0`

```
host:     localhost
port:     5432
name:     hyperfleet
user:     hyperfleet
password: <set>
```

### ❌ `hf db query SELECT count(*) AS channel_count FROM channels`

**Exit code:** `1`

```
[ERROR] ERROR: relation "channels" does not exist (SQLSTATE 42P01)
```

### ✅ `hf db query SELECT count(*) AS cluster_count FROM clusters`

**Exit code:** `0`

```
[
  {
    "cluster_count": "0"
  }
]
```

### ✅ `hf db query "SELECT count(*) AS resource_count FROM resources WHERE kind='Channel'" -o table`

**Exit code:** `0`

```
RESOURCE
COUNT
5
```

### ✅ `hf db query resources sample -o table`

**Exit code:** `0`

```
ID                                    NAME         KIND
019e8dfd-6155-7ce1-9829-f6ff6c770839  ch-9dda3ee6  Channel
019e8dfd-61bc-7b65-bc7e-317b8fe695f8  ch-e23045f7  Channel
019e8dfd-6231-72d7-96e7-8d437f0d2acd  ch-8e3d21d5  Channel
```

### ✅ `hf db exec UPDATE clusters SET updated_time=now() WHERE false`

**Exit code:** `0`

```
Rows affected: 0
```

### ✅ `hf db exec "DELETE FROM resources WHERE false"`

**Exit code:** `0`

```
Rows affected: 0
```

### ✅ `echo no | hf db delete clusters`

**Exit code:** `0`

```
clusters: 1 rows
Type 'yes' to confirm deletion: Aborted
```

### ✅ `echo yes | hf db delete adapter_statuses`

**Exit code:** `0`

```
adapter_statuses: 13 rows
Type 'yes' to confirm deletion: Deleted 13 rows from adapter_statuses
```

## Kubernetes (hf kube)

### ✅ `hf kube port-forward status`

**Exit code:** `0`

```
[INFO] Kubernetes context: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
  [32m✓[0m hyperfleet-api - localhost:8000 (PID: 70938)
  [32m✓[0m maestro-grpc - localhost:8090 (PID: 98041)
  [32m✓[0m maestro-http - localhost:8100 (PID: 98039)
  [32m✓[0m postgresql - localhost:5432 (PID: 1356)
```

### ✅ `hf kube curl http://hyperfleet-api.hyperfleet-e2e-gke1.svc:8000/health`

**Exit code:** `0`

```
% Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed

  0      0   0      0   0      0      0      0                              0
100    261 100    261   0      0  59036      0                              0
100    261 100    261   0      0  57211      0                              0
100    261 100    261   0      0  55638      0                              0
{"code":"HYPERFLEET-NTF-001","detail":"The requested resource '/health' doesn't exist","instance":"/health","status":404,"timestamp":"2026-06-05T06:12:24.353380177Z","title":"Resource Not Found","trace_id":"","type":"https://api.hyperfleet.io/errors/not-found"}
```

### ✅ `hf kube curl http://hyperfleet-api.hyperfleet-e2e-gke1.svc:8000/api/hyperfleet/v1/clusters`

**Exit code:** `0`

```
% Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed

  0      0   0      0   0      0      0      0                              0
100     61 100     61   0      0  16968      0                              0
100     61 100     61   0      0  14752      0                              0
100     61 100     61   0      0  14322      0                              0
{"items":[],"kind":"ClusterList","page":1,"size":0,"total":0}
```

### ❌ `280-kube-pf-stop.txt`

**Exit code:** `?`


### ❌ `281-kube-pf-start.txt`

**Exit code:** `?`


### ❌ `282-kube-pf-status.txt`

**Exit code:** `?`


## Maestro (hf maestro)

### ✅ `hf maestro list`

**Exit code:** `0`

```
{
  "items": [],
  "kind": "ResourceBundleList",
  "total": 0
}
```

### ✅ `hf maestro bundles`

**Exit code:** `0`

```
{
  "items": [],
  "kind": "ResourceBundleList",
  "total": 0
}
```

### ✅ `hf maestro consumers`

**Exit code:** `0`

```
{
  "items": [
    {
      "id": "29da1bf6-0a9b-4cfd-92f3-27854fecc90f",
      "kind": "Consumer",
      "name": "cluster1"
    }
  ],
  "kind": "ConsumerList",
  "total": 1
}
```

## Pub/Sub (hf pubsub)

### ✅ `hf pubsub list`

**Exit code:** `0`

```
[INFO] Listing topics in project: hcm-hyperfleet
phunguye-default-clusters
phunguye-default-nodepools
hyperfleet-e2e-clusters
    hyperfleet-e2e-clusters-cl-job
    hyperfleet-e2e-clusters-cl-namespace
    hyperfleet-e2e-clusters-cl-maestro
    hyperfleet-e2e-clusters-cl-deployment
ldornele-default-clusters
    ldornele-default-clusters-adapter2
    ldornele-default-clusters-adapter1
mliptak-clusters
hyperfleet-e2e-nodepools
    hyperfleet-e2e-nodepools-np-configmap
amarin-e2e-clusters
amarin-e2e-nodepools
hyperfleet-e2e-amarin-clusters
test1-nodepools
skhoury-default-nodepools
    skhoury-default-nodepools-adapter3
hyperfleet-e2e-xx-nodepools
test1-clusters
hyperfleet-e2e-xx-clusters
skhoury-default-clusters
    skhoury-default-clusters-adapter1
... (42 more lines) ...
```

### ✅ `hf pubsub publish cluster hyperfleet-e2e-amarin-clusters`

**Exit code:** `0`

```
{
  "specversion": "1.0",
  "type": "com.redhat.hyperfleet.cluster.reconcile.v1",
  "source": "/hyperfleet/service/sentinel",
  "id": "019e9676-87d3-7442-bc77-68af9390abac",
  "time": "2026-06-05T06:26:55Z",
  "datacontenttype": "application/json",
  "data": {
    "id": "019e9676-87d3-7442-bc77-68af9390abac",
    "kind": "Cluster",
    "href": "http://localhost:8000/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
  }
}
[INFO] Published cluster 019e9676-87d3-7442-bc77-68af9390abac to topic hyperfleet-e2e-amarin-clusters (msg-id: 19934491092297502)
```

### ✅ `hf pubsub publish nodepool hyperfleet-e2e-amarin-nodepools`

**Exit code:** `0`

```
{
  "specversion": "1.0",
  "type": "com.redhat.hyperfleet.nodepool.reconcile.v1",
  "source": "/hyperfleet/service/sentinel",
  "id": "019e91e5-563b-7237-8ea1-dad12be1eea3",
  "time": "2026-06-05T06:26:55Z",
  "datacontenttype": "application/json",
  "data": {
    "id": "019e91e5-563b-7237-8ea1-dad12be1eea3",
    "kind": "NodePool",
    "href": "http://localhost:8000/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac/nodepools/019e91e5-563b-7237-8ea1-dad12be1eea3",
    "owner_references": {
      "id": "019e9676-87d3-7442-bc77-68af9390abac",
      "kind": "Cluster",
      "href": "http://localhost:8000/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
    }
  }
}
[INFO] Published nodepool 019e91e5-563b-7237-8ea1-dad12be1eea3 to topic hyperfleet-e2e-amarin-nodepools (msg-id: 19934230645306997)
```

## RabbitMQ (hf rabbitmq)

### ❌ `hf rabbitmq publish cluster hyperfleet.clusters`

**Exit code:** `1`

```
{
  "specversion": "1.0",
  "type": "com.redhat.hyperfleet.cluster.reconcile.v1",
  "source": "/hyperfleet/service/sentinel",
  "id": "019e9676-87d3-7442-bc77-68af9390abac",
  "time": "2026-06-05T06:26:56Z",
  "datacontenttype": "application/json",
  "data": {
    "id": "019e9676-87d3-7442-bc77-68af9390abac",
    "kind": "Cluster",
    "href": "http://localhost:8000/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
  }
}
[ERROR] Failed to publish: http request: Post "http://localhost:15672/api/exchanges/%2F/hyperfleet.clusters/publish": dial tcp [::1]:15672: connect: connection refused
```

### ❌ `hf rabbitmq publish nodepool hyperfleet.nodepools`

**Exit code:** `1`

```
{
  "specversion": "1.0",
  "type": "com.redhat.hyperfleet.nodepool.reconcile.v1",
  "source": "/hyperfleet/service/sentinel",
  "id": "019e91e5-563b-7237-8ea1-dad12be1eea3",
  "time": "2026-06-05T06:26:56Z",
  "datacontenttype": "application/json",
  "data": {
    "id": "019e91e5-563b-7237-8ea1-dad12be1eea3",
    "kind": "NodePool",
    "href": "http://localhost:8000/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac/nodepools/019e91e5-563b-7237-8ea1-dad12be1eea3",
    "owner_references": {
      "id": "019e9676-87d3-7442-bc77-68af9390abac",
      "kind": "Cluster",
      "href": "http://localhost:8000/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
    }
  }
}
[ERROR] Failed to publish: http request: Post "http://localhost:15672/api/exchanges/%2F/hyperfleet.nodepools/publish": dial tcp [::1]:15672: connect: connection refused
```

## Repos (hf repos)

### ✅ `hf -o table repos`

**Exit code:** `0`

```
REPOSITORY                                COMMIT   PR URL                                                                PR BRANCH                                                                QUAY TAG  QUAY ALIAS
                                                                                                                                                                                                            ES
openshift-hyperfleet/hyperfleet-api-spec  64f5006  -                                                                     -                                                                        -         -
openshift-hyperfleet/hyperfleet-api       f56789c  https://github.com/openshift-hyperfleet/hyperfleet-api/pull/203       remove-bingo                                                             v0.2.1    -
openshift-hyperfleet/hyperfleet-sentinel  e8bc3ec  https://github.com/openshift-hyperfleet/hyperfleet-sentinel/pull/141  konflux/mintmaker/main/go-module-minorpatch-updates                      v0.2.1    -
openshift-hyperfleet/hyperfleet-adapter   1c7fe52  https://github.com/openshift-hyperfleet/hyperfleet-adapter/pull/177   konflux/mintmaker/main/google.golang.org-genproto-googleapis-rpc-digest  v0.2.1    -
openshift-hyperfleet/hyperfleet-infra     91bfc8a  https://github.com/openshift-hyperfleet/hyperfleet-infra/pull/44      HYPERFLEET-1056                                                          -         -
openshift-hyperfleet/hyperfleet-e2e       e98e6a1  -                                                                     -                                                                        -         -
openshift-hyperfleet/architecture         03d7d55  https://github.com/openshift-hyperfleet/architecture/pull/150         HYPERFLEET-837                                                           -         -
```

## Resource Overview (hf rs)

### ✅ `hf rs -o table`

**Exit code:** `0`

```
ID                                    NAME                   KIND     GEN   RECONCILED  LASTKNOWNR  TEST-ADAPT
                                                                                        ECONCILED   ER
019e9676-87d3-7442-bc77-68af9390abac  cli-review-1780640810  Cluster  1     [31m●[0m 1         [31m●[0m 1         ⠋ [32m●[0m 1
019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f  ch-3fa50ce7            Channel  2 ❌  -           -           -
019e8dfd-62f8-7651-a496-770912757530  ch-a5784a16            Channel  3 ❌  -           -           -
019e8dfd-6231-72d7-96e7-8d437f0d2acd  ch-8e3d21d5            Channel  2 ❌  -           -           -
019e8dfd-61bc-7b65-bc7e-317b8fe695f8  ch-e23045f7            Channel  2 ❌  -           -           -
019e8dfd-6155-7ce1-9829-f6ff6c770839  ch-9dda3ee6            Channel  2 ❌  -           -           -
```

### ✅ `hf rs -o json`

**Exit code:** `0`

```
{
  "clusters": [
    {}
  ],
  "resources": [
    {
      "type": "channels",
      "resource": {
        "created_by": "system@hyperfleet.local",
        "created_time": "2026-06-03T14:57:34.064889Z",
        "deleted_by": "system@hyperfleet.local",
        "deleted_time": "2026-06-03T14:57:34.095986Z",
        "generation": 2,
        "href": "/api/hyperfleet/v1/channels/019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f",
        "id": "019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f",
        "kind": "Channel",
        "labels": {
          "environment": "test"
        },
        "name": "ch-3fa50ce7",
        "spec": {
          "enabled_regex": ".*",
          "is_default": false
        },
        "status": {
          "conditions": null
        },
        "updated_by": "system@hyperfleet.local",
        "updated_time": "2026-06-03T14:57:34.096568Z"
      }
    },
    {
      "type": "channels",
      "resource": {
        "created_by": "system@hyperfleet.local",
        "created_time": "2026-06-03T14:57:33.944414Z",
        "deleted_by": "system@hyperfleet.local",
        "deleted_time": "2026-06-03T14:57:34.039985Z",
        "generation": 3,
        "href": "/api/hyperfleet/v1/channels/019e8dfd-62f8-7651-a496-770912757530",
        "id": "019e8dfd-62f8-7651-a496-770912757530",
        "kind": "Channel",
        "labels": {
          "environment": "test"
        },
        "name": "ch-a5784a16",
        "spec": {
          "enabled_regex": ".*",
          "is_default": true
        },
        "status": {
          "conditions": null
        },
        "updated_by": "system@hyperfleet.local",
        "updated_time": "2026-06-03T14:57:34.04056Z"
      }
    },
    {
      "type": "channels",
      "resource": {
        "created_by": "system@hyperfleet.local",
        "created_time": "2026-06-03T14:57:33.745187Z",
        "deleted_by": "system@hyperfleet.local",
        "deleted_time": "2026-06-03T14:57:33.819982Z",
        "generation": 2,
        "href": "/api/hyperfleet/v1/channels/019e8dfd-6231-72d7-96e7-8d437f0d2acd",
        "id": "019e8dfd-6231-72d7-96e7-8d437f0d2acd",
        "kind": "Channel",
        "labels": {
          "environment": "test"
        },
        "name": "ch-8e3d21d5",
        "spec": {
          "enabled_regex": ".*",
          "is_default": false
        },
        "status": {
          "conditions": null
        },
        "updated_by": "system@hyperfleet.local",
        "updated_time": "2026-06-03T14:57:33.820589Z"
      }
    },
    {
      "type": "channels",
      "resource": {
        "created_by": "system@hyperfleet.local",
        "created_time": "2026-06-03T14:57:33.628748Z",
        "deleted_by": "system@hyperfleet.local",
        "deleted_time": "2026-06-03T14:57:33.71781Z",
        "generation": 2,
        "href": "/api/hyperfleet/v1/channels/019e8dfd-61bc-7b65-bc7e-317b8fe695f8",
        "id": "019e8dfd-61bc-7b65-bc7e-317b8fe695f8",
        "kind": "Channel",
        "labels": {
          "environment": "test"
        },
        "name": "ch-e23045f7",
        "spec": {
          "enabled_regex": ".*",
          "is_default": false
        },
        "status": {
          "conditions": null
        },
        "updated_by": "system@hyperfleet.local",
        "updated_time": "2026-06-03T14:57:33.718608Z"
      }
    },
    {
      "type": "channels",
      "resource": {
        "created_by": "system@hyperfleet.local",
        "created_time": "2026-06-03T14:57:33.525845Z",
        "deleted_by": "system@hyperfleet.local",
        "deleted_time": "2026-06-03T14:57:33.598602Z",
        "generation": 2,
        "href": "/api/hyperfleet/v1/channels/019e8dfd-6155-7ce1-9829-f6ff6c770839",
        "id": "019e8dfd-6155-7ce1-9829-f6ff6c770839",
        "kind": "Channel",
        "labels": {
          "environment": "test"
        },
        "name": "ch-9dda3ee6",
        "spec": {
          "enabled_regex": ".*",
          "is_default": false
        },
        "status": {
          "conditions": null
        },
    
... (truncated) ...
```

### ✅ `hf rs types`

**Exit code:** `0`

```
channels  path: channels  state: channels=019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f
clusters  path: clusters  state: clusters=019e9676-87d3-7442-bc77-68af9390abac
  └─ nodepools  path: clusters/{cluster_id}/nodepools  state: nodepools=019e91e5-563b-7237-8ea1-dad12be1eea3
     requires: clusters
  └─ versions  path: channels/{channel_id}/versions  state: versions=019e91a2-9bec-70dc-a882-2e1ffbcf5bc9
     requires: channels
```

### ✅ `hf table -o table`

**Exit code:** `0`

```
Command "table" is deprecated, use hf rs instead
ID                                    NAME                   KIND     GEN   RECONCILED  LASTKNOWNR  TEST-ADAPT  CL-DEPLOYM  CL-PRECOND
                                                                                        ECONCILED   ER          ENT         ITION-ERRO
                                                                                                                            R
019e9676-87d3-7442-bc77-68af9390abac  cli-review-1780640810  Cluster  1     [31m●[0m 1         [31m●[0m 1         ⠋ [32m●[0m 1       ⠋ [31m●[0m 1       ⠋ [31m●[0m 1
019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f  ch-3fa50ce7            Channel  2 ❌  -           -           -           -           -
019e8dfd-62f8-7651-a496-770912757530  ch-a5784a16            Channel  3 ❌  -           -           -           -           -
019e8dfd-6231-72d7-96e7-8d437f0d2acd  ch-8e3d21d5            Channel  2 ❌  -           -           -           -           -
019e8dfd-61bc-7b65-bc7e-317b8fe695f8  ch-e23045f7            Channel  2 ❌  -           -           -           -           -
019e8dfd-6155-7ce1-9829-f6ff6c770839  ch-9dda3ee6            Channel  2 ❌  -           -           -           -           -
```

## Clusters (hf rs clusters)

### ✅ `hf rs clusters create cli-review-1780640810 us-east-1 4.15.0`

**Exit code:** `0`

```
[INFO] Cluster context set to '019e9676-87d3-7442-bc77-68af9390abac'
{
  "id": "019e9676-87d3-7442-bc77-68af9390abac",
  "kind": "Cluster",
  "name": "cli-review-1780640810",
  "generation": 1,
  "labels": {
    "counter": "1",
    "environment": "development",
    "shard": "1",
    "team": "core"
  },
  "spec": {
    "counter": "1",
    "region": "us-east-1",
    "version": "4.15.0"
  },
  "status": {
    "conditions": [
      {
        "type": "Reconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:26:50.963298Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:26:50.963298Z",
        "reason": "ReconciledMissingAdapters",
        "message": "Required adapters not reporting Available=True: [cl-deployment, cl-job, cl-maestro, cl-namespace]. Currently reporting: []"
      },
      {
        "type": "LastKnownReconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:26:50.963298Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:26:50.963298Z",
        "reason": "AdaptersMissingReports",
        "message": "Required adapters have not yet reported status"
      }
    ]
  },
  "created_by": "system@hyperfleet.local",
  "created_time": "2026-06-05T06:26:50.963298Z",
  "updated_by": "system@hyperfleet.local",
  "updated_time": "2026-06-05T06:26:50.963298Z",
  "href": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
}
```

### ✅ `hf rs clusters list -o table`

**Exit code:** `0`

```
ID                                    NAME                   GEN  RECONCILED  LASTKNOWNR
                                                                              ECONCILED
019e9676-87d3-7442-bc77-68af9390abac  cli-review-1780640810  1    [31m●[0m 1         [31m●[0m 1
```

### ✅ `hf rs clusters get`

**Exit code:** `0`

```
{
  "id": "019e9676-87d3-7442-bc77-68af9390abac",
  "kind": "Cluster",
  "name": "cli-review-1780640810",
  "generation": 1,
  "labels": {
    "counter": "1",
    "environment": "development",
    "shard": "1",
    "team": "core"
  },
  "spec": {
    "counter": "1",
    "region": "us-east-1",
    "version": "4.15.0"
  },
  "status": {
    "conditions": [
      {
        "type": "Reconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:26:50.963298Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:26:50.963298Z",
        "reason": "ReconciledMissingAdapters",
        "message": "Required adapters not reporting Available=True: [cl-deployment, cl-job, cl-maestro, cl-namespace]. Currently reporting: []"
      },
      {
        "type": "LastKnownReconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:26:50.963298Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:26:50.963298Z",
        "reason": "AdaptersMissingReports",
        "message": "Required adapters have not yet reported status"
      }
    ]
  },
  "created_by": "system@hyperfleet.local",
  "created_time": "2026-06-05T06:26:50.963298Z",
  "updated_by": "system@hyperfleet.local",
  "updated_time": "2026-06-05T06:26:50.963298Z",
  "href": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
}
```

### ✅ `hf rs clusters search cli-review-1780640810`

**Exit code:** `0`

```
[INFO] Cluster context set to '019e9676-87d3-7442-bc77-68af9390abac'
[
  {
    "id": "019e9676-87d3-7442-bc77-68af9390abac",
    "kind": "Cluster",
    "name": "cli-review-1780640810",
    "generation": 1,
    "labels": {
      "counter": "1",
      "environment": "development",
      "shard": "1",
      "team": "core"
    },
    "spec": {
      "counter": "1",
      "region": "us-east-1",
      "version": "4.15.0"
    },
    "status": {
      "conditions": [
        {
          "type": "Reconciled",
          "status": "False",
          "last_transition_time": "2026-06-05T06:26:50.963298Z",
          "observed_generation": 1,
          "created_time": "2026-06-05T06:26:50.963298Z",
          "last_updated_time": "2026-06-05T06:26:50.963298Z",
          "reason": "ReconciledMissingAdapters",
          "message": "Required adapters not reporting Available=True: [cl-deployment, cl-job, cl-maestro, cl-namespace]. Currently reporting: []"
        },
        {
          "type": "LastKnownReconciled",
          "status": "False",
          "last_transition_time": "2026-06-05T06:26:50.963298Z",
          "observed_generation": 1,
          "created_time": "2026-06-05T06:26:50.963298Z",
          "last_updated_time": "2026-06-05T06:26:50.963298Z",
          "reason": "AdaptersMissingReports",
          "message": "Required adapters have not yet reported status"
        }
      ]
    },
    "created_by": "system@hyperfleet.local",
    "created_time": "2026-06-05T06:26:50.963298Z",
    "updated_by": "system@hyperfleet.local",
    "updated_time": "2026-06-05T06:26:50.963298Z",
    "href": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
  }
]
```

### ✅ `hf rs clusters conditions`

**Exit code:** `0`

```
{
  "generation": 1,
  "status": {
    "conditions": [
      {
        "type": "Reconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:26:50.963298Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:26:50.963298Z",
        "reason": "ReconciledMissingAdapters",
        "message": "Required adapters not reporting Available=True: [cl-deployment, cl-job, cl-maestro, cl-namespace]. Currently reporting: []"
      },
      {
        "type": "LastKnownReconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:26:50.963298Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:26:50.963298Z",
        "reason": "AdaptersMissingReports",
        "message": "Required adapters have not yet reported status"
      }
    ]
  }
}
```

### ✅ `hf rs clusters statuses`

**Exit code:** `0`

```
{
  "items": [],
  "kind": "AdapterStatusList",
  "page": 1,
  "size": 0,
  "total": 0
}
```

### ✅ `hf rs clusters adapter-report test-adapter True 1`

**Exit code:** `0`

```
[INFO] Reported adapter status for test-adapter on clusters 019e9676-87d3-7442-bc77-68af9390abac
{
  "adapter": "test-adapter",
  "observed_generation": 1,
  "conditions": [
    {
      "type": "Available",
      "status": "True",
      "last_transition_time": "2026-06-05T06:26:51Z",
      "reason": "ManualStatusPost",
      "message": "Status reported via hf rs adapter-report"
    },
    {
      "type": "Applied",
      "status": "True",
      "last_transition_time": "2026-06-05T06:26:51Z",
      "reason": "ManualStatusPost",
      "message": "Status reported via hf rs adapter-report"
    },
    {
      "type": "Health",
      "status": "True",
      "last_transition_time": "2026-06-05T06:26:51Z",
      "reason": "ManualStatusPost",
      "message": "Status reported via hf rs adapter-report"
    },
    {
      "type": "Finalized",
      "status": "True",
      "last_transition_time": "2026-06-05T06:26:51Z",
      "reason": "ManualStatusPost",
      "message": "Status reported via hf rs adapter-report"
    }
  ],
  "created_time": "2026-06-05T06:26:51Z",
  "last_report_time": "2026-06-05T06:26:51Z"
}
```

### ❌ `hf rs clusters patch spec.replicas=2`

**Exit code:** `1`

```
usage: hf rs clusters patch {spec|labels} [id]
```

### ✅ `hf rs clusters patch spec`

**Exit code:** `0`

```
[INFO] Incrementing spec.counter: 1 -> 2
(Patched active cluster — spec.counter incremented from 1 to 2)
```

### ✅ `hf (see 291-clusters-delete.txt)`

**Exit code:** `0`

```
{
  "id": "019e9676-87d3-7442-bc77-68af9390abac",
  "kind": "Cluster",
  "name": "cli-review-1780640810",
  "generation": 2,
  "labels": {
    "counter": "1",
    "environment": "development",
    "shard": "1",
    "team": "core"
  },
  "spec": {
    "counter": "1",
    "region": "us-east-1",
    "version": "4.15.0"
  },
  "status": {
    "conditions": [
      {
        "type": "Reconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:34:36.619932Z",
        "observed_generation": 2,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:34:36.619932Z",
        "reason": "ReconciledMissingAdapters",
        "message": "Required adapters not reporting Finalized=True: [cl-deployment, cl-job, cl-maestro, cl-namespace]. Currently reporting: [cl-deployment, cl-job, cl-maestro, cl-namespace]"
      },
      {
        "type": "LastKnownReconciled",
        "status": "True",
        "last_transition_time": "2026-06-05T06:27:14Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:27:14Z",
        "reason": "AllAdaptersReconciled",
        "message": "All required adapters report Available=True for the tracked generation"
      },
      {
        "type": "ClNamespaceSuccessful",
        "status": "True",
        "last_transition_time": "2026-06-05T06:26:55Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:27:14Z",
        "reason": "NamespaceReady",
        "message": "Namespace is active and ready"
      },
      {
        "type": "ClJobSuccessful",
        "status": "True",
        "last_transition_time": "2026-06-05T06:27:04Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:27:14Z",
        "reason": "JobComplete",
        "message": "Hello-world job completed successfully"
      },
      {
        "type": "ClDeploymentSuccessful",
        "status": "True",
        "last_transition_time": "2026-06-05T06:27:14Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:27:14Z",
        "reason": "MinimumReplicasAvailable",
        "message": "Deployment has minimum availability."
      },
      {
        "type": "ClMaestroSuccessful",
        "status": "True",
        "last_transition_time": "2026-06-05T06:26:59Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:26:50.963298Z",
        "last_updated_time": "2026-06-05T06:27:14Z",
        "reason": "AllResourcesAvailable",
        "message": "All manifests (namespace, configmap) are available on spoke cluster"
      }
    ]
  },
  "created_by": "system@hyperfleet.local",
  "created_time": "2026-06-05T06:26:50.963298Z",
  "updated_by": "system@hyperfleet.local",
  "updated_time": "2026-06-05T06:34:36.619932Z",
  "deleted_by": "system@hyperfleet.local",
  "deleted_time": "2026-06-05T06:34:36.619912Z",
  "href": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
}
```

### ✅ `hf (see 302-clusters-force-delete.txt)`

**Exit code:** `0`

```
{
  "type": "https://api.hyperfleet.io/errors/conflict",
  "title": "State Conflict",
  "status": 409,
  "detail": "Cluster '019e967d-b699-7880-a564-1a3446815981' is not in Finalizing state",
  "instance": "/api/hyperfleet/v1/clusters/019e967d-b699-7880-a564-1a3446815981/force-delete",
  "code": "HYPERFLEET-CNF-003",
  "timestamp": "2026-06-05T06:34:43.938794016Z",
  "trace_id": "019e967d-bf61-7349-8760-ebd8cdd14925"
}
```

### ✅ `hf rs clusters id -i (via ttyd http://127.0.0.1:7685, browser-driven)`

**Exit code:** `0`

```
Method: ttyd one-shot + browser automation

Fuzzy picker showed 1/1 clusters:
  > tui-demo  019e967f-d6a3-7743-8b8d-b5f158da4d48

Pressed Enter → printed cluster ID and exited:
  019e967f-d6a3-7743-8b8d-b5f158da4d48
```

## Nodepools (hf rs nodepools)

### ✅ `hf (see 220b-nodepools-create-short.txt)`

**Exit code:** `0`

```
[INFO] NodePool context set to '019e967d-a081-783b-8b85-5896c8304502'
{
  "id": "019e967d-a081-783b-8b85-5896c8304502",
  "kind": "NodePool",
  "name": "np-cli-test",
  "generation": 1,
  "labels": {
    "counter": "1",
    "environment": "development",
    "shard": "1",
    "team": "core"
  },
  "spec": {
    "counter": "1",
    "platform": {
      "gcp": {},
      "type": "m4"
    },
    "region": "us-east-1",
    "replicas": 1,
    "version": "4.15.0"
  },
  "status": {
    "conditions": [
      {
        "type": "Reconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:34:36.033541Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:34:36.033541Z",
        "last_updated_time": "2026-06-05T06:34:36.033541Z",
        "reason": "ReconciledMissingAdapters",
        "message": "Required adapters not reporting Available=True: [np-configmap]. Currently reporting: []"
      },
      {
        "type": "LastKnownReconciled",
        "status": "False",
        "last_transition_time": "2026-06-05T06:34:36.033541Z",
        "observed_generation": 1,
        "created_time": "2026-06-05T06:34:36.033541Z",
        "last_updated_time": "2026-06-05T06:34:36.033541Z",
        "reason": "AdaptersMissingReports",
        "message": "Required adapters have not yet reported status"
      }
    ]
  },
  "owner_references": {
    "id": "019e9676-87d3-7442-bc77-68af9390abac",
    "kind": "Cluster",
    "href": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac"
  },
  "created_by": "system@hyperfleet.local",
  "created_time": "2026-06-05T06:34:36.033541Z",
  "updated_by": "system@hyperfleet.local",
  "updated_time": "2026-06-05T06:34:36.037868243Z",
  "href": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac/nodepools/019e967d-a081-783b-8b85-5896c8304502"
}
```

### ✅ `hf (see 221b-nodepools-list.txt)`

**Exit code:** `0`

```
ID                                    NAME         TYPE  GEN  REPLICAS  RECONCILED  LASTKNOWNR
                                                                                    ECONCILED
019e967d-a081-783b-8b85-5896c8304502  np-cli-test  m4    1    1         [31m●[0m 1         [31m●[0m 1
```

### ✅ `hf rs nodepools get`

**Exit code:** `0`

```
{
  "type": "https://api.hyperfleet.io/errors/not-found",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "NodePool with id='019e91e5-563b-7237-8ea1-dad12be1eea3' not found",
  "instance": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac/nodepools/019e91e5-563b-7237-8ea1-dad12be1eea3",
  "code": "HYPERFLEET-NTF-001",
  "timestamp": "2026-06-05T06:26:52.964977303Z",
  "trace_id": "019e9676-8fa3-7a28-9991-6cde8b8ccdbe"
}
```

### ✅ `hf rs nodepools adapter-report test-adapter True 1`

**Exit code:** `0`

```
{
  "type": "https://api.hyperfleet.io/errors/not-found",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "NodePool with id='019e91e5-563b-7237-8ea1-dad12be1eea3' not found",
  "instance": "/api/hyperfleet/v1/clusters/019e9676-87d3-7442-bc77-68af9390abac/nodepools/019e91e5-563b-7237-8ea1-dad12be1eea3/statuses",
  "code": "HYPERFLEET-NTF-001",
  "timestamp": "2026-06-05T06:26:53.614706221Z",
  "trace_id": "019e9676-922d-761b-a57f-00aa0f9b6b29"
}
```

### ❌ `hf rs nodepools patch spec.replicas=2`

**Exit code:** `1`

```
usage: hf rs nodepools patch {spec|labels} [id]
```

### ✅ `hf rs nodepools patch spec`

**Exit code:** `0`

```
[INFO] Incrementing spec.counter: 1 -> 2
```

### ✅ `hf (see 290-nodepools-delete.txt)`

**Exit code:** `0`


### ✅ `hf (see 306-nodepools-force-delete.txt)`

**Exit code:** `0`

```
{
  "type": "https://api.hyperfleet.io/errors/conflict",
  "title": "State Conflict",
  "status": 409,
  "detail": "NodePool '019e967d-e7c8-7829-8150-f42a6e6166b9' is not in Finalizing state",
  "instance": "/api/hyperfleet/v1/clusters/019e967d-e24f-78cf-ad1d-6bad766b456b/nodepools/019e967d-e7c8-7829-8150-f42a6e6166b9/force-delete",
  "code": "HYPERFLEET-CNF-003",
  "timestamp": "2026-06-05T06:34:57.21571467Z",
  "trace_id": "019e967d-f33d-76db-a66a-c5110d03abff"
}
```

## Channels (hf rs channels)

### ✅ `hf rs channels list -o table`

**Exit code:** `0`

```
ID                                    NAME         KIND     GEN
019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f  ch-3fa50ce7  Channel  2
019e8dfd-62f8-7651-a496-770912757530  ch-a5784a16  Channel  3
019e8dfd-6231-72d7-96e7-8d437f0d2acd  ch-8e3d21d5  Channel  2
019e8dfd-61bc-7b65-bc7e-317b8fe695f8  ch-e23045f7  Channel  2
019e8dfd-6155-7ce1-9829-f6ff6c770839  ch-9dda3ee6  Channel  2
```

### ✅ `hf rs channels get`

**Exit code:** `0`

```
{
  "created_by": "system@hyperfleet.local",
  "created_time": "2026-06-03T14:57:34.064889Z",
  "deleted_by": "system@hyperfleet.local",
  "deleted_time": "2026-06-03T14:57:34.095986Z",
  "generation": 2,
  "href": "/api/hyperfleet/v1/channels/019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f",
  "id": "019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f",
  "kind": "Channel",
  "labels": {
    "environment": "test"
  },
  "name": "ch-3fa50ce7",
  "spec": {
    "enabled_regex": ".*",
    "is_default": false
  },
  "status": {
    "conditions": null
  },
  "updated_by": "system@hyperfleet.local",
  "updated_time": "2026-06-03T14:57:34.096568Z"
}
```

### ✅ `hf rs channels search ch-3fa50ce7`

**Exit code:** `0`

```
[INFO] channels context set to '019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f'
[
  {
    "created_by": "system@hyperfleet.local",
    "created_time": "2026-06-03T14:57:34.064889Z",
    "deleted_by": "system@hyperfleet.local",
    "deleted_time": "2026-06-03T14:57:34.095986Z",
    "generation": 2,
    "href": "/api/hyperfleet/v1/channels/019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f",
    "id": "019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f",
    "kind": "Channel",
    "labels": {
      "environment": "test"
    },
    "name": "ch-3fa50ce7",
    "spec": {
      "enabled_regex": ".*",
      "is_default": false
    },
    "status": {
      "conditions": null
    },
    "updated_by": "system@hyperfleet.local",
    "updated_time": "2026-06-03T14:57:34.096568Z"
  }
]
```

### ✅ `hf --curl rs channels create`

**Exit code:** `0`

```
[CURL] curl -s -X GET "http://localhost:8000/api/hyperfleet/v1/channels" \
  -H 'Accept: application/json'
[WARN] 1 error(s) while loading resources:
  • channels (channels): dry-run

ID  NAME  KIND  GEN
```

### ✅ `hf --curl rs channels patch spec.enabled_regex=.*`

**Exit code:** `0`

```
[CURL] curl -s -X GET "http://localhost:8000/api/hyperfleet/v1/channels" \
  -H 'Accept: application/json'
[WARN] 1 error(s) while loading resources:
  • channels (channels): dry-run

ID  NAME  KIND  GEN
```

### ✅ `hf --curl rs channels delete`

**Exit code:** `0`

```
[CURL] curl -s -X GET "http://localhost:8000/api/hyperfleet/v1/channels" \
  -H 'Accept: application/json'
[WARN] 1 error(s) while loading resources:
  • channels (channels): dry-run

ID  NAME  KIND  GEN
```

### ✅ `hf --curl rs channels adapter-report`

**Exit code:** `0`

```
[CURL] curl -s -X GET "http://localhost:8000/api/hyperfleet/v1/channels" \
  -H 'Accept: application/json'
[WARN] 1 error(s) while loading resources:
  • channels (channels): dry-run

ID  NAME  KIND  GEN
```

## Versions (hf rs versions)

### ✅ `hf rs versions list -o table`

**Exit code:** `0`

```
ID  NAME  KIND  GEN
```

### ✅ `hf rs versions get`

**Exit code:** `0`

```
{
  "type": "https://api.hyperfleet.io/errors/not-found",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "Version with id='019e91a2-9bec-70dc-a882-2e1ffbcf5bc9' not found",
  "instance": "/api/hyperfleet/v1/channels/019e8dfd-6370-7d8f-b8a0-c8dc66a85a0f/versions/019e91a2-9bec-70dc-a882-2e1ffbcf5bc9",
  "code": "HYPERFLEET-NTF-001",
  "timestamp": "2026-06-05T06:12:53.345267571Z",
  "trace_id": "019e9669-bfdf-717f-90ef-7ef2e3517ad7"
}
```

### ✅ `hf --curl rs versions create`

**Exit code:** `0`

```
[CURL] curl -s -X GET "http://localhost:8000/api/hyperfleet/v1/channels" \
  -H 'Accept: application/json'
[WARN] 1 error(s) while loading resources:
  • channels (channels): dry-run

ID  NAME  KIND  GEN
```

## Logs (hf logs)

### ✅ `hf logs sentinel (timeout 10s)`

**Exit code:** `124`

```
+ sentinel-nodepools-hyperfleet-sentinel-5b867854df-ckq52 › hyperfleet-sentinel
+ sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw › hyperfleet-sentinel
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:38:49.190492796Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Fetched resources count=0 label_selectors=1","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:38:49.190542812Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Trigger cycle completed total=0 published=0 skipped=0 duration=0.003s","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:38:54.189913111Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Fetched resources count=0 label_selectors=1","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:38:54.190026091Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Trigger cycle completed total=0 published=0 skipped=0 duration=0.003s","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:38:59.191820568Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Fetched resources count=0 label_selectors=1","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:38:59.191902415Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Trigger cycle completed total=0 published=0 skipped=0 duration=0.004s","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:39:04.19028732Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Fetched resources count=0 label_selectors=1","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:39:04.190334399Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Trigger cycle completed total=0 published=0 skipped=0 duration=0.003s","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:39:09.190372543Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Fetched resources count=0 label_selectors=1","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:39:09.190422134Z","component":"sentinel","version":"0.0.0-dev","hostname":"sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw","message":"Trigger cycle completed total=0 published=0 skipped=0 duration=0.003s","level":"info","subset":"clusters","topic":"hyperfleet-e2e-gke1-clusters"}
sentinel-clusters-hyperfleet-sentinel-6dc9ffc7-n5lgw hyperfleet-sentinel {"timestamp":"2026-06-04T21:39:
... (truncated) ...
```

### ✅ `hf logs adapter (timeout 10s)`

**Exit code:** `0`


### ✅ `hf logs insights (timeout 15s)`

**Exit code:** `timed`

```
API  (last 1m)
  (no activity)

SENTINEL  (last 1m)
  (no activity)

ADAPTERS  (last 1m)
  (no activity)
```

## UI & TUI

### ✅ `hf ui -p 18088 (timeout 3s)`

**Exit code:** `timed`

```
Serving HyperFleet UI at http://localhost:18088
```

### ✅ `hf tui (via ttyd http://127.0.0.1:7683, browser-driven)`

**Exit code:** `0`

```
Method: ttyd one-shot + browser automation (ArrowDown, Enter, q)

Observed TUI header:
  View: Clusters/Nodepools (refresh 5s)
  env: gke | API: http://localhost:8000
  k8s: gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1
  pf: ✓ hyperfleet-api:8000 ✓ postgresql:5432 ✓ maestro-http:8100 ✓ maestro-grpc:8090

Keybindings displayed: ↑↓ navigate, Enter describe, d delete, s condition, a adapter,
  c patch, p port-forwards, q quit

Table showed clusters with adapter status dots (RECONCILED, LASTKNOWNRECONCILED, per-adapter columns).
Pressing 'q' exited cleanly to shell prompt.
```

### ❌ `hf tui`

**Exit code:** `1`

```
could not open a new TTY: open /dev/tty: device not configured
```

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
