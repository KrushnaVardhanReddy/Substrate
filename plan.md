1. Add `api/internal/config/config.go` with parsing logic for substrate.yaml:
   ```go
   package config
   import "gopkg.in/yaml.v3"
   type Discovery struct { MatchPatterns []string `yaml:"match_patterns"` }
   type SubstrateConfig struct { Discovery *Discovery `yaml:"discovery,omitempty"` }
   func Parse(content []byte) (*SubstrateConfig, error) { ... }
   ```
2. Refactor `api/internal/discovery/env_scanner.go` to accept match patterns dynamically:
   - Add `type EnvScanner struct { regex *regexp.Regexp }`
   - Add `func NewEnvScanner(patterns []string) *EnvScanner`
   - Merge provided patterns with `highSignalRegex` if they are provided, else fallback to `highSignalRegex`
   - Update `ScanEnvFile`, `ScanDockerCompose`, `ScanKubernetesManifest` to be methods on `EnvScanner`
3. Update `api/internal/webhook/handler.go` to inject the regex correctly:
   - Search through files in PushHandler to find `substrate.yaml` file
   - If found, parse it using `config.Parse`
   - Extract patterns from the parsed config. If not found, create `NewEnvScanner(nil)`
   - Call `scanner.ScanKubernetesManifest` instead of `discovery.ScanKubernetesManifest`
4. Write tests for `api/internal/discovery/env_scanner_test.go` checking custom regex implementation:
   - `TestEnvScanner_CustomPatterns` passing a custom regex like `.*_ENDPOINT$` and verifying that `PAYMENT_ENDPOINT` is discovered.
5. Complete pre commit step.
