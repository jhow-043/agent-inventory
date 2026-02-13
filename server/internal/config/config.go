// Package config loads and validates application configuration from environment variables.
package config

import (
	"log/slog"
	"os"
	"strings"
)

// Config holds all application configuration values.
type Config struct {
	DatabaseURL   string
	ServerPort    string
	LogLevel      slog.Level
	JWTSecret     string
	EnrollmentKey string
	CORSOrigins   []string
}

// Load reads configuration from environment variables and validates required fields.
func Load() *Config {
	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://inventory:changeme@localhost:5432/inventory?sslmode=disable"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		EnrollmentKey: getEnv("ENROLLMENT_KEY", ""),
		CORSOrigins:   strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
	}

	switch strings.ToLower(getEnv("LOG_LEVEL", "info")) {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		cfg.LogLevel = slog.LevelInfo
	}

	if cfg.JWTSecret == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	if len(cfg.JWTSecret) < 32 {
		slog.Warn("JWT_SECRET should be at least 32 characters for adequate security")
	}
	if cfg.EnrollmentKey == "" {
		slog.Error("ENROLLMENT_KEY environment variable is required")
		os.Exit(1)
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
