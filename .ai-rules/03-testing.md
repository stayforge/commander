# Testing Rules

## Test Coverage Requirements

### Coverage Goals
- **Overall Project**: 85%+
- **Handlers Package**: 90%+
- **New Code**: Must include tests
- **Critical Paths**: 100% coverage

### Current Status
- Overall: 64.6%
- Handlers: 75.8%
- Config: 100% ✅
- KV Interface: 100% ✅

## Test Structure

### File Naming
```
handlers.go       → handlers_test.go
kv.go            → kv_test.go
batch.go         → batch_test.go
```

### Test Function Naming
```go
// Format: TestFunctionName
func TestGetKVHandler(t *testing.T) { ... }

// With subtests: TestFunctionName_Scenario
func TestGetKVHandler_KeyNotFound(t *testing.T) { ... }

// Table-driven: TestFunctionName with t.Run
func TestGetKVHandler(t *testing.T) {
    tests := []struct{
        name string
        // ...
    }{
        {"successful get", ...},
        {"key not found", ...},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Testing Patterns

### Table-Driven Tests (Preferred)

```go
func TestSetKVHandler(t *testing.T) {
    mockKV := NewMockKV()
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/api/v1/kv/:namespace/:collection/:key", SetKVHandler(mockKV))

    tests := []struct {
        name           string
        namespace      string
        collection     string
        key            string
        body           KVRequestBody
        expectedStatus int
    }{
        {
            name:       "successful set",
            namespace:  "default",
            collection: "users",
            key:        "user1",
            body:       KVRequestBody{Value: map[string]interface{}{"name": "John"}},
            expectedStatus: http.StatusCreated,
        },
        {
            name:       "invalid namespace",
            namespace:  "",
            collection: "users",
            key:        "user1",
            body:       KVRequestBody{Value: "test"},
            expectedStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            bodyJSON, _ := json.Marshal(tt.body)
            req, _ := http.NewRequest("POST", 
                fmt.Sprintf("/api/v1/kv/%s/%s/%s", tt.namespace, tt.collection, tt.key),
                bytes.NewBuffer(bodyJSON))
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

### Mock Pattern

**MockKV Implementation**
```go
type MockKV struct {
    data map[string]map[string]map[string][]byte
}

func NewMockKV() *MockKV {
    return &MockKV{
        data: make(map[string]map[string]map[string][]byte),
    }
}

func (m *MockKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
    if ns, ok := m.data[namespace]; ok {
        if coll, ok := ns[collection]; ok {
            if val, ok := coll[key]; ok {
                return val, nil
            }
        }
    }
    return nil, kv.ErrKeyNotFound
}

// Implement other methods...
```

**Usage**
```go
func TestSomething(t *testing.T) {
    mockKV := NewMockKV()
    
    // Setup test data
    ctx := context.Background()
    testValue, _ := json.Marshal("test")
    _ = mockKV.Set(ctx, "default", "users", "user1", testValue)
    
    // Test
    value, err := mockKV.Get(ctx, "default", "users", "user1")
    assert.NoError(t, err)
    assert.NotNil(t, value)
}
```

## Test Categories

### Unit Tests
Test individual functions in isolation.

```go
func TestNormalizeNamespace(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"", "default"},
        {"custom", "custom"},
        {"default", "default"},
    }
    
    for _, tt := range tests {
        result := kv.NormalizeNamespace(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}
```

### Handler Tests
Test HTTP handlers with mock KV store.

```go
func TestGetKVHandler(t *testing.T) {
    mockKV := NewMockKV()
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.GET("/api/v1/kv/:namespace/:collection/:key", GetKVHandler(mockKV))
    
    // Setup test data
    ctx := context.Background()
    testValue, _ := json.Marshal(map[string]interface{}{"name": "test"})
    _ = mockKV.Set(ctx, "default", "users", "user1", testValue)
    
    // Make request
    req, _ := http.NewRequest("GET", "/api/v1/kv/default/users/user1", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var resp KVResponse
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NoError(t, err)
    assert.Equal(t, "user1", resp.Key)
}
```

### Integration Tests (Future)
Test with real databases (BBolt, Redis, MongoDB).

```go
// +build integration

func TestBBoltIntegration(t *testing.T) {
    // Setup real BBolt database
    tempDir := t.TempDir()
    cfg := &config.Config{
        KV: config.KVConfig{
            BackendType: config.BackendBBolt,
            BBoltPath:   tempDir,
        },
    }
    
    kvStore, err := database.NewKV(cfg)
    require.NoError(t, err)
    defer kvStore.Close()
    
    // Test actual operations
    ctx := context.Background()
    err = kvStore.Set(ctx, "test", "col", "key", []byte("value"))
    assert.NoError(t, err)
    
    value, err := kvStore.Get(ctx, "test", "col", "key")
    assert.NoError(t, err)
    assert.Equal(t, []byte("value"), value)
}
```

## Test Assertions

### Using testify/assert

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // assert - continues on failure
    assert.Equal(t, expected, actual, "should be equal")
    assert.NotNil(t, obj)
    assert.NoError(t, err)
    assert.True(t, condition)
    
    // require - stops on failure
    require.NoError(t, err, "critical error")
    require.NotNil(t, obj, "must not be nil")
}
```

### Common Assertions

```go
// Equality
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)

// Nil checks
assert.Nil(t, obj)
assert.NotNil(t, obj)

// Errors
assert.NoError(t, err)
assert.Error(t, err)
assert.EqualError(t, err, "expected error message")
assert.ErrorIs(t, err, kv.ErrKeyNotFound)

// HTTP Status
assert.Equal(t, http.StatusOK, w.Code)

// JSON
var resp Response
err := json.Unmarshal(w.Body.Bytes(), &resp)
assert.NoError(t, err)
assert.Equal(t, "expected", resp.Field)

// Collections
assert.Len(t, slice, 3)
assert.Contains(t, slice, item)
assert.Empty(t, slice)

// Types
assert.IsType(t, (*MyType)(nil), obj)
```

## Test Coverage

### Measure Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Coverage by function
go tool cover -func=coverage.out
```

### Coverage Requirements

**Must Cover**
- All exported functions
- All error paths
- All edge cases
- All HTTP status codes

**Example**
```go
func TestGetKVHandler_AllPaths(t *testing.T) {
    tests := []struct {
        name           string
        setup          func(*MockKV)
        namespace      string
        expectedStatus int
    }{
        {
            name:           "successful get",
            setup:          func(m *MockKV) { /* setup data */ },
            namespace:      "default",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "key not found",
            setup:          func(m *MockKV) { /* no data */ },
            namespace:      "default",
            expectedStatus: http.StatusNotFound,
        },
        {
            name:           "invalid parameters",
            namespace:      "",
            expectedStatus: http.StatusBadRequest,
        },
        // Cover all paths
    }
    // ...
}
```

## Test Data

### Fixtures

```go
// Test data
var (
    testUser = map[string]interface{}{
        "id":    1,
        "name":  "Test User",
        "email": "test@example.com",
    }
    
    testConfig = map[string]interface{}{
        "host": "localhost",
        "port": 8080,
    }
)

func setupTestData(mockKV *MockKV) {
    ctx := context.Background()
    userData, _ := json.Marshal(testUser)
    _ = mockKV.Set(ctx, "default", "users", "user1", userData)
}
```

### Cleanup

```go
func TestWithCleanup(t *testing.T) {
    mockKV := NewMockKV()
    
    // Setup
    setupTestData(mockKV)
    
    // Cleanup
    t.Cleanup(func() {
        mockKV.Close()
    })
    
    // Test
    // ...
}
```

## Running Tests

### Commands

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/handlers

# Run specific test
go test ./internal/handlers -run TestGetKVHandler

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...

# Race detection
go test -race ./...

# Short mode (skip long tests)
go test -short ./...

# Parallel execution
go test -parallel 4 ./...
```

### Test Modes

```go
// Skip in short mode
func TestLongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping in short mode")
    }
    // long-running test
}

