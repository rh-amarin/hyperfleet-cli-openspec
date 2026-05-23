# Spec delta: kubernetes — namespace-clean

## New requirement

**REQ-KUBE-NS-CLEAN-001**: `hf kube namespace-clean` lists all Kubernetes namespaces whose
labels contain the substring "hyperfleet" (matched against both key and value), presents the
list and count to the user, requires explicit `y` confirmation, then deletes all matching
namespaces in parallel, printing per-namespace progress and a summary.

### Matching rule

A namespace matches if `strings.Contains(labelKey, "hyperfleet") || strings.Contains(labelValue, "hyperfleet")`
is true for any entry in its `metadata.labels` map.

### Behaviour table

| Condition | Output |
|-----------|--------|
| No matching namespaces | `[INFO] No namespaces with 'hyperfleet' labels found.` |
| User answers anything except `y` | `[INFO] Aborted.` |
| All deletions succeed | `[INFO] Done. Deleted N namespace(s).` |
| Some deletions fail | per-namespace `[ERROR]` + command exits non-zero |

### Progress lines

```
[DELETING] <name>          ← printed when goroutine starts
[DELETED]  <name>          ← printed on success
[ERROR]    <name>: <msg>   ← printed on failure (to stderr)
```
