// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package init

const SubstrateYAMLTemplate = `# substrate.yaml — Substrate configuration
# Docs: https://github.com/KrushnaVardhanReddy/Substrate

service: <service-name>
schema_type: <schema-type>
spec_path: <spec-path>

owners:
  - team: your-team-name
    contact: your-team@company.com

# Custom Rules
custom_rules:
  - id: REQUIRE_SPEC_SYNC
    description: "API handler changes must be accompanied by an openapi.yaml update."
    severity: error
    match: "commit.files.contains('api/internal/handlers/') && !commit.files.contains('openapi.yaml')"

# overrides:
#   - rule_id: REQUIRE_SPEC_SYNC
#     path: "api/internal/handlers/auth.go"
#     reason: "Refactored internal DB logic; API request/response contracts remain unchanged."
#     approved_by: "team-lead-handle"
#     expires: "2026-12-31"
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
