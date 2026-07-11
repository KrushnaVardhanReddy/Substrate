---
title: AI & Shift-Left
description: AI intelligence and Shift-Left VSCode Extension
---

Substrate is powered by a local, on-premise AI intelligence layer designed to keep your code private while providing cutting-edge remediation.

### AI PR Remediation
When Substrate blocks your PR, it doesn't just leave you guessing. The AI agent analyzes the exact rule violation and your schema context to generate an **AI Impact Analysis**.
It will suggest a **Safe Remediation** code block (e.g., "Add `@deprecated: true` instead of deleting the field") that you can copy and paste to instantly resolve the CI failure.

### The AI Playground
The Substrate Dashboard features an interactive AI Playground.
Paste your current schema and proposed schema, click "Analyze with Substrate AI", and watch the Server-Sent Events (SSE) stream print out the AI's thought process, the breaking findings, and the auto-generated code fix in real-time.

### Shift-Left VSCode Extension
Don't wait for CI. Install the Substrate VSCode Extension.
When you delete a required field in your `openapi.yaml` and hit save, the extension instantly places a red squiggly line under the code and displays a tooltip warning you of the blast radius before you even type `git commit`.