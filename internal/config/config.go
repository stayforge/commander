package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Version string
	Server  ServerConfig
	KV      KVConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string
	Environment string
}

// KVConfig holds key-value storage configuration
type KVConfig struct {
	BackendType BackendType

	// MongoDB URI
	MongoURI string

	// Redis URI
	RedisURI string

	// BBolt path
	BBoltPath string
}

// BackendType represents the type of KV backend
type BackendType string

// Available KV backend types
const (
	BackendMongoDB BackendType = "mongodb"
	BackendRedis   BackendType = "redis"
	BackendBBolt   BackendType = "bbolt"
)

// LoadConfig constructs a Config populated from environment variables.
// It selects the KV backend from DATABASE (case-insensitive): "mongodb", "redis", or "bbolt" (default).
// Server port defaults to "8080" and environment to "STANDARD". MongoDB and Redis URIs are taken
// from MONGODB_URI and REDIS_URI (empty by default). BBolt data path defaults to "/var/lib/stayforge/commander".
func LoadConfig() *Config {
	// Get DATABASE type (case-insensitive), default to bbolt
	databaseType := strings.ToLower(getEnv("DATABASE", "bbolt"))

	var backendType BackendType
	switch databaseType {
	case "mongodb":
		backendType = BackendMongoDB
	case "redis":
		backendType = BackendRedis
	case "bbolt":
		backendType = BackendBBolt
	default:
		// Default to bbolt if unknown type
		backendType = BackendBBolt
	}

	return &Config{
		Version: "dev",
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "STANDARD"),
		},
		KV: KVConfig{
			BackendType: backendType,

			// MongoDB URI
			MongoURI: getEnv("MONGODB_URI", ""),

			// Redis URI
			RedisURI: getEnv("REDIS_URI", ""),

			// BBolt path (default: /var/lib/stayforge/commander)
			BBoltPath: getEnv("DATA_PATH", "/var/lib/stayforge/commander"),
		},
	}
}

// getEnv returns the value of the environment variable named by key, or defaultValue if the variable is unset or empty.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}