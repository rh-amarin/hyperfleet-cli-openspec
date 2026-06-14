#!/usr/bin/env python3
"""Regenerate cli-commands-review-gke.md from capture files."""
import os
from datetime import datetime, timezone

ROUND1 = "verification_proof/cli-review-gke"
ROUND2 = "verification_proof/cli-review-gke-round2"
DOC = "cli-commands-review-gke.md"
SEP = "────────────────────────────────────────"


def parse_capture(path):
    with open(path) as f:
        raw = f.read()
    cmd, ec, body_lines = "", "?", []
    state = "header"
    for line in raw.splitlines():
        if line.startswith("COMMAND: "):
            cmd = line[9:]
        elif line.startswith("EXIT: "):
            ec = line[6:]
        elif line == SEP:
            if state == "header":
                state = "body"
            else:
                break
        elif state == "body":
            body_lines.append(line)
    return cmd, ec, "\n".join(body_lines).strip()


def status_icon(ec):
    return "✅" if str(ec) in ("0", "timed", "124", "124 (timeout)") else "❌"


def truncate(body, limit=4000):
    if len(body) <= limit:
        return body
    return body[:limit] + "\n... (truncated) ..."


lookup = {}
for d in [ROUND1, ROUND2]:
    if os.path.isdir(d):
        for f in os.listdir(d):
            lookup[f] = os.path.join(d, f)

sections = [
    ("Meta & Help", ["00-version.txt", "01-help.txt"]),
    ("Shell Completion", ["02-completion-bash.txt", "03-completion-zsh.txt", "04-completion-fish.txt", "05-completion-powershell.txt"]),
    ("Environment (hf env)", ["10-env-list.txt", "11-env-show.txt", "12-env-show-gke.txt", "401-tty-env-picker.txt"]),
    ("Database (hf db)", ["20-db-config.txt", "21-db-query-count-channels.txt", "23-db-query-clusters.txt", "25-db-query-resources.txt", "26-db-query-resources-sample.txt", "262-db-exec-noop.txt", "27-db-exec-dryrun.txt", "263-db-delete-clusters-abort.txt", "264-db-delete-adapter-statuses.txt"]),
    ("Kubernetes (hf kube)", ["30-kube-pf-status.txt", "31-kube-curl-api-health.txt", "32-kube-curl-api-clusters.txt", "280-kube-pf-stop.txt", "281-kube-pf-start.txt", "282-kube-pf-status.txt"]),
    ("Maestro (hf maestro)", ["40-maestro-list.txt", "41-maestro-bundles.txt", "42-maestro-consumers.txt"]),
    ("Pub/Sub (hf pubsub)", ["50-pubsub-list.txt", "250-pubsub-publish-cluster.txt", "251-pubsub-publish-nodepool.txt"]),
    ("RabbitMQ (hf rabbitmq)", ["252-rabbitmq-publish-cluster.txt", "253-rabbitmq-publish-nodepool.txt"]),
    ("Repos (hf repos)", ["70-repos-table.txt"]),
    ("Resource Overview (hf rs)", ["240-rs-overview-table.txt", "241-rs-overview-json.txt", "242-rs-types.txt", "243-table-deprecated.txt"]),
    ("Clusters (hf rs clusters)", ["200-clusters-create.txt", "201-clusters-list.txt", "203-clusters-get.txt", "205-clusters-search.txt", "206-clusters-conditions.txt", "207-clusters-statuses.txt", "208-clusters-adapter-report.txt", "209-clusters-patch.txt", "304-clusters-patch-spec.txt", "291-clusters-delete.txt", "302-clusters-force-delete.txt", "402-tty-clusters-id-i.txt"]),
    ("Nodepools (hf rs nodepools)", ["220b-nodepools-create-short.txt", "221b-nodepools-list.txt", "223-nodepools-get.txt", "228-nodepools-adapter-report.txt", "229-nodepools-patch.txt", "305-nodepools-patch-spec.txt", "290-nodepools-delete.txt", "306-nodepools-force-delete.txt"]),
    ("Channels (hf rs channels)", ["90-channels-list.txt", "92-channels-get.txt", "94-channels-search.txt", "96-channels-create-curl.txt", "97-channels-patch-curl.txt", "98-channels-delete-curl.txt", "99-channels-adapter-report-curl.txt"]),
    ("Versions (hf rs versions)", ["100-versions-list.txt", "101-versions-get.txt", "103-versions-create-curl.txt"]),
    ("Logs (hf logs)", ["270-logs-sentinel.txt", "271-logs-adapter.txt", "111-logs-insights.txt"]),
    ("UI & TUI", ["120-ui-start.txt", "400-tty-tui.txt", "121-tui.txt"]),
]

