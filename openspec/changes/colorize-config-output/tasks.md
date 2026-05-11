## 1. Output Helpers

- [x] 1.1 Create `internal/output/config_display.go` with `ColorizeYAMLSections(s string, noColor bool) string` that prefixes bold-cyan ANSI codes on top-level YAML section-header lines (lines matching `^(\w[\w-]*):\s*$`)
- [x] 1.2 Add `SectionSeparator(noColor bool) string` in the same file that returns a 40-character `────────────────────────────────────────` line (no ANSI codes needed; Unicode only)

## 2. Config Command

- [x] 2.1 In `cmd/config.go` `configShowCmd.RunE`, split the YAML marshal into two passes: one for config sections (`hyperfleet`, `kubernetes`, `maestro`, `port-forward`, `database`, `rabbitmq`, `registry`) and one for the `state:` block
- [x] 2.2 Determine `noColor` by checking the `--no-color` flag AND `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))` when `cmd.OutOrStdout()` is `os.Stdout`
- [x] 2.3 Write the colorized config YAML, then `output.SectionSeparator(noColor)`, then the colorized state YAML — skip the separator when the state block is empty

## 3. Unit Tests

- [x] 3.1 In `internal/output/config_display_test.go`, test `ColorizeYAMLSections` with `noColor=true`: output MUST equal input (no ANSI injected)
- [x] 3.2 Test `ColorizeYAMLSections` with `noColor=false`: lines matching the section-header pattern MUST contain `\033[` escape codes; value lines MUST NOT
- [x] 3.3 Test `SectionSeparator`: returned string MUST be non-empty and MUST NOT contain `\033[` in either mode

## 4. Verify

- [x] 4.1 `go build ./...` succeeds
- [x] 4.2 `go vet ./...` passes
- [x] 4.3 `go test ./... 2>&1 | tee verification_proof/tests.txt`
- [x] 4.4 Run `hf config` against the live cluster and save output to `verification_proof/live.txt`
- [ ] 4.5 Commit `verification_proof/` to git
