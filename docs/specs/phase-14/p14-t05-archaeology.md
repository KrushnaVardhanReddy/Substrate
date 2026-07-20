# P14-T05: Retroactive Dependency Archaeology (Paid Onboarding Service)

## Overview
`substrate archaeology --since 2-years` scans the full git history of all connected repos and generates a paid audit report showing every historical breaking change and its estimated incident cost.

## Requirements
1. **Git History Analyzer**: Implement a script or Go routine that iterates over git tags/commits and runs a historical AST diff on the OpenAPI spec.
2. **Cost Estimation Model**: Apply the FinOps estimation logic (from P13-T01) to these historical breaking changes.
3. **CLI Command**: Add the `archaeology` command to the Substrate CLI.
4. **Report Generation**: Output a detailed PDF or Markdown report suitable for C-level executives highlighting "Hidden API Debt".
