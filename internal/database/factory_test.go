package database

import (
	"commander/internal/config"
	"strings"
	"testing"
)

func TestNewKV_BBolt(t *testing.T) {
	cfg := &config.Config{
		KV: config.KVConfig{
			BackendType: config.BackendBBolt,
			BBoltPath:   t.TempDir(), // Use temp directory for testing
		},
	}

	kv, err := NewKV(cfg)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}

	if kv == nil {
		t.Fatal("Expected non-nil KV instance")
	}

	// Clean up
	if err := kv.Close(); err != nil {
		t.Errorf("Failed to close KV: %v", err)
	}
}

func TestNewKV_MongoDB_MissingURI(t *testing.T) {
	cfg := &config.Config{
		KV: config.KVConfig{
			BackendType: config.BackendMongoDB,
			MongoURI:    "", // Empty URI
		},
	}

	kv, err := NewKV(cfg)
	if err == nil {
		t.Fatal("Expected error for missing MongoDB URI, got nil")
	}

	if kv != nil {
		t.Error("Expected nil KV instance when error occurs")
	}

	expectedMsg := "MongoDB URI is required"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewKV_Redis_MissingURI(t *testing.T) {
	cfg := &config.Config{
		KV: config.KVConfig{
			BackendType: config.BackendRedis,
			RedisURI:    "", // Empty URI
		},
	}

	kv, err := NewKV(cfg)
	if err == nil {
		t.Fatal("Expected error for missing Redis URI, got nil")
	}

	if kv != nil {
		t.Error("Expected nil KV instance when error occurs")
	}

	expectedMsg := "Redis URI is required"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewKV_UnsupportedBackend(t *testing.T) {
	cfg := &config.Config{
		KV: config.KVConfig{
			BackendType: "unsupported", // Invalid backend type
		},
	}

	kv, err := NewKV(cfg)
	if err == nil {
		t.Fatal("Expected error for unsupported backend, got nil")
	}

	if kv != nil {
		t.Error("Expected nil KV instance when error occurs")
	}

	expectedMsg := "unsupported backend type"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}
