# Design: hf kube namespace-clean

## Label matching

A namespace matches if any label key OR label value contains the substring "hyperfleet"
(case-sensitive). This is intentionally broad — operators can inspect the list before
confirming deletion.

## Packages affected

| Package | Change |
|---------|--------|
| `internal/kube` | Add `ListHyperfleetNamespaces(ctx, cs)` and `DeleteNamespace(ctx, cs, name)` |
| `cmd/kube.go` | Add `kubeNamespaceCleanCmd`, register in `init()` |
| `cmd/kube_test.go` | New file with unit tests |

## Command flow

```
hf kube namespace-clean
  1. Build clientset from kubeconfig / context
  2. Call ListHyperfleetNamespaces → []string
  3. If empty → print [INFO] and exit 0
  4. Print list (one per line, indented) + total count
  5. Prompt "Delete these N namespace(s)? [y/N]: "
  6. Read answer from stdin; anything other than "y" → [INFO] Aborted, exit 0
  7. Spawn one goroutine per namespace; each prints [DELETING] on start,
     [DELETED] on success or [ERROR] on failure (mutex-protected writes)
  8. Wait for all goroutines
  9. If any failed → return error with count; else print [INFO] Done summary
```

## Progress output format

```
[DELETING] hyperfleet-e2e-test-abc
[DELETING] hyperfleet-staging-xyz
[DELETED]  hyperfleet-e2e-test-abc
[ERROR]    hyperfleet-staging-xyz: namespaces "hyperfleet-staging-xyz" not found
[INFO] Done. Deleted 1 namespace(s), 1 error(s).
```

## Confirmation stdin handling

Uses `bufio.NewReader(cmd.InOrStdin()).ReadString('\n')` so that tests can supply a pipe
and interactive use reads the terminal line-buffered.

## Error handling

- Clientset build error → return immediately with [ERROR]
- Namespace list error → return immediately with [ERROR]
- Per-namespace delete error → printed per goroutine; final error returned after all finish
