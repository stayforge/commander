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
// Sets multiple key-value pairs in a single request
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
// Deletes multiple keys in a single request
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
// Lists all keys in a collection (backend-dependent, may not be available for all backends)
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
			_ = scanInt(offsetParam, &offset) // nolint:errcheck
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Try to list keys (this may not be supported by all backends)
		// For now, return a not-implemented response
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Message: "listing keys is not implemented for this backend",
			Code:    "NOT_IMPLEMENTED",
		})
	}
}

// Helper functions

// scanInt parses a string as an integer
func scanInt(s string, v *int) error {
	n, err := parseStringToInt(s)
	if err != nil {
		return err
	}
	*v = n
	return nil
}

// parseStringToInt parses a string to an integer using simple logic
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
