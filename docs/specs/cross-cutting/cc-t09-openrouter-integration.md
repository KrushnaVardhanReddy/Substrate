# CC-T09: LLM Aggregator Integration (OpenRouter)

## Overview
As Substrate introduces more AI-native features (like Schema Smell Detection, Copilot, and Agent Governance), the cost of running LLM inference during local development and E2E CI/CD testing can quickly escalate.

This task integrates an LLM Aggregator (OpenRouter) into the Substrate backend and GitHub App via the standard OpenAI SDK format, allowing developers to configure permanently free, open-weights models (e.g., `google/gemma-3-27b-it:free` or `meta-llama/llama-3.3-70b-instruct:free`) for non-production environments.

## Technical Requirements

### 1. Configuration (Environment Variables)
Add support for overriding the default LLM base URL and Model ID in `api/internal/config/config.go` and `.env.example`:
```bash
# LLM Configuration
LLM_PROVIDER=openrouter  # openrouter, openai, anthropic
LLM_API_KEY=sk-or-v1-...
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_MODEL=google/gemma-3-27b-it:free
```

### 2. Go Backend Integration (`github.com/sashabaranov/go-openai`)
In `api/internal/ai/client.go` (or wherever the AI client is instantiated), construct the OpenAI client using the custom `BaseURL` if `LLM_PROVIDER=openrouter`:
```go
config := openai.DefaultConfig(cfg.LLM_API_KEY)
if cfg.LLM_BASE_URL != "" {
    config.BaseURL = cfg.LLM_BASE_URL
}
client := openai.NewClientWithConfig(config)
```
Ensure all AI features (e.g., Postmortem Generator, Smell Detector) use `cfg.LLM_MODEL` dynamically rather than hardcoding `"gpt-4o"`.

### 3. CI/CD E2E Suite Updates
Update the GitHub Actions workflow (`.github/workflows/e2e.yml` or similar) to inject an `OPENROUTER_API_KEY` GitHub Secret into the E2E test environments. Ensure that the E2E testing suite runs exclusively on free OpenRouter models to prevent draining the OpenAI balance during daily PR runs.

## Deliverables
1. Update `config.go` and `.env.example` with LLM endpoint overrides.
2. Refactor existing Go AI clients to accept the dynamic `BaseURL` and `Model`.
3. Update GitHub Actions workflows to use the OpenRouter key for E2E testing.
4. Provide a quick CLI test command (e.g., `substrate ai test`) to verify the OpenRouter connection.
