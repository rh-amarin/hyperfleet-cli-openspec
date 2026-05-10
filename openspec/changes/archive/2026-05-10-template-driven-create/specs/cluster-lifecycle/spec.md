# Cluster Lifecycle — Template-Driven Create Delta

## MODIFIED Requirements

### Requirement: Create Cluster

`hf cluster create` SHALL load the request body from a JSON template file rather than hardcoded defaults. The binary embeds a built-in default template (`cmd/assets/cluster-template.json`) and auto-creates `<config-dir>/cluster-template.json` on first use.

#### Scenario: Create cluster with default template (no template file exists) (MODIFIED)

- GIVEN no `cluster-template.json` exists in the config dir
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST write the built-in default to `<config-dir>/cluster-template.json`
- AND print `[INFO] Created default cluster template at <path>`
- AND create the cluster using the default payload (`kind=Cluster`, `name=my-cluster`, default labels and spec)

#### Scenario: Create cluster using existing config-dir template (MODIFIED)

- GIVEN `<config-dir>/cluster-template.json` exists with a custom payload
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST use the custom payload as the request body

#### Scenario: Create cluster with `-f` file override (ADDED)

- GIVEN a file at `<path>` containing a valid JSON cluster payload
- WHEN the user runs `hf cluster create -f <path>`
- THEN the CLI MUST use that file's content as the request body
- AND MUST NOT read or write the config-dir template

#### Scenario: Create cluster with name positional argument (MODIFIED)

- GIVEN a template (any source) with `"name": "<template-name>"`
- WHEN the user runs `hf cluster create <name>`
- THEN the CLI MUST set `name` to `<name>` in the request body, overriding the template value
- AND positional arg takes precedence over `--name` flag

#### Scenario: Create cluster with positional arguments (PRESERVED)

- WHEN the user runs `hf cluster create <name> [region] [version]`
- THEN positional args override the corresponding template fields (`name`, `spec.region`, `spec.version`)

#### Scenario: Malformed template file (ADDED)

- GIVEN a template file containing invalid JSON
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST exit with `[ERROR] loading template: <reason>` and code 1

#### Scenario: Create cluster with no arguments (MODIFIED)

- GIVEN the config-dir template exists (or is auto-created on first use)
- WHEN the user runs `hf cluster create` with no arguments
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the template payload
