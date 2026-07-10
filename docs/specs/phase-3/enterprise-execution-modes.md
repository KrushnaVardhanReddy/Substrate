# Enterprise Execution Modes (P3-T13)

## Objective
Provide enterprise teams with a way to adopt Substrate progressively without instantly blocking all their CI pipelines on day one. We will introduce a `mode` configuration (CLI flag and `substrate.yaml`) with two settings: `strict` and `legacy`.

## The Modes

1. **`strict` (Default)**
   - The standard Substrate behavior.
   - Analyzes schemas, and if any unacknowledged breaking changes are found, exits with Code `2`.
   - CI pipeline fails, blocking the merge.

2. **`legacy` (Shadow Mode / Dry Run)**
   - Runs the exact same diff engine and reports all breaking changes.
   - **Crucially:** Always exits with Code `0` (Success), even if breaking changes are detected.
   - Outputs a warning banner: `⚠️ LEGACY MODE: Breaking changes detected, but merge is not blocked.`
   - This allows large orgs to install Substrate globally, gather analytics/visibility on breaking changes, but not abruptly block developers until they are ready.

## Configuration

The mode can be set in two ways (CLI overrides file):

1. **Config File (`substrate.yaml`)**
   ```yaml
   mode: legacy  # default is strict
   ```

2. **CLI Flag**
   ```bash
   substrate diff --base=base.yaml --head=head.yaml --mode=legacy
   ```

## Implementation Requirements

1. **Config Parsing:**
   - Update `engine/internal/config/config.go` to support `Mode string \`yaml:"mode"\``.
   - Default to `strict` if empty or invalid.

2. **CLI Flag:**
   - Add `--mode` flag to `engine/cmd/substrate/main.go`.
   - Ensure CLI flag takes precedence over `substrate.yaml`.

3. **Exit Code Logic:**
   - In `main.go`, after the diff report is generated, if `mode == "legacy"`, force the exit code to `0` instead of `2`.
   - Ensure stdout still prints the breaking changes, but prepends/appends a clear warning that it's running in legacy mode.

4. **HTTP Server Mode (`/diff`)**
   - Update `engine/internal/server/server.go` (if applicable) or the diff JSON output to include a field `"mode": "legacy"` so that the Cloudflare Worker PR formatter knows to style the GitHub comment differently (e.g., green checkmark on the PR status, but warning in the comment body).
