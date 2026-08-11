// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadLLMConfig(t *testing.T) {
	// Clear environment variables to test fallbacks
	clearEnv := func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("SUBSTRATE_AI_PROVIDER")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("SUBSTRATE_AI_API_KEY")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
		os.Unsetenv("SUBSTRATE_AI_BASE_URL")
		os.Unsetenv("LLM_MODEL")
		os.Unsetenv("SUBSTRATE_AI_MODEL")
	}

	t.Run("default values", func(t *testing.T) {
		clearEnv()
		cfg := LoadLLMConfig()
		assert.Equal(t, "openai", cfg.Provider)
		assert.Equal(t, "gpt-4o", cfg.Model)
		assert.Equal(t, "", cfg.APIKey)
		assert.Equal(t, "", cfg.BaseURL)
	})

	t.Run("primary LLM variables", func(t *testing.T) {
		clearEnv()
		os.Setenv("LLM_PROVIDER", "openrouter")
		os.Setenv("LLM_API_KEY", "sk-or-v1-xxx")
		os.Setenv("LLM_BASE_URL", "https://openrouter.ai/api/v1")
		os.Setenv("LLM_MODEL", "google/gemma-3-27b-it:free")

		cfg := LoadLLMConfig()
		assert.Equal(t, "openrouter", cfg.Provider)
		assert.Equal(t, "sk-or-v1-xxx", cfg.APIKey)
		assert.Equal(t, "https://openrouter.ai/api/v1", cfg.BaseURL)
		assert.Equal(t, "google/gemma-3-27b-it:free", cfg.Model)
	})

	t.Run("fallback to SUBSTRATE_AI variables", func(t *testing.T) {
		clearEnv()
		os.Setenv("SUBSTRATE_AI_PROVIDER", "azure")
		os.Setenv("SUBSTRATE_AI_API_KEY", "azure-key")
		os.Setenv("SUBSTRATE_AI_BASE_URL", "https://azure.com")
		os.Setenv("SUBSTRATE_AI_MODEL", "azure-gpt")

		cfg := LoadLLMConfig()
		assert.Equal(t, "azure", cfg.Provider)
		assert.Equal(t, "azure-key", cfg.APIKey)
		assert.Equal(t, "https://azure.com", cfg.BaseURL)
		assert.Equal(t, "azure-gpt", cfg.Model)
	})

	t.Run("fallback to OPENAI_API_KEY", func(t *testing.T) {
		clearEnv()
		os.Setenv("OPENAI_API_KEY", "openai-key")

		cfg := LoadLLMConfig()
		assert.Equal(t, "openai-key", cfg.APIKey)
	})
}
