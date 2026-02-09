package handlers

import (
	"errors"
	"net/http"
	"time"

	"commander/internal/kv"

	"github.com/gin-gonic/gin"
)

// BatchSetRequest represents a batch set operation request
type BatchSetRequest struct {
	Operations []BatchSetOperation `json:"operations" binding:"required,min=1,max=1000"`
}

// BatchSetOperation represents a single set operation in a batch
type BatchSetOperation struct {
	Namespace  string      `json:"namespace" binding:"required"`
	Collection string      `json:"collection" binding:"required"`
	Key        string      `json:"key" binding:"required"`
	Value      interface{} `json:"value" binding:"required"`
}

// BatchDeleteRequest represents a batch delete operation request
type BatchDeleteRequest struct {
	Operations []BatchDeleteOperation `json:"operations" binding:"required,min=1,max=1000"`
}

// BatchDeleteOperation represents a single delete operation in a batch
type BatchDeleteOperation struct {
	Namespace  string `json:"namespace" binding:"required"`
	Collection string `json:"collection" binding:"required"`
	Key        string `json:"key" binding:"required"`
}

// BatchOperationResult represents the result of a single batch operation
type BatchOperationResult struct {
	Namespace  string `json:"namespace"`
	Collection string `json:"collection"`
	Key        string `json:"key"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

// BatchSetResponse represents the response for a batch set operation
type BatchSetResponse struct {
	Message      string                 `json:"message"`
	Results      []BatchOperationResult `json:"results"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Timestamp    string                 `json:"timestamp"`
}

// BatchDeleteResponse represents the response for a batch delete operation
type BatchDeleteResponse struct {
	Message      string                 `json:"message"`
	Results      []BatchOperationResult `json:"results"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Timestamp    string                 `json:"timestamp"`
}

// BatchSetHandler handles POST /api/v1/kv/batch (set)
// BatchSetHandler returns a Gin handler that performs multiple set operations against the provided KV store.
// It accepts a JSON BatchSetRequest containing one or more operations and responds with a BatchSetResponse
// that includes per-operation results, aggregate success and failure counts, and a UTC timestamp.
// The handler responds with HTTP 400 for an invalid request body or when the operations list is empty;
// individual operation failures are reported in the returned Results slice.
func BatchSetHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BatchSetRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    "INVALID_BODY",
			})
			return
		}

		// Validate that we don't have too many operations
		if len(req.Operations) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "at least one operation is required",
				Code:    "EMPTY_OPERATIONS",
			})
			return
		}

		results := make([]BatchOperationResult, 0, len(req.Operations))
		successCount := 0
		failureCount := 0
		ctx := c.Request.Context()

		// Process each operation
		for _, op := range req.Operations {
			result := BatchOperationResult{
				Namespace:  op.Namespace,
				Collection: op.Collection,
				Key:        op.Key,
				Success:    false,
			}

			// Validate operation
			if op.Namespace == "" || op.Collection == "" || op.Key == "" {
				result.Error = "namespace, collection, and key are required"
				failureCount++
				results = append(results, result)
				continue
			}

			// Normalize namespace
			namespace := kv.NormalizeNamespace(op.Namespace)

			// Marshal value to JSON
			valueJSON, err := marshalJSON(op.Value)
			if err != nil {
				result.Error = "failed to encode value: " + err.Error()
				failureCount++
				results = append(results, result)
				continue
			}

			// Set value in KV store
			if err := kvStore.Set(ctx, namespace, op.Collection, op.Key, valueJSON); err != nil {
				result.Error = "failed to set key: " + err.Error()
				failureCount++
				results = append(results, result)
				continue
			}

			result.Success = true
			successCount++
			results = append(results, result)
		}

		c.JSON(http.StatusOK, BatchSetResponse{
			Message:      "Batch operation completed",
			Results:      results,
			SuccessCount: successCount,
			FailureCount: failureCount,
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// BatchDeleteHandler handles DELETE /api/v1/kv/batch (delete)
// BatchDeleteHandler returns a gin handler that processes a batch delete request using the provided KV store.
// 
// The handler accepts a JSON BatchDeleteRequest containing one or more delete operations, validates each
// operation (namespace, collection, key), normalizes namespaces, and attempts to delete each key from the
// KV store. The response is a BatchDeleteResponse containing per-operation results, aggregate success and
// failure counts, and a UTC timestamp. The handler responds with 400 for invalid request bodies or empty
// operations and 200 when the batch has been processed.
func BatchDeleteHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BatchDeleteRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    "INVALID_BODY",
			})
			return
		}

		// Validate that we don't have too many operations
		if len(req.Operations) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "at least one operation is required",
				Code:    "EMPTY_OPERATIONS",
			})
			return
		}

		results := make([]BatchOperationResult, 0, len(req.Operations))
		successCount := 0
		failureCount := 0
		ctx := c.Request.Context()

		// Process each operation
		for _, op := range req.Operations {
			result := BatchOperationResult{
				Namespace:  op.Namespace,
				Collection: op.Collection,
				Key:        op.Key,
				Success:    false,
			}

			// Validate operation
			if op.Namespace == "" || op.Collection == "" || op.Key == "" {
				result.Error = "namespace, collection, and key are required"
				failureCount++
				results = append(results, result)
				continue
			}

			// Normalize namespace
			namespace := kv.NormalizeNamespace(op.Namespace)

			// Delete value from KV store
			if err := kvStore.Delete(ctx, namespace, op.Collection, op.Key); err != nil {
				result.Error = "failed to delete key: " + err.Error()
				failureCount++
				results = append(results, result)
				continue
			}

			result.Success = true
			successCount++
			results = append(results, result)
		}

		c.JSON(http.StatusOK, BatchDeleteResponse{
			Message:      "Batch operation completed",
			Results:      results,
			SuccessCount: successCount,
			FailureCount: failureCount,
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// ListKeysRequest represents a request to list keys in a collection
type ListKeysRequest struct {
	Limit  int `json:"limit,omitempty" binding:"max=10000"`
	Offset int `json:"offset,omitempty"`
}

// ListKeysResponse represents the response for listing keys
type ListKeysResponse struct {
	Message    string   `json:"message"`
	Namespace  string   `json:"namespace"`
	Collection string   `json:"collection"`
	Keys       []string `json:"keys"`
	Total      int      `json:"total"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	Timestamp  string   `json:"timestamp"`
}

