# OpenSpec Workflow Evaluation Log

This file tracks what works and what doesn't in the OpenSpec workflow for this repo, across iteration cycles.

---

## Cycle 1 — 2026-05-09

### Setup performed

1. `npm install -g @fission-ai/openspec@latest` (v1.3.1)
2. `openspec init --tools claude --force` in repo root
3. Moved `specs/` → `openspec/specs/` (18 domain specs)
4. Moved `config.yaml` → `openspec/config.yaml` (merged context + rules)
5. Initial commit pushed to `main`

### Test task

> Spawn a fresh Claude Code agent and ask it to start implementing the `technical-architecture` spec using the OpenSpec workflow.

### What WORKS ✅

| Item | Detail |
|---|---|
| CLI auto-install check | Agent correctly ran `openspec --version` first; skipped install since v1.3.1 was present |
| `openspec new change` | Created `openspec/changes/technical-architecture-init/` with `.openspec.yaml` scaffold |
| `openspec status --json` | Agent used this after each artifact to determine next build step |
| `openspec instructions <artifact> --json` | Used correctly before each artifact write; extracted `outputPath`, `template`, `context`, `rules` |
| Artifact sequencing | proposal → design → specs → tasks, honoring dependency order |
| Context/rules separation | Agent applied `context` and `rules` as constraints without copying them into artifact files |
| Tasks checked off immediately | `[x]` marks applied in real-time, not batched |
| No code before planning | All 4 planning artifacts created before any Go files were written |
| Go scaffold is valid | `go build ./...` passes with no errors |
| `go.mod` correct | Module name `github.com/rh-amarin/hyperfleet-cli`, Go 1.22, cobra dependency |

### What DOESN'T WORK / Gaps ❌

| Item | Detail |
|---|---|
| Verification proof not generated | Agent did not reach the Verify section (tasks 7.x) — only 5 of 20 tasks completed. Expected for "start implementation" scope but worth tracking. |
| go.sum not in git | `go.sum` generated locally but not committed — needs explicit staging |
| Delta spec quality unclear | The `specs/technical-architecture/spec.md` in the change folder adds new requirements; not yet validated against the main spec format |

### Observations

- CLAUDE.md already had complete OpenSpec workflow instructions — the agent followed them without needing extra prompting
- The `openspec/config.yaml` context (project stack, DoD) was correctly picked up by `openspec instructions --json`
- The skill files in `.claude/skills/` are the key enablers: they encode the exact step sequence the agent must follow
- Skills work even in a spawned sub-agent that doesn't have the parent session's context

### Verdict: PASS ✅

The workflow is functional. The agent followed the spec-driven methodology end to end: CLI install check → scaffold change → artifacts in order → code only after planning → immediate task check-off.

---

## Improvement Notes for Cycle 2 (if needed)

- Add a `.gitignore` to avoid committing binaries from `go build`
- Consider adding `openspec validate` step to CLAUDE.md workflow instructions
- Add explicit reminder in tasks template that `go.sum` must be committed
- Consider adding `openspec archive` step test to verify full lifecycle

---

## Cycle 2 — 2026-05-09 — ACP Session Orchestration

### Context

Attempting to coordinate implementation via ACP (Ambient Code Platform) sessions — one session per phase, with a separate review session per PR.

### ACP Session Observations ⚠️

| Observation | Detail |
|---|---|
| `totalMessages` stays 0 | The API does not surface assistant messages during autonomous initial-prompt processing. Messages only count after explicit `send_message` exchanges. Coordinator cannot see agent progress mid-run. |
| `lastActivityTime` updates slowly | Timestamp only increments every 30–90 seconds during long tool-call sequences (file writes, bash). Long gaps do NOT mean the session is stuck. |
| `recentMessages` always empty | During the initial prompt run, no messages appear even after `send_message` nudges — those are queued behind the running task. |
| `restart` does not re-launch | `acp_restart_session` stopped session v1 but it transitioned to `Stopped` without re-entering `Running`. A new session must be created instead. |
| Branch push as progress signal | The most reliable external signal is watching for the git branch to appear on GitHub and new commits to land. `lastActivityTime` alone is insufficient. |
| Sessions run but are opaque | The agent IS working (activity time ticks, branch appears), but there is no way to see intermediate output or tool calls during the run. |

### Workarounds adopted

- Added `git push -u origin <branch>` immediately after `git checkout -b` so branch visibility is the progress indicator
- Added frequent WIP commits (`git add -A && git commit -m "wip: <step>"`) after each major step
- Watching `gh api repos/.../commits?sha=<branch>` for new commits as a heartbeat
- Polling both session status AND GitHub branch simultaneously to reduce false negatives

### What WORKS in ACP sessions ✅

- Sessions start and clone the repo correctly
- `openspec --version` check works (CLI persists between steps)
- `git checkout -b` and `git push -u origin` work
- Agent reads CLAUDE.md and follows OpenSpec workflow instructions

### Open questions

- Is there a maximum execution time per session?
- Does `maxTokens: 0` (seen in session spec) impose a hidden limit?
- Can hooks be used to surface session output to the coordinator in real time?
