# Database Rules

## KV Interface

### Interface Definition

All database implementations must satisfy the `kv.KV` interface:

```go
type KV interface {
    Get(ctx context.Context, namespace, collection, key string) ([]byte, error)
    Set(ctx context.Context, namespace, collection, key string, value []byte) error
    Delete(ctx context.Context, namespace, collection, key string) error
    Exists(ctx context.Context, namespace, collection, key string) (bool, error)
    Close() error
    Ping(ctx context.Context) error
}
```

### Implementation Guidelines

**Context Handling**
- Always respect context cancellation
- Use context for timeout control
- Pass context to underlying operations

**Error Handling**
- Return `kv.ErrKeyNotFound` for missing keys
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Don't panic on errors

**Resource Management**
- Implement proper `Close()` method
- Clean up connections in `Close()`
- Use `defer` for cleanup

## Three Backend Implementations

### 1. BBolt (Embedded Database)

**Data Mapping**
- Namespace → Separate `.db` file
- Collection → Bucket within file
- Key → Bucket key
- Value → Bucket value (JSON bytes)

**File Structure**
```
/var/lib/stayforge/commander/
├── default.db       # default namespace
├── production.db    # production namespace
└── test.db          # test namespace
```

**Configuration**
```go
type KVConfig struct {
    BackendType BackendType
    BBoltPath   string  // e.g., "/var/lib/stayforge/commander"
}
```

**Best For**
- Edge devices
- Single-node deployments
- No external dependencies
- Development environments

**Limitations**
- Single-node only
- No built-in replication
- File-based locking

### 2. Redis (In-Memory Database)

**Data Mapping**
- Key format: `{namespace}:{collection}:{key}`
- Value: JSON string
- Example: `default:users:user1` → `{"name":"John"}`

**Configuration**
```go
type KVConfig struct {
    BackendType BackendType
    RedisURI    string  // e.g., "redis://localhost:6379/0"
}
```

**Connection String Examples**
```
redis://localhost:6379/0
redis://:password@localhost:6379/0
redis://user:pass@redis-server:6380/1
```

**Best For**
- High-performance caching
- Session storage
- Distributed systems
- High concurrency

**Limitations**
- In-memory (potential data loss)
- Memory constraints
- Requires external Redis server

### 3. MongoDB (Cloud Database)

**Data Mapping**
- Namespace → Database
- Collection → Collection
- Document: `{"key": "user1", "value": "{...}"}`

**Configuration**
```go
type KVConfig struct {
    BackendType BackendType
    MongoURI    string  // e.g., "mongodb+srv://..."
}
```

**Connection String**
```
mongodb+srv://username:password@cluster.mongodb.net/
```

**Best For**
- Cloud deployments
- Distributed systems
- Complex queries
- Automatic backups

**Limitations**
- Requires external MongoDB service
- Network latency
- Cost considerations

## Factory Pattern

### Database Selection

```go
func NewKV(cfg *config.Config) (kv.KV, error) {
    switch cfg.KV.BackendType {
    case config.BackendBBolt:
        return bbolt.NewBBoltKV(cfg.KV.BBoltPath)
    case config.BackendMongoDB:
        return mongodb.NewMongoKV(cfg.KV.MongoURI)
    case config.BackendRedis:
        return redis.NewRedisKV(cfg.KV.RedisURI)
    default:
        return nil, fmt.Errorf("unsupported backend: %s", cfg.KV.BackendType)
    }
}
```

### Configuration

```bash
# BBolt (default)
DATABASE=bbolt
DATA_PATH=/var/lib/stayforge/commander

# MongoDB
DATABASE=mongodb
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/

# Redis
DATABASE=redis
REDIS_URI=redis://localhost:6379/0
```

## Data Organization

### Namespace Guidelines

**Naming**
- Lowercase, alphanumeric
- Use hyphens, not underscores
- Meaningful names: `production`, `staging`, `test`
- Default namespace: `"default"`

**Examples**
```
default       # Default namespace
production    # Production environment
staging       # Staging environment
user-123      # User-specific namespace
```

### Collection Guidelines

**Purpose**
- Group related data
- Logical categorization
- Similar to database tables