// ListKeysHandler handles GET /api/v1/kv/{namespace}/{collection}
// ListKeysHandler returns a gin.HandlerFunc that handles requests to list keys in a collection.
// It validates required path parameters `namespace` and `collection`, parses optional `limit`
// (default 1000, capped at 10000) and `offset` (default 0) query parameters, and responds with
// HTTP 501 Not Implemented indicating that key listing is not supported by the backend.
func ListKeysHandler(kvStore kv.KV) gin.HandlerFunc {
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

		// Parse query parameters
		limit := 1000
		offset := 0
		if limitParam := c.Query("limit"); limitParam != "" {
			if err := scanInt(limitParam, &limit); err != nil || limit > 10000 {
				limit = 1000
			}
		}
		if offsetParam := c.Query("offset"); offsetParam != "" {
			_ = scanInt(offsetParam, &offset) //nolint:errcheck // offset parsing failure is intentionally ignored, default 0 is used
		}

		// Try to list keys (this may not be supported by all backends)
		// For now, return a not-implemented response
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "listing keys is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// Helper functions

// scanInt parses s as a base-10 integer and stores the result in v.
// It returns an error if s is not a valid integer representation.
func scanInt(s string, v *int) error {
	n, err := parseStringToInt(s)
	if err != nil {
		return err
	}
	*v = n
	return nil
}

// parseStringToInt parses s as a base-10 integer and returns the integer value or an error.
// It accepts an optional leading '-' for negative values. It returns an error if s is empty
// or if any non-digit character (other than a leading '-') is present.
func parseStringToInt(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}

	result := 0
	negative := false

	// Check for negative sign
	start := 0
	if s[0] == '-' {
		negative = true
		start = 1
	}

	// Parse digits
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errors.New("invalid character in number")
		}
		result = result*10 + int(s[i]-'0')
	}

	if negative {
		result = -result
	}

	return result, nil
}