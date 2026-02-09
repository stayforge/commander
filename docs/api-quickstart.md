# Commander API Quick Start Guide

A quick reference guide to get started with the Commander KV Storage API.

## Prerequisites

- Commander service running on `http://localhost:8080`
- `curl` command-line tool (or any HTTP client like Postman)
- Basic understanding of JSON and HTTP methods

## Service Health Check

Before making requests, verify the service is running:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "environment": "STANDARD",
  "message": "Commander service is running",
  "timestamp": "2026-02-03T12:34:56Z"
}
```

## Basic Operations (5 minutes)

### 1. Set a Value (Create/Update)

Store a key-value pair in a namespace and collection:

```bash
curl -X POST http://localhost:8080/api/v1/kv/default/users/user1 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30
    }
  }'
```

Response:
```json
{
  "message": "Successfully",
  "namespace": "default",
  "collection": "users",
  "key": "user1",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  },
  "timestamp": "2026-02-03T12:34:56Z"
}
```

### 2. Get a Value (Read)

Retrieve a previously stored value:

```bash
curl http://localhost:8080/api/v1/kv/default/users/user1
```

Response:
```json
{
  "message": "Successfully",
  "namespace": "default",
  "collection": "users",
  "key": "user1",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  },
  "timestamp": "2026-02-03T12:34:56Z"
}
```

### 3. Check Key Existence (HEAD)

Check if a key exists without retrieving its value:

```bash
curl -I http://localhost:8080/api/v1/kv/default/users/user1
```

Response:
- **HTTP 200**: Key exists
- **HTTP 404**: Key not found

### 4. Delete a Value (Remove)

Delete a key-value pair:

```bash
curl -X DELETE http://localhost:8080/api/v1/kv/default/users/user1
```

Response:
```json
{
  "message": "Successfully",
  "namespace": "default",
  "collection": "users",
  "key": "user1",
  "timestamp": "2026-02-03T12:34:56Z"
}
```

## Batch Operations

### Batch Set Multiple Keys

Set multiple key-value pairs in a single request:

```bash
curl -X POST http://localhost:8080/api/v1/kv/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "namespace": "default",
        "collection": "users",
        "key": "user1",
        "value": {"name": "Alice", "age": 25}
      },
      {
        "namespace": "default",
        "collection": "users",
        "key": "user2",
        "value": {"name": "Bob", "age": 28}
      },
      {
        "namespace": "default",
        "collection": "config",
        "key": "app_name",
        "value": "My Application"
      }
    ]
  }'
```

Response:
```json
{
  "message": "Batch operation completed",
  "results": [
    {
      "namespace": "default",
      "collection": "users",
      "key": "user1",
      "success": true
    },
    {
      "namespace": "default",
      "collection": "users",
      "key": "user2",
      "success": true
    },
    {
      "namespace": "default",
      "collection": "config",
      "key": "app_name",
      "success": true
    }
  ],
  "success_count": 3,
  "failure_count": 0,
  "timestamp": "2026-02-03T12:34:56Z"
}
```

### Batch Delete Multiple Keys

Delete multiple keys in a single request:

```bash
curl -X DELETE http://localhost:8080/api/v1/kv/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "namespace": "default",
        "collection": "users",
        "key": "user1"
      },
      {
        "namespace": "default",
        "collection": "users",
        "key": "user2"
      }
    ]
  }'
```

## Data Organization

### Namespaces

Namespaces are top-level groupings for organizing data. If you don't specify a namespace, it defaults to `"default"`.

```bash
# These are equivalent:
curl http://localhost:8080/api/v1/kv/default/users/user1
curl http://localhost:8080/api/v1/kv//users/user1  # defaults to "default"

# Custom namespace:
curl http://localhost:8080/api/v1/kv/production/users/user1
```

### Collections

Collections are second-level groupings within namespaces. They help organize related data.

```bash
# Store user data in "users" collection
POST /api/v1/kv/default/users/user1

# Store configuration in "config" collection
POST /api/v1/kv/default/config/database_url

