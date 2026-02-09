# Testing Guide

## Test Structure

Use table-driven tests with `testify/assert`:

```go
func TestSomeHandler(t *testing.T) {
    mockKV := NewMockKV()
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.GET("/api/v1/kv/:namespace/:collection/:key", GetKVHandler(mockKV))

    tests := []struct {
        name           string
        namespace      string
        collection     string
        key            string
        expectedStatus int
    }{
        {"successful get", "default", "users", "user1", http.StatusOK},
        {"key not found", "default", "users", "missing", http.StatusNotFound},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, _ := http.NewRequest("GET",
                fmt.Sprintf("/api/v1/kv/%s/%s/%s", tt.namespace, tt.collection, tt.key), nil)
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

## MockKV

Use `MockKV` (in-memory map) for unit tests. For integration tests, use `t.TempDir()` with real BBolt.

## Coverage

- Goal: 85%+ overall
- New code must have tests
- Cover all error paths and edge cases

## Commands

```bash
go test ./...                          # Run all
go test -v ./internal/handlers         # Specific package
go test -run TestGetKVHandler ./...    # Specific test
go test -cover ./...                   # With coverage
go test -race ./...                    # Race detection
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out  # Report
```
