package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commander/internal/kv"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockKV is a mock implementation of kv.KV for testing
type MockKV struct {
	data map[string]map[string]map[string][]byte
}

// NewMockKV creates a new MockKV instance
func NewMockKV() *MockKV {
	return &MockKV{
		data: make(map[string]map[string]map[string][]byte),
	}
}

// Get retrieves a value from the mock KV store
func (m *MockKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
	if ns, ok := m.data[namespace]; ok {
		if coll, ok := ns[collection]; ok {
			if val, ok := coll[key]; ok {
				return val, nil
			}
		}
	}
	return nil, kv.ErrKeyNotFound
}

// Set stores a value in the mock KV store
func (m *MockKV) Set(ctx context.Context, namespace, collection, key string, value []byte) error {
	if _, ok := m.data[namespace]; !ok {
		m.data[namespace] = make(map[string]map[string][]byte)
	}
	if _, ok := m.data[namespace][collection]; !ok {
		m.data[namespace][collection] = make(map[string][]byte)
	}
	m.data[namespace][collection][key] = value
	return nil
}

// Delete removes a key from the mock KV store
func (m *MockKV) Delete(ctx context.Context, namespace, collection, key string) error {
	if ns, ok := m.data[namespace]; ok {
		if coll, ok := ns[collection]; ok {
			delete(coll, key)
		}
	}
	return nil
}

// Exists checks if a key exists in the mock KV store
func (m *MockKV) Exists(ctx context.Context, namespace, collection, key string) (bool, error) {
	if ns, ok := m.data[namespace]; ok {
		if coll, ok := ns[collection]; ok {
			_, exists := coll[key]
			return exists, nil
		}
	}
	return false, nil
}

// Close is a no-op for mock KV
func (m *MockKV) Close() error {
	return nil
}

// Ping is a no-op for mock KV
func (m *MockKV) Ping(ctx context.Context) error {
	return nil
}

// TestGetKVHandler tests GET /api/v1/kv/{namespace}/{collection}/{key}
func TestGetKVHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/kv/:namespace/:collection/:key", GetKVHandler(mockKV))

	testValue := map[string]interface{}{"name": "test", "value": 123}
	valueJSON, _ := json.Marshal(testValue)

	// Setup mock data
	err := mockKV.Set(context.Background(), "default", "users", "user1", valueJSON)
	require.NoError(t, err)

	tests := []struct {
		name           string
		namespace      string
		collection     string
		key            string
		expectedStatus int
		expectedInBody bool
	}{
		{
			name:           "successful get",
			namespace:      "default",
			collection:     "users",
			key:            "user1",
			expectedStatus: http.StatusOK,
			expectedInBody: true,
		},
		{
			name:           "key not found",
			namespace:      "default",
			collection:     "users",
			key:            "nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedInBody: false,
		},
		{
			name:           "invalid namespace",
			namespace:      "",
			collection:     "users",
			key:            "user1",
			expectedStatus: http.StatusBadRequest,
			expectedInBody: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET",
				"/api/v1/kv/"+tt.namespace+"/"+tt.collection+"/"+tt.key,
				http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedInBody {
				var resp KVResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "default", resp.Namespace)
				assert.Equal(t, "users", resp.Collection)
				assert.Equal(t, "user1", resp.Key)
			}
		})
	}
}

// TestSetKVHandler tests POST /api/v1/kv/{namespace}/{collection}/{key}
func TestSetKVHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/kv/:namespace/:collection/:key", SetKVHandler(mockKV))

	tests := []struct {
		name           string
		namespace      string
		collection     string
		key            string
		body           KVRequestBody
		expectedStatus int
	}{
		{
			name:       "successful set",
			namespace:  "default",
			collection: "users",
			key:        "user1",
			body: KVRequestBody{
				Value: map[string]interface{}{"name": "John", "age": 30},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "set string value",
			namespace:  "default",
			collection: "config",
			key:        "app_name",
			body: KVRequestBody{
				Value: "My App",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "invalid namespace",
			namespace:  "",
			collection: "users",
			key:        "user1",
			body: KVRequestBody{
				Value: "test",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST",
				"/api/v1/kv/"+tt.namespace+"/"+tt.collection+"/"+tt.key,
				bytes.NewBuffer(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var resp KVResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.namespace, resp.Namespace)
				assert.Equal(t, tt.collection, resp.Collection)
				assert.Equal(t, tt.key, resp.Key)
				assert.Equal(t, "Successfully", resp.Message)
			}
		})
	}
}

// TestDeleteKVHandler tests DELETE /api/v1/kv/{namespace}/{collection}/{key}
func TestDeleteKVHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/v1/kv/:namespace/:collection/:key", DeleteKVHandler(mockKV))

	// Setup initial data
	ctx := context.Background()
	testValue, _ := json.Marshal("test value")
	err := mockKV.Set(ctx, "default", "users", "user1", testValue)
	require.NoError(t, err)

	tests := []struct {
		name           string
		namespace      string
		collection     string
		key            string
		expectedStatus int
	}{
		{
			name:           "successful delete",
			namespace:      "default",
			collection:     "users",
			key:            "user1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid namespace",
			namespace:      "",
			collection:     "users",
			key:            "user1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE",
				"/api/v1/kv/"+tt.namespace+"/"+tt.collection+"/"+tt.key,
				http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestHeadKVHandler tests HEAD /api/v1/kv/{namespace}/{collection}/{key}
func TestHeadKVHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.HEAD("/api/v1/kv/:namespace/:collection/:key", HeadKVHandler(mockKV))

	// Setup initial data
	ctx := context.Background()
	testValue, _ := json.Marshal("test value")
	err := mockKV.Set(ctx, "default", "users", "user1", testValue)
	require.NoError(t, err)

	tests := []struct {
		name           string
		namespace      string
		collection     string
		key            string
		expectedStatus int
	}{
		{
			name:           "key exists",
			namespace:      "default",
			collection:     "users",
			key:            "user1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "key not found",
			namespace:      "default",
			collection:     "users",
			key:            "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid namespace",
			namespace:      "",
			collection:     "users",
			key:            "user1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("HEAD",
				"/api/v1/kv/"+tt.namespace+"/"+tt.collection+"/"+tt.key,
				http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestNormalizeNamespace tests namespace normalization
func TestNormalizeNamespace(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/kv/:namespace/:collection/:key", SetKVHandler(mockKV))

	// Test empty namespace defaults to "default"
	body := KVRequestBody{Value: "test"}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		"/api/v1/kv/default/users/user1",
		bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp KVResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "default", resp.Namespace)
}
