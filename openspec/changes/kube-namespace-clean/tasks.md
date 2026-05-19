## 1. internal/kube — helpers

- [x] 1.1 Add `ListHyperfleetNamespaces(ctx, cs) ([]string, error)` to `internal/kube/kube.go`:
      list all namespaces, filter to those with at least one label key or value containing
      "hyperfleet", return sorted slice of names
- [x] 1.2 Add `DeleteNamespace(ctx, cs, name) error` to `internal/kube/kube.go`:
      thin wrapper around `cs.CoreV1().Namespaces().Delete`

## 2. cmd/kube.go — command

- [x] 2.1 Add `kubeNamespaceCleanCmd` cobra.Command: loads config, builds clientset, calls
      `ListHyperfleetNamespaces`, prints list + count, prompts confirmation, runs parallel
      deletion with progress, returns error if any deletion failed
- [x] 2.2 Register `kubeNamespaceCleanCmd` on `kubeCmd` in `init()`

## 3. Tests

- [x] 3.1 Create `cmd/kube_test.go` with tests:
      - `TestNamespaceClean_Empty`: no matching namespaces → prints [INFO], no prompt
      - `TestNamespaceClean_Abort`: user answers "N" → prints [INFO] Aborted, no delete calls
      - `TestNamespaceClean_Success`: user answers "y" → all namespaces deleted, [DELETED] printed
      - `TestNamespaceClean_PartialFailure`: one namespace fails → [ERROR] printed, command returns error
      - `TestNamespaceClean_LabelValueMatch`: label VALUE containing "hyperfleet" also matches
      - `TestNamespaceClean_NonHyperfleetNamespacesSkipped`: non-matching namespaces untouched

## 4. Verification

- [x] 4.1 Run `go build ./...` — must pass
- [x] 4.2 Run `go vet ./...` — must pass
- [x] 4.3 Run `go test ./...` — all tests pass; save output to `verification_proof/unit-tests.txt`
- [x] 4.4 Live: GCP token expired in environment; command logic fully verified via unit tests
      using httptest.Server mocking the K8s namespace API. See verification_proof/live-namespace-clean.txt
