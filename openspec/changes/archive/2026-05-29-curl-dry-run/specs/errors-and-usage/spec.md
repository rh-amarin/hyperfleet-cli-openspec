## MODIFIED Requirements

### Requirement: Duplicate Creation Prevention

The CLI SHALL prevent duplicate resource creation.

#### Scenario: Create cluster with existing name

- GIVEN a cluster with the same name already exists
- WHEN the user runs `hf cluster create <existing-name>`
- THEN the CLI MUST search for an existing cluster with that name
- AND if found MUST print `[WARN] Cluster '<name>' already exists, skipping creation`
- AND exit with code 0
- AND MUST NOT send a POST create request

#### Scenario: Create cluster with existing name under dry-run

- GIVEN a cluster with the same name already exists
- WHEN the user runs `hf cluster create <existing-name> --curl`
- THEN the CLI MUST NOT perform the duplicate-check GET request
- AND MUST print the POST create curl to stderr
- AND MUST NOT print cluster data to stdout
- AND MUST NOT mutate state
- AND exit with code 0

#### Scenario: Create nodepool with existing name under dry-run

- GIVEN a nodepool with the same name already exists in the active cluster
- WHEN the user runs `hf nodepool create <existing-name> --curl`
- THEN the CLI MUST NOT perform the duplicate-check GET request
- AND MUST print the POST create curl to stderr
- AND MUST NOT mutate state
- AND exit with code 0
