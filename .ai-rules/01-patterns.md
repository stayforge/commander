# Commander Code Patterns

Project-specific patterns and conventions. For general Go best practices, follow standard Go conventions.

## Handler Pattern

All HTTP handlers follow this structure:

```go
func SomeHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract parameters
        namespace := c.Param("namespace")
        collection := c.Param("collection")
        key := c.Param("key")

        // 2. Validate
        if namespace == "" || collection == "" || key == "" {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "namespace, collection, and key are required",
                Code:    "INVALID_PARAMS",
            })
            return
        }

        // 3. Normalize
        namespace = kv.NormalizeNamespace(namespace)

        // 4. Process
        ctx := c.Request.Context()
        result, err := kvStore.Get(ctx, namespace, collection, key)
        if err != nil {
            if errors.Is(err, kv.ErrKeyNotFound) {
                c.JSON(http.StatusNotFound, ErrorResponse{
                    Message: "key not found",
                    Code:    "KEY_NOT_FOUND",
                })
                return
            }
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Message: "internal error",
                Code:    "INTERNAL_ERROR",
            })
            return
        }

        // 5. Respond
        c.JSON(http.StatusOK, KVResponse{
            Message:   "Successfully",
            Namespace: namespace,
            Key:       key,
            Value:     result,
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
    }
}
```

## Response Formats

**Success**: include message, namespace, key, value, timestamp (RFC3339 UTC).

**Error**: include message (user-friendly) and code (machine-readable).

Error codes: `KEY_NOT_FOUND`, `INVALID_PARAMS`, `INVALID_BODY`, `INTERNAL_ERROR`

## KV Backend Implementation

Each backend implements `kv.KV` interface with this data mapping:

| Backend | Namespace        | Collection           | Key         |
|---------|------------------|----------------------|-------------|
| BBolt   | Separate .db file | Bucket              | Bucket key  |
| MongoDB | Database         | Collection           | Document    |
| Redis   | Key prefix       | Key prefix           | Full key    |

Redis key format: `{namespace}:{collection}:{key}`

## API URL Structure

```
GET/POST/DELETE/HEAD /api/v1/kv/{namespace}/{collection}/{key}
POST/DELETE          /api/v1/kv/batch
GET                  /api/v1/namespaces
GET                  /api/v1/namespace/{namespace}/collections
```

Card verification (MongoDB only):
```
POST /api/v1/namespace/{namespace}
POST /api/v1/namespace/{namespace}/device/{device_name}/vguang
```

## Import Order

```go
import (
    // Standard library
    "context"
    "net/http"

    // Internal packages
    "commander/internal/kv"

    // External packages
    "github.com/gin-gonic/gin"
)
```
