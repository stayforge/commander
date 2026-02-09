# Database Abstraction Layer Architecture Documentation

## Overview

Commander employs a **Hexagonal Architecture (also known as Ports & Adapters Pattern)** to design its database abstraction layer, enabling unified support for multiple database backends. This document details the overall design, implementation specifics of each adapter, and how to extend the architecture with new backends.

---

## 1. Overall Architecture Diagram

### Hexagonal Architecture Design

```mermaid
graph TB
    subgraph "API Layer"
        HTTP["HTTP Request<br/>Gin Router"]
    end
    
    subgraph "Handler Layer"
        Handler["Handlers<br/>GetKVHandler<br/>SetKVHandler<br/>DeleteKVHandler<br/>..."]
    end
    
    subgraph "Port (Interface)"
        Port["🔌 KV Interface<br/><br/>Get()<br/>Set()<br/>Delete()<br/>Exists()<br/>Close()<br/>Ping()"]
    end
    
    subgraph "Adapter Layer"
        BBolt["🔌 BBolt Adapter<br/>File System<br/>namespace → file<br/>collection → bucket"]
        Redis["🔌 Redis Adapter<br/>In-Memory Cache<br/>key: ns:coll:key"]
        MongoDB["🔌 MongoDB Adapter<br/>Document Database<br/>ns → db<br/>coll → collection"]
    end
    
    subgraph "Backend Layer"
        BBoltDB["BBolt Database<br/>*.db files"]
        RedisDB["Redis Server<br/>Memory"]
        MongoDB_Actual["MongoDB Server<br/>Cloud/On-Premise"]
    end
    
    HTTP --> Handler
    Handler --> |Depends on<br/>Interface, not impl| Port
    Port --> |Implements| BBolt
    Port --> |Implements| Redis
    Port --> |Implements| MongoDB
    BBolt --> BBoltDB
    Redis --> RedisDB
    MongoDB --> MongoDB_Actual
    
    style Port fill:#4CAF50,stroke:#2E7D32,color:#fff
    style HTTP fill:#2196F3,stroke:#1565C0,color:#fff
    style Handler fill:#FF9800,stroke:#E65100,color:#fff
    style BBolt fill:#9C27B0,stroke:#6A1B9A,color:#fff
    style Redis fill:#FF5722,stroke:#D84315,color:#fff
    style MongoDB fill:#009688,stroke:#00695C,color:#fff
```

**Core Design Principles**:
- **Port (Interface)**: The `kv.KV` interface defines a unified contract for data access
- **Adapters**: Three independent implementations adapting different database backends
- **Dependency Direction**: Handlers depend on the interface (Port), not concrete implementations (Adapters)
- **Benefits**:
  - ✅ Support runtime database switching via environment variables
  - ✅ Easy to test (can mock KV interface)
  - ✅ Easy to add new backends (just implement KV interface)
  - ✅ Business logic decoupled from data storage

---

## 2. KV Interface Definition

### Interface Signature

Location: `internal/kv/kv.go`

```go
type KV interface {
    // Get retrieves a JSON value by key from namespace and collection
    Get(ctx context.Context, namespace, collection, key string) ([]byte, error)
    
    // Set stores a JSON value by key in namespace and collection
    Set(ctx context.Context, namespace, collection, key string, value []byte) error
    
    // Delete removes a key-value pair from namespace and collection
    Delete(ctx context.Context, namespace, collection, key string) error
    
    // Exists checks if a key exists in namespace and collection
    Exists(ctx context.Context, namespace, collection, key string) (bool, error)
    
    // Close closes the connection to the backend
    Close() error
    
    // Ping checks if the connection is alive
    Ping(ctx context.Context) error
}
```

### Interface Methods Reference

