## ADDED Requirements

### Requirement: config-get

`hf config get <section> <key>` SHALL print a single configuration value.

#### Scenario: Key found

- WHEN the user runs `hf config get hyperfleet api-url`
- THEN the CLI MUST print the resolved value of that key
- AND exit with code 0

#### Scenario: Key not found

- WHEN the user runs `hf config get <section> <unknown-key>`
- THEN the CLI MUST print `[ERROR] Config key '<section>.<key>' not found`
- AND exit with code 1

#### Scenario: Unknown section

- WHEN the user runs `hf config get <unknown-section> <key>`
- THEN the CLI MUST print `[ERROR] Config key '<unknown-section>.<key>' not found`
- AND exit with code 1

---

### Requirement: config-set-section-validation

`hf config set` SHALL validate that the target section is one of the known sections.

#### Scenario: Valid section

- WHEN the user runs `hf config set hyperfleet api-url http://newapi:8000`
- THEN the value MUST be written to config.yaml
- AND the command MUST exit with code 0 and no output

#### Scenario: Invalid section

- WHEN the user runs `hf config set badSection someKey value`
- THEN the CLI MUST print `[ERROR] Unknown config section 'badSection'`
- AND exit with code 1

---

### Requirement: config-env-create

`hf config env create <name>` SHALL create a new environment profile file.

#### Scenario: Create with flags

- WHEN the user runs `hf config env create dev --api-url http://dev:8000 --api-token tok123`
- THEN a file MUST be created at `~/.config/hf/environments/dev.yaml`
- AND it MUST contain the provided values under the appropriate YAML sections
- AND the CLI MUST print a success hint to run `hf config env activate dev`

#### Scenario: Already exists

- GIVEN an environment named `dev` already exists
- WHEN the user runs `hf config env create dev`
- THEN the CLI MUST print `[ERROR] Environment 'dev' already exists`
- AND exit with code 1

---

### Requirement: config-env-delete

`hf config env delete <name>` (alias: `rm`) SHALL remove an environment profile.

#### Scenario: Delete existing environment

- WHEN the user runs `hf config env delete dev`
- THEN the environment file MUST be removed from disk
- AND the CLI MUST exit with code 0

#### Scenario: Delete active environment

- GIVEN `dev` is the currently active environment
- WHEN the user runs `hf config env delete dev`
- THEN the file MUST be removed
- AND `active-environment` MUST be cleared from state.yaml

#### Scenario: Delete non-existent environment

- WHEN the user runs `hf config env delete nonexistent`
- THEN the CLI MUST print `[ERROR] Environment 'nonexistent' not found`
- AND exit with code 1

---

### Requirement: config-doctor

`hf config doctor` SHALL check connectivity to the configured API server.

#### Scenario: API reachable

- GIVEN the API server is running at the configured URL
- WHEN the user runs `hf config doctor`
- THEN the CLI MUST print `[OK] API server reachable at <url>`
- AND exit with code 0

#### Scenario: API unreachable

- GIVEN the API server is not reachable
- WHEN the user runs `hf config doctor`
- THEN the CLI MUST print `[ERROR] Cannot reach API server at <url>: <reason>`
- AND exit with code 1

---

### Requirement: active-env-guard-implementation

The root `PersistentPreRunE` SHALL enforce that an active environment is set before executing any command except the defined bypass list.

#### Scenario: Bypass list commands succeed without active env

- GIVEN no active environment is set
- WHEN the user runs any of: `hf config env list`, `hf config env create`, `hf config env activate`, `hf config env show`, `hf config env delete`, `hf version`, `hf completion`, `hf help`
- THEN the command MUST execute normally

#### Scenario: Non-bypass command fails without active env

- GIVEN no active environment is set
- WHEN the user runs `hf config show` or `hf config set`
- THEN the CLI MUST print `[ERROR] No active environment. Run 'hf config env activate <name>' to set one.`
- AND exit with code 1

## MODIFIED Requirements

## Requirement: Show Configuration

`hf config show` (and `hf config` with no args) SHALL display the resolved configuration.

#### Scenario: Active environment set

GIVEN an active environment is configured in state.yaml
WHEN the user runs `hf config show`
THEN the active environment name MUST be printed at the top
AND config values MUST be shown grouped by section with no source annotations
AND secrets (token, database.password, rabbitmq.password) MUST be shown as `<set>` or `<not set>`

#### Scenario: Show config for a named environment

GIVEN environment profiles exist in `~/.config/hf/environments/`
WHEN the user runs `hf config show <env-name>`
THEN the CLI MUST display the file path of that environment file (e.g., `~/.config/hf/environments/<env-name>.yaml`)
AND MUST display the values defined in that environment file grouped by section
AND secrets MUST be shown as `<set>` or `<not set>`
AND the active environment MUST NOT be changed

#### Scenario: Show config for a named environment that is also the active environment

GIVEN the named environment is also the currently active environment
WHEN the user runs `hf config show <env-name>`
THEN the CLI MUST display the file path prefixed with `[active] `
AND MUST display the values as usual
AND the active environment MUST NOT be changed

#### Scenario: No active environment

GIVEN no active environment is configured
WHEN the user runs `hf config show`
THEN the command MUST exit with code 1
AND MUST print the no-active-environment error with guidance

---

## Requirement: hf-config-env-show

`hf config env show <name>` SHALL display the file location and values of a named environment profile without activating it.

#### Scenario: Show named environment

GIVEN an environment named `<name>` exists at `~/.config/hf/environments/<name>.yaml`
WHEN the user runs `hf config env show <name>`
THEN the CLI MUST print the absolute file path of the environment file
AND MUST display all key-value pairs defined in that file, grouped by section
AND secrets MUST be shown as `<set>` or `<not set>`
AND the currently active environment MUST NOT be changed

#### Scenario: Non-existent environment

GIVEN no environment named `<name>` exists
WHEN the user runs `hf config env show <name>`
THEN the CLI MUST print `[ERROR] environment '<name>' not found`
AND MUST exit with code 1

---

## Requirement: active-env-guard

Commands that require a configured target MUST fail when no active environment is set.

#### Scenario: Command requires active env, none set

GIVEN no active environment is configured
WHEN the user runs any of: `hf config show`, `hf config set`
THEN the command MUST exit non-zero
AND MUST print the no-active-environment error with guidance

#### Scenario: Always-available commands

GIVEN no active environment is configured
WHEN the user runs any of: `hf config env list`, `hf config env new`, `hf config env activate`, `hf config env show`, `hf config env create`, `hf config env delete`
THEN the command MUST succeed normally
