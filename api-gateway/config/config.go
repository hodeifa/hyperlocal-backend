package config

import (
	"os"
)

// Config menyimpan semua konfigurasi API Gateway.
type Config struct {
	Env       string // "development", "staging", "production"
	Port      string // default "8080"
	JWTSecret string // dari env JWT_SECRET
}

// Load membaca konfigurasi dari environment variables.
func Load() Config {
	return Config{
		Env:       getEnv("APP_ENV", "development"),
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}