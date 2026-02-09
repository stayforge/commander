package bbolt

import (
	"bytes"
	"commander/internal/kv"
	"context"
	"testing"
)

func TestNewBBoltKV(t *testing.T) {
	tempDir := t.TempDir()

	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	if store.baseDir != tempDir {
		t.Errorf("Expected baseDir %s, got %s", tempDir, store.baseDir)
	}
}

func TestBBoltKV_SetAndGet(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
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

	if !bytes.Equal(retrieved, value) {
		t.Errorf("Expected value %s, got %s", value, retrieved)
	}
}

func TestBBoltKV_GetNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Get non-existent key
	_, err = store.Get(ctx, "testdb", "users", "nonexistent")
	if err != kv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestBBoltKV_Delete(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
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

func TestBBoltKV_DeleteNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Delete non-existent key
	err = store.Delete(ctx, "testdb", "users", "nonexistent")
	if err != kv.ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestBBoltKV_Exists(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
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

func TestBBoltKV_MultipleNamespaces(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	key := "shared_key"
	value1 := []byte(`{"db":"db1"}`)
	value2 := []byte(`{"db":"db2"}`)

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
	if !bytes.Equal(retrieved1, value1) {
		t.Errorf("Expected value %s, got %s", value1, retrieved1)
	}

	// Get from namespace2
	retrieved2, err := store.Get(ctx, "namespace2", "collection", key)
	if err != nil {
		t.Fatalf("Failed to get value from namespace2: %v", err)
	}
	if !bytes.Equal(retrieved2, value2) {
		t.Errorf("Expected value %s, got %s", value2, retrieved2)
	}
}

func TestBBoltKV_MultipleCollections(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
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
	if !bytes.Equal(retrieved1, value1) {
		t.Errorf("Expected value %s, got %s", value1, retrieved1)
	}

	// Get from collection2
	retrieved2, err := store.Get(ctx, namespace, "posts", key)
	if err != nil {
		t.Fatalf("Failed to get value from posts: %v", err)
	}
	if !bytes.Equal(retrieved2, value2) {
		t.Errorf("Expected value %s, got %s", value2, retrieved2)
	}
}

func TestBBoltKV_DefaultNamespace(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
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

	if !bytes.Equal(retrieved, value) {
		t.Errorf("Expected value %s, got %s", value, retrieved)
	}
}

func TestBBoltKV_Ping(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	err = store.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestBBoltKV_Close(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
	}

	ctx := context.Background()

	// Create some databases
	_ = store.Set(ctx, "db1", "col1", "key1", []byte("value1"))
	_ = store.Set(ctx, "db2", "col2", "key2", []byte("value2"))

	// Close
	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify databases are closed (map should be empty)
	if len(store.dbs) != 0 {
		t.Errorf("Expected empty dbs map after close, got %d entries", len(store.dbs))
	}
}

func TestBBoltKV_UpdateValue(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBBoltKV(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BBolt KV: %v", err)
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

	if !bytes.Equal(retrieved, value2) {
		t.Errorf("Expected updated value %s, got %s", value2, retrieved)
	}
}
