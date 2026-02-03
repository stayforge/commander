package handlers

import (
	"net/http"
	"time"

	"commander/internal/kv"

	"github.com/gin-gonic/gin"
)

// ListNamespacesResponse represents the response for listing namespaces
type ListNamespacesResponse struct {
	Message    string   `json:"message"`
	Namespaces []string `json:"namespaces"`
	Count      int      `json:"count"`
	Timestamp  string   `json:"timestamp"`
}

// ListCollectionsResponse represents the response for listing collections
type ListCollectionsResponse struct {
	Message     string   `json:"message"`
	Namespace   string   `json:"namespace"`
	Collections []string `json:"collections"`
	Count       int      `json:"count"`
	Timestamp   string   `json:"timestamp"`
}

// DeleteNamespaceResponse represents the response for deleting a namespace
type DeleteNamespaceResponse struct {
	Message   string `json:"message"`
	Namespace string `json:"namespace"`
	Timestamp string `json:"timestamp"`
}

// DeleteCollectionResponse represents the response for deleting a collection
type DeleteCollectionResponse struct {
	Message    string `json:"message"`
	Namespace  string `json:"namespace"`
	Collection string `json:"collection"`
	Timestamp  string `json:"timestamp"`
}

// ListNamespacesHandler handles GET /api/v1/namespaces
// Lists all namespaces (returns empty list with not-implemented message)
func ListNamespacesHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Note: Listing namespaces is not implemented for all backends
		// Each backend would need to implement namespace listing separately
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "listing namespaces is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// ListCollectionsHandler handles GET /api/v1/namespaces/{namespace}/collections
// Lists all collections in a namespace (returns empty list with not-implemented message)
func ListCollectionsHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")

		// Validate parameters
		if namespace == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace is required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Note: Listing collections is not implemented for all backends
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "listing collections is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// DeleteNamespaceHandler handles DELETE /api/v1/namespaces/{namespace}
// Deletes an entire namespace (backend-dependent)
// Note: For BBolt, this would delete the entire .db file
// For MongoDB, this would drop the database
// For Redis, this would delete all keys with the namespace prefix
func DeleteNamespaceHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")

		// Validate parameters
		if namespace == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace is required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Normalize namespace (but prevent deletion of empty string)
		if namespace != "default" && namespace != kv.DefaultNamespace {
			// For safety, we require explicit namespace name, not empty string
		}

		// Note: Namespace deletion is not implemented for all backends
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "deleting namespaces is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// DeleteCollectionHandler handles DELETE /api/v1/namespaces/{namespace}/collections/{collection}
// Deletes all keys in a collection
func DeleteCollectionHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		collection := c.Param("collection")

		// Validate parameters
		if namespace == "" || collection == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace and collection are required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Note: Collection deletion is not implemented for all backends
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "deleting collections is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// NamespaceInfoResponse represents information about a namespace
type NamespaceInfoResponse struct {
	Message     string   `json:"message"`
	Namespace   string   `json:"namespace"`
	Collections []string `json:"collections,omitempty"`
	KeyCount    int      `json:"key_count,omitempty"`
	Size        int64    `json:"size,omitempty"`
	Timestamp   string   `json:"timestamp"`
}

// GetNamespaceInfoHandler handles GET /api/v1/namespaces/{namespace}/info
// Returns information about a namespace
func GetNamespaceInfoHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")

		// Validate parameters
		if namespace == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace is required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		c.JSON(http.StatusOK, NamespaceInfoResponse{
			Message:   "Namespace information retrieved",
			Namespace: namespace,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}