| Method | Parameters | Returns | Description |
|--------|-----------|---------|-------------|
| **Get** | namespace, collection, key | ([]byte, error) | Retrieves JSON value; returns `ErrKeyNotFound` if not exists |
| **Set** | namespace, collection, key, value | error | Saves JSON value; overwrites existing value |
| **Delete** | namespace, collection, key | error | Deletes key; returns success even if not exists |
| **Exists** | namespace, collection, key | (bool, error) | Checks if key exists |
| **Close** | - | error | Closes connection and cleans up resources |
| **Ping** | ctx | error | Health check; verifies connection is alive |

### Error Definitions

```go
var (
    ErrKeyNotFound = errors.New("key not found")
    ErrConnectionFailed = errors.New("connection failed")
)
```

### Data Organization

All adapters use a unified logical hierarchy:

```
Namespace (logical isolation)
  ├── Collection 1 (data category)
  │   ├── Key 1 → Value (JSON bytes)
  │   ├── Key 2 → Value (JSON bytes)
  │   └── ...
  ├── Collection 2
  │   ├── Key 1 → Value (JSON bytes)
  │   └── ...
  └── ...
```

**Design Rationale**:
- Namespace: Isolates data for different applications/modules (e.g., app, mobile, admin)
- Collection: Categorizes data within the same namespace (e.g., users, cards, settings)
- Key: Unique identifier for specific data (e.g., user_id, card_number)

---

## 3. Factory Pattern (Dynamic Backend Selection)

### Backend Selection Logic

Location: `internal/database/factory.go`

```go
func NewKV(cfg *config.Config) (kv.KV, error) {
    switch cfg.KV.BackendType {
    case config.BackendMongoDB:
        return mongodb.NewMongoDBKV(cfg.KV.MongoURI)
    case config.BackendRedis:
        return redis.NewRedisKV(cfg.KV.RedisURI)
    case config.BackendBBolt:
        return bbolt.NewBBoltKV(cfg.KV.BBoltPath)
    default:
        return nil, fmt.Errorf("unsupported backend type: %s", cfg.KV.BackendType)
    }
}
```

### Configuration-Driven Selection

```bash
# Select backend in .env file
KV_BACKEND_TYPE=mongodb  # or redis, bbolt

# MongoDB backend configuration
MONGODB_URI=mongodb://localhost:27017

# Redis backend configuration
REDIS_URI=redis://localhost:6379

# BBolt backend configuration
BBOLT_PATH=/data/kv
```

**Advantage**: Switch backends without recompilation; environment variable based selection.

---

## 4. Adapter Implementation Comparison

### Mapping Strategy Comparison Table

| Concept | BBolt | Redis | MongoDB |
|---------|-------|-------|---------|
| **Namespace** | `.db` file in filesystem | Key prefix (1st segment) | Database |
| **Collection** | BBolt Bucket | Key prefix (2nd segment) | Collection |
| **Key** | Key within bucket | Redis Key (3rd segment) | Document `key` field |
| **Value** | Binary bytes | Redis String (bytes) | Document `value` field (string) |
| **Storage Location** | `{BBoltPath}/{namespace}.db` | Single Redis server | MongoDB server |
| **Concurrency Control** | `sync.RWMutex` (per adapter) | Redis atomic ops | MongoDB transactions |
| **Indexing** | No index (O(1) lookup) | Key unique | Auto unique index on `key` |
| **Distributed** | No (local files) | Yes (clustering) | Yes (replica sets) |
| **Use Cases** | Edge devices, development | High-performance cache, real-time | Production, cloud, distributed |

---

## 5. Complete Data Flow - GET Request Example

### Example: GET /api/v1/kv/default/users/user1

