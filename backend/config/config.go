package config

import (
	"os"
)

type Config struct {
	Port          string
	DBPath        string
	JWTSecret     string
	LLMAPIKey     string
	LLMBaseURL    string
	LLMModel      string
	MaxHearts     int
	XPPerLesson   int
	XPPerExercise int
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "codelearn.db"),
		JWTSecret:     getEnv("JWT_SECRET", "codelearn-dev-secret-change-me"),
		LLMAPIKey:     getEnv("LLM_API_KEY", ""),
		LLMBaseURL:    getEnv("LLM_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
		LLMModel:      getEnv("LLM_MODEL", "doubao-1-5-pro-32k-250115"),
		MaxHearts:     5,
		XPPerLesson:   20,
		XPPerExercise: 10,
	}
}

func (c *Config) LLMEnabled() bool {
	return c.LLMAPIKey != ""
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
