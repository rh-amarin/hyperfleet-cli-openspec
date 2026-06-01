# rs-entity-commands Specification

## Purpose
TBD - created by archiving change rs-entity-commands. Update Purpose after archive.
## Requirements
### Requirement: RS Entity Command Tree

For each `resource-types` entry, the CLI SHALL register `hf rs <entity>` (alias `hf resource <entity>`) with subcommands: `list`, `table`, `get`, `create`, `search`, `patch`, `delete`, `conditions`, `statuses`, `adapter-report`, and `id`.

Types that support force-delete (`clusters`, `nodepools` in the bundled template) SHALL also register `force-delete`.

#### Scenario: Entity help lists subcommands

- GIVEN `clusters` and `nodepools` are configured under `resource-types`
- WHEN the user runs `hf rs clusters --help`
- THEN help MUST list `list`, `table`, `get`, `create`, `search`, `patch`, `delete`, `conditions`, `statuses`, `adapter-report`, `id`, and `force-delete` where applicable

### Requirement: Table Subcommand

`hf rs <entity> table` SHALL be equivalent to `hf rs <entity> list --output table`.

#### Scenario: Table alias

- WHEN the user runs `hf rs clusters table`
- THEN the CLI MUST render the same output as `hf rs clusters list --output table`

### Requirement: Reconciled Entity List Table

For entities `clusters` and `nodepools`, list/table output with `--output table` SHALL follow the Table Column Architecture in `tables-and-lists` (fixed columns, dynamic condition columns, dynamic adapter columns, GEN deletion marker, watch spinners).

#### Scenario: Nodepool list table columns

- GIVEN `clusters` is set in state and nodepools exist
- WHEN the user runs `hf rs nodepools table`
- THEN fixed columns MUST include `ID`, `NAME`, `TYPE`, `GEN`, `REPLICAS` where `TYPE` is `spec.platform.type` and `REPLICAS` is `spec.replicas`

### Requirement: Conditions Subcommand

`hf rs <entity> conditions [id]` SHALL display `status.conditions` for the resolved resource.

#### Scenario: Conditions table output

- WHEN the user runs `hf rs clusters conditions [id] --output table`
- THEN columns MUST be `TYPE`, `STATUS`, `LAST TRANSITION`, `REASON`, `MESSAGE` with colored dots on `STATUS`

### Requirement: Statuses Subcommand

`hf rs <entity> statuses [id]` SHALL list adapter statuses from GET `{collection-path}/{id}/statuses`.

#### Scenario: Statuses table

- WHEN the user runs `hf rs <entity> statuses [id] --output table`
- THEN columns MUST be `ADAPTER`, `GEN`, `AVAILABLE`, `FINALIZED` with status dots

#### Scenario: Statuses filter

- GIVEN `--filter` is set
- WHEN the user runs `hf rs <entity> statuses [id] --filter`
- THEN the CLI MUST open the interactive status filter UI (same as legacy cluster/nodepool statuses)

### Requirement: Create Entity Parity

`hf rs <entity> create` SHALL support template-based create with entity-specific overrides for `clusters` and `nodepools`.

#### Scenario: Duplicate name guard

- GIVEN a non-empty create name and an existing resource with that name
- WHEN the user runs `hf rs <entity> create` without `--curl`
- THEN the CLI MUST warn and skip POST without error exit

#### Scenario: Clusters create positional args

- WHEN the user runs `hf rs clusters create [name] [region] [version]`
- THEN positional args MUST override `name`, `spec.region`, and `spec.version` after template merge
- AND flags `--replicas` and `--nodepool-id` MAY override template fields

#### Scenario: Nodepools create flags

- WHEN the user runs `hf rs nodepools create [name]`
- THEN `--type` MUST set `spec.platform.type` and `--replicas` MUST set `spec.replicas`
- AND parent `cluster-id` MUST be set in state

### Requirement: Patch Dual Mode

`hf rs <entity> patch {spec|labels} [id]` SHALL support counter increment when `--file` is omitted and file-based patch when `--file` is set.

#### Scenario: Counter patch default

- GIVEN no `--file` flag
- WHEN the user runs `hf rs clusters patch spec [id]`
- THEN the CLI MUST increment the counter field and PATCH using the same body shape as legacy `hf cluster patch`

#### Scenario: File patch

- GIVEN `--file` is set
- WHEN the user runs `hf rs <entity> patch spec [id] --file <path>`
- THEN the CLI MUST PATCH `{collection-path}/{id}/spec` with the file body

### Requirement: Delete and Force Delete

`hf rs <entity> delete [id]` SHALL DELETE the resource. Force-delete SHALL be available for `nodepools` via `force-delete` and for `clusters` via `delete --force`.

#### Scenario: Cluster delete returns body

- WHEN `hf rs clusters delete [id]` succeeds
- THEN the CLI MUST print the deleted resource JSON (legacy parity)

#### Scenario: Nodepool force-delete

- WHEN the user runs `hf rs nodepools force-delete [id] --reason <text>`
- THEN the CLI MUST POST `{path}/{id}/force-delete` with `{"reason":"<text>"}` and require `--reason`

#### Scenario: Delete not found

- GIVEN HTTP 404 from DELETE
- THEN the CLI MUST print `[ERROR] <Entity> '<id>' not found` and exit 1

### Requirement: Adapter Report Command

`hf rs <entity> adapter-report <adapter> <True|False|Unknown> <gen> [id]` SHALL PUT adapter status to `{collection-path}/{id}/statuses` (replacing legacy `adapter post-status` for config-defined types).

#### Scenario: Adapter report PUT body

- WHEN the user runs `hf rs clusters adapter-report cl-deployment True 3`
- THEN the PUT body MUST include four conditions with `ManualStatusPost` reason and message referencing `hf rs adapter-report`

### Requirement: Hierarchical Overview

`hf rs` with no subcommand SHALL be the canonical combined operational view. It MUST render a table overview of configured types with parent/child nesting, default output `table`, `--watch` support, partial-failure warnings at the top, and GEN deletion markers.

For environments with `clusters` and `nodepools` in `resource-types`, overview MUST also support the adapter-rich combined table (equivalent to the former `hf table` / `hf resources --output table`): dynamic condition and adapter columns, dot rendering, and watch spinners.

#### Scenario: Overview replaces hf table

- GIVEN `clusters` and `nodepools` are configured
- WHEN the user runs `hf rs --output table --watch`
- THEN output MUST be equivalent to the former `hf table --output table --watch` for the same cluster data

#### Scenario: Overview tolerates fetch errors

- GIVEN a child list fails for one parent resource
- WHEN the user runs `hf rs`
- THEN the CLI MUST still print successfully loaded rows and `[WARN]` lines for failures at the top of output

### Requirement: Legacy Command Deprecation

After implementation is complete, `hf cluster`, `hf nodepool`, `hf table`, and `hf resources` MUST NOT remain as separate command implementations.

#### Scenario: Canonical documentation

- GIVEN the change is archived
- WHEN operators read CLI documentation
- THEN cluster and nodepool workflows MUST be documented only under `hf rs clusters` and `hf rs nodepools`

### Requirement: Search Query Format

`hf rs <entity> search <name>` SHALL use `search=name='<name>'` on the collection path (legacy cluster search convention).

#### Scenario: Search by name

- WHEN the user runs `hf rs clusters search prod-eu`
- THEN the CLI MUST query with `search=name='prod-eu'`

