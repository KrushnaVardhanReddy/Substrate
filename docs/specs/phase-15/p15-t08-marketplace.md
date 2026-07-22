# P15-T08: Substrate Marketplace (Community Rules & Plugins)

## Overview
A community marketplace for governance rule packs (e.g., `substrate-plugin-hipaa`, `substrate-plugin-pci`, `substrate-plugin-owasp`). Network effects compound as every contributed rule pack increases value for all organizations.

## Requirements
1. **Plugin System Architecture**: Define a standard JSON schema for rule packs.
2. **CLI Integration**: Add `substrate plugin publish` to package and upload a rule pack, and `substrate plugin install <name>` to fetch and apply it to a local `substrate.yaml`.
3. **API Endpoints**: Create `POST /api/marketplace/publish` and `GET /api/marketplace/plugins`.
4. **Dashboard View**: Add a "Marketplace" tab in the dashboard listing community plugins with "Install" buttons.