```mermaid
sequenceDiagram
    participant Client as HTTP Client
    participant Router as Gin Router
    participant Handler as GetKVHandler
    participant KV_Interface as KV Interface
    participant Adapter as Backend Adapter
    participant DB as Database
    
    Client->>Router: GET /api/v1/kv/default/users/user1
    Router->>Handler: Route Match
    activate Handler
    Handler->>Handler: Parse params<br/>ns=default, coll=users, key=user1
    Handler->>KV_Interface: kvStore.Get(ctx, "default", "users", "user1")
    activate KV_Interface
    
    alt Backend == BBolt
        KV_Interface->>Adapter: Open default.db
        Adapter->>DB: Read users bucket
        DB->>Adapter: Return user1 value
    else Backend == Redis
        KV_Interface->>Adapter: GET "default:users:user1"
        Adapter->>DB: Redis GET command
        DB->>Adapter: Return bytes
    else Backend == MongoDB
        KV_Interface->>Adapter: db.default.users.findOne({key: "user1"})
        Adapter->>DB: MongoDB Query
        DB->>Adapter: Return document
    end
    
    Adapter->>KV_Interface: Return []byte value
    deactivate KV_Interface
    Handler->>Handler: Unmarshal JSON
    Handler->>Handler: Build response
    Handler->>Client: HTTP 200 + JSON
    deactivate Handler
```

---

## 6. BBolt Adapter Implementation Details

### Architecture Characteristics

```mermaid
graph TB
    subgraph "BBolt KV Instance"
        BboltStore["BBoltKV struct<br/>baseDir: string<br/>dbs: map[ns]*bbolt.DB<br/>mu: sync.RWMutex"]
    end
    
    subgraph "Namespace 1 (File)"
        File1["default.db<br/>binary file"]
        Buckets1["Buckets in default.db<br/>├─ users (bucket)<br/>├─ cards (bucket)<br/>└─ settings (bucket)"]
    end
    
    subgraph "Namespace 2 (File)"
        File2["mobile.db<br/>binary file"]
    end
    
    subgraph "Namespace N (File)"
        FileN["..."]
    end
    
    BboltStore -->|lazy load| File1
    BboltStore -->|lazy load| File2
    BboltStore -->|lazy load| FileN
    File1 --> Buckets1
    
    style BboltStore fill:#9C27B0,stroke:#6A1B9A,color:#fff
    style File1 fill:#CE93D8,stroke:#8E24AA,color:#fff
    style Buckets1 fill:#F3E5F5,stroke:#9C27B0
```

### Data Organization

```
{BBoltPath}/
├── default.db              # Namespace: default
│   ├── users bucket        # Collection: users
│   │   ├── user1 → {"name":"Alice","age":30}
│   │   ├── user2 → {"name":"Bob","age":25}
│   │   └── ...
│   └── cards bucket        # Collection: cards
│       ├── card001 → {"room":"101","valid":true}
│       └── ...
├── mobile.db               # Namespace: mobile
│   └── ...
└── admin.db                # Namespace: admin
    └── ...
```

### Key Implementation Details

Location: `internal/database/bbolt/bbolt.go`

**Concurrency Control**:
```go
type BBoltKV struct {
    baseDir string
    dbs map[string]*bbolt.DB  // One connection per namespace
    mu sync.RWMutex           // Protects dbs map
}
```

**Lazy Loading**:
```go
// Opens file only on first namespace access
func (b *BBoltKV) getDB(namespace string) (*bbolt.DB, error) {
    // Read lock for fast path
    b.mu.RLock()
    if db, exists := b.dbs[namespace]; exists {
        b.mu.RUnlock()
        return db, nil
    }
    b.mu.RUnlock()
    
    // Write lock for opening new db
    b.mu.Lock()
    defer b.mu.Unlock()
    
    dbPath := filepath.Join(b.baseDir, fmt.Sprintf("%s.db", namespace))
    db, _ := bbolt.Open(dbPath, 0o600, nil)
    b.dbs[namespace] = db
    return db, nil
}
```

**Advantages**:
- ✅ No external dependencies (no server required)
- ✅ Ideal for edge devices and development environments
- ✅ Native filesystem support with data persistence
- ✅ Low latency (local disk access)

**Limitations**:
- ❌ No distributed support
- ❌ Single-process locking (conflicts with multi-process access)
- ❌ Performance limited by local disk speed

---

## 7. Redis Adapter Implementation Details

