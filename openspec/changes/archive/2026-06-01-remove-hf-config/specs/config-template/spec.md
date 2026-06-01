## MODIFIED Requirements

### Requirement: Bundled environment template

The CLI SHALL ship a default environment template embedded in the binary. The template serves two purposes: it seeds new environment files created by `hf env create`, and it provides the built-in default values for the `internal/config` package.

#### Scenario: Template file location and embedding

- **WHEN** the binary is compiled
- **THEN** the file at `internal/config/assets/config-template.yaml` MUST be embedded using `//go:embed` in the `internal/config` package
- **AND** the embedded bytes MUST be exported as `config.ConfigTemplateYAML []byte`
- **AND** `hf env create` MUST use `config.ConfigTemplateYAML` when seeding new environment files
