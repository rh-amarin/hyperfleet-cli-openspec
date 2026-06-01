## MODIFIED Requirements

### Requirement: Cobra Command Tree

The CLI SHALL use [spf13/cobra](https://github.com/spf13/cobra) for command routing, flag parsing, and help generation.

#### Scenario: Command hierarchy

- GIVEN Cobra is the CLI framework
- WHEN commands are registered after migration is complete
- THEN the command tree MUST include:
  ```
  hf
  ├── env                          # create | list | show | activate | delete
  ├── resource | rs                # overview when invoked with no subcommand
  │   ├── types
  │   └── <entity>                 # from resource-types (includes clusters, nodepools)
  │       ├── list
  │       ├── table
  │       ├── get
  │       ├── create
  │       ├── search
  │       ├── patch
  │       ├── delete
  │       ├── force-delete         # nodepools; clusters use delete --force
  │       ├── conditions
  │       ├── statuses
  │       ├── adapter-report
  │       └── id
  ```
- AND dynamic `<entity>` groups MUST be registered after the active environment is loaded
- AND `hf config` MUST NOT be registered
- AND `hf cluster`, `hf nodepool`, `hf resources`, and `hf table` MUST NOT be registered

#### Scenario: Legacy commands removed from root help

- GIVEN this change is complete
- WHEN the user runs `hf --help`
- THEN `hf config`, `hf cluster`, and `hf nodepool` MUST NOT appear
- AND `hf table` and `hf resources` MUST NOT appear
- AND cluster/nodepool operations MUST appear only under `hf rs`
