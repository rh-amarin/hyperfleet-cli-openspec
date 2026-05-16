## ADDED Requirements

### Requirement: Config template namespace placement

The bundled config template SHALL place the HyperFleet application namespace under the `hyperfleet` section, not the `kubernetes` section.

#### Scenario: Template hyperfleet section includes namespace

- **WHEN** a new environment is created from the template
- **THEN** the `hyperfleet:` section MUST contain a `namespace` key with default value `"hyperfleet"`
- **AND** the `kubernetes:` section MUST NOT contain a `namespace` key
