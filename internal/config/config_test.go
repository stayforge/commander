package config

import (
	"os"
	"testing"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	cfg := LoadConfig()

	if cfg.Version != "dev" {
		t.Errorf("Expected version 'dev', got '%s'", cfg.Version)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected port '8080', got '%s'", cfg.Server.Port)
	}

	if cfg.Server.Environment != "STANDARD" {
		t.Errorf("Expected environment 'STANDARD', got '%s'", cfg.Server.Environment)
	}

	if cfg.KV.BackendType != BackendBBolt {
		t.Errorf("Expected backend type 'bbolt', got '%s'", cfg.KV.BackendType)
	}

	if cfg.KV.BBoltPath != "/var/lib/stayforge/commander" {
		t.Errorf("Expected BBolt path '/var/lib/stayforge/commander', got '%s'", cfg.KV.BBoltPath)
	}
}

func TestLoadConfig_MongoDB(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE", "mongodb")
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("ENVIRONMENT", "TESTING")

	cfg := LoadConfig()

	if cfg.KV.BackendType != BackendMongoDB {
		t.Errorf("Expected backend type 'mongodb', got '%s'", cfg.KV.BackendType)
	}

	if cfg.KV.MongoURI != "mongodb://localhost:27017" {
		t.Errorf("Expected MongoDB URI 'mongodb://localhost:27017', got '%s'", cfg.KV.MongoURI)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected port '9090', got '%s'", cfg.Server.Port)
	}

	if cfg.Server.Environment != "TESTING" {
		t.Errorf("Expected environment 'TESTING', got '%s'", cfg.Server.Environment)
	}
}

func TestLoadConfig_Redis(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE", "redis")
	os.Setenv("REDIS_URI", "redis://localhost:6379")

	cfg := LoadConfig()

	if cfg.KV.BackendType != BackendRedis {
		t.Errorf("Expected backend type 'redis', got '%s'", cfg.KV.BackendType)
	}

	if cfg.KV.RedisURI != "redis://localhost:6379" {
		t.Errorf("Expected Redis URI 'redis://localhost:6379', got '%s'", cfg.KV.RedisURI)
	}
}

func TestLoadConfig_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected BackendType
	}{
		{"lowercase", "mongodb", BackendMongoDB},
		{"uppercase", "MONGODB", BackendMongoDB},
		{"mixed case", "MongoDb", BackendMongoDB},
		{"redis lowercase", "redis", BackendRedis},
		{"redis uppercase", "REDIS", BackendRedis},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("DATABASE", tt.input)

			cfg := LoadConfig()

			if cfg.KV.BackendType != tt.expected {
				t.Errorf("Expected backend type '%s', got '%s'", tt.expected, cfg.KV.BackendType)
			}
		})
	}
}

func TestLoadConfig_UnknownBackend(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE", "unknown")

	cfg := LoadConfig()

	// Should default to bbolt
	if cfg.KV.BackendType != BackendBBolt {
		t.Errorf("Expected backend type 'bbolt' for unknown backend, got '%s'", cfg.KV.BackendType)
	}
}

func TestGetEnv(t *testing.T) {
	os.Clearenv()

	// Test with set value
	os.Setenv("TEST_KEY", "test_value")
	if got := getEnv("TEST_KEY", "default"); got != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", got)
	}

	// Test with default value
	if got := getEnv("UNSET_KEY", "default"); got != "default" {
		t.Errorf("Expected 'default', got '%s'", got)
	}
}
