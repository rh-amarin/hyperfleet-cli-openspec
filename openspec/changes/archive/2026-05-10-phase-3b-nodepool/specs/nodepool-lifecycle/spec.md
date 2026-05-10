# NodePool Lifecycle — Phase 3b Delta

## ADDED Requirements

### Requirement: CLI NodePool List Command

The CLI SHALL implement `hf nodepool list` to list all nodepools via GET /nodepools.

#### Scenario: List nodepools as JSON

- GIVEN an active environment is configured
- WHEN the user runs `hf nodepool list`
- THEN the CLI MUST send GET to /nodepools and output ListResponse[NodePool] as JSON

#### Scenario: List nodepools as table

- GIVEN nodepools exist
- WHEN the user runs `hf nodepool list --output table`
- THEN the CLI MUST display columns: ID, NAME, TYPE, GEN, REPLICAS, STATUS

### Requirement: CLI NodePool Get Command

The CLI SHALL implement `hf nodepool get [id]` to retrieve a single nodepool.

#### Scenario: Get nodepool by explicit ID

- GIVEN a valid nodepool ID is provided
- WHEN the user runs `hf nodepool get <id>`
- THEN the CLI MUST send GET to /nodepools/<id> and output the NodePool JSON

#### Scenario: Get nodepool from state

- GIVEN nodepool-id is set in state and no argument is provided
- WHEN the user runs `hf nodepool get`
- THEN the CLI MUST use the state nodepool-id

### Requirement: CLI NodePool Create Command

The CLI SHALL implement `hf nodepool create` to create a new nodepool.

#### Scenario: Create with defaults

- GIVEN no flags are provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST POST with name="my-nodepool", type="m4", replicas=1

#### Scenario: Create with explicit flags

- GIVEN --name, --type, --replicas flags are provided
- WHEN the user runs `hf nodepool create --name workers --type n2-standard-4 --replicas 3`
- THEN the CLI MUST POST with the provided values and persist nodepool-id to state

#### Scenario: Duplicate guard

- GIVEN a nodepool with the same name already exists
- WHEN the user runs `hf nodepool create --name existing`
- THEN the CLI MUST warn and skip the POST

### Requirement: CLI NodePool Update Command

The CLI SHALL implement `hf nodepool update <id>` to update a nodepool via PATCH.

#### Scenario: Update nodepool name

- GIVEN a valid nodepool ID is provided
- WHEN the user runs `hf nodepool update <id> --name new-name`
- THEN the CLI MUST send PATCH to /nodepools/<id> with the new name and output the updated NodePool

### Requirement: CLI NodePool Delete Command

The CLI SHALL implement `hf nodepool delete <id>` to delete a nodepool.

#### Scenario: Delete existing nodepool

- GIVEN a nodepool exists
- WHEN the user runs `hf nodepool delete <id>`
- THEN the CLI MUST send DELETE to /nodepools/<id> with no output on success

#### Scenario: Delete non-existent nodepool

- GIVEN no nodepool exists with the given ID
- WHEN the user runs `hf nodepool delete <id>`
- THEN the CLI MUST display `[ERROR] NodePool '<id>' not found` and exit with code 1

### Requirement: CLI NodePool Conditions Command

The CLI SHALL implement `hf nodepool conditions [id]` to display status conditions.

#### Scenario: Get conditions as JSON

- GIVEN cluster-id and nodepool-id are set in state
- WHEN the user runs `hf nodepool conditions`
- THEN the CLI MUST output `{generation, status: {conditions: [...]}}` as JSON

#### Scenario: Get conditions as table

- GIVEN a nodepool exists
- WHEN the user runs `hf nodepool conditions <id> --output table`
- THEN the CLI MUST display columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE

### Requirement: CLI NodePool Statuses Command

The CLI SHALL implement `hf nodepool statuses [id]` to display adapter statuses.

#### Scenario: Get statuses as JSON

- GIVEN nodepool-id is set in state
- WHEN the user runs `hf nodepool statuses`
- THEN the CLI MUST send GET to /nodepools/<id>/statuses and output ListResponse[AdapterStatus] as JSON

#### Scenario: Get statuses as table

- GIVEN adapter statuses exist for the nodepool
- WHEN the user runs `hf nodepool statuses --output table`
- THEN the CLI MUST display columns: ADAPTER, GEN, AVAILABLE
