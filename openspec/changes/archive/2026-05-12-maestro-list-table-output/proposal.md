# Proposal: maestro-list-table-output

## What

Add `--output table` support to `hf maestro list` that renders a hierarchical, human-readable tree of resource bundles and their manifests.

Each bundle appears on one line showing its `id`, `name`, and `version`. Its manifests are printed on indented child lines, each showing `kind`, `name`, and `namespace`.

Example output:

```
b1d2c3e4-0000-0000-0000-000000000001  mw-cluster1  v3
  Deployment/my-app  default
  ConfigMap/my-cfg   default
```

## Why

The JSON default for `hf maestro list` is complete but not quickly scannable. Operators want a glanceable view of which bundles are deployed to the consumer and what Kubernetes resources each one manages — without parsing nested JSON. The table format gives that at a glance.

The existing `--output` flag already accepts `table` as a value (it is the project-wide convention per CLAUDE.md: "All data-producing commands support `--output json|table|yaml`"). Currently passing `--output table` to `hf maestro list` silently falls through to JSON. This change makes it render the hierarchical view instead.
