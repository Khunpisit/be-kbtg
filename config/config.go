package config

// Package config provides application configuration loading.

import (
	"os"
	"time"
)

// Environment variable names.
const (
	EnvDBPath    = "DB_PATH"
	EnvJWTSecret = "JWT_SECRET"
	EnvPort      = "PORT"
)

// Config holds runtime configuration.
type Config struct {
	DBPath       string
	JWTSecret    string
	JWTExpiresIn time.Duration
	Port         string
}

// Load reads configuration from environment with reasonable defaults.
func Load() Config {
	cfg := Config{
		DBPath:       getEnv(EnvDBPath, "app.db"),
		JWTSecret:    getEnv(EnvJWTSecret, "dev-secret-change"), // NOTE: override in production
		JWTExpiresIn: time.Hour,
		Port:         getEnv(EnvPort, "3000"),
	}
	return cfg
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
