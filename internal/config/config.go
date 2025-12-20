package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	MongoDBURI        string
	MongoDBDatabase   string
	MongoDBCollection string
	ServerPort        string
	Environment       string
}

// Load reads configuration from environment variables
// It will attempt to load .env file if it exists, but won't fail if it doesn't
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	config := &Config{
		MongoDBURI:      os.Getenv("MONGODB_URI"),
		MongoDBDatabase: os.Getenv("MONGODB_DATABASE"),
		ServerPort:      os.Getenv("SERVER_PORT"),
		Environment:     os.Getenv("ENVIRONMENT"),
	}

	// Set defaults
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	if config.Environment == "" {
		config.Environment = "STANDARD"
	}

	// Validate required fields
	if config.MongoDBURI == "" {
		return nil, fmt.Errorf("MONGODB_URI is required")
	}
	if config.MongoDBDatabase == "" {
		return nil, fmt.Errorf("MONGODB_DATABASE is required")
	}

	return config, nil
}
