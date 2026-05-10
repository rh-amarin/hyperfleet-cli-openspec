# Shell Completions Spec

## ADDED Requirements

### Requirement: Shell completion command

The CLI SHALL provide a `completion` subcommand that generates shell completion scripts.

#### Scenario: Generate bash completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion bash`
- THEN the CLI MUST output a bash completion script to stdout
- AND the output MUST be compatible with `source <(hf completion bash)`

#### Scenario: Generate zsh completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion zsh`
- THEN the CLI MUST output a zsh completion script to stdout
- AND the output MUST begin with `#compdef hf`

#### Scenario: Generate fish completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion fish`
- THEN the CLI MUST output a non-empty fish completion script to stdout

#### Scenario: Generate powershell completion

- GIVEN the user has installed `hf`
- WHEN the user runs `hf completion powershell`
- THEN the CLI MUST output a non-empty PowerShell completion script to stdout

#### Scenario: Reject invalid shell argument

- GIVEN the user runs `hf completion invalid`
- THEN the CLI MUST exit with a non-zero exit code
- AND MUST NOT output a completion script

### Requirement: Output flag tab completion

The CLI SHALL provide tab completions for the `--output` persistent flag.

#### Scenario: --output completions

- GIVEN a shell with tab completion enabled
- WHEN the user types `hf <command> --output <TAB>`
- THEN the shell MUST offer `json`, `table`, and `yaml` as completions
- AND MUST NOT offer file paths
