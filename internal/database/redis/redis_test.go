package redis

import (
	"commander/internal/kv"
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, string) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	return mr, "redis://" + mr.Addr()
}

func TestNewRedisKV(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("Expected non-nil store")
	}
}

func TestNewRedisKV_EmptyURI(t *testing.T) {
	_, err := NewRedisKV("")
	if err == nil {
		t.Fatal("Expected error for empty URI")
	}

	expectedMsg := "Redis URI is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestNewRedisKV_InvalidURI(t *testing.T) {
	_, err := NewRedisKV("://invalid")
	if err == nil {
		t.Fatal("Expected error for invalid URI")
	}
}

func TestNewRedisKV_ConnectionFailed(t *testing.T) {
	// Use an invalid address
	_, err := NewRedisKV("redis://localhost:99999")
	if err == nil {
		t.Fatal("Expected connection error")
	}
}

func TestNewRedisKV_URIParsing(t *testing.T) {
	mr, _ := setupMiniredis(t)
	defer mr.Close()

	tests := []struct {
		name string
		uri  string
	}{
		{"simple", "redis://" + mr.Addr()},
		{"with db", "redis://" + mr.Addr() + "/0"},
		{"with different db", "redis://" + mr.Addr() + "/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewRedisKV(tt.uri)
			if err != nil {
				t.Fatalf("Failed to create Redis KV: %v", err)
			}
			defer store.Close()

			if store == nil {
				t.Fatal("Expected non-nil store")
			}
		})
	}
}

func TestRedisKV_SetAndGet(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	namespace := "testdb"
	collection := "users"
	key := "user1"
	value := []byte(`{"name":"John","age":30}`)

	// Set value
	err = store.Set(ctx, namespace, collection, key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Get value
	retrieved, err := store.Get(ctx, namespace, collection, key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected value %s, got %s", value, retrieved)
	}
}

func TestRedisKV_GetNonExistent(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Get non-existent key
	_, err = store.Get(ctx, "testdb", "users", "nonexistent")
	if err != kv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestRedisKV_Delete(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	namespace := "testdb"
	collection := "users"
	key := "user1"
	value := []byte(`{"name":"John"}`)

	// Set value
	err = store.Set(ctx, namespace, collection, key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Delete value
	err = store.Delete(ctx, namespace, collection, key)
	if err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	// Verify deletion
	_, err = store.Get(ctx, namespace, collection, key)
	if err != kv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound after deletion, got %v", err)
	}
}

func TestRedisKV_DeleteNonExistent(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Delete non-existent key
	err = store.Delete(ctx, "testdb", "users", "nonexistent")
	if err != kv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestRedisKV_Exists(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	namespace := "testdb"
	collection := "users"
	key := "user1"
	value := []byte(`{"name":"John"}`)

	// Check non-existent key
	exists, err := store.Exists(ctx, namespace, collection, key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Expected key to not exist")
	}

	// Set value
	err = store.Set(ctx, namespace, collection, key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Check existing key
	exists, err = store.Exists(ctx, namespace, collection, key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist")
	}
}

func TestRedisKV_BuildKey(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	tests := []struct {
		namespace  string
		collection string
		key        string
		expected   string
	}{
		{"myapp", "users", "user1", "myapp:users:user1"},
		{"", "posts", "post1", "default:posts:post1"}, // Empty namespace becomes "default"
		{"app", "cache", "key123", "app:cache:key123"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := store.buildKey(tt.namespace, tt.collection, tt.key)
			if result != tt.expected {
				t.Errorf("buildKey(%q, %q, %q) = %q, want %q",
					tt.namespace, tt.collection, tt.key, result, tt.expected)
			}
		})
	}
}

func TestRedisKV_NamespaceIsolation(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	key := "shared_key"
	value1 := []byte(`{"namespace":"ns1"}`)
	value2 := []byte(`{"namespace":"ns2"}`)

	// Set value in namespace1
	err = store.Set(ctx, "namespace1", "collection", key, value1)
	if err != nil {
		t.Fatalf("Failed to set value in namespace1: %v", err)
	}

	// Set value in namespace2
	err = store.Set(ctx, "namespace2", "collection", key, value2)
	if err != nil {
		t.Fatalf("Failed to set value in namespace2: %v", err)
	}

	// Get from namespace1
	retrieved1, err := store.Get(ctx, "namespace1", "collection", key)
	if err != nil {
		t.Fatalf("Failed to get value from namespace1: %v", err)
	}
	if string(retrieved1) != string(value1) {
		t.Errorf("Expected value %s, got %s", value1, retrieved1)
	}

	// Get from namespace2
	retrieved2, err := store.Get(ctx, "namespace2", "collection", key)
	if err != nil {
		t.Fatalf("Failed to get value from namespace2: %v", err)
	}
	if string(retrieved2) != string(value2) {
		t.Errorf("Expected value %s, got %s", value2, retrieved2)
	}
}

func TestRedisKV_CollectionIsolation(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	namespace := "testdb"
	key := "shared_key"
	value1 := []byte(`{"collection":"users"}`)
	value2 := []byte(`{"collection":"posts"}`)

	// Set value in collection1
	err = store.Set(ctx, namespace, "users", key, value1)
	if err != nil {
		t.Fatalf("Failed to set value in users: %v", err)
	}

	// Set value in collection2
	err = store.Set(ctx, namespace, "posts", key, value2)
	if err != nil {
		t.Fatalf("Failed to set value in posts: %v", err)
	}

	// Get from collection1
	retrieved1, err := store.Get(ctx, namespace, "users", key)
	if err != nil {
		t.Fatalf("Failed to get value from users: %v", err)
	}
	if string(retrieved1) != string(value1) {
		t.Errorf("Expected value %s, got %s", value1, retrieved1)
	}

	// Get from collection2
	retrieved2, err := store.Get(ctx, namespace, "posts", key)
	if err != nil {
		t.Fatalf("Failed to get value from posts: %v", err)
	}
	if string(retrieved2) != string(value2) {
		t.Errorf("Expected value %s, got %s", value2, retrieved2)
	}
}

func TestRedisKV_DefaultNamespace(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	collection := "users"
	key := "user1"
	value := []byte(`{"name":"John"}`)

	// Set with empty namespace
	err = store.Set(ctx, "", collection, key, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Get with explicit default namespace
	retrieved, err := store.Get(ctx, "default", collection, key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected value %s, got %s", value, retrieved)
	}
}

func TestRedisKV_Ping(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	err = store.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestRedisKV_Close(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}

	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestRedisKV_UpdateValue(t *testing.T) {
	mr, uri := setupMiniredis(t)
	defer mr.Close()

	store, err := NewRedisKV(uri)
	if err != nil {
		t.Fatalf("Failed to create Redis KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	namespace := "testdb"
	collection := "users"
	key := "user1"
	value1 := []byte(`{"name":"John","age":30}`)
	value2 := []byte(`{"name":"John","age":31}`)

	// Set initial value
	err = store.Set(ctx, namespace, collection, key, value1)
	if err != nil {
		t.Fatalf("Failed to set initial value: %v", err)
	}

	// Update value
	err = store.Set(ctx, namespace, collection, key, value2)
	if err != nil {
		t.Fatalf("Failed to update value: %v", err)
	}

	// Get updated value
	retrieved, err := store.Get(ctx, namespace, collection, key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrieved) != string(value2) {
		t.Errorf("Expected updated value %s, got %s", value2, retrieved)
	}
}
