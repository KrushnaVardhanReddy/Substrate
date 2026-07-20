# P9-T02: Continuous AI Sync (`watch`)

## Overview
A background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. Runs as `substrate watch` in the background.

## Requirements
1. **File Watcher**: Use `fsnotify` to watch Go/TypeScript source files for changes.
2. **Debounce**: Debounce rapid file changes with a 500ms window to avoid excessive re-runs.
3. **Spec Update**: On change detection, re-run the spec extraction logic and write the updated OpenAPI spec to the local file.
4. **Output**: Print a timestamped log line on every update, e.g. `[substrate] spec updated: 2026-07-20T10:00:00Z`.
5. **CLI**: Add `substrate watch` command to the CLI.
