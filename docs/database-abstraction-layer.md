# 數據庫抽象層架構設計文檔

## 概述

Commander 採用 **六邊形架構 (Hexagonal Architecture，又稱端口與適配器模式)** 設計數據庫抽象層，實現對多個數據庫後端的統一支持。本文詳細介紹整體設計、各個適配器的實現細節，以及如何基於此架構擴展新的功能。

---

## 1. 整體架構圖

### 六邊形架構設計

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
        BBolt["🔌 BBolt Adapter<br/>文件系統<br/>namespace → file<br/>collection → bucket"]
        Redis["🔌 Redis Adapter<br/>內存緩存<br/>key: ns:coll:key"]
        MongoDB["🔌 MongoDB Adapter<br/>文檔數據庫<br/>ns → db<br/>coll → collection"]
    end
    
    subgraph "Backend Layer"
        BBoltDB["BBolt Database<br/>*.db files"]
        RedisDB["Redis Server<br/>Memory"]
        MongoDB_Actual["MongoDB Server<br/>Cloud/On-Premise"]
    end
    
    HTTP --> Handler
    Handler --> |依賴<br/>接口不是實現| Port
    Port --> |實現| BBolt
    Port --> |實現| Redis
    Port --> |實現| MongoDB
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

**核心設計理念**：
- **Port (端口)**：`kv.KV` 接口定義了統一的數據訪問契約
- **Adapters (適配器)**：三個獨立的實現，分別適配不同的數據庫後端
- **依賴方向**：Handlers 只依賴接口 (Port)，不依賴具體實現 (Adapters)
- **優勢**：
  - ✅ 支持運行時切換數據庫（通過環境變量）
  - ✅ 易於測試（可以 mock KV 接口）
  - ✅ 易於添加新的後端（只需實現 KV 接口）
  - ✅ 業務邏輯與數據存儲解耦

---

## 2. KV 接口定義

### Interface 簽名

位置：`internal/kv/kv.go`

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

### 接口方法詳解

| 方法 | 參數 | 返回值 | 說明 |
|------|------|--------|------|
| **Get** | namespace, collection, key | ([]byte, error) | 讀取 JSON 值，不存在返回 `ErrKeyNotFound` |
| **Set** | namespace, collection, key, value | error | 保存 JSON 值，會覆蓋舊值 |
| **Delete** | namespace, collection, key | error | 刪除鍵，不存在也返回成功 |
| **Exists** | namespace, collection, key | (bool, error) | 檢查鍵是否存在 |
| **Close** | - | error | 關閉連接，清理資源 |
| **Ping** | ctx | error | 健康檢查，驗證連接可用 |

### 錯誤定義

```go
var (
    ErrKeyNotFound = errors.New("key not found")
    ErrConnectionFailed = errors.New("connection failed")
)
```

### 數據結構

所有適配器統一使用以下邏輯層次：

```
Namespace（命名空間）
  ├── Collection 1（集合）
  │   ├── Key 1 → Value (JSON bytes)
  │   ├── Key 2 → Value (JSON bytes)
  │   └── ...
  ├── Collection 2
  │   ├── Key 1 → Value (JSON bytes)
  │   └── ...
  └── ...
```

**設計理由**：
- Namespace 用於不同的應用/模塊隔離（如：app, mobile, admin）
- Collection 用於同一命名空間內的數據分類（如：users, cards, settings）
- Key 為具體的數據標識符（如：user_id, card_number）

---

## 3. 工廠模式 (Factory Pattern)

### 動態後端選擇

位置：`internal/database/factory.go`

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

### 配置驅動

```bash
# .env 文件中選擇後端
KV_BACKEND_TYPE=mongodb  # 或 redis, bbolt

# MongoDB 後端配置
MONGODB_URI=mongodb://localhost:27017

# Redis 後端配置
REDIS_URI=redis://localhost:6379

# BBolt 後端配置
BBOLT_PATH=/data/kv
```

**優勢**：無需重新編譯代碼，通過環境變量切換後端

---

## 4. 三個適配器實現對比

### 映射策略對比表

| 概念 | BBolt | Redis | MongoDB |
|------|-------|-------|---------|
| **Namespace** | 文件系統目錄中的 `.db` 文件 | Key 前綴 (1st segment) | Database |
| **Collection** | BBolt Bucket | Key 前綴 (2nd segment) | Collection |
| **Key** | Bucket 內的鍵 | Redis Key (3rd segment) | Document `key` field |
| **Value** | 二進制字節 | Redis String (字節) | Document `value` field (字符串) |
| **存儲位置** | `{BBoltPath}/{namespace}.db` | 單一 Redis 服務器 | MongoDB 服務器 |
| **並發控制** | `sync.RWMutex` (per adapter) | Redis 原子操作 | MongoDB 事務 |
| **索引** | 無索引 (O(1) 查找) | Key 唯一 | 自動建立 unique index |
| **分佈式** | 否（本地文件） | 是（可集群） | 是（可副本集） |
| **適用場景** | 邊界設備、開發環境 | 高性能緩存、實時應用 | 生產環境、雲部署 |

