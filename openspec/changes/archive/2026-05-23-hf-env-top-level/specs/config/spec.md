# Config — delta spec for hf-env-top-level

Only the scenarios below change. All other scenarios from `openspec/specs/config/spec.md` remain in force.

## MODIFIED Requirements

### Requirement: hf-config-env-create

Environment management commands MUST be accessible exclusively at the top level as `hf env <subcommand>`. The `hf config env` group is removed. `hf env` with no arguments MUST launch an interactive fuzzy picker.

#### Scenario: Create environment via hf env create

- GIVEN no environment named `<name>` exists
- WHEN the user runs `hf env create <name>`
- THEN the CLI MUST copy the bundled template to `~/.config/hf/environments/<name>.yaml`
- AND set `active-environment: <name>` in `state.yaml`
- AND print the absolute path of the new file with a message to edit it

#### Scenario: Name already exists

- GIVEN an environment named `<name>` already exists
- WHEN the user runs `hf env create <name>`
- THEN the CLI MUST exit with code 1
- AND MUST print `[ERROR] environment '<name>' already exists`

#### Scenario: Name not provided

- GIVEN the user runs `hf env create` with no argument
- THEN the CLI MUST show the command usage and exit with code 1

#### Scenario: List environments via hf env list

- GIVEN one or more environment profiles exist
- WHEN the user runs `hf env list` (or `hf env ls`)
- THEN the CLI MUST display a table of environment names with an active marker (✓)

#### Scenario: Activate environment via hf env activate

- GIVEN the user runs `hf env activate <name>` and the environment exists
- THEN the CLI MUST set `<name>` as the active environment and print a confirmation

#### Scenario: Delete environment via hf env delete

- GIVEN the user runs `hf env delete <name>` (or `hf env rm <name>`)
- THEN the CLI MUST delete the environment file; if it was active, clear the active-environment state

#### Scenario: hf env bare — no environments

- GIVEN no environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST print the command help block
- AND MUST print a message directing the user to run `hf env create <name>`

#### Scenario: hf env bare — interactive picker

- GIVEN one or more environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST display a fuzzy-searchable list of environment names in the left panel
- AND MUST display the full colorized YAML of the highlighted environment in the right preview panel
- AND the currently active environment MUST be marked with ✓ in the list
- WHEN the user selects an environment
- THEN the CLI MUST activate it and print the full active config (same as `hf config show`)
- WHEN the user aborts (Esc / Ctrl+C)
- THEN the CLI MUST exit cleanly with no changes

### Requirement: hf-config-env-show

`hf env show <name>` MUST display the file location and values of a named environment profile without activating it. Behavior is identical to the former `hf config env show <name>`.

#### Scenario: Show named environment

- GIVEN an environment named `<name>` exists
- WHEN the user runs `hf env show <name>`
- THEN the CLI MUST print the absolute file path of the environment file
- AND display all key-value pairs defined in that file, with secrets redacted
- AND the currently active environment MUST NOT be changed
- AND if `<name>` is the active environment the path MUST be prefixed with `[active]`

#### Scenario: Non-existent environment

- GIVEN no environment named `<name>` exists
- WHEN the user runs `hf env show <name>`
- THEN the CLI MUST print `[ERROR] environment '<name>' not found`
- AND exit with code 1

### Requirement: active-env-guard

Commands that require a configured target MUST fail when no active environment is set.

#### Scenario: Command requires active env, none set

- GIVEN no active environment is configured
- WHEN the user runs any of: `hf config show`, `hf config set`
- THEN the command MUST exit with code 1
- AND MUST print:
  ```
  [ERROR] no active environment
    → run 'hf env create <name>' to create one
    → run 'hf env activate <name>' to activate an existing one
  ```

#### Scenario: Always-available commands

- GIVEN no active environment is configured
- WHEN the user runs any of: `hf env`, `hf env list`, `hf env create`, `hf env activate`, `hf env show`
- THEN the command MUST succeed normally
