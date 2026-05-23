# Config — delta spec for picker-header-and-config-set-preview

Only the scenarios below change. All other scenarios from `openspec/specs/config/spec.md` remain in force.

## MODIFIED Requirements

### Requirement: Set Configuration Value

The CLI SHALL allow setting individual configuration properties using dotted section.key notation. When called with no arguments the command SHALL launch an interactive fuzzy-finder to select the parameter. The picker MUST use a split-screen layout: the left panel lists filterable section.key pairs with their current values; the right panel MUST display the full active configuration as colorized YAML. A header MUST be shown above the item list explaining the operation and keyboard shortcuts. After any successful set the full active configuration MUST be displayed.

#### Scenario: Interactive set — split-screen preview

- GIVEN an active environment is configured
- WHEN the user runs `hf config set` with no arguments
- THEN the CLI MUST display a split-screen picker
- AND the left panel MUST list all known section.key pairs with their current values
- AND the right panel MUST display the full active configuration as colorized YAML
- AND a header MUST appear above the item list describing the operation and keyboard shortcuts

#### Scenario: Interactive set — fuzzy-find parameter selection

- GIVEN an active environment is configured
- WHEN the user runs `hf config set` with no arguments
- THEN the CLI MUST open an interactive fuzzy-finder listing all known `section.key` pairs with their current values (secrets shown as `<set>` or `<not set>`)
- AND the user can type to filter and press Enter to select a parameter
- AND after selection the CLI MUST prompt for the new value on stdin
- AND after the user enters a value the CLI MUST write it to the active environment file
- AND the CLI MUST display the full active configuration after the successful write

#### Scenario: Interactive set — user aborts selection

- GIVEN an active environment is configured
- WHEN the user runs `hf config set` with no arguments
- AND presses Escape or Ctrl+C in the fuzzy-finder
- THEN the CLI MUST exit with code 0 without modifying any configuration

#### Scenario: Invalid key format

- GIVEN the user runs `hf config set` without a dotted key
- WHEN the key does not contain a `.` separator
- THEN the command MUST exit with code 1
- AND MUST print `[ERROR] key must be in section.key format (e.g. hyperfleet.api-url)`

#### Scenario: Unknown section

- GIVEN the user runs `hf config set` with an unknown section prefix
- WHEN the section is not one of: hyperfleet, kubernetes, maestro, port-forward, database, rabbitmq, registry
- THEN the command MUST exit with code 1
- AND MUST print `[ERROR] unknown config section '<section>'`

#### Scenario: No active environment

- GIVEN no active environment is configured
- WHEN the user runs `hf config set <section.key> <value>`
- THEN the command MUST exit with code 1
- AND MUST print the no-active-environment error with guidance

### Requirement: hf-config-env-create

Environment management commands MUST be accessible exclusively at the top level as `hf env <subcommand>`. The `hf config env` group is removed. `hf env` with no arguments MUST launch an interactive fuzzy picker with a split-screen layout: left panel lists environments; right panel MUST display the highlighted environment's full YAML. A header MUST be shown above the item list explaining the operation and keyboard shortcuts.

#### Scenario: hf env bare — interactive picker

- GIVEN one or more environment profiles exist
- WHEN the user runs `hf env` with no arguments
- THEN the CLI MUST display a fuzzy-searchable list of environment names in the left panel
- AND MUST display the full colorized YAML of the highlighted environment in the right preview panel
- AND the currently active environment MUST be marked with a checkmark in the list
- AND a header MUST appear above the item list describing the operation and keyboard shortcuts
- WHEN the user selects an environment
- THEN the CLI MUST activate it and print the full active config (same as `hf config show`)
- WHEN the user aborts (Esc / Ctrl+C)
- THEN the CLI MUST exit cleanly with no changes

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
- THEN the CLI MUST display a table of environment names with an active marker

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
