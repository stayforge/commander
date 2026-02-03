# Performance Rules

## Performance Targets

### Edge Device Constraints
- **Memory**: 512MB RAM
- **CPU**: ARM64 (e.g., Raspberry Pi 4)
- **Storage**: Flash-based (SD card)
- **Network**: Intermittent connectivity

### Performance Goals
- **Response Time**: <50ms p99 latency
- **Binary Size**: <20MB (target <15MB)
- **Memory Usage**: <100MB runtime
- **Startup Time**: <1 second

## Binary Size Optimization

### Build Flags

**Standard Build**
```bash
go build -o bin/server ./cmd/server
# Result: ~15-20MB
```

**Optimized Build**
```bash
# Strip debug symbols and disable DWARF
go build -ldflags="-s -w" -trimpath -o bin/server ./cmd/server
# Result: ~10-15MB

# With UPX compression (optional)
upx --best --lzma bin/server
# Result: ~5-8MB (slower startup)
```

### Reduce Dependencies

**Avoid Heavy Dependencies**
```go
// Bad - large dependency for simple task
import "github.com/huge-framework/everything"

// Good - use standard library
import "encoding/json"
```

**Review go.mod Regularly**
```bash
# Check dependency sizes
go mod graph | awk '{print $1}' | sort -u

# Remove unused dependencies
go mod tidy
```

## Memory Optimization

### Avoid Unnecessary Allocations

**String Concatenation**
```go
// Bad - creates many intermediate strings
result := ""
for _, s := range strings {
    result = result + s  // allocates new string each time
}

// Good - pre-allocate buffer
var builder strings.Builder
builder.Grow(estimatedSize)
for _, s := range strings {
    builder.WriteString(s)
}
result := builder.String()
```

**Slice Pre-allocation**
```go
// Bad - grows slice repeatedly
var results []Result
for _, item := range items {
    results = append(results, process(item))
}

// Good - pre-allocate capacity
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
```

### Use sync.Pool for Temporary Objects

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processRequest() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    // Use buffer
}
```

### Limit Memory Growth

**Batch Operations**
```go
// Limit batch size to control memory usage
const MaxBatchSize = 1000

func BatchSetHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req BatchSetRequest
        if err := c.BindJSON(&req); err != nil {
            return
        }
        
        // Enforce limit
        if len(req.Operations) > MaxBatchSize {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: fmt.Sprintf("batch size exceeds maximum of %d", MaxBatchSize),
                Code:    "BATCH_SIZE_EXCEEDED",
            })
            return
        }
        
        // Process in chunks if needed
        chunkSize := 100
        for i := 0; i < len(req.Operations); i += chunkSize {
            end := i + chunkSize
            if end > len(req.Operations) {
                end = len(req.Operations)
            }
            chunk := req.Operations[i:end]
            // Process chunk
        }
    }
}
```

## CPU Optimization

### Avoid Unnecessary Work

**Conditional Execution**
```go
// Bad - always computes, even if not needed
result := expensiveComputation()
if condition {
    use(result)
}

// Good - compute only when needed
if condition {
    result := expensiveComputation()
    use(result)
}
```

**Short-circuit Evaluation**
```go
// Check cheap conditions first
if cheapCheck() && expensiveCheck() {
    // ...
}
```

### Use Goroutines Wisely

**Don't Overuse Goroutines**
```go
// Bad - goroutine overhead for small tasks
for _, item := range items {
    go processItem(item)
}

// Good - use goroutines for I/O-bound operations
results := make(chan Result, len(items))
for _, item := range items {
    go func(item Item) {
        results <- fetchFromNetwork(item)
    }(item)
}
```

## I/O Optimization

### BBolt (Flash Storage)

**Batch Writes**
```go
// Bad - many small writes
for key, value := range data {
    db.Update(func(tx *bolt.Tx) error {
        bucket := tx.Bucket([]byte("data"))
        return bucket.Put([]byte(key), value)
    })
}

// Good - single batch write
db.Update(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("data"))
    for key, value := range data {
        if err := bucket.Put([]byte(key), value); err != nil {
            return err
        }
    }
    return nil
})
```

**Optimize for Flash Storage**
```go
// Configure BBolt for SD cards
db, err := bolt.Open(path, 0600, &bolt.Options{
    NoSync:     false,  // Ensure durability
    NoGrowSync: true,   // Reduce sync on growth
    FreelistType: bolt.FreelistMapType,
})
```

### Network I/O

**Connection Pooling**
```go
// Redis connection pool
client := redis.NewClient(&redis.Options{
    Addr:         uri,
    PoolSize:     10,      // Limit connections
    MinIdleConns: 2,       // Keep some ready
    MaxRetries:   3,       // Retry on failure
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})

// MongoDB connection pool
clientOpts := options.Client().
    ApplyURI(uri).
    SetMaxPoolSize(50).    // Limit for edge devices
    SetMinPoolSize(5)
