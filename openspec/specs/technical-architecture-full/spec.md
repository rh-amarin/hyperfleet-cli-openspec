# technical-architecture-full Specification

## Purpose
TBD - created by archiving change technical-architecture-full. Update Purpose after archive.
## Requirements
### Requirement: Command stub registration

Every domain command group SHALL be registered with the root command so that `hf --help` lists all top-level subcommands.

#### Scenario: Stub commands visible in help

- **WHEN** the user runs `hf --help`
- **THEN** the following commands MUST appear in the output: `cluster`, `nodepool`, `config`, `db`, `maestro`, `pubsub`, `rabbitmq`, `kube`, `logs`, `repos`, `resources`, `version`, `completion`

#### Scenario: Stub group commands show help when invoked directly

- **WHEN** the user runs `hf cluster` (or any other stub group command) with no subcommand
- **THEN** the CLI MUST print help text for that command and exit 0

### Requirement: Version package

The CLI SHALL expose a canonical version string through `internal/version`.

#### Scenario: Default version

- **WHEN** the binary is built without `-ldflags` injection
- **THEN** `internal/version.Version` MUST equal `"dev"`

#### Scenario: Injected version

- **WHEN** the binary is built with `-ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=<tag>"`
- **THEN** `internal/version.Version` MUST equal `<tag>`

#### Scenario: Version command output

- **WHEN** the user runs `hf version`
- **THEN** the CLI MUST print the version string to stdout and exit 0

### Requirement: Makefile build targets

The repository SHALL provide a `Makefile` with standard developer targets.

#### Scenario: Build target

- **WHEN** the developer runs `make build`
- **THEN** the binary `bin/hf` MUST be produced with no errors

#### Scenario: Test target

- **WHEN** the developer runs `make test`
- **THEN** `go test ./...` MUST run and exit 0 when all tests pass

#### Scenario: Vet target

- **WHEN** the developer runs `make vet`
- **THEN** `go vet ./...` MUST run and exit 0 when no issues are found

### Requirement: Unit test coverage for scaffold

Every new package introduced by this change SHALL have unit tests.

#### Scenario: Version package tests

- **WHEN** `go test ./internal/version/...` is run
- **THEN** all tests MUST pass, verifying that `Version` is non-empty and `String()` returns it

#### Scenario: Root command tests

- **WHEN** `go test ./cmd/...` is run
- **THEN** all tests MUST pass, verifying `Execute()` returns nil for `--help` and all global flags are registered

