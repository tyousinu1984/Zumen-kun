package config

import (
	"log"
	"os"
)

type Config struct {
	SupabaseURL string
	SupabaseKey string
	RedisURL    string
	Port        string
	Env         string
}

func Load() *Config {
	cfg := &Config{
		SupabaseURL: getEnv("SUPABASE_URL", ""),
		SupabaseKey: getEnv("SUPABASE_KEY", ""),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:        getEnv("GO_PORT", "8080"),
		Env:         getEnv("GO_ENV", "development"),
	}

	// Validate required fields
	if cfg.SupabaseURL == "" {
		log.Println("WARNING: SUPABASE_URL not set")
	}
	if cfg.SupabaseKey == "" {
		log.Println("WARNING: SUPABASE_KEY not set")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
