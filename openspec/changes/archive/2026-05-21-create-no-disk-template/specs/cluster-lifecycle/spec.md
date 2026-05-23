# Cluster Lifecycle — delta spec for create-no-disk-template

Only the scenarios below change. All other scenarios from `openspec/specs/cluster-lifecycle/spec.md` remain in force.

## MODIFIED Requirements

### Requirement: Create Cluster

`hf cluster create` SHALL load the request body from a JSON template. The binary embeds a built-in default template (`cmd/assets/cluster-template.json`). When no `--file` flag is given, the CLI MUST use the embedded default bytes directly in memory — it MUST NOT read from or write to `<config-dir>`.

#### Scenario: Create cluster with default template (no template file exists)

- GIVEN no `--file` flag is provided
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST use the built-in embedded template in memory
- AND MUST NOT write any file to `<config-dir>`
- AND MUST NOT print `[INFO] Created default cluster template at …`
- AND MUST create the cluster using the default payload (`kind=Cluster`, `name=my-cluster`, default labels and spec)

#### Scenario: Create cluster using existing config-dir template

- GIVEN `<config-dir>/cluster-template.json` exists on disk
- WHEN the user runs `hf cluster create` without `--file`
- THEN the CLI MUST use the embedded default in memory
- AND MUST NOT read the on-disk file
- AND MUST NOT write any file to `<config-dir>`

#### Scenario: Create cluster with no arguments

- GIVEN no `--file` flag and no positional arguments
- WHEN the user runs `hf cluster create`
- THEN the CLI MUST NOT show a usage message
- AND MUST proceed with creation using the embedded template payload
- AND MUST NOT write any file to `<config-dir>`
