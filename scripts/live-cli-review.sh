#!/usr/bin/env bash
# Live CLI command review against GKE — captures all command outputs.
set -uo pipefail

HF="${HF_BIN:-./bin/hf}"
OUT_DIR="${OUT_DIR:-verification_proof/cli-review-gke}"
mkdir -p "$OUT_DIR"

run_cmd() {
  local name="$1"
  shift
  local outfile="$OUT_DIR/${name}.txt"
  {
    echo "COMMAND: hf $*"
    echo "EXIT: pending"
    echo "────────────────────────────────────────"
    set +e
    "$HF" "$@" 2>&1
    local ec=$?
    set -e
    echo
    echo "────────────────────────────────────────"
    echo "EXIT: $ec"
  } > "$outfile" 2>&1 || true
  # patch exit code line
  sed -i '' "s/EXIT: pending/EXIT: $(tail -1 "$outfile" 2>/dev/null | grep -o '[0-9]*' || echo '?')/" "$outfile" 2>/dev/null || true
  echo "$name: done (exit $?)"
}

# Re-capture with proper exit code tracking
run_cmd_v2() {
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
}

echo "=== Live CLI review starting ==="
echo "HF=$HF OUT_DIR=$OUT_DIR"
echo "Active env: $($HF env list 2>/dev/null | grep '✓' || echo unknown)"

# ── Meta ──
run_cmd_v2 "00-version" version
run_cmd_v2 "01-help" --help

# ── Completion (sample) ──
run_cmd_v2 "02-completion-bash" completion bash
run_cmd_v2 "03-completion-zsh" completion zsh
run_cmd_v2 "04-completion-fish" completion fish
run_cmd_v2 "05-completion-powershell" completion powershell

# ── Environment ──
run_cmd_v2 "10-env-list" env list
run_cmd_v2 "11-env-show" env show
run_cmd_v2 "12-env-show-gke" env show gke

# ── Database ──
run_cmd_v2 "20-db-config" db config
run_cmd_v2 "21-db-query-count-channels" db query "SELECT count(*) AS channel_count FROM channels"
run_cmd_v2 "22-db-query-channels-table" -o table db query "SELECT id, name, kind FROM channels LIMIT 5"
run_cmd_v2 "23-db-query-clusters" db query "SELECT count(*) AS cluster_count FROM clusters"
run_cmd_v2 "24-db-exec-curl-dryrun" --curl db exec "UPDATE channels SET updated_time=now() WHERE false"

# ── Kube ──
run_cmd_v2 "30-kube-pf-status" kube port-forward status
run_cmd_v2 "31-kube-curl-api-health" kube curl http://hyperfleet-api.hyperfleet-e2e-gke1.svc:8000/health
run_cmd_v2 "32-kube-curl-api-clusters" kube curl http://hyperfleet-api.hyperfleet-e2e-gke1.svc:8000/api/hyperfleet/v1/clusters
run_cmd_v2 "33-kube-debug-help" kube debug --help
run_cmd_v2 "34-kube-namespace-clean-help" kube namespace-clean --help

# ── Maestro ──
run_cmd_v2 "40-maestro-list" maestro list
run_cmd_v2 "41-maestro-bundles" maestro bundles
run_cmd_v2 "42-maestro-consumers" maestro consumers

# ── Pub/Sub ──
run_cmd_v2 "50-pubsub-list" pubsub list
run_cmd_v2 "51-pubsub-publish-cluster-curl" --curl pubsub publish cluster
run_cmd_v2 "52-pubsub-publish-nodepool-curl" --curl pubsub publish nodepool

# ── RabbitMQ ──
run_cmd_v2 "60-rabbitmq-publish-cluster-curl" --curl rabbitmq publish cluster
run_cmd_v2 "61-rabbitmq-publish-nodepool-curl" --curl rabbitmq publish nodepool

# ── Repos ──
run_cmd_v2 "70-repos-table" -o table repos

