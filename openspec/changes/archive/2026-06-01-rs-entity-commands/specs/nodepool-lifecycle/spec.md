## REMOVED Requirements

### Requirement: Create NodePool

**Reason:** Nodepool CLI operations are normative under `hf rs nodepools` (see `rs-entity-commands`). The `hf nodepool` command group is no longer registered.

**Migration:** Use `hf rs nodepools create [name]` with templates at `~/.config/hf/templates/nodepools.json`.

### Requirement: List NodePools

**Reason:** Superseded by `hf rs nodepools list` / `hf rs nodepools table`.

**Migration:** Use `hf rs nodepools list` or `hf rs nodepools table`.

### Requirement: Search NodePool

**Reason:** Superseded by `hf rs nodepools search`.

**Migration:** Use `hf rs nodepools search <name>`.

### Requirement: Get NodePool

**Reason:** Superseded by `hf rs nodepools get`.

**Migration:** Use `hf rs nodepools get [id]`.

### Requirement: Interactive NodePool View

**Reason:** Legacy `hf nodepool` entrypoint removed.

**Migration:** Use `hf rs nodepools get -i` when interactive view is supported.

### Requirement: Patch NodePool

**Reason:** Superseded by `hf rs nodepools patch`.

**Migration:** Use `hf rs nodepools patch {spec|labels} [id]`.

### Requirement: Delete NodePool

**Reason:** Superseded by `hf rs nodepools delete` and `hf rs nodepools force-delete`.

**Migration:** Use `hf rs nodepools delete [id]` or `hf rs nodepools force-delete [id] --reason <text>`.

### Requirement: Get NodePool Conditions

**Reason:** Superseded by `hf rs nodepools conditions`.

**Migration:** Use `hf rs nodepools conditions [id]`.

### Requirement: Get NodePool Conditions Table

**Reason:** Superseded by `hf rs nodepools conditions --output table`.

**Migration:** Use `hf rs nodepools conditions [id] --output table`.

### Requirement: Get NodePool Adapter Statuses

**Reason:** Superseded by `hf rs nodepools statuses`.

**Migration:** Use `hf rs nodepools statuses [id]`.

### Requirement: Interactive NodePool Status Filter

**Reason:** Superseded by `hf rs nodepools statuses --filter`.

**Migration:** Use `hf rs nodepools statuses --filter`.

### Requirement: Display NodePool Table

**Reason:** Superseded by `hf rs nodepools table`.

**Migration:** Use `hf rs nodepools table`.

### Requirement: NodePool Create Dry-Run

**Reason:** Dry-run remains on global `--curl` for `hf rs nodepools create`.

**Migration:** Use `hf rs nodepools create ... --curl`.

## ADDED Requirements

### Requirement: NodePool Lifecycle via hf rs nodepools

The CLI SHALL expose all nodepool lifecycle operations under `hf rs nodepools` (alias `hf resource nodepools`). API paths, request bodies, and response handling SHALL match the behaviors previously normative for `hf nodepool` commands, as specified in `rs-entity-commands`.

#### Scenario: hf nodepool command removed

- GIVEN the rs-entity-commands change is complete
- WHEN the user runs `hf nodepool list`
- THEN the CLI MUST NOT register `hf nodepool`
- AND the user MUST use `hf rs nodepools list` instead