```

**Timeouts**
```go
// Set reasonable timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := kvStore.Get(ctx, namespace, collection, key)
```

## Caching

### In-Memory Cache (Future)

**LRU Cache**
```go
type Cache struct {
    data     map[string]*CacheEntry
    maxSize  int
    lruList  *list.List
}

type CacheEntry struct {
    key       string
    value     []byte
    element   *list.Element
    expiresAt time.Time
}

func (c *Cache) Get(key string) ([]byte, bool) {
    if entry, ok := c.data[key]; ok {
        if time.Now().Before(entry.expiresAt) {
            // Move to front (most recently used)
            c.lruList.MoveToFront(entry.element)
            return entry.value, true
        }
        // Expired, remove
        c.remove(key)
    }
    return nil, false
}
```

**Cache Middleware**
```go
func CacheMiddleware(cache *Cache) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only cache GET requests
        if c.Request.Method != "GET" {
            c.Next()
            return
        }
        
        key := c.Request.URL.String()
        if value, ok := cache.Get(key); ok {
            c.Data(http.StatusOK, "application/json", value)
            return
        }
        
        // Proceed with handler
        c.Next()
        
        // Cache response
        if c.Writer.Status() == http.StatusOK {
            cache.Set(key, c.Writer.Body(), 60*time.Second)
        }
    }
}
```

## Profiling

### CPU Profiling

```bash
# Start server with profiling
go run cmd/server/main.go -cpuprofile=cpu.prof

# Or use pprof endpoint
import _ "net/http/pprof"
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Generate profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze
go tool pprof -http=:8080 cpu.prof
```

### Memory Profiling

```bash
# Heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof -http=:8080 heap.prof

# Allocation profile
curl http://localhost:6060/debug/pprof/allocs > allocs.prof
go tool pprof allocs.prof
```

### Benchmarking

```go
func BenchmarkGetKV(b *testing.B) {
    mockKV := NewMockKV()
    ctx := context.Background()
    
    // Setup
    testValue := []byte("test value")
    mockKV.Set(ctx, "default", "test", "key1", testValue)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = mockKV.Get(ctx, "default", "test", "key1")
    }
}

// Run benchmark
go test -bench=. -benchmem ./internal/handlers
```

## Monitoring

### Metrics to Track

**Response Time**
```go
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        // Record metric
        recordLatency(c.Request.URL.Path, duration)
    }
}
```

**Memory Usage**
```go
import "runtime"

func getMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Alloc = %v MB", m.Alloc / 1024 / 1024)
    log.Printf("TotalAlloc = %v MB", m.TotalAlloc / 1024 / 1024)
    log.Printf("Sys = %v MB", m.Sys / 1024 / 1024)
    log.Printf("NumGC = %v", m.NumGC)
}
```

**Connection Pool**
```go
// Redis
stats := client.PoolStats()
log.Printf("Hits=%d Misses=%d Timeouts=%d TotalConns=%d IdleConns=%d",
    stats.Hits, stats.Misses, stats.Timeouts,
    stats.TotalConns, stats.IdleConns)
```

## Load Testing

### Test Scenarios

```bash
# Using vegeta
echo "GET http://localhost:8080/api/v1/kv/default/users/user1" | \
  vegeta attack -duration=30s -rate=100 | \
  vegeta report

# Using ab (Apache Bench)
ab -n 10000 -c 100 http://localhost:8080/health

# Using k6
k6 run --vus 100 --duration 30s load-test.js
```

### k6 Script Example

```javascript
import http from 'k6/http';
import { check } from 'k6';

export default function() {
  const res = http.get('http://localhost:8080/api/v1/kv/default/users/user1');
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 50ms': (r) => r.timings.duration < 50,
  });
}
```

## Best Practices

### DO
- ✅ Profile before optimizing
- ✅ Set timeouts on all operations
- ✅ Use connection pooling
- ✅ Batch operations when possible
- ✅ Pre-allocate slices and maps
- ✅ Monitor memory usage
- ✅ Test on target hardware (Raspberry Pi)
- ✅ Use benchmarks to verify improvements

### DON'T
- ❌ Premature optimization
- ❌ Allocate in hot paths
- ❌ Block on I/O without timeout
- ❌ Create goroutines without limits
- ❌ Ignore memory pressure
- ❌ Skip profiling
- ❌ Assume cloud performance
- ❌ Forget about startup time

## Edge Device Specific

### Raspberry Pi Optimization

**ARM64 Build**
```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/server-arm64 ./cmd/server
```

**Systemd Resource Limits**
```ini
[Service]
MemoryMax=256M
MemoryHigh=200M
CPUQuota=50%
```

**Monitoring**
```bash
# Check memory
free -h

# Check CPU
top -p $(pgrep server)

# Check disk I/O
iostat -x 1

# Check network
iftop
```

## References

- [Go Performance Tips](https://github.com/golang/go/wiki/Performance)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
- [pprof Documentation](https://pkg.go.dev/runtime/pprof)
