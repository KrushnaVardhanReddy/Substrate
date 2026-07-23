# P9-T02: Continuous AI Sync (`watch`)

## Overview
A background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. Runs as `substrate watch` in the background.

## Requirements
1. **File Watcher**: Use `fsnotify` to watch Go/TypeScript source files for changes.
2. **Debounce**: Debounce rapid file changes with a 500ms window to avoid excessive re-runs.
3. **Spec Update**: On change detection, re-run the spec extraction logic and write the updated OpenAPI spec to the local file.
4. **Output**: Print a timestamped log line on every update, e.g. `[substrate] spec updated: 2026-07-20T10:00:00Z`.
5. **CLI**: Add `substrate watch` command to the CLI.

## Implementation Status: ✅ Implemented

### CLI Flags
The `substrate watch` command (`engine/cmd/substrate/watch.go`) supports the following flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `.` | Directory to watch for source file changes |
| `--spec` | `openapi.yaml` | Output path for the generated OpenAPI spec file |

### Watch Options (`engine/watcher/watcher.go`)
The `WatchOptions` struct includes an optional `ExtractFunc` callback:
```go
type WatchOptions struct {
    Dir         string
    SpecPath    string
    ExtractFunc func(outputPath string) error // Custom post-extraction hook
}
```
This allows callers to inject custom logic after extraction (e.g., renaming the output file).

### E2E Validation
- Covered by `TestPhase9SystemE2E/Scenario_3:_Watch_Daemon` in `scripts/e2e/phase9_e2e_test.go`.
- Test starts the watcher, triggers a file change, and asserts the spec is updated within 3 seconds.
