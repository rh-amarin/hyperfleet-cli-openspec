## 1. cmd/config.go — hf config bare (help then show)

- [x] 1.1 Update `configCmd.RunE` to call `cmd.Help()` then `configShowCmd.RunE(cmd, args)`

## 2. cmd/config.go — hf config set interactive mode

- [x] 2.1 Add `"bufio"` and `"github.com/rh-amarin/hyperfleet-cli/internal/selector"` to imports
- [x] 2.2 Add package-level `var configSetSel selector.Selector = selector.FuzzySelector{}`
- [x] 2.3 Update `configSetCmd.Use` to `"set [section.key value]"` and `Short` to mention interactive mode
- [x] 2.4 Replace `Args: helpOnNoArgs(2)` with a custom validator that accepts 0 or 2 args (shows help otherwise)
- [x] 2.5 Add interactive path in `configSetCmd.RunE` (0 args): build items from `knownKeysForSection`, call `configSetSel.Select`, prompt for value via `bufio.Scanner(cmd.InOrStdin())`, call `s.Set`
- [x] 2.6 Add `configShowCmd.RunE(cmd, nil)` after successful `s.Set` in both the interactive and 2-arg paths

## 3. cmd/kube.go — hf kube port-forward bare (help then status + start prompt)

- [x] 3.1 Add `RunE` to `portForwardCmd`: call `cmd.Help()`, then `loadConfig()`, resolve context, call `kube.ListPortForwards()`
- [x] 3.2 In the `RunE`: if no port-forwards tracked, print `No port-forwards tracked.` and return
- [x] 3.3 In the `RunE`: check connectivity for each tracked port-forward, print status via `printPortForwardStatus`
- [x] 3.4 In the `RunE`: if any down, print prompt, read answer via `bufio.Scanner(cmd.InOrStdin())`; if `y`, call `pfStartCmd.RunE(cmd, nil)`

## 4. Tests — cmd/config_test.go

- [x] 4.1 Add `TestConfigNoArgs_ShowsHelpBeforeConfig`: verify "Usage:" appears in output before "hyperfleet:"
- [x] 4.2 Add `TestConfigSet_ShowsConfigAfterSet`: run 2-arg set, verify output contains config sections
- [x] 4.3 Add `TestConfigSet_Interactive`: inject `configSetSel = mockSel{idx: 0}`, set `rootCmd.SetIn` with value, run `hf config set`, verify value written and config shown

## 5. Tests — cmd/kube_test.go

- [x] 5.1 Add `TestPortForwardNoArgs_NoForwards`: set `XDG_CACHE_HOME` to temp dir, verify output contains "Usage:" and "No port-forwards tracked."

## 6. Verification

- [x] 6.1 Run `go build ./...` and confirm zero errors; save output to `verification_proof/build.txt`
- [x] 6.2 Run `go vet ./...` and confirm zero errors; save output to `verification_proof/vet.txt`
- [x] 6.3 Run `go test ./cmd/...` and confirm all tests pass; save output to `verification_proof/test.txt`
