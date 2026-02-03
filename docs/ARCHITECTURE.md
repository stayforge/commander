# Commander Architecture

Comprehensive architecture documentation for the Commander KV Storage Abstraction Service.

## Table of Contents

1. [System Overview](#system-overview)
2. [High-Level Architecture](#high-level-architecture)
3. [Component Architecture](#component-architecture)
4. [Data Flow](#data-flow)
5. [Database Backend Architecture](#database-backend-architecture)
6. [Deployment Architecture](#deployment-architecture)
7. [Security Architecture](#security-architecture)
8. [Technology Stack](#technology-stack)

---

## System Overview

Commander is a unified KV storage abstraction service that provides a single REST API interface for multiple database backends.

### Key Characteristics

- **Language**: Go 1.25.5
- **Framework**: Gin Web Framework
- **Backends**: BBolt (embedded), Redis (in-memory), MongoDB (cloud)
- **Target**: Edge devices and cloud deployments
- **Architecture Pattern**: Hexagonal (Ports and Adapters)

---

## High-Level Architecture

```mermaid
graph TB
    subgraph "External Clients"
        Client1[Web Browser]
        Client2[Mobile App]
        Client3[IoT Device]
        Client4[Microservice]
    end

    subgraph "Commander Service"
        API[REST API<br/>Gin Router]
        
        subgraph "Business Logic"
            Handlers[HTTP Handlers<br/>CRUD + Batch + Management]
            Validation[Input Validation]
            ErrorHandler[Error Handler]
        end
        
        subgraph "Abstraction Layer"
            KVInterface[KV Interface<br/>Get/Set/Delete/Exists]
            Factory[Database Factory<br/>Runtime Selection]
        end
        
        subgraph "Backend Implementations"
            BBolt[BBolt KV<br/>Embedded DB]
            Redis[Redis KV<br/>In-Memory]
            MongoDB[MongoDB KV<br/>Cloud DB]
        end
    end

    subgraph "Storage Backends"
        BBoltDB[(BBolt Files<br/>*.db)]
        RedisDB[(Redis Server<br/>6379)]
        MongoDBAtlas[(MongoDB Atlas<br/>Cloud)]
    end

    Client1 --> API
    Client2 --> API
    Client3 --> API
    Client4 --> API
    
    API --> Handlers
    Handlers --> Validation
    Handlers --> ErrorHandler
    Handlers --> KVInterface
    
    KVInterface --> Factory
    Factory --> BBolt
    Factory --> Redis
    Factory --> MongoDB
    
    BBolt --> BBoltDB
    Redis --> RedisDB
    MongoDB --> MongoDBAtlas

    style API fill:#4A90E2
    style KVInterface fill:#7ED321
    style Factory fill:#F5A623
    style BBolt fill:#BD10E0
    style Redis fill:#B8E986
    style MongoDB fill:#50E3C2
```

---

## Component Architecture

### Layer-by-Layer View

```mermaid
graph TD
    subgraph "Layer 1: HTTP Layer"
        Router[Gin Router]
        Middleware[Middleware<br/>Logger, Recovery, CORS]
        Routes[Route Registration<br/>12 Endpoints]
    end

    subgraph "Layer 2: Handler Layer"
        CRUD[CRUD Handlers<br/>Get/Set/Delete/Head]
        Batch[Batch Handlers<br/>BatchSet/BatchDelete]
        Mgmt[Management Handlers<br/>Namespace/Collection]
        Health[Health Handlers<br/>Health/Root]
    end

    subgraph "Layer 3: Business Logic"
        Validation[Parameter Validation]
        Normalization[Namespace Normalization]
        Serialization[JSON Serialization]
        ContextMgmt[Context Management]
    end

    subgraph "Layer 4: Abstraction Layer"
        KVInterface[KV Interface]
        Factory[Factory Pattern]
    end

    subgraph "Layer 5: Database Layer"
        BBoltImpl[BBolt Implementation]
        RedisImpl[Redis Implementation]
        MongoImpl[MongoDB Implementation]
    end

    subgraph "Layer 6: Storage Layer"
        Files[(Local Files)]
        RedisServer[(Redis Server)]
        MongoServer[(MongoDB Atlas)]
    end

    Router --> Middleware
    Middleware --> Routes
    Routes --> CRUD
    Routes --> Batch
    Routes --> Mgmt
    Routes --> Health
    
    CRUD --> Validation
    Batch --> Validation
    Mgmt --> Validation
    
    Validation --> Normalization
    Normalization --> Serialization
    Serialization --> ContextMgmt
    
    ContextMgmt --> KVInterface
    KVInterface --> Factory
    
    Factory -.->|Runtime Selection| BBoltImpl
    Factory -.->|Runtime Selection| RedisImpl
    Factory -.->|Runtime Selection| MongoImpl
    
    BBoltImpl --> Files
    RedisImpl --> RedisServer
    MongoImpl --> MongoServer

    style KVInterface fill:#7ED321
    style Factory fill:#F5A623
```

### Package Structure

```mermaid
graph LR
    subgraph "cmd/"
        Main[main.go<br/>Entry Point]
    end

    subgraph "internal/"
        subgraph "config/"
            Config[config.go<br/>Configuration]
        end
        
        subgraph "handlers/"
            KVHandlers[kv.go<br/>CRUD Handlers]
            BatchHandlers[batch.go<br/>Batch Ops]
            NSHandlers[namespace.go<br/>Management]
            HealthHandlers[health.go<br/>Health Check]
        end
        
        subgraph "kv/"
            Interface[kv.go<br/>KV Interface]
        end
        
        subgraph "database/"
            Factory[factory.go<br/>Factory]
            
            subgraph "bbolt/"
                BBoltKV[bbolt.go<br/>Implementation]
            end
            
            subgraph "redis/"
                RedisKV[redis.go<br/>Implementation]
            end
            
            subgraph "mongodb/"
                MongoKV[mongodb.go<br/>Implementation]
            end
        end
    end

    Main --> Config
    Main --> KVHandlers
    Main --> Factory
    
    KVHandlers --> Interface
    BatchHandlers --> Interface
    NSHandlers --> Interface
    
    Interface -.->|implements| BBoltKV
    Interface -.->|implements| RedisKV
    Interface -.->|implements| MongoKV
    
    Factory --> BBoltKV
    Factory --> RedisKV
    Factory --> MongoKV

    style Interface fill:#7ED321
    style Factory fill:#F5A623
```

---

## Data Flow

### GET Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Router as Gin Router
    participant Handler as GetKVHandler
    participant KV as KV Interface
    participant Backend as Database Backend
    participant Storage as Storage

    Client->>Router: GET /api/v1/kv/default/users/user1
    Router->>Handler: Route to handler
    
    Handler->>Handler: Extract parameters<br/>(namespace, collection, key)
    Handler->>Handler: Validate parameters
    Handler->>Handler: Normalize namespace
    
    Handler->>KV: Get(ctx, "default", "users", "user1")
    KV->>Backend: Get(ctx, "default", "users", "user1")
    Backend->>Storage: Read data
    Storage-->>Backend: Return bytes
    Backend-->>KV: Return []byte
    KV-->>Handler: Return []byte
    
    Handler->>Handler: Unmarshal JSON
    Handler->>Handler: Build response
    Handler-->>Router: JSON response
    Router-->>Client: 200 OK + data
```

### POST Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Router as Gin Router
    participant Handler as SetKVHandler
    participant KV as KV Interface
    participant Backend as Database Backend
    participant Storage as Storage

    Client->>Router: POST /api/v1/kv/default/users/user1<br/>{"value": {...}}
    Router->>Handler: Route to handler
    
    Handler->>Handler: Extract parameters
    Handler->>Handler: Parse JSON body
    Handler->>Handler: Validate input
    Handler->>Handler: Normalize namespace
    Handler->>Handler: Marshal value to JSON bytes
    
    Handler->>KV: Set(ctx, "default", "users", "user1", []byte)
    KV->>Backend: Set(ctx, "default", "users", "user1", []byte)
    Backend->>Storage: Write data
    Storage-->>Backend: Success
    Backend-->>KV: nil (success)
    KV-->>Handler: nil (success)
    
    Handler->>Handler: Build response
    Handler-->>Router: JSON response
    Router-->>Client: 201 Created
```

### Batch Operation Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler as BatchSetHandler
    participant KV as KV Interface
    participant Backend as Database Backend

    Client->>Handler: POST /api/v1/kv/batch<br/>{"operations": [{...}, {...}]}
    
    Handler->>Handler: Parse batch request
    Handler->>Handler: Validate (max 1000 ops)
    
    loop For each operation
        Handler->>Handler: Validate operation
        Handler->>Handler: Normalize namespace
        Handler->>Handler: Marshal value
        Handler->>KV: Set(ctx, ns, col, key, value)
        KV->>Backend: Set(...)
        Backend-->>KV: Result
        KV-->>Handler: Result
        Handler->>Handler: Record result<br/>(success/failure)
    end
    
    Handler->>Handler: Build batch response<br/>(results + counts)
    Handler-->>Client: 200 OK + batch results
```

---

## Database Backend Architecture

### Three Backend Implementations

```mermaid
graph TB
    subgraph "KV Interface Contract"
        Interface[Interface: KV<br/>Get/Set/Delete/Exists/Close/Ping]
    end

    subgraph "BBolt Backend"
        BBoltKV[BBolt KV Implementation]
        BBoltConn[File-based Connection]
        BBoltData[(Namespace Files<br/>default.db<br/>production.db)]
        
        BBoltKV --> BBoltConn
        BBoltConn --> BBoltData
        
        BBoltNote[Data Model:<br/>Namespace → File<br/>Collection → Bucket<br/>Key → Bucket Key]
    end

    subgraph "Redis Backend"
        RedisKV[Redis KV Implementation]
        RedisPool[Connection Pool]
        RedisServer[(Redis Server<br/>:6379)]
        
        RedisKV --> RedisPool
        RedisPool --> RedisServer
        
        RedisNote[Data Model:<br/>Key: ns:col:key<br/>Value: JSON string]
    end

    subgraph "MongoDB Backend"
        MongoKV[MongoDB KV Implementation]
        MongoPool[Connection Pool]
        MongoAtlas[(MongoDB Atlas<br/>Cloud)]
        
        MongoKV --> MongoPool
        MongoPool --> MongoAtlas
        
        MongoNote[Data Model:<br/>Namespace → Database<br/>Collection → Collection<br/>Doc: {key, value}]
    end

    Interface -.->|implements| BBoltKV
    Interface -.->|implements| RedisKV
    Interface -.->|implements| MongoKV

    style Interface fill:#7ED321
    style BBoltKV fill:#BD10E0
    style RedisKV fill:#B8E986
    style MongoKV fill:#50E3C2
```

### Data Organization Comparison

```mermaid
graph TD
    subgraph "Logical Structure"
        NS[Namespace: 'default']
        COL[Collection: 'users']
        KEY[Key: 'user1']
        VAL[Value: JSON]
        
        NS --> COL
        COL --> KEY
        KEY --> VAL
    end

    subgraph "BBolt Mapping"
        BBoltFile[File: default.db]
        BBoltBucket[Bucket: users]
        BBoltKey[Key: user1]
        BBoltVal[Value: JSON bytes]
        
        BBoltFile --> BBoltBucket
        BBoltBucket --> BBoltKey
        BBoltKey --> BBoltVal
    end

    subgraph "Redis Mapping"
        RedisKey[Key: 'default:users:user1']
        RedisVal[Value: JSON string]
        
        RedisKey --> RedisVal
    end

    subgraph "MongoDB Mapping"
        MongoDB[Database: default]
        MongoColl[Collection: users]
        MongoDoc[Document:<br/>{key: 'user1',<br/>value: '{...}'}]
        
        MongoDB --> MongoColl
        MongoColl --> MongoDoc
    end

    NS -.->|maps to| BBoltFile
    NS -.->|maps to| RedisKey
    NS -.->|maps to| MongoDB
```

---

## Deployment Architecture

### Edge Device Deployment (BBolt)

```mermaid
graph TB
    subgraph "Edge Device (Raspberry Pi)"
        subgraph "Commander Service"
            API[REST API<br/>:8080]
            Handler[Handlers]
            BBolt[BBolt KV]
        end
        
        subgraph "Storage"
            Files[(*.db Files<br/>/var/lib/stayforge/commander/)]
        end
        
        subgraph "System"
            Systemd[systemd Service]
            Monitor[Health Monitor]
        end
        
        API --> Handler
        Handler --> BBolt
        BBolt --> Files
        
        Systemd -.->|manages| API
        Monitor -.->|checks| API
    end

    subgraph "External"
        LocalApp[Local Application]
        RemoteApp[Remote Application<br/>Intermittent Network]
    end

    LocalApp -->|HTTP| API
    RemoteApp -.->|HTTP<br/>when connected| API

    style API fill:#4A90E2
    style BBolt fill:#BD10E0
    style Files fill:#F5A623
```

### Cloud Deployment (MongoDB/Redis)

```mermaid
graph TB
    subgraph "Cloud Environment (AWS/GCP/Azure)"
        subgraph "Compute"
            LB[Load Balancer]
            
            subgraph "Commander Instances"
                API1[Commander 1<br/>:8080]
                API2[Commander 2<br/>:8080]
                API3[Commander 3<br/>:8080]
            end
        end
        
        subgraph "Data Tier"
            RedisCluster[(Redis Cluster<br/>Cache Layer)]
            MongoCluster[(MongoDB Atlas<br/>Primary Storage)]
        end
        
        subgraph "Monitoring"
            Prometheus[Prometheus]
            Grafana[Grafana]
        end
        
        LB --> API1
        LB --> API2
        LB --> API3
        
        API1 --> RedisCluster
        API2 --> RedisCluster
        API3 --> RedisCluster
        
        API1 --> MongoCluster
        API2 --> MongoCluster
        API3 --> MongoCluster
        
        API1 -.->|metrics| Prometheus
        API2 -.->|metrics| Prometheus
        API3 -.->|metrics| Prometheus
        
        Prometheus --> Grafana
    end

    Client[External Clients] -->|HTTPS| LB

    style LB fill:#4A90E2
    style RedisCluster fill:#B8E986
    style MongoCluster fill:#50E3C2
```

### Hybrid Deployment

```mermaid
graph TB
    subgraph "Edge Layer"
        Edge1[Edge Device 1<br/>BBolt]
        Edge2[Edge Device 2<br/>BBolt]
        Edge3[Edge Device 3<br/>BBolt]
    end

    subgraph "Aggregation Layer"
        Gateway[API Gateway<br/>Commander + Redis]
    end

    subgraph "Cloud Layer"
        Cloud[Cloud Commander<br/>MongoDB Atlas]
        Analytics[Analytics Service]
    end

    Edge1 -.->|Sync when online| Gateway
    Edge2 -.->|Sync when online| Gateway
    Edge3 -.->|Sync when online| Gateway
    
    Gateway --> Cloud
    Cloud --> Analytics

    style Edge1 fill:#BD10E0
    style Edge2 fill:#BD10E0
    style Edge3 fill:#BD10E0
    style Gateway fill:#B8E986
    style Cloud fill:#50E3C2
```

---

## Security Architecture

### Security Layers

```mermaid
graph TB
    subgraph "External Layer"
        Client[Client Application]
        HTTPS[HTTPS/TLS 1.2+]
    end

    subgraph "API Layer Security"
        RateLimit[Rate Limiting<br/>Per IP/User]
        Auth[Authentication<br/>Basic/API Key]
        InputVal[Input Validation<br/>All Parameters]
    end

    subgraph "Application Layer Security"
        ContextTimeout[Context Timeouts<br/>5s default]
        ErrorSanitize[Error Sanitization<br/>No info leak]
        SecretsMgmt[Secrets Management<br/>Environment vars]
    end

    subgraph "Data Layer Security"
        Encryption[Data Encryption<br/>TLS for network]
        FilePerms[File Permissions<br/>0600 for BBolt]
        ConnSecurity[Secure Connections<br/>Authenticated]
    end

    subgraph "Monitoring & Audit"
        Logging[Security Logging]
        Metrics[Security Metrics]
        Alerts[Security Alerts]
    end

    Client --> HTTPS
    HTTPS --> RateLimit
    RateLimit --> Auth
    Auth --> InputVal
    
    InputVal --> ContextTimeout
    ContextTimeout --> ErrorSanitize
    ErrorSanitize --> SecretsMgmt
    
    SecretsMgmt --> Encryption
    Encryption --> FilePerms
    FilePerms --> ConnSecurity
    
    ConnSecurity -.-> Logging
    ConnSecurity -.-> Metrics
    Metrics -.-> Alerts

    style Auth fill:#E74C3C
    style InputVal fill:#E74C3C
    style Encryption fill:#E74C3C
```

### Authentication Flow (Future)

```mermaid
sequenceDiagram
    participant Client
    participant Auth as Auth Middleware
    participant Handler as Handler
    participant KV as KV Store

    Client->>Auth: Request + API Key
    Auth->>Auth: Validate API Key
    
    alt Valid Key
        Auth->>Auth: Extract User Context
        Auth->>Handler: Forward Request + Context
        Handler->>KV: Process Operation
        KV-->>Handler: Result
        Handler-->>Client: 200 OK + Response
    else Invalid Key
        Auth-->>Client: 401 Unauthorized
    end
```

---

## Technology Stack

### Complete Stack Overview

```mermaid
graph TB
    subgraph "Programming"
        Go[Go 1.25.5<br/>Primary Language]
    end

    subgraph "Web Framework"
        Gin[Gin v1.11.0<br/>HTTP Router & Middleware]
    end

    subgraph "Database Drivers"
        BBoltLib[etcd-io/bbolt v1.4.3<br/>Embedded KV]
        RedisLib[go-redis v9.17.2<br/>Redis Client]
        MongoLib[mongo-driver v1.17.6<br/>MongoDB Driver]
    end

    subgraph "Testing"
        Testify[testify v1.11.1<br/>Test Framework]
        MiniRedis[miniredis v2.36.1<br/>Redis Mock]
    end

    subgraph "Build & Deploy"
        Docker[Docker<br/>Containerization]
        Systemd[systemd<br/>Service Management]
    end

    subgraph "Documentation"
        OpenAPI[OpenAPI 3.0<br/>API Specification]
        Markdown[Markdown<br/>Documentation]
    end

    subgraph "CI/CD"
        GitHub[GitHub Actions<br/>Automation]
        GoLint[golangci-lint<br/>Code Quality]
        Codecov[Codecov<br/>Coverage Tracking]
    end

    subgraph "Monitoring (Future)"
        Prometheus[Prometheus<br/>Metrics]
        Grafana[Grafana<br/>Dashboards]
    end

    Go --> Gin
    Gin --> BBoltLib
    Gin --> RedisLib
    Gin --> MongoLib
    
    Go --> Testify
    Testify --> MiniRedis
    
    Go --> Docker
    Docker --> Systemd
    
    OpenAPI -.-> Markdown
    
    GitHub --> GoLint
    GitHub --> Codecov

    style Go fill:#00ADD8
    style Gin fill:#4A90E2
    style BBoltLib fill:#BD10E0
    style RedisLib fill:#B8E986
    style MongoLib fill:#50E3C2
```

### Dependency Graph

```mermaid
graph LR
    subgraph "Core Dependencies"
        Gin[gin-gonic/gin]
        BBolt[etcd-io/bbolt]
        Redis[redis/go-redis]
        Mongo[mongodb/mongo-driver]
    end

    subgraph "Testing Dependencies"
        Testify[stretchr/testify]
        MiniRedis[alicebob/miniredis]
    end

    subgraph "Commander Application"
        Main[cmd/server/main.go]
        Handlers[internal/handlers]
        Database[internal/database]
        Config[internal/config]
    end

    Main --> Gin
    Main --> Config
    Main --> Handlers
    Main --> Database
    
    Handlers --> Gin
    
    Database --> BBolt
    Database --> Redis
    Database --> Mongo
    
    Handlers -.->|testing| Testify
    Database -.->|testing| Testify
    Database -.->|testing| MiniRedis

    style Main fill:#4A90E2
    style Handlers fill:#7ED321
    style Database fill:#F5A623
```

---

## Performance Characteristics

### Response Time Budget (Edge Device)

```mermaid
graph LR
    subgraph "Total: <50ms (p99)"
        A[Network<br/>~5ms]
        B[Routing<br/>~1ms]
        C[Validation<br/>~1ms]
        D[KV Operation<br/>~10ms]
        E[Serialization<br/>~2ms]
        F[Response<br/>~1ms]
    end

    A --> B
    B --> C
    C --> D
    D --> E
    E --> F

    style D fill:#E74C3C
```

### Resource Usage Targets

| Resource | Target | Maximum |
|----------|--------|---------|
| **Memory** | 50MB | 256MB |
| **Binary Size** | <15MB | <20MB |
| **Startup Time** | <1s | <2s |
| **CPU (idle)** | <5% | <10% |
| **Disk I/O** | Minimal | Moderate |

---

## Design Principles

### SOLID Principles Applied

1. **Single Responsibility**: Each handler does one thing
2. **Open/Closed**: New backends via interface implementation
3. **Liskov Substitution**: All KV implementations interchangeable
4. **Interface Segregation**: KV interface is minimal and focused
5. **Dependency Inversion**: Handlers depend on KV interface, not concrete implementations

### Architectural Patterns

- **Hexagonal Architecture**: Ports (KV interface) and Adapters (implementations)
- **Factory Pattern**: Runtime backend selection
- **Dependency Injection**: KV store injected into handlers
- **Repository Pattern**: KV interface abstracts storage

---

## Future Architecture Evolution

### Phase 2 Enhancements

```mermaid
graph TB
    Current[Current Architecture<br/>Phase 1]
    
    subgraph "Phase 2 Additions"
        Cache[LRU Cache Layer]
        Metrics[Prometheus Metrics]
        Swagger[Swagger UI]
    end
    
    subgraph "Phase 3 Additions"
        Offline[Offline Mode]
        Sync[Data Sync]
        Compress[Binary Compression]
    end

    Current --> Cache
    Current --> Metrics
    Current --> Swagger
    
    Cache --> Offline
    Metrics --> Sync
    Swagger --> Compress

    style Current fill:#4A90E2
    style Cache fill:#F5A623
    style Metrics fill:#7ED321
```

---

## References

- **Project Plan**: `PROJECT_MANAGEMENT_PLAN.md`
- **API Specification**: `api-specification.yaml`
- **KV Usage**: `kv-usage.md`
- **Phase 1 Report**: `PHASE1_COMPLETION.md`
- **Code Rules**: `../.clinerules` and `../.ai-rules/`

---

**Version**: 1.0.0  
**Last Updated**: 2026-02-03  
**Status**: Phase 1 Complete ✅