**Naming**
- Plural nouns: `users`, `sessions`, `configs`
- Lowercase
- Descriptive

**Examples**
```
users         # User data
sessions      # Session data
configs       # Configuration
cache         # Cached data
```

### Key Guidelines

**Format**
- Any string
- Use meaningful identifiers
- Consider prefixes for organization

**Examples**
```
user_123
session_abc123def456
config:app:database
cache:report:2026-02
```

## Context Usage

### Timeout Control

```go
// Set timeout for operation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := kvStore.Get(ctx, namespace, collection, key)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
}
```

### Cancellation

```go
// Respect context cancellation
func (k *KVStore) Get(ctx context.Context, ns, col, key string) ([]byte, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Proceed with operation
}
```

## Error Handling

### Standard Errors

```go
var (
    ErrKeyNotFound = errors.New("key not found")
    ErrConnectionFailed = errors.New("connection failed")
)
```

### Error Checking

```go
value, err := kvStore.Get(ctx, ns, col, key)
if err != nil {
    if errors.Is(err, kv.ErrKeyNotFound) {
        // Handle key not found
        return nil, fmt.Errorf("key not found: %s", key)
    }
    // Handle other errors
    return nil, fmt.Errorf("get failed: %w", err)
}
```

### Error Wrapping

```go
// Wrap errors with context
return fmt.Errorf("failed to get key %s from collection %s: %w", key, collection, err)
```

## Transaction Handling (Future)

### BBolt Transactions

```go
// Read-write transaction
err := db.Update(func(tx *bolt.Tx) error {
    bucket, err := tx.CreateBucketIfNotExists([]byte(collection))
    if err != nil {
        return err
    }
    return bucket.Put([]byte(key), value)
})
```

### MongoDB Transactions

```go
// Multi-document transaction
session, err := client.StartSession()
defer session.EndSession(ctx)

err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
    // Operations in transaction
})
```

## Connection Management

### Connection Pooling

**Redis**
```go
// Configure connection pool
client := redis.NewClient(&redis.Options{
    Addr:         uri,
    PoolSize:     10,
    MinIdleConns: 2,
})
```

**MongoDB**
```go
// Configure connection pool
clientOpts := options.Client().
    ApplyURI(uri).
    SetMaxPoolSize(100).
    SetMinPoolSize(10)
```

### Health Checks

```go
func (k *KVStore) Ping(ctx context.Context) error {
    // Implement health check
    // Return error if connection is down
}
```

## Best Practices

### DO
- ✅ Always use context for operations
- ✅ Handle `ErrKeyNotFound` explicitly
- ✅ Close connections properly
- ✅ Implement health checks
- ✅ Use connection pooling
- ✅ Normalize namespace to "default" if empty
- ✅ Store values as JSON bytes

### DON'T
- ❌ Ignore context cancellation
- ❌ Leave connections open
- ❌ Hard-code database paths
- ❌ Skip error handling
- ❌ Use blocking operations without timeout
- ❌ Store binary data without encoding

## Testing Database Implementations

### Mock Implementation

```go
type MockKV struct {
    data map[string]map[string]map[string][]byte
}

func (m *MockKV) Get(ctx context.Context, ns, col, key string) ([]byte, error) {
    if val, ok := m.data[ns][col][key]; ok {
        return val, nil
    }
    return nil, kv.ErrKeyNotFound
}
```

### Integration Tests

```go
// +build integration

func TestBBoltIntegration(t *testing.T) {
    tempDir := t.TempDir()
    kvStore, err := bbolt.NewBBoltKV(tempDir)
    require.NoError(t, err)
    defer kvStore.Close()
    
    // Test operations
}
```

## Performance Considerations

### BBolt
- Optimize for sequential writes
- Use batching for bulk operations
- Consider page size for flash storage

### Redis
- Use pipelining for multiple operations
- Consider memory limits
- Monitor connection pool

### MongoDB
- Create indexes for frequently accessed fields
- Use bulk operations
- Monitor connection pool size

## References

- [BBolt Documentation](https://github.com/etcd-io/bbolt)
- [Redis Go Client](https://github.com/redis/go-redis)
- [MongoDB Go Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver)
