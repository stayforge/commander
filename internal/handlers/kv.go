package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"commander/internal/kv"

	"github.com/gin-gonic/gin"
)

// KVRequestBody represents the JSON body for KV operations
type KVRequestBody struct {
	Value interface{} `json:"value" binding:"required"` // The value to store (will be JSON-encoded)
}

// KVResponse represents a standard KV response
type KVResponse struct {
	Message    string      `json:"message"`
	Namespace  string      `json:"namespace"`
	Collection string      `json:"collection"`
	Key        string      `json:"key"`
	Value      interface{} `json:"value,omitempty"`
	Exists     bool        `json:"exists,omitempty"`
	Timestamp  string      `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// GetKVHandler handles GET /api/v1/kv/{namespace}/{collection}/{key}
// GetKVHandler produces an HTTP handler for GET /api/v1/kv/{namespace}/{collection}/{key} that retrieves a JSON-decoded value from the provided KV store.
// 
// The handler validates that namespace, collection, and key are present, normalizes the namespace, and attempts to fetch the value from the KV store.
// On a missing key it responds with 404 and code "KEY_NOT_FOUND". On JSON decode failures it responds with 500 and code "DECODE_ERROR". On other retrieval failures it responds with 500 and code "INTERNAL_ERROR".
// On success the handler responds with 200 and a KVResponse containing the decoded value, request identifiers, and a UTC RFC3339 timestamp.
func GetKVHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		collection := c.Param("collection")
		key := c.Param("key")

		// Validate parameters
		if namespace == "" || collection == "" || key == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace, collection, and key are required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Get value from KV store
		ctx := c.Request.Context()
		value, err := kvStore.Get(ctx, namespace, collection, key)
		if err != nil {
			if errors.Is(err, kv.ErrKeyNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Message: "key not found",
					Code:    "KEY_NOT_FOUND",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "failed to retrieve key: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
			return
		}

		// Decode value as JSON for response
		var decodedValue interface{}
		if err := unmarshalJSON(value, &decodedValue); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "failed to decode value",
				Code:    "DECODE_ERROR",
			})
			return
		}

		c.JSON(http.StatusOK, KVResponse{
			Message:    "Successfully",
			Namespace:  namespace,
			Collection: collection,
			Key:        key,
			Value:      decodedValue,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// SetKVHandler handles POST /api/v1/kv/{namespace}/{collection}/{key}
// SetKVHandler returns a gin.HandlerFunc that handles POST /api/v1/kv/{namespace}/{collection}/{key} requests and stores the provided JSON-encodable value in the specified namespace, collection, and key of the given KV store.
// The handler validates path parameters and request body, normalizes the namespace, encodes the value to JSON, writes it to the KV store, and responds with a KVResponse containing the stored value and a UTC timestamp on success or an ErrorResponse with an appropriate HTTP status on failure.
func SetKVHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		collection := c.Param("collection")
		key := c.Param("key")

		// Validate parameters
		if namespace == "" || collection == "" || key == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace, collection, and key are required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Parse request body
		var req KVRequestBody
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    "INVALID_BODY",
			})
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Marshal value to JSON
		valueJSON, err := marshalJSON(req.Value)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "failed to encode value: " + err.Error(),
				Code:    "ENCODE_ERROR",
			})
			return
		}

		// Set value in KV store
		ctx := c.Request.Context()
		if err := kvStore.Set(ctx, namespace, collection, key, valueJSON); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "failed to set key: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
			return
		}

		c.JSON(http.StatusCreated, KVResponse{
			Message:    "Successfully",
			Namespace:  namespace,
			Collection: collection,
			Key:        key,
			Value:      req.Value,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// DeleteKVHandler handles DELETE /api/v1/kv/{namespace}/{collection}/{key}
// DeleteKVHandler returns a gin.HandlerFunc that handles DELETE requests to remove a key from the KV store.
// It validates the namespace, collection, and key parameters, normalizes the namespace, and calls the store's
// Delete method. On success it responds with a 200 JSON KVResponse containing namespace, collection, key and a
// UTC timestamp. If parameters are missing it responds with 400 and an ErrorResponse (code "INVALID_PARAMS"),
// and if the store delete fails it responds with 500 and an ErrorResponse (code "INTERNAL_ERROR").
func DeleteKVHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		collection := c.Param("collection")
		key := c.Param("key")

		// Validate parameters
		if namespace == "" || collection == "" || key == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "namespace, collection, and key are required",
				Code:    "INVALID_PARAMS",
			})
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Delete value from KV store
		ctx := c.Request.Context()
		if err := kvStore.Delete(ctx, namespace, collection, key); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "failed to delete key: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
			return
		}

		c.JSON(http.StatusOK, KVResponse{
			Message:    "Successfully",
			Namespace:  namespace,
			Collection: collection,
			Key:        key,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// HeadKVHandler handles HEAD /api/v1/kv/{namespace}/{collection}/{key}
// when the key exists, 404 when it does not, 400 when required parameters are missing, and 500 on internal errors.
func HeadKVHandler(kvStore kv.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		collection := c.Param("collection")
		key := c.Param("key")

		// Validate parameters
		if namespace == "" || collection == "" || key == "" {
			c.String(http.StatusBadRequest, "namespace, collection, and key are required")
			return
		}

		// Normalize namespace
		namespace = kv.NormalizeNamespace(namespace)

		// Check if key exists
		ctx := c.Request.Context()
		exists, err := kvStore.Exists(ctx, namespace, collection, key)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to check key existence")
			return
		}

		if exists {
			c.Status(http.StatusOK)
		} else {
			c.Status(http.StatusNotFound)
		}
	}
}

// Helper functions

// marshalJSON converts v to JSON bytes. If v is a string it is returned unchanged (treated as pre-encoded JSON); otherwise v is encoded using json.Marshal.
func marshalJSON(value interface{}) ([]byte, error) {
	// If already a string, assume it's JSON
	if str, ok := value.(string); ok {
		return []byte(str), nil
	}

	// Otherwise use Go's JSON marshaling
	return json.Marshal(value)
}

// unmarshalJSON unmarshals JSON-encoded data into v.
// v must be a pointer to the value to populate; returns an error if decoding fails.
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}