### Architecture Characteristics

```mermaid
graph TB
    subgraph "Redis KV Instance"
        RedisStore["RedisKV struct<br/>client: *redis.Client"]
    end
    
    subgraph "Redis Server"
        Server["Redis Memory<br/>Single Process Store"]
    end
    
    subgraph "Key Space (Virtual)"
        Keys["All Keys in Memory<br/>default:users:user1<br/>default:users:user2<br/>default:cards:card001<br/>mobile:settings:theme<br/>..."]
    end
    
    RedisStore -->|TCP Connection| Server
    Server --> Keys
    
    style RedisStore fill:#FF5722,stroke:#D84315,color:#fff
    style Server fill:#FFAB91,stroke:#E64A19,color:#fff
    style Keys fill:#FFF3E0,stroke:#FF6E40
```

### Key Naming Convention

```
Namespace:Collection:Key

Examples:
├── default:users:user1
├── default:users:user2
├── default:cards:card001
├── default:cards:card002
├── mobile:settings:theme
├── mobile:settings:language
└── admin:logs:2024-02-01
```

### Key Implementation Details

Location: `internal/database/redis/redis.go`

**Connection Pool**:
```go
type RedisKV struct {
    client *redis.Client  // Manages connection pool
}
```

**Key Formatting**:
```go
func makeKey(namespace, collection, key string) string {
    return fmt.Sprintf("%s:%s:%s", namespace, collection, key)
}
```

**Operation Examples**:
```go
// Set: Redis SET namespace:collection:key value
func (r *RedisKV) Set(ctx context.Context, ns, coll, key string, value []byte) error {
    redisKey := makeKey(ns, coll, key)
    return r.client.Set(ctx, redisKey, value, 0).Err()
}

// Get: Redis GET namespace:collection:key
func (r *RedisKV) Get(ctx context.Context, ns, coll, key string) ([]byte, error) {
    redisKey := makeKey(ns, coll, key)
    val, err := r.client.Get(ctx, redisKey).Result()
    if err == redis.Nil {
        return nil, kv.ErrKeyNotFound
    }
    return []byte(val), err
}
```

**Advantages**:
- ✅ Ultra-high performance (<1ms latency)
- ✅ Cluster support (distributed caching)
- ✅ Rich data structures (List, Set, Hash, etc.)
- ✅ Native transaction support

**Limitations**:
- ❌ Memory capacity constraints
- ❌ Data loss risk (requires persistence configuration)
- ❌ Requires external Redis server

**Use Cases**:
- Real-time applications, high-concurrency read/write
- Cache layer
- Session storage
- Queue systems

---

## 8. MongoDB Adapter Implementation Details

### Architecture Characteristics

```mermaid
graph TB
    subgraph "MongoDB KV Instance"
        MongoStore["MongoDBKV struct<br/>client: *mongo.Client<br/>uri: string"]
    end
    
    subgraph "MongoDB Server"
        Server["MongoDB Atlas / On-Prem"]
    end
    
    subgraph "Database Level"
        DB1["Database: default<br/>Collections:<br/>├─ users<br/>├─ cards<br/>└─ settings"]
        DB2["Database: mobile<br/>Collections:<br/>├─ settings<br/>└─ logs"]
    end
    
    subgraph "Collection Level (Example)"
        Coll["Collection: users<br/>Documents:<br/>├─ {_id:..., key:'user1', value:'...'}<br/>├─ {_id:..., key:'user2', value:'...'}<br/>└─ ..."]
    end
    
    MongoStore -->|TCP Connection| Server
    Server -->|ns=default| DB1
    Server -->|ns=mobile| DB2
    DB1 --> Coll
    
    style MongoStore fill:#009688,stroke:#00695C,color:#fff
    style Server fill:#80CBC4,stroke:#00897B,color:#fff
    style DB1 fill:#B2DFDB,stroke:#26A69A
    style Coll fill:#E0F2F1,stroke:#00ACC1
```

