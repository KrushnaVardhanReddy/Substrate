# P15-T05: AI Incident Post-Mortem Generator

## Overview
`substrate postmortem --incident <date>` correlates the incident window with schema changes, lists every breaking change and blast radius, and estimates incident cost. Outputs a ready-to-share Markdown document.

## Requirements
1. **CLI Command**: Add `postmortem` command to the Substrate CLI.
2. **Database Query**: Query the PostgreSQL backend for all breaking changes merged within 24 hours prior to the incident date.
3. **LLM Generation**: Use the AI engine to generate a Blameless Post-Mortem document outlining the root cause (the breaking schema change), the blast radius (downstream consumers affected), and recommended remediation.
4. **Output Format**: Ensure the output is formatted as Markdown and easily exportable to Notion/Confluence.
