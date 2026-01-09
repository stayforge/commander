package kv

import (
	"context"
	"errors"
)

var (
	// ErrKeyNotFound is returned when a key does not exist
	ErrKeyNotFound = errors.New("key not found")
	// ErrConnectionFailed is returned when connection to backend fails
	ErrConnectionFailed = errors.New("connection failed")

	// DefaultNamespace is the default namespace used when namespace is empty
	DefaultNamespace = "default"
)

// NormalizeNamespace returns the namespace, or "default" if empty
func NormalizeNamespace(namespace string) string {
	if namespace == "" {
		return DefaultNamespace
	}
	return namespace
}

// KV is the interface for key-value storage backends
// Key is string, Value is JSON bytes
// Supports namespace and collection for data organization
type KV interface {
	// Get retrieves a JSON value by key from namespace and collection
	Get(ctx context.Context, namespace, collection, key string) ([]byte, error)

	// Set stores a JSON value by key in namespace and collection
	Set(ctx context.Context, namespace, collection, key string, value []byte) error

	// Delete removes a key-value pair from namespace and collection
	Delete(ctx context.Context, namespace, collection, key string) error

	// Exists checks if a key exists in namespace and collection
	Exists(ctx context.Context, namespace, collection, key string) (bool, error)

	// Close closes the connection to the backend
	Close() error

	// Ping checks if the connection is alive
	Ping(ctx context.Context) error
}
