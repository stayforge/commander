# API Design Rules

## RESTful Principles

### HTTP Methods
- **GET**: Retrieve resources (idempotent, safe)
- **POST**: Create resources or actions (not idempotent)
- **PUT**: Replace entire resource (idempotent)
- **PATCH**: Partial update (not used in Commander)
- **DELETE**: Remove resource (idempotent)
- **HEAD**: Check resource existence (idempotent, safe)

### URL Structure

**Format**: `/api/v1/{resource}/{id}`

```
GET    /api/v1/kv/{namespace}/{collection}/{key}
POST   /api/v1/kv/{namespace}/{collection}/{key}
DELETE /api/v1/kv/{namespace}/{collection}/{key}
HEAD   /api/v1/kv/{namespace}/{collection}/{key}

POST   /api/v1/kv/batch
DELETE /api/v1/kv/batch

GET    /api/v1/namespaces
GET    /api/v1/namespaces/{namespace}/collections
```

**Guidelines**
- Use lowercase
- Use hyphens, not underscores
- Resource names plural where appropriate
- Hierarchical structure for nested resources

## Request/Response Format

### Request Body (POST/PUT)

**Structure**
```json
{
  "value": {
    "key": "value"
  }
}
```

**Validation**
```go
type KVRequestBody struct {
    Value interface{} `json:"value" binding:"required"`
}
```

### Response Format

**Success Response**
```json
{
  "message": "Successfully",
  "namespace": "default",
  "collection": "users",
  "key": "user1",
  "value": {
    "name": "John"
  },
  "timestamp": "2026-02-03T12:34:56Z"
}
```

**Error Response**
```json
{
  "message": "key not found",
  "code": "KEY_NOT_FOUND"
}
```

### Status Codes

**Success**
- `200 OK`: Successful GET, DELETE
- `201 Created`: Successful POST (create)
- `204 No Content`: Successful DELETE (no body)

**Client Errors**
- `400 Bad Request`: Invalid parameters or body
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource doesn't exist
- `409 Conflict`: Resource conflict

**Server Errors**
- `500 Internal Server Error`: Unexpected server error
- `501 Not Implemented`: Feature not available
- `503 Service Unavailable`: Temporary unavailability

### Error Codes

**Format**: `CATEGORY_DETAIL`

**Codes**
- `KEY_NOT_FOUND`: Key doesn't exist
- `INVALID_PARAMS`: Missing or invalid parameters
- `INVALID_BODY`: Invalid request body
- `DECODE_ERROR`: Failed to decode value
- `ENCODE_ERROR`: Failed to encode value
- `INTERNAL_ERROR`: Server error
- `NOT_IMPLEMENTED`: Feature not available

**Implementation**
```go
type ErrorResponse struct {
    Message string `json:"message"`
    Code    string `json:"code"`
}

c.JSON(http.StatusNotFound, ErrorResponse{
    Message: "key not found",
    Code:    "KEY_NOT_FOUND",
})
```

## Parameter Handling

### Path Parameters

```go
// Extract from URL
namespace := c.Param("namespace")
collection := c.Param("collection")
key := c.Param("key")

// Validate
if namespace == "" || collection == "" || key == "" {
    c.JSON(http.StatusBadRequest, ErrorResponse{
        Message: "namespace, collection, and key are required",
        Code:    "INVALID_PARAMS",
    })
    return
}

// Normalize
namespace = kv.NormalizeNamespace(namespace)
```

### Query Parameters

```go
// Optional parameters with defaults
limit := 1000
if limitParam := c.Query("limit"); limitParam != "" {
    if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
        limit = parsedLimit
    }
}

// Validation
if limit > 10000 {
    limit = 10000
}
```

### Request Body

```go
// Parse JSON
var req KVRequestBody
if err := c.BindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{
        Message: "invalid request body: " + err.Error(),
        Code:    "INVALID_BODY",
    })
    return
}

// Validate
if req.Value == nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{
        Message: "value is required",
        Code:    "INVALID_BODY",
    })
    return
}
```

## Gin Handler Pattern

### Standard Handler Structure

```go
func SomeHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract parameters
        namespace := c.Param("namespace")
        collection := c.Param("collection")
        key := c.Param("key")
        
        // 2. Validate parameters
        if namespace == "" || collection == "" || key == "" {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "namespace, collection, and key are required",
                Code:    "INVALID_PARAMS",
            })
            return
        }
        
        // 3. Normalize/transform
        namespace = kv.NormalizeNamespace(namespace)
        
        // 4. Process request
        ctx := c.Request.Context()
        result, err := kvStore.Operation(ctx, namespace, collection, key)
        if err != nil {
            if errors.Is(err, kv.ErrKeyNotFound) {
                c.JSON(http.StatusNotFound, ErrorResponse{
                    Message: "key not found",
                    Code:    "KEY_NOT_FOUND",
                })
                return
            }
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Message: "internal error: " + err.Error(),
                Code:    "INTERNAL_ERROR",
            })
            return
        }
        
        // 5. Return response
        c.JSON(http.StatusOK, Response{
            Message:   "Successfully",
            Namespace: namespace,
            Key:       key,
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
    }
}
```

### Context Usage

```go
// Use request context
ctx := c.Request.Context()

// Pass to KV operations
value, err := kvStore.Get(ctx, namespace, collection, key)

// Respect context cancellation
select {
case <-ctx.Done():
    c.JSON(http.StatusRequestTimeout, ErrorResponse{
        Message: "request timeout",
        Code:    "TIMEOUT",
    })
    return
default:
}
```

