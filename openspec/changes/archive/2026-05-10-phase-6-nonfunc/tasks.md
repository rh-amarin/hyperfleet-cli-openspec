## 1. Shell Completions

- [x] 1.1 Update `cmd/completion.go`: use `cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)`, `cmd.OutOrStdout()`, and `GenBashCompletionV2`
- [x] 1.2 Register `--output` flag completion in `cmd/root.go` via `rootCmd.RegisterFlagCompletionFunc`
- [x] 1.3 Write `cmd/completion_test.go` with tests for all 4 shells

## 2. GoReleaser

- [x] 2.1 Create `.goreleaser.yaml` at repo root

## 3. GitHub Actions

- [x] 3.1 Create `.github/workflows/ci.yml`
- [x] 3.2 Create `.github/workflows/release.yml`

## 4. Makefile & .gitignore

- [x] 4.1 Add `completions` and `lint` targets to `Makefile`
- [x] 4.2 Create `.gitignore` with `bin/`, `dist/`, `*.exe`, `*.test`

## 5. Verification

- [x] 5.1 Run `go build ./...` and save output to `verification_proof/build.txt`
- [x] 5.2 Run `go vet ./...` and save output to `verification_proof/vet.txt`
- [x] 5.3 Run `go test ./...` and save output to `verification_proof/test.txt`
- [x] 5.4 Write `verification_proof/live_verification_note.txt`
