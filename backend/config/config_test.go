package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DB_PATH")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("LLM_API_KEY")
	os.Unsetenv("LLM_BASE_URL")
	os.Unsetenv("LLM_MODEL")

	cfg := Load()

	assertEquals(t, "8080", cfg.Port)
	assertEquals(t, "codelearn.db", cfg.DBPath)
	assertEquals(t, "codelearn-dev-secret-change-me", cfg.JWTSecret)
	assertEquals(t, "", cfg.LLMAPIKey)
	assertEquals(t, "https://ark.cn-beijing.volces.com/api/v3", cfg.LLMBaseURL)
	assertEquals(t, "doubao-1-5-pro-32k-250115", cfg.LLMModel)
	assertEquals(t, 5, cfg.MaxHearts)
	assertEquals(t, 20, cfg.XPPerLesson)
	assertEquals(t, 10, cfg.XPPerExercise)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("JWT_SECRET", "my-secret")
	t.Setenv("LLM_API_KEY", "sk-test-key")
	t.Setenv("LLM_BASE_URL", "https://api.openai.com/v1")
	t.Setenv("LLM_MODEL", "gpt-4")

	cfg := Load()

	assertEquals(t, "3000", cfg.Port)
	assertEquals(t, "/tmp/test.db", cfg.DBPath)
	assertEquals(t, "my-secret", cfg.JWTSecret)
	assertEquals(t, "sk-test-key", cfg.LLMAPIKey)
	assertEquals(t, "https://api.openai.com/v1", cfg.LLMBaseURL)
	assertEquals(t, "gpt-4", cfg.LLMModel)
}

func TestLLMEnabled(t *testing.T) {
	t.Run("disabled when no key", func(t *testing.T) {
		t.Setenv("LLM_API_KEY", "")
		cfg := Load()
		if cfg.LLMEnabled() {
			t.Error("expected LLM disabled")
		}
	})
	t.Run("enabled when key set", func(t *testing.T) {
		t.Setenv("LLM_API_KEY", "sk-test")
		cfg := Load()
		if !cfg.LLMEnabled() {
			t.Error("expected LLM enabled")
		}
	})
}

func TestGetEnvFallback(t *testing.T) {
	assertEquals(t, "default", getEnv("NONEXISTENT_VAR_12345", "default"))
	t.Setenv("TEST_ENV_VAR", "custom")
	assertEquals(t, "custom", getEnv("TEST_ENV_VAR", "default"))
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
