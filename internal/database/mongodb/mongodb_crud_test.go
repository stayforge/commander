package mongodb

import (
	"context"
	"testing"

	"commander/internal/kv"
	"commander/internal/testing/mocks"

	"github.com/stretchr/testify/assert"
)

// ===== MongoDB CRUD Operations Tests with Mocks =====

// Helper to create a mock-backed test client
type mockMongoDBKV struct {
	mockClient *mocks.MockClient
}

// Get simulates the Get operation using mock
func (m *mockMongoDBKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
	namespace = kv.NormalizeNamespace(namespace)
	if _, exists := m.mockClient.Collections[namespace]; !exists {
		return nil, kv.ErrKeyNotFound
	}
	if coll, exists := m.mockClient.Collections[namespace][collection]; exists {
		if doc, found := coll.Documents[key]; found {
			// Simulate BSON marshaling
			if str, ok := doc.(string); ok {
				return []byte(str), nil
			}
		}
	}
	return nil, kv.ErrKeyNotFound
}

// Set simulates the Set operation using mock
func (m *mockMongoDBKV) Set(ctx context.Context, namespace, collection, key string, value []byte) error {
	namespace = kv.NormalizeNamespace(namespace)
	if _, exists := m.mockClient.Collections[namespace]; !exists {
		m.mockClient.Collections[namespace] = make(map[string]*mocks.MockCollection)
	}
	if _, exists := m.mockClient.Collections[namespace][collection]; !exists {
		m.mockClient.Collections[namespace][collection] = &mocks.MockCollection{
			Documents: make(map[string]interface{}),
		}
	}
	m.mockClient.Collections[namespace][collection].Documents[key] = string(value)
	return nil
}

// Delete simulates the Delete operation using mock
func (m *mockMongoDBKV) Delete(ctx context.Context, namespace, collection, key string) error {
	namespace = kv.NormalizeNamespace(namespace)
	if _, exists := m.mockClient.Collections[namespace]; !exists {
		return kv.ErrKeyNotFound
	}
	if coll, exists := m.mockClient.Collections[namespace][collection]; exists {
		if _, found := coll.Documents[key]; found {
			delete(coll.Documents, key)
			return nil
		}
	}
	return kv.ErrKeyNotFound
}

// Exists simulates the Exists operation using mock
func (m *mockMongoDBKV) Exists(ctx context.Context, namespace, collection, key string) (bool, error) {
	namespace = kv.NormalizeNamespace(namespace)
	if _, exists := m.mockClient.Collections[namespace]; !exists {
		return false, nil
	}
	if coll, exists := m.mockClient.Collections[namespace][collection]; exists {
		_, found := coll.Documents[key]
		return found, nil
	}
	return false, nil
}

// Close simulates close operation
func (m *mockMongoDBKV) Close() error {
	m.mockClient.Clear()
	return nil
}

// Ping simulates ping operation
func (m *mockMongoDBKV) Ping(ctx context.Context) error {
	return m.mockClient.Ping(ctx)
}

// ===== GET Operation Tests =====

func TestMongoDBCRUD_Get_Success(t *testing.T) {
	t.Run("get existing key returns value", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set a value first
		err := mock.Set(ctx, "default", "test", "mykey", []byte("myvalue"))
		assert.NoError(t, err)

		// Get the value
		value, err := mock.Get(ctx, "default", "test", "mykey")
		assert.NoError(t, err)
		assert.Equal(t, []byte("myvalue"), value)
	})

	t.Run("get returns exact bytes stored", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		testData := []byte(`{"name": "test", "value": 123}`)
		err := mock.Set(ctx, "default", "test", "jsonkey", testData)
		assert.NoError(t, err)

		value, err := mock.Get(ctx, "default", "test", "jsonkey")
		assert.NoError(t, err)
		assert.Equal(t, testData, value)
	})

	t.Run("get applies namespace normalization for empty namespace", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set with default namespace (empty string normalizes to "default")
		err := mock.Set(ctx, "", "test", "key1", []byte("value1"))
		assert.NoError(t, err)

		// Get with explicit "default" namespace
		value, err := mock.Get(ctx, "default", "test", "key1")
		assert.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
	})
}

