## 1. Implement table output in cmd/maestro.go

- [x] 1.1 In `maestroListCmd.RunE`: after fetching `list`, add a branch — if `outputFmt == "table"`, call `printMaestroListTable(cmd.OutOrStdout(), list)` and return; otherwise fall through to `p.Print(list)` as before
- [x] 1.2 Add `printMaestroListTable(w io.Writer, list *maestro.ResourceBundleList)`:
  - If `list.Items` is empty, print `No resource bundles.\n` and return
  - For each bundle: print `<id>  <name>  v<version>\n`
  - For each manifest in the bundle: print `  <kind>/<name>  <namespace>\n`

## 2. Unit tests in cmd/maestro_test.go

- [x] 2.1 `TestMaestroListTable_WithManifests` — start an `httptest.NewServer` returning one bundle with two manifests; run `maestroListCmd` with `--output table`; assert output contains the bundle line and both indented manifest lines
- [x] 2.2 `TestMaestroListTable_Empty` — start an `httptest.NewServer` returning an empty items list; run `maestroListCmd` with `--output table`; assert output is `No resource bundles.\n`

## 3. Spec delta

- [x] 3.1 Write `openspec/changes/maestro-list-table-output/specs/maestro/spec.md` — MODIFIED delta for the "List Maestro Resources" requirement adding the table output scenario

## 4. Verify

- [x] 4.1 `go build ./...` succeeds
- [x] 4.2 `go vet ./...` reports no issues
- [x] 4.3 `go test ./...` passes — capture output to `verification_proof/tests.txt`
- [x] 4.4 Live verification: run `hf maestro list --output table` against the real Maestro instance; save output to `verification_proof/live.txt`