### Document Structure

**MongoDB Document Structure**:
```json
{
    "_id": ObjectId("..."),        // Auto-generated by MongoDB
    "key": "user1",                // Our key field
    "value": "{\"name\":\"Alice\"}", // JSON string
    "created_at": ISODate("..."),  // Creation timestamp
    "updated_at": ISODate("...")   // Update timestamp
}
```

**Multi-level Mapping**:
```
MongoDB Layer          | KV Layer
namespace → Database
collection → Collection
key → Document.key field
value → Document.value field
```

### Key Implementation Details

Location: `internal/database/mongodb/mongodb.go`

**Connection Management**:
```go
type MongoDBKV struct {
    client *mongo.Client  // Single connection managing all operations
    uri    string
}
```

**Index Creation**:
```go
// Create unique index on key for each collection
func (m *MongoDBKV) ensureIndex(ctx context.Context, coll *mongo.Collection) error {
    indexModel := mongo.IndexModel{
        Keys: bson.D{{Key: "key", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    _, err := coll.Indexes().CreateOne(ctx, indexModel)
    return err
}
```

**Get Operation**:
```go
func (m *MongoDBKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
    coll := m.getCollection(namespace, collection)
    m.ensureIndex(ctx, coll)
    
    var doc struct {
        Key   string `bson:"key"`
        Value string `bson:"value"`
    }
    
    err := coll.FindOne(ctx, bson.M{"key": key}).Decode(&doc)
    if err == mongo.ErrNoDocuments {
        return nil, kv.ErrKeyNotFound
    }
    
    return []byte(doc.Value), err
}
```

**Advantages**:
- ✅ Fully managed (cloud services like Atlas)
- ✅ Auto replica sets and failover
- ✅ Support for complex queries (extensible features)
- ✅ High availability and security
- ✅ Unlimited capacity

**Limitations**:
- ❌ Network latency (compared to local storage)
- ❌ Requires external service
- ❌ Potentially higher costs

**Use Cases**:
- Production environments
- Cloud deployment
- Distributed systems
- Applications requiring high availability

---

## 9. Complete Data Flow Example

### Scenario: Storing Card Data

#### Step 1: Configuration Selection (main.go)

```go
cfg := config.LoadConfig()
// KV_BACKEND_TYPE=mongodb read from .env
kvStore, _ := database.NewKV(cfg)
// Returns MongoDBKV instance
```

#### Step 2: HTTP Request

```bash
POST /api/v1/kv/default/cards/card001
Content-Type: application/json

{
  "value": {
    "room_number": "101",
    "valid_from": "2026-02-01",
    "valid_until": "2026-02-05",
    "status": "active"
  }
}
```

#### Step 3: Handler Processing

```go
// handlers/kv.go
func SetKVHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse parameters
        ns := c.Param("namespace")        // "default"
        coll := c.Param("collection")     // "cards"
        key := c.Param("key")             // "card001"
        
        // Parse JSON body
        var req KVRequestBody
        c.BindJSON(&req)
        
        // Encode to JSON bytes
        valueBytes, _ := json.Marshal(req.Value)
        
        // Call KV interface (agnostic to implementation)
        err := kvStore.Set(c.Request.Context(), ns, coll, key, valueBytes)
        
        // Return result
        c.JSON(200, KVResponse{...})
    }
}
```

#### Step 4: MongoDB Adapter Execution

```go
// internal/database/mongodb/mongodb.go
func (m *MongoDBKV) Set(ctx context.Context, ns, coll, key string, value []byte) error {
    collection := m.getCollection(ns, coll)     // db: default, collection: cards
    m.ensureIndex(ctx, collection)              // Ensure key uniqueness
    
    doc := bson.M{
        "key": key,                             // "card001"
        "value": string(value),                 // JSON string
        "created_at": time.Now(),
        "updated_at": time.Now(),
    }
    
    // MongoDB operation: upsert
    opts := options.Update().SetUpsert(true)
    _, err := collection.UpdateOne(
        ctx,
        bson.M{"key": key},
        bson.D{{Key: "$set", Value: doc}},
        opts,
    )
    return err
}
```

