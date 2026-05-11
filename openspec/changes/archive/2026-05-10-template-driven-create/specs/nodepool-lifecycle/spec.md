# NodePool Lifecycle — Template-Driven Create Delta

## MODIFIED Requirements

### Requirement: Create NodePool

`hf nodepool create` SHALL load the request body from a JSON template file rather than hardcoded defaults. The binary embeds a built-in default template (`cmd/assets/nodepool-template.json`) and auto-creates `<config-dir>/nodepool-template.json` on first use.

#### Scenario: Create nodepool with default template (no template file exists) (MODIFIED)

- GIVEN no `nodepool-template.json` exists in the config dir
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST write the built-in default to `<config-dir>/nodepool-template.json`
- AND print `[INFO] Created default nodepool template at <path>`
- AND create the nodepool using the default payload (`kind=NodePool`, `name=my-nodepool`, default labels and spec)

#### Scenario: Create nodepool using existing config-dir template (MODIFIED)

- GIVEN `<config-dir>/nodepool-template.json` exists with a custom payload
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST use the custom payload as the request body

#### Scenario: Create nodepool with `-f` file override (ADDED)

- GIVEN a file at `<path>` containing a valid JSON nodepool payload
- WHEN the user runs `hf nodepool create -f <path>`
- THEN the CLI MUST use that file's content as the request body
- AND MUST NOT read or write the config-dir template

#### Scenario: Create nodepool with name positional argument (ADDED)

- GIVEN a template (any source) with `"name": "<template-name>"`
- WHEN the user runs `hf nodepool create <name>`
- THEN the CLI MUST set `name` to `<name>` in the request body, overriding the template value
- AND positional arg takes precedence over `--name` flag

#### Scenario: Create nodepool with no arguments (PRESERVED)

- GIVEN the config-dir template exists (or is auto-created on first use)
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the template payload

#### Scenario: Malformed template file (ADDED)

- GIVEN a template file containing invalid JSON
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST exit with `[ERROR] loading template: <reason>` and code 1