## Response Patterns

### Success Response

```go
c.JSON(http.StatusOK, KVResponse{
    Message:    "Successfully",
    Namespace:  namespace,
    Collection: collection,
    Key:        key,
    Value:      decodedValue,
    Timestamp:  time.Now().UTC().Format(time.RFC3339),
})
```

### Error Response

```go
// Not found
if errors.Is(err, kv.ErrKeyNotFound) {
    c.JSON(http.StatusNotFound, ErrorResponse{
        Message: "key not found",
        Code:    "KEY_NOT_FOUND",
    })
    return
}

// Bad request
c.JSON(http.StatusBadRequest, ErrorResponse{
    Message: "invalid parameters",
    Code:    "INVALID_PARAMS",
})

// Internal error
c.JSON(http.StatusInternalServerError, ErrorResponse{
    Message: "failed to process request: " + err.Error(),
    Code:    "INTERNAL_ERROR",
})
```

### Batch Response

```go
c.JSON(http.StatusOK, BatchSetResponse{
    Message:      "Batch operation completed",
    Results:      results,
    SuccessCount: successCount,
    FailureCount: failureCount,
    Timestamp:    time.Now().UTC().Format(time.RFC3339),
})
```

## Data Organization

### Three-Level Hierarchy

**Namespace** → **Collection** → **Key**

```
default/users/user1
default/config/app_name
production/sessions/sess_abc123
```

**Namespace**
- Top-level isolation
- Maps to different storage units (BBolt files, MongoDB databases)
- Defaults to "default" if empty

**Collection**
- Group related data
- Like tables or buckets
- Examples: users, sessions, config

**Key**
- Individual item identifier
- Unique within a collection
- Any string format

### Namespace Normalization

```go
// Empty namespace → "default"
func NormalizeNamespace(namespace string) string {
    if namespace == "" {
        return "default"
    }
    return namespace
}

// Usage
namespace = kv.NormalizeNamespace(c.Param("namespace"))
```

## Batch Operations

### Batch Request Format

```json
{
  "operations": [
    {
      "namespace": "default",
      "collection": "users",
      "key": "user1",
      "value": {"name": "Alice"}
    },
    {
      "namespace": "default",
      "collection": "users",
      "key": "user2",
      "value": {"name": "Bob"}
    }
  ]
}
```

### Batch Response Format

```json
{
  "message": "Batch operation completed",
  "results": [
    {
      "namespace": "default",
      "collection": "users",
      "key": "user1",
      "success": true
    },
    {
      "namespace": "default",
      "collection": "users",
      "key": "user2",
      "success": false,
      "error": "validation failed"
    }
  ],
  "success_count": 1,
  "failure_count": 1,
  "timestamp": "2026-02-03T12:34:56Z"
}
```

### Batch Limits

- Maximum 1000 operations per batch
- Individual operation failures don't stop batch
- Return detailed results for each operation

## Versioning

### API Versioning Strategy

**URL-based** (current)
```
/api/v1/kv/{namespace}/{collection}/{key}
/api/v2/kv/{namespace}/{collection}/{key}  # future
```

**Guidelines**
- Major version in URL path
- Breaking changes require version bump
- Maintain backwards compatibility in same version
- Deprecate old versions with notice period

### Breaking Changes

**Examples**
- Changing response structure
- Removing fields
- Changing field types
- Changing error codes

**Non-Breaking Changes**
- Adding new endpoints
- Adding optional parameters
- Adding new response fields
- Adding new error codes

## Documentation

### OpenAPI Specification

All endpoints must be documented in `docs/api-specification.yaml`:

```yaml
/api/v1/kv/{namespace}/{collection}/{key}:
  get:
    summary: Get a value
    parameters:
      - name: namespace
        in: path
        required: true
        schema:
          type: string
    responses:
      '200':
        description: Value retrieved successfully
      '404':
        description: Key not found
```

### Code Examples

Provide examples in multiple languages:
- curl (command line)
- Python (requests)
- JavaScript (fetch)

See `docs/api-examples.md`

## Best Practices

### DO
- ✅ Use consistent response formats
- ✅ Validate all input parameters
- ✅ Return appropriate HTTP status codes
- ✅ Provide helpful error messages
- ✅ Include timestamps in responses
- ✅ Use idempotent operations where possible
- ✅ Document all endpoints
- ✅ Version your API

### DON'T
- ❌ Expose internal errors to clients
- ❌ Use different response formats for same endpoint
- ❌ Ignore error cases
- ❌ Return 200 for errors
- ❌ Make breaking changes without version bump
- ❌ Skip input validation
- ❌ Leak sensitive information in errors

## Commander-Specific Patterns

### Response Timestamps
Always include RFC3339 timestamps:
```go
Timestamp: time.Now().UTC().Format(time.RFC3339)
```

### Error Message Format
Clear, actionable error messages:
```go
// Good
"failed to set key: namespace 'invalid' contains special characters"

// Bad
"error"
"invalid input"
```

### Success Message
Consistent "Successfully" message:
```go
Message: "Successfully"
```

## References

- [REST API Tutorial](https://restfulapi.net/)
- [HTTP Status Codes](https://httpstatuses.com/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [OpenAPI Specification](https://swagger.io/specification/)
