package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Clear env vars for clean test
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")

	// Test missing required fields
	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing MONGODB_URI, got nil")
	}

	// Test with required fields
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("MONGODB_DATABASE", "testdb")
	defer os.Unsetenv("MONGODB_URI")
	defer os.Unsetenv("MONGODB_DATABASE")

	cfg, err := Load()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if cfg.MongoDBURI != "mongodb://localhost:27017" {
		t.Errorf("Expected MongoDBURI 'mongodb://localhost:27017', got '%s'", cfg.MongoDBURI)
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("Expected default ServerPort '8080', got '%s'", cfg.ServerPort)
	}
}