lines = [
    "# HyperFleet CLI — Live Command Review (GKE)",
    "",
    f"Generated: {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M UTC')}",
    "",
    "> **Round 2 update:** Environment includes `clusters` and `nodepools` in `resource-types`.",
    "> Destructive commands executed live. Interactive commands tested via **ttyd** + browser.",
    "",
    "## Test Environment",
    "",
    "| Setting | Value |",
    "|---------|-------|",
    "| Active environment | `gke` |",
    "| Kubernetes context | `gke_hcm-hyperfleet_europe-southwest1-a_hyperfleet-dev-amarin-eu1` |",
    "| Namespace | `hyperfleet-e2e-gke1` |",
    "| Resource types | `clusters`, `nodepools`, `channels`, `versions` |",
    "| Port-forwards | API `:8000`, PostgreSQL `:5432`, Maestro HTTP `:8100`, gRPC `:8090` |",
    "| Interactive testing | ttyd ports 7683–7685, browser-driven |",
    "",
    "## Results Summary",
    "",
]

ok = fail = 0
for _, names in sections:
    for n in names:
        if n not in lookup:
            continue
        _, ec, _ = parse_capture(lookup[n])
        if status_icon(ec) == "✅":
            ok += 1
        else:
            fail += 1
lines += [f"| ✅ Success / expected | {ok} |", f"| ❌ Error / missing prereq | {fail} |", ""]
lines += [
    "### Key Findings",
    "",
    "- **Clusters/nodepools** — full CRUD against live GKE API",
    "- **Patch** — `hf rs clusters patch spec` increments counter (not key=value syntax)",
    "- **Force-delete** — 409 when not in Finalizing state",
    "- **db delete** — requires typing `yes`; deleted 13 adapter_statuses rows live",
    "- **RabbitMQ** — connection refused (no :15672 port-forward)",
    "- **TUI/env/`-i`** — verified via ttyd + browser",
    "",
    "---",
    "",
]

for title, names in sections:
    lines.append(f"## {title}")
    lines.append("")
    for n in names:
        path = lookup.get(n)
        if not path:
            lines.append(f"### ⚠️ `{n}` — capture missing")
            lines.append("")
            continue
        cmd, ec, body = parse_capture(path)
        if "completion" in n and len(body) > 800:
            bl = body.splitlines()
            body = "\n".join(bl[:8]) + f"\n... ({len(bl)-16} more lines) ...\n" + "\n".join(bl[-8:])
        elif n == "50-pubsub-list.txt" and len(body) > 1500:
            bl = body.splitlines()
            body = "\n".join(bl[:25]) + f"\n... ({len(bl)-25} more lines) ..."
        else:
            body = truncate(body)
        lines.append(f"### {status_icon(ec)} `{cmd or n}`")
        lines.append("")
        lines.append(f"**Exit code:** `{ec}`")
        lines.append("")
        if body:
            lines.append("```")
            lines.append(body)
            lines.append("```")
        lines.append("")

lines += [
    "## Destructive Commands Tested",
    "",
    "| Command | Result |",
    "|---------|--------|",
    "| `hf rs clusters create` | Created test clusters on GKE |",
    "| `hf rs nodepools create` | Created `np-cli-test` (name ≤15 chars) |",
    "| `hf rs clusters/nodepools patch spec` | Counter 1→2 |",
    "| `hf rs nodepools delete` | Soft-delete nodepool |",
    "| `hf rs clusters delete` | Soft-delete cluster |",
    "| `hf rs * force-delete` | 409 conflict (not Finalizing) |",
    "| `echo yes \\| hf db delete adapter_statuses` | Deleted 13 rows |",
    "| `echo no \\| hf db delete clusters` | Aborted |",
    "| `hf pubsub publish cluster/nodepool` | Published to GCP |",
    "| `hf kube port-forward stop/start` | Cycled 4 forwards |",
    "",
    "## Interactive (ttyd + browser)",
    "",
    "| Command | Port | Result |",
    "|---------|------|--------|",
    "| `hf tui` | 7683 | Cluster table + keybindings; `q` quits |",
    "| `hf env` | 7684 | 9-env fuzzy picker + YAML preview |",
    "| `hf rs clusters id -i` | 7685 | Selected cluster from fuzzy list |",
    "",
    "## Raw Captures",
    "",
    f"- `{ROUND1}/` — round 1 (meta, channels, kube, maestro, …)",
    f"- `{ROUND2}/` — round 2 (clusters, nodepools, destructive, ttyd)",
    "",
]

with open(DOC, "w") as f:
    f.write("\n".join(lines))
print(f"Wrote {DOC} ({len(lines)} lines)")
