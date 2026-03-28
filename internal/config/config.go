package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort   string
	DatabasePath string
	JWTSecret    string
	JWTExpiry    time.Duration
	Environment  string
}

func Load() *Config {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./inventory.db"),
		JWTSecret:    getEnv("JWT_SECRET", "inventory-secret-key-change-in-production"),
		JWTExpiry:    24 * time.Hour,
		Environment:  getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