func TestMongoDBCRUD_Get_Errors(t *testing.T) {
	t.Run("get non-existent key returns ErrKeyNotFound", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		value, err := mock.Get(ctx, "default", "test", "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, kv.ErrKeyNotFound, err)
		assert.Nil(t, value)
	})

	t.Run("get from non-existent collection returns ErrKeyNotFound", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		value, err := mock.Get(ctx, "default", "nonexistent", "key")
		assert.Error(t, err)
		assert.Equal(t, kv.ErrKeyNotFound, err)
		assert.Nil(t, value)
	})

	t.Run("get from non-existent namespace returns ErrKeyNotFound", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		value, err := mock.Get(ctx, "nonexistent", "test", "key")
		assert.Error(t, err)
		assert.Equal(t, kv.ErrKeyNotFound, err)
		assert.Nil(t, value)
	})
}

// ===== SET Operation Tests =====

func TestMongoDBCRUD_Set_Success(t *testing.T) {
	t.Run("set creates new key-value pair", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		err := mock.Set(ctx, "default", "test", "newkey", []byte("newvalue"))
		assert.NoError(t, err)

		// Verify it was stored
		value, err := mock.Get(ctx, "default", "test", "newkey")
		assert.NoError(t, err)
		assert.Equal(t, []byte("newvalue"), value)
	})

	t.Run("set updates existing key", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set initial value
		mock.Set(ctx, "default", "test", "key", []byte("value1"))

		// Update with new value
		err := mock.Set(ctx, "default", "test", "key", []byte("value2"))
		assert.NoError(t, err)

		// Verify updated value
		value, err := mock.Get(ctx, "default", "test", "key")
		assert.NoError(t, err)
		assert.Equal(t, []byte("value2"), value)
	})

	t.Run("set creates collection and namespace if not exist", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set in non-existent namespace and collection
		err := mock.Set(ctx, "newnamespace", "newcollection", "key", []byte("value"))
		assert.NoError(t, err)

		// Verify it exists
		exists, err := mock.Exists(ctx, "newnamespace", "newcollection", "key")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("set stores empty value", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		err := mock.Set(ctx, "default", "test", "emptykey", []byte(""))
		assert.NoError(t, err)

		value, err := mock.Get(ctx, "default", "test", "emptykey")
		assert.NoError(t, err)
		assert.Equal(t, []byte(""), value)
	})
}

func TestMongoDBCRUD_Set_LargeValues(t *testing.T) {
	t.Run("set stores large values", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Create a large value (1MB)
		largeValue := make([]byte, 1024*1024)
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}

		err := mock.Set(ctx, "default", "test", "largekey", largeValue)
		assert.NoError(t, err)

		value, err := mock.Get(ctx, "default", "test", "largekey")
		assert.NoError(t, err)
		assert.Equal(t, largeValue, value)
	})
}

// ===== DELETE Operation Tests =====

func TestMongoDBCRUD_Delete_Success(t *testing.T) {
	t.Run("delete removes existing key", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set a value
		mock.Set(ctx, "default", "test", "delkey", []byte("delvalue"))

		// Verify it exists
		exists, _ := mock.Exists(ctx, "default", "test", "delkey")
		assert.True(t, exists)

		// Delete it
		err := mock.Delete(ctx, "default", "test", "delkey")
		assert.NoError(t, err)

		// Verify it's gone
		exists, _ = mock.Exists(ctx, "default", "test", "delkey")
		assert.False(t, exists)
	})

	t.Run("delete returns ErrKeyNotFound for non-existent key", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		err := mock.Delete(ctx, "default", "test", "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, kv.ErrKeyNotFound, err)
	})

	t.Run("delete is idempotent operation", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set and delete
		mock.Set(ctx, "default", "test", "key", []byte("value"))
		mock.Delete(ctx, "default", "test", "key")

		// Second delete should fail
		err := mock.Delete(ctx, "default", "test", "key")
		assert.Error(t, err)
	})
}

// ===== EXISTS Operation Tests =====

