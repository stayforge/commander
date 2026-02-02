package mongodb

import (
	"commander/internal/kv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test NewMongoDBKV with invalid connection
func TestNewMongoDBKV_ConnectionFailed(t *testing.T) {
	// Test with invalid URI - should fail to connect
	_, err := NewMongoDBKV("mongodb://invalid-host:99999")
	assert.Error(t, err)
	assert.ErrorIs(t, err, kv.ErrConnectionFailed)
}

// Test NewMongoDBKV with malformed URI
func TestNewMongoDBKV_InvalidURI(t *testing.T) {
	// Test with malformed URI
	_, err := NewMongoDBKV("://invalid")
	assert.Error(t, err)
}

// Test NewMongoDBKV with empty URI
func TestNewMongoDBKV_EmptyURI(t *testing.T) {
	_, err := NewMongoDBKV("")
	assert.Error(t, err)
}

// Note: Full integration tests for MongoDB CRUD operations should be run
// with a real MongoDB instance or testcontainers. The current implementation
// focuses on testing connection errors and URI validation.
//
// For production use, consider adding integration tests with:
// - testcontainers-go for spinning up real MongoDB instances
// - or a dedicated test MongoDB server
//
// These tests would cover:
// - Get/Set/Delete/Exists operations
// - Namespace and collection isolation
// - Default namespace handling
// - Ping and Close operations
