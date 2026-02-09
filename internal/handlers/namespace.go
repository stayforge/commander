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
// ListNamespacesHandler returns a gin.HandlerFunc that always responds with HTTP 501 Not Implemented.
// The handler sends an ErrorResponse with Message "listing namespaces is not implemented for this backend" and Code "NOT_IMPLEMENTED".
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
// ListCollectionsHandler provides a Gin handler that validates a namespace path parameter and responds with a not-implemented error for listing collections.
// 
// If the "namespace" path parameter is empty the handler responds with HTTP 400 and an ErrorResponse containing Message "namespace is required" and Code "INVALID_PARAMS".
// If the parameter is present the handler responds with HTTP 501 and an ErrorResponse containing Message "listing collections is not implemented for this backend" and Code "NOT_IMPLEMENTED".
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
// DeleteNamespaceHandler returns a gin.HandlerFunc that handles HTTP requests to delete a namespace.
// It validates the "namespace" path parameter and responds with HTTP 400 and an error when the parameter is missing.
// If a namespace is provided the handler responds with HTTP 501 and an error indicating namespace deletion is not implemented for this backend.
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

		// Note: Namespace deletion is not implemented for all backends
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "deleting namespaces is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// DeleteCollectionHandler handles DELETE /api/v1/namespaces/{namespace}/collections/{collection}
// DeleteCollectionHandler returns a gin.HandlerFunc that validates the "namespace" and
// "collection" path parameters and handles collection deletion requests.
// If either parameter is missing it responds with HTTP 400 and an ErrorResponse with
// Message "namespace and collection are required" and Code "INVALID_PARAMS".
// For supported backends this handler would perform collection deletion; currently it
// responds with HTTP 501 and an ErrorResponse with Message "deleting collections is not
// implemented for this backend" and Code "NOT_IMPLEMENTED".
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
// GetNamespaceInfoHandler returns a gin.HandlerFunc that handles requests for namespace information.
// It validates that the "namespace" path parameter is present (responding 400 with an error if missing), normalizes the namespace using kv.NormalizeNamespace, and responds 200 with a NamespaceInfoResponse containing the normalized namespace and a timestamp.
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