# ── Resource overview ──
run_cmd_v2 "80-rs-overview-table" rs -o table
run_cmd_v2 "81-rs-overview-json" rs -o json
run_cmd_v2 "82-rs-types" rs types
run_cmd_v2 "83-table-deprecated" table -o table
run_cmd_v2 "84-resources-deprecated" resources -o table

# ── Channels ──
run_cmd_v2 "90-channels-list" rs channels list -o table
run_cmd_v2 "91-channels-table" rs channels table
run_cmd_v2 "92-channels-get" rs channels get
run_cmd_v2 "93-channels-id" rs channels id
run_cmd_v2 "94-channels-search" rs channels search ch-3fa50ce7
run_cmd_v2 "95-channels-list-json" rs channels list -o json
run_cmd_v2 "96-channels-create-curl" --curl rs channels create
run_cmd_v2 "97-channels-patch-curl" --curl rs channels patch spec.enabled_regex=.*
run_cmd_v2 "98-channels-delete-curl" --curl rs channels delete
run_cmd_v2 "99-channels-adapter-report-curl" --curl rs channels adapter-report

# ── Versions ──
run_cmd_v2 "100-versions-list" rs versions list -o table
run_cmd_v2 "101-versions-get" rs versions get
run_cmd_v2 "102-versions-id" rs versions id
run_cmd_v2 "103-versions-create-curl" --curl rs versions create
run_cmd_v2 "104-versions-search-help" rs versions search --help

# ── Logs (bounded) ──
timeout 8 "$HF" logs hyperfleet-api 2>&1 | head -40 > "$OUT_DIR/110-logs-hyperfleet-api.txt" || true
{
  echo "COMMAND: hf logs hyperfleet-api (timeout 8s, first 40 lines)"
  echo "EXIT: $(test -s "$OUT_DIR/110-logs-hyperfleet-api.txt" && echo 0 || echo 124)"
  echo "────────────────────────────────────────"
  cat "$OUT_DIR/110-logs-hyperfleet-api.txt" 2>/dev/null || echo "(no output)"
} > "$OUT_DIR/110-logs-hyperfleet-api-wrapped.txt"
mv "$OUT_DIR/110-logs-hyperfleet-api-wrapped.txt" "$OUT_DIR/110-logs-hyperfleet-api.txt"

timeout 15 "$HF" logs insights 2>&1 > "$OUT_DIR/111-logs-insights.txt" || true
{
  echo "COMMAND: hf logs insights (timeout 15s)"
  echo "EXIT: timed"
  echo "────────────────────────────────────────"
  cat "$OUT_DIR/111-logs-insights.txt" 2>/dev/null | head -80
} > "$OUT_DIR/111-logs-insights-wrapped.txt"
mv "$OUT_DIR/111-logs-insights-wrapped.txt" "$OUT_DIR/111-logs-insights.txt"

run_cmd_v2 "112-logs-adapter" logs adapter

# ── UI (start briefly) ──
timeout 3 "$HF" ui -p 18088 2>&1 > "$OUT_DIR/120-ui-start.txt" || true
{
  echo "COMMAND: hf ui -p 18088 (timeout 3s)"
  echo "EXIT: timed"
  echo "────────────────────────────────────────"
  cat "$OUT_DIR/120-ui-start.txt"
} > "$OUT_DIR/120-ui-start-wrapped.txt"
mv "$OUT_DIR/120-ui-start-wrapped.txt" "$OUT_DIR/120-ui-start.txt"

# ── TUI (non-interactive fails expected) ──
run_cmd_v2 "121-tui" tui

# ── Maestro get (first bundle if any) ──
BUNDLE=$("$HF" maestro list -o json 2>/dev/null | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    items = d if isinstance(d, list) else d.get('items', d.get('bundles', []))
    if items:
        print(items[0].get('name', items[0].get('metadata', {}).get('name', '')))
except: pass
" 2>/dev/null || true)
if [[ -n "${BUNDLE:-}" ]]; then
  run_cmd_v2 "43-maestro-get" maestro get "$BUNDLE"
fi

echo "=== Done. Outputs in $OUT_DIR ==="
ls -la "$OUT_DIR"
