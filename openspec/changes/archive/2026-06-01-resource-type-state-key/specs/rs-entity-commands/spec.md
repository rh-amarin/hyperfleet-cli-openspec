## MODIFIED Requirements

### Requirement: Reconciled Entity List Table

For entities `clusters` and `nodepools`, list/table output with `--output table` SHALL follow the Table Column Architecture in `tables-and-lists` (fixed columns, dynamic condition columns, dynamic adapter columns, GEN deletion marker, watch spinners).

#### Scenario: Nodepool list table columns

- GIVEN `clusters` is set in state and nodepools exist
- WHEN the user runs `hf rs nodepools table`
- THEN fixed columns MUST include `ID`, `NAME`, `TYPE`, `GEN`, `REPLICAS` where `TYPE` is `spec.platform.type` and `REPLICAS` is `spec.replicas`
