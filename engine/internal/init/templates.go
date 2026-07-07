package init

const SubstrateYAMLTemplate = `# substrate.yaml — Substrate configuration
# Docs: https://github.com/KrushnaVardhanReddy/Substrate

service: <service-name>
schema_type: <schema-type>
spec_path: <spec-path>

owners:
  - team: your-team-name
    contact: your-team@company.com

# Overrides — acknowledge intentional breaking changes
# overrides:
#   - rule_id: ENDPOINT_REMOVED
#     path: paths./your-endpoint
#     reason: "Reason this is safe (min 20 chars)"
#     approved_by: you@company.com
#     expires: YYYY-MM-DD
`

const WorkflowTemplate = `# Substrate API Contract Guard
# Docs: https://github.com/KrushnaVardhanReddy/Substrate

name: Substrate API Contract Guard

on:
  pull_request:
    branches: [<branch>]

jobs:
  check-api-contracts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Check API Breaking Changes
        uses: KrushnaVardhanReddy/Substrate@v0.1.0
        with:
          base_schema: <spec-path>
          head_schema: <spec-path>
          config: substrate.yaml
`
