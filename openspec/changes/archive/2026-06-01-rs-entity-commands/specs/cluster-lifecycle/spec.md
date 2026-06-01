## REMOVED Requirements

### Requirement: Create Cluster

**Reason:** Cluster CLI operations are normative under `hf rs clusters` (see `rs-entity-commands`). The `hf cluster` command group is no longer registered.

**Migration:** Use `hf rs clusters create [name] [region] [version]` with templates at `~/.config/hf/templates/clusters.json`.

### Requirement: Search Cluster

**Reason:** Superseded by `hf rs clusters search`.

**Migration:** Use `hf rs clusters search <name>`.

### Requirement: Get Cluster

**Reason:** Superseded by `hf rs clusters get`.

**Migration:** Use `hf rs clusters get [id]`.

### Requirement: Interactive Cluster View

**Reason:** Interactive cluster viewer remains available via `hf rs clusters get -i` when implemented; legacy `hf cluster` entrypoint removed.

**Migration:** Use `hf rs clusters get -i`.

### Requirement: Patch Cluster

**Reason:** Superseded by `hf rs clusters patch`.

**Migration:** Use `hf rs clusters patch {spec|labels} [id]`.

### Requirement: Delete Cluster

**Reason:** Superseded by `hf rs clusters delete`.

**Migration:** Use `hf rs clusters delete [id]`; force-delete via `--force --reason`.

### Requirement: Get Cluster Conditions

**Reason:** Superseded by `hf rs clusters conditions`.

**Migration:** Use `hf rs clusters conditions [id]`.

### Requirement: Get Cluster Conditions Table

**Reason:** Table output is covered by `hf rs clusters conditions --output table`.

**Migration:** Use `hf rs clusters conditions [id] --output table`.

### Requirement: Get Cluster Adapter Statuses

**Reason:** Superseded by `hf rs clusters statuses`.

**Migration:** Use `hf rs clusters statuses [id]`.

### Requirement: Interactive Cluster Status Filter

**Reason:** Superseded by `hf rs clusters statuses --filter`.

**Migration:** Use `hf rs clusters statuses --filter`.

### Requirement: List Clusters

**Reason:** Superseded by `hf rs clusters list` / `hf rs clusters table`.

**Migration:** Use `hf rs clusters list` or `hf rs clusters table`.

### Requirement: Cluster Create Dry-Run

**Reason:** Dry-run remains on global `--curl` for `hf rs clusters create`.

**Migration:** Use `hf rs clusters create ... --curl`.

## ADDED Requirements

### Requirement: Cluster Lifecycle via hf rs clusters

The CLI SHALL expose all cluster lifecycle operations under `hf rs clusters` (alias `hf resource clusters`). API paths, request bodies, and response handling SHALL match the behaviors previously normative for `hf cluster` commands, as specified in `rs-entity-commands`.

#### Scenario: hf cluster command removed

- GIVEN the rs-entity-commands change is complete
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST NOT register `hf cluster`
- AND the user MUST use `hf rs clusters list` instead
