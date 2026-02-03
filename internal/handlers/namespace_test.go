package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestListNamespacesHandler tests GET /api/v1/namespaces
func TestListNamespacesHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/namespaces", ListNamespacesHandler(mockKV))

	req, _ := http.NewRequest("GET", "/api/v1/namespaces", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "NOT_IMPLEMENTED", resp.Code)
}

// TestListCollectionsHandler tests GET /api/v1/namespaces/{namespace}/collections
func TestListCollectionsHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/namespaces/:namespace/collections", ListCollectionsHandler(mockKV))

	tests := []struct {
		name           string
		namespace      string
		expectedStatus int
	}{
		{
			name:           "list collections in namespace",
			namespace:      "default",
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "invalid namespace (empty)",
			namespace:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/namespaces/"+tt.namespace+"/collections", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteNamespaceHandler tests DELETE /api/v1/namespaces/{namespace}
func TestDeleteNamespaceHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/v1/namespaces/:namespace", DeleteNamespaceHandler(mockKV))

	tests := []struct {
		name           string
		namespace      string
		expectedStatus int
	}{
		{
			name:           "delete namespace",
			namespace:      "custom",
			expectedStatus: http.StatusNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/v1/namespaces/"+tt.namespace, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteCollectionHandler tests DELETE /api/v1/namespaces/{namespace}/collections/{collection}
func TestDeleteCollectionHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/v1/namespaces/:namespace/collections/:collection", DeleteCollectionHandler(mockKV))

	tests := []struct {
		name           string
		namespace      string
		collection     string
		expectedStatus int
	}{
		{
			name:           "delete collection",
			namespace:      "default",
			collection:     "users",
			expectedStatus: http.StatusNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE",
				"/api/v1/namespaces/"+tt.namespace+"/collections/"+tt.collection, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetNamespaceInfoHandler tests GET /api/v1/namespaces/{namespace}/info
func TestGetNamespaceInfoHandler(t *testing.T) {
	mockKV := NewMockKV()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/namespaces/:namespace/info", GetNamespaceInfoHandler(mockKV))

	tests := []struct {
		name           string
		namespace      string
		expectedStatus int
	}{
		{
			name:           "get namespace info",
			namespace:      "default",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid namespace (empty)",
			namespace:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/namespaces/"+tt.namespace+"/info", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp NamespaceInfoResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.namespace, resp.Namespace)
			}
		})
	}
}
