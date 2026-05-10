# Delta spec: config

## Modified Requirement: Show Configuration

Amends the "Active environment set" scenario in `openspec/specs/config/spec.md`.

#### Scenario: Active environment set — state variables included

GIVEN an active environment is configured and state.yaml contains any of: cluster-id, cluster-name, nodepool-id
WHEN the user runs `hf config show`
THEN a `state:` section MUST appear first in the output
AND it MUST list all non-empty state keys (active-environment, cluster-id, cluster-name, nodepool-id) as YAML key-value pairs
AND the existing config sections (hyperfleet, kubernetes, …) follow the state block unchanged

---

## New Requirement: Get Configuration Value

`hf config get <key>` retrieves a single value using dot notation for config or a plain key for state.

#### Scenario: Get a config value

GIVEN an active environment is set
WHEN the user runs `hf config get hyperfleet.api-url`
THEN the CLI MUST print the resolved value of `api-url` in the `hyperfleet` section

#### Scenario: Get a state value

GIVEN an active environment is set and cluster-id is present in state.yaml
WHEN the user runs `hf config get cluster-id`
THEN the CLI MUST print the value of `cluster-id` from state.yaml

#### Scenario: Key not found

GIVEN an active environment is set
WHEN the user runs `hf config get hyperfleet.nonexistent`
THEN the CLI MUST print `[ERROR] Config key 'hyperfleet.nonexistent' not found`
AND exit with code 1

#### Scenario: No arguments provided

GIVEN the user runs `hf config get` with no arguments
THEN the CLI MUST display the command's full help text
AND exit with code 1
