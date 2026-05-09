## ADDED Requirements

### Requirement: Go Module Scaffold
The CLI SHALL be initialized as a Go module at `github.com/rh-amarin/hyperfleet-cli` with Go 1.22 as the minimum language version, producing a single self-contained binary named `hf`.

#### Scenario: Module initialization
- **WHEN** the repository is cloned fresh
- **THEN** `go build ./...` SHALL succeed and produce a runnable `hf` binary under `bin/`

#### Scenario: Entry point delegation
- **WHEN** `main.go` is compiled
- **THEN** `main()` SHALL call `cmd.Execute()` and nothing else; all logic SHALL reside in `cmd/` or `internal/`

### Requirement: Root Cobra Command with Global Flags
The CLI SHALL define a root Cobra command in `cmd/root.go` that registers all global persistent flags before any subcommand runs.

#### Scenario: Global flags available on every subcommand
- **WHEN** any `hf` subcommand is invoked
- **THEN** the following persistent flags SHALL be available: `--config`, `--output`, `--no-color`, `--verbose` (`-v`), `--api-url`, `--api-token`

#### Scenario: PersistentPreRunE wires flag values
- **WHEN** a subcommand executes
- **THEN** `PersistentPreRunE` on the root command SHALL populate package-level variables from parsed flag values before `RunE` fires

### Requirement: Command Stub Registration
All top-level command groups SHALL be registered with the root command via `init()` in their respective `cmd/<domain>.go` files, even if the subcommand bodies are empty stubs.

#### Scenario: Stub commands appear in help output
- **WHEN** `hf --help` is run
- **THEN** all domain commands (cluster, nodepool, config, db, maestro, pubsub, rabbitmq, kube, logs, repos, resources, version, completion) SHALL appear in the help output

#### Scenario: Stub commands call Help on invocation
- **WHEN** a stub domain command is run without a recognized subcommand
- **THEN** it SHALL print its usage/help and exit 0

### Requirement: Build-time Version Package
The CLI SHALL provide an `internal/version` package with a `Version` variable that defaults to `"dev"` and can be overridden at build time via linker flags.

#### Scenario: Default version
- **WHEN** the binary is built without ldflags
- **THEN** `hf version` SHALL print `dev`

#### Scenario: Release version via ldflags
- **WHEN** the binary is built with `-ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=1.2.3"`
- **THEN** `hf version` SHALL print `1.2.3`

### Requirement: Makefile Build Targets
The repository SHALL include a `Makefile` with standard targets used by CI and the Definition of Done verification steps.

#### Scenario: build target
- **WHEN** `make build` is run
- **THEN** the binary SHALL be produced at `./bin/hf` with the git-describe version embedded via ldflags

#### Scenario: test target
- **WHEN** `make test` is run
- **THEN** `go test ./...` SHALL execute and exit 0 if all tests pass

#### Scenario: vet target
- **WHEN** `make vet` is run
- **THEN** `go vet ./...` SHALL execute and exit 0 if no issues are found
