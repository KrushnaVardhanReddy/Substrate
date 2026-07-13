# V1-T08: Custom Discovery Rules via YAML

## Objective
The current dependency discovery engine hardcodes a `highSignalRegex` that only flags Kubernetes environment variables ending in `_API_URL`, `_SERVICE_URL`, etc. Enterprise users have varying naming conventions (e.g., `_ENDPOINT`, `_HOST`). This task exposes the dependency matching rules to the user via the `substrate.yaml` configuration file, making Substrate flexible enough to fit any organization's naming standards.

## Requirements

### 1. Update `substrate.yaml` Parser
- Update the `SubstrateConfig` struct (likely in `api/internal/config/` or wherever `substrate.yaml` is parsed).
- Add a new `discovery` block that accepts `match_patterns`:
  ```yaml
  discovery:
    match_patterns:
      - '.*_ENDPOINT$'
      - '.*_HOST$'
  ```

### 2. Update the Environment Scanner
- Target: `api/internal/discovery/env_scanner.go`
- Currently, it uses a hardcoded regex (e.g., `(?i)(API_URL|SERVICE_URL)$`).
- Refactor the scanner to accept the parsed `substrate.yaml` config (or the extracted `match_patterns`).
- If `match_patterns` are provided in the config, the scanner MUST compile and use them alongside (or instead of) the default regex to identify high-signal dependencies.

### 3. Graceful Fallback
- If the user does not provide a `discovery.match_patterns` array in `substrate.yaml`, the scanner MUST default to the existing `_API_URL` / `_SERVICE_URL` behavior so we don't break existing integrations.

### 4. Tests
- Add a unit test to `env_scanner_test.go` that explicitly passes a custom regex like `.*_ENDPOINT$` and verifies that an environment variable like `PAYMENT_ENDPOINT` is successfully extracted as a dependency edge.

### 5. E2E Validation
- Update `scripts/e2e/scale_generator.go` to explicitly include a `substrate.yaml` file in the generated `PushPayload` with `discovery.match_patterns` defined (e.g., `.*_CUSTOM_ENDPOINT$`).
- Modify the `getMockContent` function to output `_CUSTOM_ENDPOINT` instead of the legacy `_API_URL` to prove that the webhook handler correctly parses the incoming config and instantiates the `env_scanner` with the custom overrides.
- Fix the JSON unmarshalling in `runAssertionAndReporting` to correctly expect an array of `DependencyEdge` structs, rather than an object containing `nodes` and `edges`, to ensure accurate E2E reporting.
