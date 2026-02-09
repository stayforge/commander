package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestBatchSetHandler tests POST /api/v1/kv/batch (set)
func TestBatchSetHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/kv/batch", BatchSetHandler(mockKV))

	tests := []struct {
		name           string
		request        BatchSetRequest
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful batch set",
			request: BatchSetRequest{
				Operations: []BatchSetOperation{
					{
						Namespace:  "default",
						Collection: "users",
						Key:        "user1",
						Value: map[string]interface{}{
							"name": "John",
							"age":  30,
						},
					},
					{
						Namespace:  "default",
						Collection: "users",
						Key:        "user2",
						Value:      "simple string",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "batch set with single operation",
			request: BatchSetRequest{
				Operations: []BatchSetOperation{
					{
						Namespace:  "config",
						Collection: "app",
						Key:        "name",
						Value:      "MyApp",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "batch set with invalid operation (missing key)",
			request: BatchSetRequest{
				Operations: []BatchSetOperation{
					{
						Namespace:  "default",
						Collection: "users",
						Key:        "",
						Value:      "test",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/kv/batch", bytes.NewBuffer(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp BatchSetResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.request.Operations), len(resp.Results))
				assert.Equal(t, tt.expectedCount, resp.SuccessCount+resp.FailureCount)
			}
		})
	}
}

// TestBatchDeleteHandler tests DELETE /api/v1/kv/batch (delete)
func TestBatchDeleteHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/v1/kv/batch", BatchDeleteHandler(mockKV))

	// Setup initial data
	ctx := context.Background()
	testValue, _ := json.Marshal("test value")
	_ = mockKV.Set(ctx, "default", "users", "user1", testValue)
	_ = mockKV.Set(ctx, "default", "users", "user2", testValue)

	tests := []struct {
		name           string
		request        BatchDeleteRequest
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful batch delete",
			request: BatchDeleteRequest{
				Operations: []BatchDeleteOperation{
					{
						Namespace:  "default",
						Collection: "users",
						Key:        "user1",
					},
					{
						Namespace:  "default",
						Collection: "users",
						Key:        "user2",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "batch delete with single operation",
			request: BatchDeleteRequest{
				Operations: []BatchDeleteOperation{
					{
						Namespace:  "default",
						Collection: "config",
						Key:        "setting1",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("DELETE", "/api/v1/kv/batch", bytes.NewBuffer(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp BatchDeleteResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.request.Operations), len(resp.Results))
				assert.Equal(t, tt.expectedCount, resp.SuccessCount+resp.FailureCount)
			}
		})
	}
}

// TestListKeysHandler tests GET /api/v1/kv/{namespace}/{collection}
func TestListKeysHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/kv/:namespace/:collection", ListKeysHandler(mockKV))

	tests := []struct {
		name           string
		namespace      string
		collection     string
		expectedStatus int
	}{
		{
			name:           "list keys in collection",
			namespace:      "default",
			collection:     "users",
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "invalid namespace",
			namespace:      "",
			collection:     "users",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET",
				"/api/v1/kv/"+tt.namespace+"/"+tt.collection,
				http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestParseStringToInt tests the integer parsing function
func TestParseStringToInt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		shouldError bool
	}{
		{
			name:        "parse positive number",
			input:       "123",
			expected:    123,
			shouldError: false,
		},
		{
			name:        "parse negative number",
			input:       "-456",
			expected:    -456,
			shouldError: false,
		},
		{
			name:        "parse zero",
			input:       "0",
			expected:    0,
			shouldError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "invalid characters",
			input:       "12a3",
			expected:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseStringToInt(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