#### Step 5: MongoDB Storage Result

```javascript
// MongoDB database view
use default
db.cards.find()
// Result:
{
  "_id": ObjectId("67b12345..."),
  "key": "card001",
  "value": "{\"room_number\":\"101\",\"valid_from\":\"2026-02-01\",...}",
  "created_at": ISODate("2026-02-03T..."),
  "updated_at": ISODate("2026-02-03T...")
}
```

---

## 10. Extending with New Backends

### How to Add a PostgreSQL Adapter

#### Step 1: Create Adapter Files

```
internal/database/postgres/
├── postgres.go          # Implement KV interface
└── postgres_test.go     # Unit tests
```

#### Step 2: Implement KV Interface

```go
package postgres

import "commander/internal/kv"

type PostgresKV struct {
    db *sql.DB
}

// Implement all 6 methods
func (p *PostgresKV) Get(ctx context.Context, ns, coll, key string) ([]byte, error) {
    query := `SELECT value FROM kv_store WHERE namespace=$1 AND collection=$2 AND key=$3`
    var value []byte
    err := p.db.QueryRowContext(ctx, query, ns, coll, key).Scan(&value)
    if err == sql.ErrNoRows {
        return nil, kv.ErrKeyNotFound
    }
    return value, err
}

func (p *PostgresKV) Set(ctx context.Context, ns, coll, key string, value []byte) error {
    // INSERT OR UPDATE logic
    ...
}

// Implement remaining 4 methods...
```

#### Step 3: Update Config

```go
// internal/config/config.go
const BackendPostgres = "postgres"

type KVConfig struct {
    BackendType string
    PostgresURI string `envconfig:"POSTGRES_URI"`
    // ...
}
```

#### Step 4: Update Factory

```go
// internal/database/factory.go
func NewKV(cfg *config.Config) (kv.KV, error) {
    switch cfg.KV.BackendType {
    case config.BackendPostgres:
        return postgres.NewPostgresKV(cfg.KV.PostgresURI)
    // ... other cases
    }
}
```

#### Step 5: Update .env.example

```bash
# Add PostgreSQL configuration
KV_BACKEND_TYPE=postgres
POSTGRES_URI=postgresql://user:pass@localhost:5432/kv_store
```

Done! No business logic code changes required.

---

## 11. Design Principles Explained

### Dependency Inversion Principle (DIP)

```
❌ Wrong approach (tight coupling):
Handler → MongoDBKV → mongo-driver

✅ Correct approach (loose coupling):
Handler → KV Interface ← MongoDBKV
                      ← RedisKV
                      ← BBoltKV
```

**Benefits**:
- Upper layers don't depend on lower layers; both depend on abstraction
- Switching implementations requires no upper-layer code changes

### Open/Closed Principle (OCP)

```
Open for extension: Can add new adapters (e.g., PostgreSQL)
Closed for modification: No need to modify existing code
```

### Single Responsibility Principle (SRP)

```
Each adapter is responsible for one database implementation only:
- BBoltKV: Only handles filesystem operations
- RedisKV: Only handles Redis protocol
- MongoDBKV: Only handles MongoDB protocol
```

### Interface Segregation Principle (ISP)

```
KV interface contains only 6 necessary methods:
- Doesn't force implementation of unnecessary methods
- Keeps interface minimal and focused
```

---

## 12. Integration with MVP Card Verification System

### Scenario: Card Validity Verification

#### Option A: Direct MongoDB Adapter Usage (Quick MVP)