---

## 5. 數據流圖 - 完整的 GET 請求

### 示例：GET /api/v1/kv/default/users/user1

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

## 6. BBolt 適配器實現細節

### 架構特點

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

### 數據組織

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

### 關鍵實現

位置：`internal/database/bbolt/bbolt.go`

**並發控制**：
```go
type BBoltKV struct {
    baseDir string
    dbs map[string]*bbolt.DB  // 每個 namespace 一個連接
    mu sync.RWMutex           // 保護 dbs map
}
```

**Lazy Loading**：
```go
// 首次訪問 namespace 時才打開文件
func (b *BBoltKV) getDB(namespace string) (*bbolt.DB, error) {
    // 讀鎖查詢
    b.mu.RLock()
    if db, exists := b.dbs[namespace]; exists {
        b.mu.RUnlock()
        return db, nil
    }
    b.mu.RUnlock()
    
    // 寫鎖打開
    b.mu.Lock()
    defer b.mu.Unlock()
    
    dbPath := filepath.Join(b.baseDir, fmt.Sprintf("%s.db", namespace))
    db, _ := bbolt.Open(dbPath, 0o600, nil)
    b.dbs[namespace] = db
    return db, nil
}
```

**優勢**：
- ✅ 無外部依賴（無需服務器）
- ✅ 適合邊界設備和開發環境
- ✅ 文件系統原生支持，數據持久化
- ✅ 低延遲（本地磁盤訪問）

**限制**：
- ❌ 不支持分佈式
- ❌ 單進程鎖定（多進程會衝突）
- ❌ 性能受限於本地磁盤

---

## 7. Redis 適配器實現細節

### 架構特點

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

### Key 命名規則

```
Namespace:Collection:Key

示例：
├── default:users:user1
├── default:users:user2
├── default:cards:card001
├── default:cards:card002
├── mobile:settings:theme
├── mobile:settings:language
└── admin:logs:2024-02-01
```

### 關鍵實現

位置：`internal/database/redis/redis.go`

**連接池**：
```go
type RedisKV struct {
    client *redis.Client  // 管理連接池
}
```

**Key 格式化**：
```go
func makeKey(namespace, collection, key string) string {
    return fmt.Sprintf("%s:%s:%s", namespace, collection, key)
}
```

**操作示例**：
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

**優勢**：
- ✅ 超高性能（內存訪問，<1ms）
- ✅ 支持集群（分佈式緩存）
- ✅ 豐富的數據結構（List, Set, Hash 等）
- ✅ 原生事務支持

**限制**：
- ❌ 內存容量有限
- ❌ 數據易丟失（需要配置持久化）
- ❌ 需要獨立的 Redis 服務器

**適用場景**：
- 實時應用、高並發讀寫
- 緩存層
- 會話存儲
- 排隊系統

---

## 8. MongoDB 適配器實現細節

### 架構特點

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

### 數據結構

**MongoDB 文檔結構**：
```json
{
    "_id": ObjectId("..."),        // MongoDB 自動生成
    "key": "user1",                // 我們的 key 字段
    "value": "{\"name\":\"Alice\"}", // JSON 字符串
    "created_at": ISODate("..."),  // 創建時間
    "updated_at": ISODate("...")   // 更新時間
}
```

**多層次映射**：
```
MongoDB 層次        | KV 層次
namespace → Database
collection → Collection
key → Document.key field
value → Document.value field
```

### 關鍵實現

位置：`internal/database/mongodb/mongodb.go`

**連接管理**：
```go
type MongoDBKV struct {
    client *mongo.Client  // 單一連接管理所有操作
    uri    string
}
```

**索引創建**：
```go
// 為每個 collection 建立唯一索引，確保 key 唯一
func (m *MongoDBKV) ensureIndex(ctx context.Context, coll *mongo.Collection) error {
    indexModel := mongo.IndexModel{
        Keys: bson.D{{Key: "key", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    _, err := coll.Indexes().CreateOne(ctx, indexModel)
    return err
}
```

**Get 操作**：
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

