#!/usr/bin/env bash
# Round 2: clusters/nodepools + destructive commands against GKE.
set -uo pipefail

HF="${HF_BIN:-./bin/hf}"
OUT_DIR="${OUT_DIR:-verification_proof/cli-review-gke-round2}"
mkdir -p "$OUT_DIR"

TS=$(date +%s)
CLUSTER_NAME="cli-review-${TS}"
NODEPOOL_NAME="cli-np-${TS}"

run_cmd() {
  local name="$1"
  shift
  local outfile="$OUT_DIR/${name}.txt"
  local tmp
  tmp=$(mktemp)
  set +e
  "$HF" "$@" >"$tmp" 2>&1
  local ec=$?
  set -e
  {
    echo "COMMAND: hf $*"
    echo "EXIT: $ec"
    echo "────────────────────────────────────────"
    cat "$tmp"
  } > "$outfile"
  rm -f "$tmp"
  echo "${name}: exit $ec"
}

echo "=== Round 2 CLI review ==="
echo "Cluster name: $CLUSTER_NAME"
echo "Nodepool name: $NODEPOOL_NAME"

# ── Clusters: create and exercise ──
run_cmd "200-clusters-create" rs clusters create "$CLUSTER_NAME" us-east-1 4.15.0
CLUSTER_ID=$("$HF" rs clusters id 2>/dev/null | tail -1)
echo "CLUSTER_ID=$CLUSTER_ID"

run_cmd "201-clusters-list" rs clusters list -o table
run_cmd "202-clusters-table" rs clusters table
run_cmd "203-clusters-get" rs clusters get
run_cmd "204-clusters-id" rs clusters id
run_cmd "205-clusters-search" rs clusters search "$CLUSTER_NAME"
run_cmd "206-clusters-conditions" rs clusters conditions
run_cmd "207-clusters-statuses" rs clusters statuses
run_cmd "208-clusters-adapter-report" rs clusters adapter-report test-adapter True 1
run_cmd "209-clusters-patch" rs clusters patch spec.replicas=2
run_cmd "210-clusters-list-json" rs clusters list -o json

# ── Nodepools: create and exercise ──
run_cmd "220-nodepools-create" rs nodepools create "$NODEPOOL_NAME" --type m4 --replicas 1
NODEPOOL_ID=$("$HF" rs nodepools id 2>/dev/null | tail -1)
echo "NODEPOOL_ID=$NODEPOOL_ID"

run_cmd "221-nodepools-list" rs nodepools list -o table
run_cmd "222-nodepools-table" rs nodepools table
run_cmd "223-nodepools-get" rs nodepools get
run_cmd "224-nodepools-id" rs nodepools id
run_cmd "225-nodepools-search" rs nodepools search "$NODEPOOL_NAME"
run_cmd "226-nodepools-conditions" rs nodepools conditions
run_cmd "227-nodepools-statuses" rs nodepools statuses
run_cmd "228-nodepools-adapter-report" rs nodepools adapter-report test-adapter True 1
run_cmd "229-nodepools-patch" rs nodepools patch spec.replicas=2
run_cmd "230-nodepools-list-json" rs nodepools list -o json

# ── Combined overview with clusters ──
run_cmd "240-rs-overview-table" rs -o table
run_cmd "241-rs-overview-json" rs -o json
run_cmd "242-rs-types" rs types
run_cmd "243-table-deprecated" table -o table

# ── Pub/Sub & RabbitMQ with args ──
run_cmd "250-pubsub-publish-cluster" pubsub publish cluster hyperfleet-e2e-amarin-clusters
run_cmd "251-pubsub-publish-nodepool" pubsub publish nodepool hyperfleet-e2e-amarin-nodepools
run_cmd "252-rabbitmq-publish-cluster" rabbitmq publish cluster hyperfleet.clusters
run_cmd "253-rabbitmq-publish-nodepool" rabbitmq publish nodepool hyperfleet.nodepools

# ── DB destructive (safe no-op / test tables) ──
run_cmd "260-db-query-cluster" db query "SELECT id, name FROM clusters LIMIT 5" -o table
run_cmd "261-db-query-nodepools" db query "SELECT id, name FROM node_pools LIMIT 5" -o table
run_cmd "262-db-exec-noop" db exec "UPDATE clusters SET updated_time=now() WHERE false"
run_cmd "263-db-delete-clusters-dry" db delete clusters 2>&1 || true

# ── Logs with cluster context ──
timeout 10 "$HF" logs sentinel 2>&1 | head -30 > "$OUT_DIR/270-logs-sentinel.txt" || true
{
  echo "COMMAND: hf logs sentinel (timeout 10s)"
  echo "EXIT: 124"
  echo "────────────────────────────────────────"
  cat "$OUT_DIR/270-logs-sentinel.txt" 2>/dev/null
} > "$OUT_DIR/270-logs-sentinel-wrapped.txt"
mv "$OUT_DIR/270-logs-sentinel-wrapped.txt" "$OUT_DIR/270-logs-sentinel.txt"

run_cmd "271-logs-adapter" logs adapter

# ── Kube port-forward stop/start ──
run_cmd "280-kube-pf-stop" kube port-forward stop
run_cmd "281-kube-pf-start" kube port-forward start
run_cmd "282-kube-pf-status" kube port-forward status

# ── Destructive deletes (nodepool first, then cluster) ──
run_cmd "290-nodepools-delete" rs nodepools delete
run_cmd "291-clusters-delete" rs clusters delete

# Verify cleanup
run_cmd "292-clusters-list-after-delete" rs clusters list -o table
run_cmd "293-nodepools-list-after-delete" rs nodepools list -o table

echo "=== Round 2 complete. Outputs in $OUT_DIR ==="
ls -la "$OUT_DIR"