// Parallel test
func TestParallel(t *testing.T) {
    t.Parallel()
    // test logic
}
```

## Best Practices

### DO
- ✅ Write tests before or with implementation
- ✅ Use table-driven tests for multiple scenarios
- ✅ Test all error paths
- ✅ Use meaningful test names
- ✅ Keep tests focused and isolated
- ✅ Use mocks for external dependencies
- ✅ Clean up resources (defer, t.Cleanup)
- ✅ Test edge cases (empty strings, nil, etc.)

### DON'T
- ❌ Skip writing tests
- ❌ Test implementation details
- ❌ Depend on test execution order
- ❌ Use real databases in unit tests
- ❌ Ignore race conditions
- ❌ Write flaky tests
- ❌ Test private functions directly
- ❌ Commit commented-out tests

## Commander-Specific Guidelines

### Handler Testing Pattern
1. Create MockKV
2. Set up Gin test mode
3. Create router with handler
4. Prepare request
5. Execute request
6. Assert response

### Test Coverage Priority
1. New features: 100% coverage required
2. Bug fixes: Add test that reproduces bug
3. Refactoring: Maintain existing coverage
4. Documentation: Update examples

### Example Test Structure
```go
func TestBatchSetHandler(t *testing.T) {
    // Setup
    mockKV := NewMockKV()
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/api/v1/kv/batch", BatchSetHandler(mockKV))

    // Test cases (table-driven)
    tests := []struct {
        name           string
        request        BatchSetRequest
        expectedStatus int
        expectedCount  int
    }{
        // ... test cases
    }

    // Execute tests
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test logic
        })
    }
}
```

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