**優勢**：
- ✅ 完全托管（云服務如 Atlas）
- ✅ 自動副本集、故障轉移
- ✅ 支持複雜查詢（可擴展功能）
- ✅ 高可用性、安全性
- ✅ 無容量限制

**限制**：
- ❌ 網絡延遲（相比本地存儲）
- ❌ 需要外部服務
- ❌ 成本可能更高

**適用場景**：
- 生產環境
- 雲部署
- 分佈式系統
- 需要高可用性的應用

---

## 9. 完整的數據流示例

### 場景：存儲房卡數據

#### Step 1: 配置選擇 (main.go)

```go
cfg := config.LoadConfig()
// KV_BACKEND_TYPE=mongodb 從 .env 讀取
kvStore, _ := database.NewKV(cfg)
// 返回 MongoDBKV instance
```

#### Step 2: HTTP 請求

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

#### Step 3: Handler 處理

```go
// handlers/kv.go
func SetKVHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 解析參數
        ns := c.Param("namespace")        // "default"
        coll := c.Param("collection")     // "cards"
        key := c.Param("key")             // "card001"
        
        // 解析 JSON body
        var req KVRequestBody
        c.BindJSON(&req)
        
        // 編碼為 JSON bytes
        valueBytes, _ := json.Marshal(req.Value)
        
        // 調用 KV 接口（不知道具體實現）
        err := kvStore.Set(c.Request.Context(), ns, coll, key, valueBytes)
        
        // 返回結果
        c.JSON(200, KVResponse{...})
    }
}
```

#### Step 4: MongoDB 適配器執行

```go
// internal/database/mongodb/mongodb.go
func (m *MongoDBKV) Set(ctx context.Context, ns, coll, key string, value []byte) error {
    collection := m.getCollection(ns, coll)     // db: default, collection: cards
    m.ensureIndex(ctx, collection)              // 確保 key 唯一
    
    doc := bson.M{
        "key": key,                             // "card001"
        "value": string(value),                 // JSON 字符串
        "created_at": time.Now(),
        "updated_at": time.Now(),
    }
    
    // MongoDB 操作：upsert
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

#### Step 5: MongoDB 存儲結果

```javascript
// MongoDB 數據庫視圖
use default
db.cards.find()
// 結果：
{
  "_id": ObjectId("67b12345..."),
  "key": "card001",
  "value": "{\"room_number\":\"101\",\"valid_from\":\"2026-02-01\",...}",
  "created_at": ISODate("2026-02-03T..."),
  "updated_at": ISODate("2026-02-03T...")
}
```

---

## 10. 擴展新的後端

### 如何添加 PostgreSQL 適配器

#### Step 1: 創建適配器文件

```
internal/database/postgres/
├── postgres.go          # 實現 KV 接口
└── postgres_test.go     # 單元測試
```

#### Step 2: 實現 KV 接口

```go
package postgres

import "commander/internal/kv"

type PostgresKV struct {
    db *sql.DB
}

// 實現所有 6 個方法
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
    // INSERT OR UPDATE 邏輯
    ...
}

// 其他 4 個方法...
```

#### Step 3: 更新 Config

```go
// internal/config/config.go
const BackendPostgres = "postgres"

type KVConfig struct {
    BackendType string
    PostgresURI string `envconfig:"POSTGRES_URI"`
    // ...
}
```

#### Step 4: 更新 Factory

```go
// internal/database/factory.go
func NewKV(cfg *config.Config) (kv.KV, error) {
    switch cfg.KV.BackendType {
    case config.BackendPostgres:
        return postgres.NewPostgresKV(cfg.KV.PostgresURI)
    // ... 其他 cases
    }
}
```

#### Step 5: 更新 .env.example

```bash
# 新增 PostgreSQL 配置
KV_BACKEND_TYPE=postgres
POSTGRES_URI=postgresql://user:pass@localhost:5432/kv_store
```

完成！無需修改任何業務邏輯代碼。

---

## 11. 設計原則詳解

### 依賴倒置原則 (DIP - Dependency Inversion Principle)

```
❌ 錯誤做法 (強耦合)：
Handler → MongoDBKV → mongo-driver

✅ 正確做法 (弱耦合)：
Handler → KV Interface ← MongoDBKV
                      ← RedisKV
                      ← BBoltKV