# Store settings in "settings" collection
POST /api/v1/kv/default/settings/theme
```

## Error Handling

### Key Not Found (404)

```bash
curl http://localhost:8080/api/v1/kv/default/users/nonexistent
```

Response (HTTP 404):
```json
{
  "message": "key not found",
  "code": "KEY_NOT_FOUND"
}
```

### Invalid Parameters (400)

```bash
curl http://localhost:8080/api/v1/kv//collection/key
```

Response (HTTP 400):
```json
{
  "message": "namespace, collection, and key are required",
  "code": "INVALID_PARAMS"
}
```

### Invalid Request Body (400)

```bash
curl -X POST http://localhost:8080/api/v1/kv/default/users/user1 \
  -H "Content-Type: application/json" \
  -d '{"invalid": "body"}'
```

Response (HTTP 400):
```json
{
  "message": "invalid request body: Key: 'KVRequestBody.Value' Error:Field validation for 'Value' failed on the 'required' tag",
  "code": "INVALID_BODY"
}
```

## Namespace Management

### Get Namespace Info

```bash
curl http://localhost:8080/api/v1/namespaces/default/info
```

Response:
```json
{
  "message": "Namespace information retrieved",
  "namespace": "default",
  "timestamp": "2026-02-03T12:34:56Z"
}
```

## Data Types

The value in a KV pair can be:

- **Object**: `{"key": "value", "nested": {"key": "value"}}`
- **String**: `"simple text"`
- **Number**: `123` or `45.67`
- **Boolean**: `true` or `false`
- **Array**: `["item1", "item2"]`
- **Null**: `null`

## Limits

- **Batch operations**: Maximum 1000 operations per request
- **Key/Collection/Namespace length**: No strict limit (backend-dependent)
- **Value size**: Depends on backend configuration (typically 1MB-16MB)

## Environment Variables

Configure Commander with environment variables:

```bash
# Database backend (mongodb, redis, bbolt)
export DATABASE=bbolt

# Server port
export SERVER_PORT=8080

# Environment (STANDARD or PRODUCTION)
export ENVIRONMENT=STANDARD

# BBolt data path
export DATA_PATH=/var/lib/stayforge/commander

# Redis URI
export REDIS_URI=redis://localhost:6379/0

# MongoDB URI
export MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/
```

## Next Steps

1. **Read the full API specification**: See `api-specification.yaml` for detailed endpoint documentation
2. **Edge deployment**: See `edge-deployment.md` for deploying on Raspberry Pi and IoT devices
3. **KV Library**: See `../docs/kv-usage.md` for programmatic access patterns

## Common Use Cases

### 1. Configuration Management

Store application configuration:

```bash
curl -X POST http://localhost:8080/api/v1/kv/production/config/database \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "host": "db.example.com",
      "port": 5432,
      "pool_size": 10
    }
  }'
```

### 2. Session Storage

Store user sessions:

```bash
curl -X POST http://localhost:8080/api/v1/kv/default/sessions/sess_abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "user_id": 42,
      "login_time": "2026-02-03T10:00:00Z",
      "ip": "192.168.1.1"
    }
  }'
```

### 3. Cache Data

Cache computation results:

```bash
curl -X POST http://localhost:8080/api/v1/kv/default/cache/report_q1_2026 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "generated_at": "2026-02-03T12:00:00Z",
      "total_revenue": 1500000,
      "total_users": 5000
    }
  }'
```

## Performance Tips

1. **Use batch operations** for multiple writes instead of individual requests
2. **Normalize your data structure** to avoid deeply nested objects
3. **Use appropriate collection names** to logically group related data
4. **Monitor response times** - target <50ms latency for edge devices
5. **Enable caching** in your application for frequently accessed data

## Troubleshooting

**Service not responding?**
```bash
curl -v http://localhost:8080/health
```

**Port already in use?**
```bash
lsof -i :8080  # Linux/Mac
netstat -ano | findstr :8080  # Windows
```

**Invalid JSON in request?**
```bash
# Validate JSON syntax
echo '{"value": {"key": "test"}}' | jq .
```

For more help, see the [troubleshooting guide](troubleshooting.md).
