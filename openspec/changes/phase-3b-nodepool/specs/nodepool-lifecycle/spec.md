# NodePool Lifecycle — Phase 3b Delta

This change implements the CLI layer from the nodepool-lifecycle spec.

## Commands implemented

- `hf nodepool list` — GET /nodepools, table columns: ID NAME TYPE GEN REPLICAS STATUS
- `hf nodepool get [id]` — GET /nodepools/<id>, falls back to state nodepool-id
- `hf nodepool create` — POST /nodepools, flags: --name --type --replicas, persists nodepool-id
- `hf nodepool update <id>` — PATCH /nodepools/<id>, flags: --name --replicas
- `hf nodepool delete <id>` — DELETE /nodepools/<id>, silent success, [ERROR] on 404
- `hf nodepool conditions [id]` — GET /nodepools/<id>, renders conditions, falls back to state
- `hf nodepool statuses [id]` — GET /nodepools/<id>/statuses, table: ADAPTER GEN AVAILABLE
