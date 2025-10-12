package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	Addr         string
	Model        string
	Language     string
	ModelBaseURL string
}

// FromEnv loads the configuration from environment variables.
func FromEnv() *Config {
	cfg := &Config{
		Addr:         ":8080",
		Model:        "base.en",
		Language:     "en",
		ModelBaseURL: "",
	}

	if p := os.Getenv("PORT"); p != "" {
		if _, err := strconv.Atoi(p); err == nil {
			cfg.Addr = ":" + p
		}
	}
	if m := os.Getenv("GOSPER_MODEL"); m != "" {
		cfg.Model = m
	}
	if l := os.Getenv("GOSPER_LANG"); l != "" {
		cfg.Language = l
	}
	if u := os.Getenv("MODEL_BASE_URL"); u != "" {
		cfg.ModelBaseURL = u
	}

	return cfg
}
