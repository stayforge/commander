# KV Storage Usage Guide

This document provides a comprehensive guide on how to use the KV (Key-Value) storage abstraction layer in the Commander project.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Backend Implementations](#backend-implementations)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

The KV storage abstraction layer provides a unified interface for key-value storage operations across multiple backend implementations. It supports:

- **Multiple Backends**: MongoDB, Redis, and BBolt
- **Namespace Support**: Organize data by namespace (defaults to "default")
- **Collection Support**: Further organize data within namespaces
- **JSON Values**: Store and retrieve JSON-encoded data
- **Type Safety**: Compile-time type checking with Go interfaces

## Architecture

```
┌─────────────────┐
│   Application   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   KV Interface  │  (internal/kv)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Factory        │  (internal/database)
└────────┬────────┘
         │
    ┌────┴────┬──────────┐
    ▼         ▼          ▼
┌────────┐ ┌──────┐ ┌────────┐
│MongoDB │ │Redis │ │ BBolt  │
└────────┘ └──────┘ └────────┘
```

## Configuration

### Environment Variables

The KV storage is configured through environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE` | Backend type (`mongodb`, `redis`, `bbolt`) | `bbolt` | No |
| `MONGODB_URI` | MongoDB connection URI | - | Yes (if DATABASE=mongodb) |
| `REDIS_URI` | Redis connection URI | - | Yes (if DATABASE=redis) |
| `DATA_PATH` | BBolt storage path | `/var/lib/stayforge/commander` | No |

### Configuration Examples

#### MongoDB
```bash
export DATABASE=mongodb
export MONGODB_URI="mongodb+srv://user:password@cluster.mongodb.net/"
```

#### Redis
```bash
export DATABASE=redis
export REDIS_URI="redis://:password@localhost:6379/0"
# Or without password:
export REDIS_URI="redis://localhost:6379/0"
```

#### BBolt (Default)
```bash
export DATABASE=bbolt
export DATA_PATH="/var/lib/stayforge/commander"
# Or use default (no export needed)
```

## Basic Usage

### 1. Initialize KV Store

```go
package main

import (
    "context"
    "commander/internal/config"
    "commander/internal/database"
    "commander/internal/kv"
)

func main() {
    // Load configuration from environment variables
    cfg := config.LoadConfig()
    
    // Create KV store instance
    store, err := database.NewKV(cfg)
    if err != nil {
        log.Fatalf("Failed to create KV store: %v", err)
    }
    defer store.Close()
    
    // Verify connection
    ctx := context.Background()
    if err := store.Ping(ctx); err != nil {
        log.Fatalf("Failed to ping KV store: %v", err)
    }
}
```

### 2. Store Data (Set)

```go
import "encoding/json"

// Prepare data as JSON
data := map[string]interface{}{
    "name":        "Fire Dragon",
    "card_number": "ABC123DEF456",
    "devices":     []string{"device-001", "device-002"},
    "status":      "active",
}

// Marshal to JSON bytes
valueBytes, err := json.Marshal(data)
if err != nil {
    log.Fatalf("Failed to marshal data: %v", err)
}

// Store with namespace, collection, and key
namespace := "commander"
collection := "cards"
key := "card_001"

err = store.Set(ctx, namespace, collection, key, valueBytes)
if err != nil {
    log.Fatalf("Failed to set value: %v", err)
}
```

### 3. Retrieve Data (Get)

```go
// Retrieve data
valueBytes, err := store.Get(ctx, namespace, collection, key)
if err != nil {
    if errors.Is(err, kv.ErrKeyNotFound) {
        log.Println("Key not found")
    } else {
        log.Fatalf("Failed to get value: %v", err)
    }
}

// Unmarshal JSON
var data map[string]interface{}
if err := json.Unmarshal(valueBytes, &data); err != nil {
    log.Fatalf("Failed to unmarshal data: %v", err)
}

fmt.Printf("Retrieved: %+v\n", data)
```

### 4. Check Existence

```go
exists, err := store.Exists(ctx, namespace, collection, key)
if err != nil {
    log.Fatalf("Failed to check existence: %v", err)
}

if exists {
    fmt.Println("Key exists")
} else {
    fmt.Println("Key does not exist")
}
```

### 5. Delete Data

```go
err := store.Delete(ctx, namespace, collection, key)
if err != nil {
    if errors.Is(err, kv.ErrKeyNotFound) {
        log.Println("Key not found, nothing to delete")
    } else {
        log.Fatalf("Failed to delete: %v", err)
    }
}
```

## Advanced Usage

### Namespace and Collection

The KV storage uses a three-level hierarchy:

1. **Namespace**: Top-level organization (defaults to "default" if empty)
2. **Collection**: Second-level organization within namespace
3. **Key**: Individual key within collection

#### Default Namespace

If you pass an empty string for namespace, it automatically uses "default":

```go
// These are equivalent:
store.Set(ctx, "", "cards", "card_001", valueBytes)
store.Set(ctx, "default", "cards", "card_001", valueBytes)
```

#### Namespace Examples

```go
// Production data
store.Set(ctx, "production", "users", "user_123", userData)

// Staging data
store.Set(ctx, "staging", "users", "user_123", userData)

// Development data (using default namespace)
store.Set(ctx, "", "users", "user_123", userData)
```

### Working with Structs

```go
type Card struct {
    Name        string   `json:"name"`
    CardNumber  string   `json:"card_number"`
    Devices     []string `json:"devices"`
    Status      string   `json:"status"`
}

// Store struct
card := Card{
    Name:       "Fire Dragon",
    CardNumber: "ABC123DEF456",
    Devices:    []string{"device-001", "device-002"},
    Status:     "active",
}

valueBytes, _ := json.Marshal(card)
store.Set(ctx, "commander", "cards", "card_001", valueBytes)

// Retrieve struct
valueBytes, _ := store.Get(ctx, "commander", "cards", "card_001")
var retrievedCard Card
json.Unmarshal(valueBytes, &retrievedCard)
```

### Context Usage

Always use context for operations to support cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := store.Get(ctx, namespace, collection, key)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(10 * time.Second)
    cancel() // Cancel operation after 10 seconds
}()
value, err := store.Get(ctx, namespace, collection, key)
```

## Backend Implementations

### MongoDB

**Storage Mapping:**
- Namespace → Database
- Collection → Collection
- Key → Document field
- Value → JSON string in document

**Example Document:**
```json
{
  "key": "card_001",
  "value": "{\"name\":\"Fire Dragon\",\"card_number\":\"ABC123DEF456\"}"
}
```

**URI Format:**
```
mongodb+srv://username:password@cluster.mongodb.net/
mongodb://username:password@host:port/
```

### Redis

**Storage Mapping:**
- Namespace:Collection:Key → Redis key
- Value → JSON string

**Key Format:**
```
<namespace>:<collection>:<key>
```

**Example:**
```
commander:cards:card_001 → {"name":"Fire Dragon",...}
```

**URI Format:**
```
redis://:password@host:port/db
redis://host:port/db
redis://localhost:6379/0
```

### BBolt

**Storage Mapping:**
- Namespace → Database file (`<namespace>.db`)
- Collection → Bucket
- Key → Bucket key
- Value → JSON bytes

**File Structure:**
```
/var/lib/stayforge/commander/
├── default.db      (namespace: "")
├── commander.db    (namespace: "commander")
└── production.db   (namespace: "production")
```

**Path Configuration:**
- Default: `/var/lib/stayforge/commander`
- Configurable via `DATA_PATH` environment variable

## Error Handling

### Common Errors

```go
import (
    "errors"
    "commander/internal/kv"
)

value, err := store.Get(ctx, namespace, collection, key)
if err != nil {
    if errors.Is(err, kv.ErrKeyNotFound) {
        // Key does not exist
        log.Println("Key not found")
    } else if errors.Is(err, kv.ErrConnectionFailed) {
        // Connection to backend failed
        log.Fatalf("Connection failed: %v", err)
    } else {
        // Other errors
        log.Fatalf("Unexpected error: %v", err)
    }
}
```

### Error Types

| Error | Description | When It Occurs |
|-------|-------------|----------------|
| `kv.ErrKeyNotFound` | Key does not exist | Get/Delete on non-existent key |
| `kv.ErrConnectionFailed` | Backend connection failed | Initial connection or ping failure |

## Best Practices

### 1. Always Close the Store

```go
store, err := database.NewKV(cfg)
if err != nil {
    return err
}
defer store.Close() // Always close to release resources
```

### 2. Use Context for Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := store.Get(ctx, namespace, collection, key)
```

### 3. Handle Errors Properly

```go
value, err := store.Get(ctx, namespace, collection, key)
if err != nil {
    if errors.Is(err, kv.ErrKeyNotFound) {
        // Handle not found case
        return nil, nil
    }
    return nil, err
}
```

### 4. Validate JSON Before Storing

```go
// Validate JSON before storing
if !json.Valid(valueBytes) {
    return fmt.Errorf("invalid JSON")
}
store.Set(ctx, namespace, collection, key, valueBytes)
```

### 5. Use Meaningful Namespaces and Collections

```go
// Good: Clear organization
store.Set(ctx, "production", "users", "user_123", data)
store.Set(ctx, "production", "devices", "device_456", data)

// Bad: Unclear organization
store.Set(ctx, "data", "stuff", "item1", data)
```

### 6. Ping Before Critical Operations

```go
if err := store.Ping(ctx); err != nil {
    log.Fatalf("KV store is not available: %v", err)
}
// Proceed with operations
```

## Examples

### Complete Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "commander/internal/config"
    "commander/internal/database"
    "commander/internal/kv"
)

func main() {
    // Load configuration
    cfg := config.LoadConfig()
    
    // Create KV store
    store, err := database.NewKV(cfg)
    if err != nil {
        log.Fatalf("Failed to create KV store: %v", err)
    }
    defer store.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Verify connection
    if err := store.Ping(ctx); err != nil {
        log.Fatalf("Failed to ping KV store: %v", err)
    }
    
    // Define namespace and collection
    namespace := "commander"
    collection := "cards"
    key := "card_001"
    
    // Create and store data
    card := map[string]interface{}{
        "name":        "Fire Dragon",
        "card_number": "ABC123DEF456",
        "devices":     []string{"device-001", "device-002"},
        "status":      "active",
        "created_at":  time.Now().UTC().Format(time.RFC3339),
    }
    
    valueBytes, err := json.Marshal(card)
    if err != nil {
        log.Fatalf("Failed to marshal: %v", err)
    }
    
    // Store
    if err := store.Set(ctx, namespace, collection, key, valueBytes); err != nil {
        log.Fatalf("Failed to set: %v", err)
    }
    fmt.Printf("Stored: %s\n", key)
    
    // Check existence
    exists, err := store.Exists(ctx, namespace, collection, key)
    if err != nil {
        log.Fatalf("Failed to check existence: %v", err)
    }
    fmt.Printf("Exists: %v\n", exists)
    
    // Retrieve
    retrievedBytes, err := store.Get(ctx, namespace, collection, key)
    if err != nil {
        log.Fatalf("Failed to get: %v", err)
    }
    
    var retrievedCard map[string]interface{}
    if err := json.Unmarshal(retrievedBytes, &retrievedCard); err != nil {
        log.Fatalf("Failed to unmarshal: %v", err)
    }
    
    fmt.Printf("Retrieved: %+v\n", retrievedCard)
    
    // Delete
    if err := store.Delete(ctx, namespace, collection, key); err != nil {
        log.Fatalf("Failed to delete: %v", err)
    }
    fmt.Printf("Deleted: %s\n", key)
}
```

### Batch Operations Example

```go
func batchStore(ctx context.Context, store kv.KV, namespace, collection string, items map[string]interface{}) error {
    for key, value := range items {
        valueBytes, err := json.Marshal(value)
        if err != nil {
            return fmt.Errorf("failed to marshal %s: %w", key, err)
        }
        
        if err := store.Set(ctx, namespace, collection, key, valueBytes); err != nil {
            return fmt.Errorf("failed to set %s: %w", key, err)
        }
    }
    return nil
}
```

### Conditional Update Example

```go
func conditionalUpdate(ctx context.Context, store kv.KV, namespace, collection, key string, newValue interface{}) error {
    // Check if key exists
    exists, err := store.Exists(ctx, namespace, collection, key)
    if err != nil {
        return err
    }
    
    if !exists {
        return fmt.Errorf("key %s does not exist", key)
    }
    
    // Update
    valueBytes, err := json.Marshal(newValue)
    if err != nil {
        return err
    }
    
    return store.Set(ctx, namespace, collection, key, valueBytes)
}
```

## Migration Between Backends

The KV abstraction layer makes it easy to switch between backends:

```bash
# Switch from BBolt to MongoDB
export DATABASE=mongodb
export MONGODB_URI="mongodb+srv://..."

# Switch from MongoDB to Redis
export DATABASE=redis
export REDIS_URI="redis://..."

# Switch back to BBolt
export DATABASE=bbolt
# or just unset DATABASE (defaults to bbolt)
```

No code changes are required - the same interface works with all backends!

## Troubleshooting

### Connection Issues

```go
// Always check connection before use
if err := store.Ping(ctx); err != nil {
    log.Fatalf("Connection failed: %v", err)
}
```

### Invalid JSON

```go
// Validate JSON before storing
if !json.Valid(valueBytes) {
    return fmt.Errorf("invalid JSON")
}
```

### Namespace Not Found

Remember that empty namespace defaults to "default":
```go
// These are equivalent:
store.Set(ctx, "", "collection", "key", value)
store.Set(ctx, "default", "collection", "key", value)
```

## See Also

- [Configuration Guide](../README.md#configuration)
- [Docker Deployment](../README.Docker.md)
- [API Reference](../internal/kv/kv.go)

