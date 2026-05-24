## 1. Dependencies and scaffolding

- [x] 1.1 Add bubbletea, bubbles, and lipgloss to go.mod
- [x] 1.2 Create `internal/tui/` package skeleton with DataFetcher interface

## 2. Data layer

- [x] 2.1 Implement table row builder reusing resources.go logic (headers, rows, adapter cols)
- [x] 2.2 Implement detail content renderers (json, yaml, overview, statuses)
- [x] 2.3 Implement smart search filter (lowercase adapter / uppercase condition type)

## 3. TUI model

- [x] 3.1 Implement bubbletea Model with main table, selection, and refresh loop
- [x] 3.2 Implement detail panel with scroll, format cycling (V), statuses toggle (S)
- [x] 3.3 Implement search overlay (/) and Esc close behavior
- [x] 3.4 Export ColorizeJSONBytes helper in internal/output if needed

## 4. Command wiring

- [x] 4.1 Add `cmd/tui.go` with `hf tui` command and `-s` flag
- [x] 4.2 Wire DataFetcher to real API via fetchResourceEntries

## 5. Tests and verification

- [x] 5.1 Unit tests for filter logic, detail rendering, and model state transitions
- [x] 5.2 Unit tests for data fetcher with httptest.NewServer
- [x] 5.3 cmd/tui_test.go for command registration and flag defaults
- [x] 5.4 Run go test, go build, go vet; save output to verification_proof/
- [x] 5.5 Live verification via ttyd; save output to verification_proof/