```

**優勢**：
- 上層模塊不依賴下層模塊，都依賴抽象
- 切換實現無需修改上層代碼

### 開閉原則 (OCP - Open/Closed Principle)

```
開放於擴展：可以添加新的適配器（如 PostgreSQL）
對修改封閉：不需要修改已有代碼
```

### 單一職責原則 (SRP - Single Responsibility Principle)

```
每個適配器只負責一種數據庫的實現
- BBoltKV: 僅處理文件系統操作
- RedisKV: 僅處理 Redis 協議
- MongoDBKV: 僅處理 MongoDB 協議
```

### 接口隔離原則 (ISP - Interface Segregation Principle)

```
KV 接口只包含必要的 6 個方法
- 不強制實現不需要的方法
- 保持接口最小化
```

---

## 12. 與 MVP 房卡驗證系統的結合

### 場景：房卡有效性驗證

#### 方案 A：直接使用 MongoDB Adapter（快速 MVP）

```go
// 優勢：快速、靈活
// 劣勢：與 KV 抽象分離

func VerifyCard(ctx context.Context, cardID string) (bool, error) {
    // 直接訪問 MongoDB
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
    
    // 驗證邏輯
    return card.Status == "active" && time.Now().Before(card.ExpireAt), nil
}
```

#### 方案 B：擴展 KV 接口（長期解決方案）

```go
// 在 kv.KV 接口中添加查詢方法
type KV interface {
    // ... 原有 6 個方法
    
    // 新增查詢方法
    Query(ctx context.Context, ns, coll string, filter map[string]interface{}) ([]map[string]interface{}, error)
}
```

#### 方案 C：並行架構（推薦用於生產）

```
KV 層（通用數據存儲）
  ├─ 存儲通用配置、設置、日誌

Card Service 層（業務邏輯）
  ├─ 讀取 MongoDB（直接查詢）
  ├─ 驗證房卡邏輯
  └─ 寫入 Redis 緩存（熱數據）

HTTP API
  └─ /api/v1/cards/verify (房卡驗證)
```

---

## 13. 性能特性對比

### 延遲對比 (Latency)

```
操作：Get 單個鍵值

BBolt:   1-5ms      (本地磁盤)
Redis:   <1ms       (內存，網絡延遲)
MongoDB: 5-50ms     (網絡延遲 + 查詢)
```

### 吞吐量對比 (Throughput)

```
假設：64 核 CPU，網絡帶寬充足

BBolt:   ~10K ops/sec    (磁盤 I/O 限制)
Redis:   ~100K ops/sec   (內存操作)
MongoDB: ~50K ops/sec    (網絡限制)
```

### 存儲容量對比

```
BBolt:   取決於磁盤空間 (可達 TB 級)
Redis:   取決於內存大小 (通常 GB 級)
MongoDB: 可達 PB 級 (分佈式存儲)
```

### 成本對比

```
BBolt:   $0          (開源，無服務器成本)
Redis:   低-中       (需要服務器)
MongoDB: 低-高       (Atlas 按使用量計費)
```

---

## 14. 選擇指南

### 何時使用 BBolt？

```
✅ 邊界設備 (Raspberry Pi, IoT)
✅ 開發環境
✅ 簡單的單機應用
✅ 對成本敏感
❌ 高並發應用
❌ 分佈式系統
```

### 何時使用 Redis？

```
✅ 高性能實時應用
✅ 緩存層
✅ 會話存儲
✅ 排隊系統
❌ 長期數據存儲 (需要持久化)
❌ 複雜查詢
```

### 何時使用 MongoDB？

```
✅ 生產環境
✅ 云部署 (Atlas)
✅ 分佈式系統
✅ 複雜數據結構
✅ 高可用性要求
❌ 超低延遲要求 (<1ms)
❌ 內存有限的環境
```

---

## 15. 監控和調試

### 健康檢查

```bash
# 所有後端都支持 Ping 方法
curl http://localhost:8080/health

# 響應示例
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2026-02-03T12:00:00Z"
}
```

### 日誌記錄

```go
// 所有操作都記錄日誌
log.Printf("KV Get: namespace=%s, collection=%s, key=%s", ns, coll, key)
log.Printf("KV Set: namespace=%s, collection=%s, key=%s, size=%d bytes", ns, coll, key, len(value))
```

### 性能監控

建議添加指標：
- 請求延遲 (p50, p95, p99)
- 每秒操作數 (OPS)
- 错誤率
- 連接池使用率

---

## 總結

Commander 的數據庫抽象層提供：

1. **統一接口**：通過 `kv.KV` 接口隱藏實現細節
2. **多後端支持**：支持 BBolt、Redis、MongoDB 三種主流方案
3. **運行時切換**：通過環境變量動態選擇後端
4. **易於擴展**：添加新後端只需實現接口
5. **設計模式**：遵循 SOLID 原則，代碼高內聚、低耦合
6. **性能優化**：針對不同場景選擇最優方案

這個設計為 MVP 房卡驗證系統、生產環境部署、邊界設備支持提供了堅實的基礎。

