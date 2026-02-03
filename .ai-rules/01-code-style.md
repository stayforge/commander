# Code Style Rules

## Go Best Practices

### Naming Conventions

**Variables**
- Use camelCase for local variables: `userName`, `kvStore`
- Use descriptive names: `getUserByID` not `getU`
- Avoid single-letter names except in loops

**Functions**
- Exported functions: PascalCase: `GetKVHandler`, `NewMockKV`
- Unexported functions: camelCase: `marshalJSON`, `parseStringToInt`
- Verb-first for actions: `setKV`, `deleteNamespace`, `validateInput`

**Types**
- Exported types: PascalCase: `KVResponse`, `ErrorResponse`
- Suffix with purpose: `KVRequestBody`, `BatchSetRequest`

**Constants**
- ALL_CAPS with underscores: `DEFAULT_NAMESPACE`
- Or package-scoped: `DefaultNamespace`

### File Organization

**Package Structure**
```go
// 1. Package declaration
package handlers

// 2. Imports (grouped)
import (
    // Standard library
    "context"
    "errors"
    "net/http"
    
    // Internal packages
    "commander/internal/kv"
    
    // External packages
    "github.com/gin-gonic/gin"
)

// 3. Constants
const (
    DefaultTimeout = 5 * time.Second
)

// 4. Types
type KVResponse struct { ... }

// 5. Functions (public first, then private)
func GetKVHandler() { ... }
func validateParams() { ... }
```

### Function Guidelines

**Length**
- Keep functions under 50 lines
- Extract complex logic into helper functions
- One function = one responsibility

**Parameters**
- Max 3-4 parameters
- Use structs for complex parameter groups
- Context always first: `func DoSomething(ctx context.Context, ...)`

**Return Values**
- Error always last: `func Get() ([]byte, error)`
- Use named returns for documentation
- Don't ignore errors

**Example**
```go
// Good
func GetKVHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Handler logic
    }
}

// Bad - too many parameters
func GetKV(ctx context.Context, ns string, col string, key string, db kv.KV, timeout time.Duration) error
```

### Error Handling

**Never Ignore Errors**
```go
// Good
if err := kvStore.Set(ctx, ns, col, key, value); err != nil {
    return fmt.Errorf("failed to set key: %w", err)
}

// Bad
kvStore.Set(ctx, ns, col, key, value) // ignoring error
```

**Custom Errors**
```go
var (
    ErrKeyNotFound = errors.New("key not found")
    ErrInvalidParams = errors.New("invalid parameters")
)

// Use errors.Is for checking
if errors.Is(err, kv.ErrKeyNotFound) {
    // handle
}
```

**Error Wrapping**
```go
return fmt.Errorf("failed to get value from %s: %w", collection, err)
```

### Comments

**Package Comments**
```go
// Package handlers provides HTTP request handlers for the Commander API.
// It includes CRUD operations, batch operations, and namespace management.
package handlers
```

**Function Comments** (exported only)
```go
// GetKVHandler handles GET /api/v1/kv/{namespace}/{collection}/{key}
// Retrieves a value from the KV store by namespace, collection, and key.
// Returns 404 if the key does not exist.
func GetKVHandler(kvStore kv.KV) gin.HandlerFunc {
```

**Inline Comments** (sparingly)
```go
// Normalize namespace to "default" if empty
namespace = kv.NormalizeNamespace(namespace)
```

### Code Formatting

**Use gofmt**
```bash
go fmt ./...
gofmt -w .
```

**Line Length**
- Aim for 100 characters
- Break at logical points
- Align parameters/arguments

**Spacing**
```go
// Good
if condition {
    doSomething()
}

for i := 0; i < n; i++ {
    process(i)
}

// Group related declarations
type (
    Request  struct { ... }
    Response struct { ... }
)
```

### Imports

**Order**
1. Standard library
2. Internal packages
3. External packages

**Grouping**
```go
import (
    "context"
    "errors"
    "net/http"
    
    "commander/internal/kv"
    "commander/internal/config"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)
```

**Avoid dot imports**
```go
// Bad
import . "github.com/gin-gonic/gin"

// Good
import "github.com/gin-gonic/gin"
```

### JSON Handling

**Struct Tags**
```go
type KVResponse struct {
    Message   string      `json:"message"`
    Namespace string      `json:"namespace"`
    Value     interface{} `json:"value,omitempty"` // omit if empty
    Timestamp string      `json:"timestamp"`
}
```

**Validation Tags**
```go
type KVRequestBody struct {
    Value interface{} `json:"value" binding:"required"`
}
```

### Concurrency

**Use Context**
```go
func GetValue(ctx context.Context, key string) ([]byte, error) {
    // Respect context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Actual work
}
```

**Avoid Goroutine Leaks**
```go
// Good - with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

## Project-Specific Patterns

### Handler Pattern
```go
func SomeHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract parameters
        namespace := c.Param("namespace")
        
        // 2. Validate
        if namespace == "" {
            c.JSON(http.StatusBadRequest, ErrorResponse{...})
            return
        }
        
        // 3. Process
        ctx := c.Request.Context()
        result, err := kvStore.Get(ctx, namespace, collection, key)
        if err != nil {
            // Handle error
            return
        }
        
        // 4. Respond
        c.JSON(http.StatusOK, KVResponse{...})
    }
}
```

### Response Pattern
```go
// Success
c.JSON(http.StatusOK, KVResponse{
    Message:   "Successfully",
    Namespace: namespace,
    Key:       key,
    Value:     value,
    Timestamp: time.Now().UTC().Format(time.RFC3339),
})

// Error
c.JSON(http.StatusBadRequest, ErrorResponse{
    Message: "detailed error message",
    Code:    "ERROR_CODE",
})
```

### Testing Pattern
```go
func TestSomething(t *testing.T) {
    // Setup
    mockKV := NewMockKV()
    
    // Test cases
    tests := []struct {
        name           string
        input          string
        expectedStatus int
    }{
        {"valid input", "test", http.StatusOK},
        {"invalid input", "", http.StatusBadRequest},
    }
    
    // Run tests
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
            assert.Equal(t, tt.expectedStatus, actualStatus)
        })
    }
}
```

## Linting

Must pass `golangci-lint`:
```bash
golangci-lint run
```

Configuration in `.golangci.yml`

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
