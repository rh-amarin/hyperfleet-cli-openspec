# Spec: config

## Requirement: Show Configuration

`hf config show` (and `hf config` with no args) displays the resolved configuration.

#### Scenario: Active environment set

GIVEN an active environment is configured in state.yaml
WHEN the user runs `hf config show`
THEN the active environment name is printed at the top
AND config values are shown grouped by section with no source annotations
AND secrets (token, database.password, rabbitmq.password) are shown as `<set>` or `<not set>`

#### Scenario: Show config for a named environment

GIVEN environment profiles exist in `~/.config/hf/environments/`
WHEN the user runs `hf config show <env-name>`
THEN the CLI MUST display the file path of that environment file (e.g., `~/.config/hf/environments/<env-name>.yaml`)
AND display the values defined in that environment file grouped by section
AND secrets MUST be shown as `<set>` or `<not set>`
AND the active environment is NOT changed

#### Scenario: Show config for a named environment that is also the active environment

GIVEN the named environment is also the currently active environment
WHEN the user runs `hf config show <env-name>`
THEN the CLI MUST display the file path prefixed with `[active] ` (e.g., `[active] ~/.config/hf/environments/<env-name>.yaml`)
AND display the values as usual
AND the active environment is NOT changed

#### Scenario: No active environment

GIVEN no active environment is configured
WHEN the user runs `hf config show`
THEN the command exits with code 1
AND prints:
```
error: no active environment
  → run 'hf config env new' to create one
  → run 'hf config env activate <name>' to activate an existing one
```

---

## Requirement: Set Configuration Value

The CLI SHALL allow setting individual configuration properties using dotted section.key notation.

#### Scenario: Set a config property

GIVEN an active environment is configured
WHEN the user runs `hf config set <section.key> <value>`
THEN the value MUST be written into the correct section of `config.yaml`
AND subsequent reads MUST return the new value

#### Scenario: No active environment

GIVEN no active environment is configured
WHEN the user runs `hf config set <section.key> <value>`
THEN the command exits non-zero with the no-active-environment error

---

## Requirement: hf-config-env-new

`hf config env new [name]` creates a new named environment profile.

#### Scenario: Name provided as argument

GIVEN the user runs `hf config env new dev`
WHEN the command starts
THEN no name prompt is shown
AND the user is prompted for configuration values with defaults
AND a sparse YAML file is saved to `environments/dev.yaml`
AND the success message instructs the user to run `hf config env activate dev`

#### Scenario: Name not provided

GIVEN the user runs `hf config env new` with no argument
WHEN the command starts
THEN the user is prompted to enter an environment name
AND after entering the name the config value prompts follow
AND behaviour is otherwise identical to the name-provided scenario

#### Scenario: Prompt defaults

GIVEN the user presses enter on every prompt
THEN the saved environment file contains:
- database.user = hyperfleet
- database.name = hyperfleet
- database.password = foobar-bizz-buzz
- rabbitmq.host = rabbitmq
- rabbitmq.user = guest
- rabbitmq.password = guest

---

## Requirement: hf-config-env-show

`hf config env show <name>` displays the file location and values of a named environment profile without activating it.

#### Scenario: Show named environment

GIVEN an environment named `<name>` exists at `~/.config/hf/environments/<name>.yaml`
WHEN the user runs `hf config env show <name>`
THEN the CLI MUST print the absolute file path of the environment file
AND display all key-value pairs defined in that file, grouped by section
AND secrets MUST be shown as `<set>` or `<not set>`
AND the currently active environment MUST NOT be changed

#### Scenario: Non-existent environment

GIVEN no environment named `<name>` exists
WHEN the user runs `hf config env show <name>`
THEN the CLI MUST print `[ERROR] environment '<name>' not found`
AND exit with code 1

---

## Requirement: Show Current Cluster and NodePool IDs

`hf cluster id` and `hf nodepool id` display the currently active resource IDs from state.

NOTE on severity: `hf cluster id` and `hf nodepool id` use `[WARN]` + exit 0 when no ID is set because they are purely informational — the absence of a stored ID is not an error condition for these commands. All other commands that *require* a cluster-id or nodepool-id to function use `[ERROR]` + exit 1.

#### Scenario: Display cluster ID

GIVEN a cluster-id is set in state.yaml
WHEN the user runs `hf cluster id`
THEN the CLI MUST print the value of `cluster-id` from `~/.config/hf/state.yaml`

#### Scenario: No cluster ID set

GIVEN no cluster-id is set in state.yaml
WHEN the user runs `hf cluster id`
THEN the CLI MUST print `[WARN] No cluster-id set in state`
AND exit with code 0

#### Scenario: Display nodepool ID

GIVEN a nodepool-id is set in state.yaml
WHEN the user runs `hf nodepool id`
THEN the CLI MUST print the value of `nodepool-id` from `~/.config/hf/state.yaml`

#### Scenario: No nodepool ID set

GIVEN no nodepool-id is set in state.yaml
WHEN the user runs `hf nodepool id`
THEN the CLI MUST print `[WARN] No nodepool-id set in state`
AND exit with code 0

---

## Requirement: active-env-guard

Commands that require a configured target must fail when no active environment is set.

#### Scenario: Command requires active env, none set

GIVEN no active environment is configured
WHEN the user runs any of: `hf config show`, `hf config set`
THEN the command exits non-zero
AND prints the no-active-environment error with guidance

#### Scenario: Always-available commands

GIVEN no active environment is configured
WHEN the user runs any of: `hf config env list`, `hf config env new`, `hf config env activate`, `hf config env show`
THEN the command succeeds normally