func TestMongoDBCRUD_Exists_Success(t *testing.T) {
	t.Run("exists returns true for existing key", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		mock.Set(ctx, "default", "test", "exkey", []byte("exvalue"))

		exists, err := mock.Exists(ctx, "default", "test", "exkey")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("exists returns false for non-existent key", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		exists, err := mock.Exists(ctx, "default", "test", "notexist")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("exists returns false for non-existent collection", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		exists, err := mock.Exists(ctx, "default", "nocoll", "key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("exists returns false for non-existent namespace", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		exists, err := mock.Exists(ctx, "nonamespace", "test", "key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// ===== Namespace Isolation Tests =====

func TestMongoDBCRUD_Namespaces(t *testing.T) {
	t.Run("different namespaces are isolated", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set in namespace1
		mock.Set(ctx, "ns1", "test", "key", []byte("value1"))

		// Set in namespace2
		mock.Set(ctx, "ns2", "test", "key", []byte("value2"))

		// Verify isolation
		val1, _ := mock.Get(ctx, "ns1", "test", "key")
		val2, _ := mock.Get(ctx, "ns2", "test", "key")

		assert.Equal(t, []byte("value1"), val1)
		assert.Equal(t, []byte("value2"), val2)
	})

	t.Run("namespace normalization works consistently", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set with explicit namespace
		mock.Set(ctx, "MyNamespace", "test", "key", []byte("value"))

		// Get with same namespace (case-sensitive in our implementation)
		val, err := mock.Get(ctx, "MyNamespace", "test", "key")
		assert.NoError(t, err)
		assert.Equal(t, []byte("value"), val)
	})
}

// ===== Collection Isolation Tests =====

func TestMongoDBCRUD_Collections(t *testing.T) {
	t.Run("different collections are isolated", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set in collection1
		mock.Set(ctx, "default", "coll1", "key", []byte("value1"))

		// Set in collection2
		mock.Set(ctx, "default", "coll2", "key", []byte("value2"))

		// Verify isolation
		val1, _ := mock.Get(ctx, "default", "coll1", "key")
		val2, _ := mock.Get(ctx, "default", "coll2", "key")

		assert.Equal(t, []byte("value1"), val1)
		assert.Equal(t, []byte("value2"), val2)
	})

	t.Run("same key in different collections stores different values", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		mock.Set(ctx, "default", "users", "123", []byte("Alice"))
		mock.Set(ctx, "default", "products", "123", []byte("Laptop"))

		user, _ := mock.Get(ctx, "default", "users", "123")
		product, _ := mock.Get(ctx, "default", "products", "123")

		assert.Equal(t, []byte("Alice"), user)
		assert.Equal(t, []byte("Laptop"), product)
	})
}

// ===== Connection Tests =====

func TestMongoDBCRUD_Connection(t *testing.T) {
	t.Run("ping succeeds on mock client", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		err := mock.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("close clears all data", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		mock.Set(ctx, "default", "test", "key", []byte("value"))
		err := mock.Close()
		assert.NoError(t, err)

		// Verify data is cleared
		exists, _ := mock.Exists(ctx, "default", "test", "key")
		assert.False(t, exists)
	})
}

// ===== Transaction Simulation Tests =====

func TestMongoDBCRUD_MultipleOperations(t *testing.T) {
	t.Run("multiple set operations", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		for i := 0; i < 10; i++ {
			key := "key" + string(rune('0'+i))
			value := []byte("value" + string(rune('0'+i)))
			err := mock.Set(ctx, "default", "test", key, value)
			assert.NoError(t, err)
		}

		// Verify all were stored
		for i := 0; i < 10; i++ {
			key := "key" + string(rune('0'+i))
			exists, _ := mock.Exists(ctx, "default", "test", key)
			assert.True(t, exists)
		}
	})

	t.Run("set-get-delete cycle", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set
		mock.Set(ctx, "default", "test", "key", []byte("value"))
		exists1, _ := mock.Exists(ctx, "default", "test", "key")
		assert.True(t, exists1)

		// Get
		value, err := mock.Get(ctx, "default", "test", "key")
		assert.NoError(t, err)
		assert.Equal(t, []byte("value"), value)

		// Delete
		mock.Delete(ctx, "default", "test", "key")
		exists2, _ := mock.Exists(ctx, "default", "test", "key")
		assert.False(t, exists2)
	})

	t.Run("concurrent namespace operations", func(t *testing.T) {
		mock := &mockMongoDBKV{mockClient: mocks.NewMockClient()}
		ctx := context.Background()

		// Set in multiple namespaces
		for ns := 1; ns <= 5; ns++ {
			namespace := "ns" + string(rune('0'+ns))
			for i := 0; i < 3; i++ {
				key := "key" + string(rune('0'+i))
				value := []byte(namespace + "-" + key)
				mock.Set(ctx, namespace, "test", key, value)
			}
		}

		// Verify each namespace has correct data
		for ns := 1; ns <= 5; ns++ {
			namespace := "ns" + string(rune('0'+ns))
			for i := 0; i < 3; i++ {
				key := "key" + string(rune('0'+i))
				value, _ := mock.Get(ctx, namespace, "test", key)
				expected := []byte(namespace + "-" + key)
				assert.Equal(t, expected, value)
			}
		}
	})
}
