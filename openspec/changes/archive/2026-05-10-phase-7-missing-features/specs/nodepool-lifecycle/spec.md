# NodePool Lifecycle — Phase 7 Delta

## MODIFIED Requirements

### Requirement: Get NodePool Adapter Statuses

The statuses table SHALL include a FINALIZED column in addition to AVAILABLE.

#### Scenario: Get statuses table (MODIFIED — add FINALIZED column)

- WHEN the user runs `hf nodepool statuses --output table`
- THEN the CLI MUST output columns: ADAPTER, GEN, AVAILABLE, FINALIZED
- AND AVAILABLE and FINALIZED columns MUST be color-coded dots: green=True, red=False, `-`=not present