```go
// Advantages: Fast, flexible
// Disadvantages: Separated from KV abstraction

func VerifyCard(ctx context.Context, cardID string) (bool, error) {
    // Direct MongoDB access
    collection := mongoClient.Database("default").Collection("cards")
    
    var card struct {
        CardID string `bson:"card_id"`
        Status string `bson:"status"`
        ExpireAt time.Time `bson:"expire_at"`
    }
    
    err := collection.FindOne(ctx, bson.M{"card_id": cardID}).Decode(&card)
    if err != nil {
        return false, err
    }
    
    // Verification logic
    return card.Status == "active" && time.Now().Before(card.ExpireAt), nil
}
```

#### Option B: Extend KV Interface (Long-term Solution)

```go
// Add query method to kv.KV interface
type KV interface {
    // ... original 6 methods
    
    // New query method
    Query(ctx context.Context, ns, coll string, filter map[string]interface{}) ([]map[string]interface{}, error)
}
```

#### Option C: Parallel Architecture (Production Recommended)

```
KV Layer (general-purpose storage)
  ├─ Store general config, settings, logs

Card Service Layer (business logic)
  ├─ Read MongoDB (direct queries)
  ├─ Verify card logic
  └─ Write Redis cache (hot data)

HTTP API
  └─ /api/v1/cards/verify (card verification)
```

---

## 13. Performance Characteristics Comparison

### Latency Comparison (Latency)

```
Operation: Get single key-value

BBolt:   1-5ms      (local disk)
Redis:   <1ms       (memory, network latency)
MongoDB: 5-50ms     (network latency + query)
```

### Throughput Comparison (Throughput)

```
Assumption: 64-core CPU, sufficient network bandwidth

BBolt:   ~10K ops/sec    (disk I/O limited)
Redis:   ~100K ops/sec   (memory operations)
MongoDB: ~50K ops/sec    (network limited)
```

### Storage Capacity Comparison

```
BBolt:   Disk space dependent (up to TB scale)
Redis:   Memory size dependent (typically GB scale)
MongoDB: Up to PB scale (distributed storage)
```

### Cost Comparison

```
BBolt:   $0          (open source, no server costs)
Redis:   Low-Medium  (requires server)
MongoDB: Low-High    (Atlas pay-per-use model)
```

---

## 14. Selection Guide

### When to Use BBolt?

```
✅ Edge devices (Raspberry Pi, IoT)
✅ Development environments
✅ Simple single-machine applications
✅ Cost-sensitive projects
❌ High-concurrency applications
❌ Distributed systems
```

### When to Use Redis?

```
✅ High-performance real-time applications
✅ Cache layer
✅ Session storage
✅ Queue systems
❌ Long-term data storage (requires persistence)
❌ Complex queries
```

### When to Use MongoDB?

```
✅ Production environments
✅ Cloud deployment (Atlas)
✅ Distributed systems
✅ Complex data structures
✅ High availability requirements
❌ Sub-millisecond latency requirements (<1ms)
❌ Memory-constrained environments
```

---

## 15. Monitoring and Debugging

### Health Checks

```bash
# All backends support Ping method
curl http://localhost:8080/health

# Example response
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2026-02-03T12:00:00Z"
}
```

### Logging

```go
// All operations are logged
log.Printf("KV Get: namespace=%s, collection=%s, key=%s", ns, coll, key)
log.Printf("KV Set: namespace=%s, collection=%s, key=%s, size=%d bytes", ns, coll, key, len(value))
```

### Performance Monitoring

Recommended metrics:
- Request latency (p50, p95, p99)
- Operations per second (OPS)
- Error rate
- Connection pool utilization

---

## Summary

Commander's database abstraction layer provides:

1. **Unified Interface**: The `kv.KV` interface hides implementation details
2. **Multi-Backend Support**: Supports BBolt, Redis, MongoDB—three mainstream solutions
3. **Runtime Switching**: Dynamically select backend via environment variables
4. **Extensibility**: Adding new backends requires only interface implementation
5. **Design Patterns**: Follows SOLID principles with high cohesion and low coupling
6. **Performance Optimization**: Choose optimal solution for each specific scenario

This design provides a solid foundation for MVP card verification systems, production deployments, and edge device support.
