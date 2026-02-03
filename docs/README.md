# Commander API Documentation

Complete documentation for the Commander Unified KV Storage API.

## 📚 Quick Links

### Getting Started (5 minutes)
- **[API Quick Start](api-quickstart.md)** - 5-minute setup guide with examples
- **[Health Check](#health-check)** - Verify service is running
- **[Basic CRUD Operations](#basic-crud)** - Get, Set, Delete, Exists

### API Reference
- **[OpenAPI 3.0 Specification](api-specification.yaml)** - Complete API specification
- **[API Examples](api-examples.md)** - Python, JavaScript, curl examples
- **[Error Handling](#error-handling)** - Error codes and responses

### Project Management
- **[Project Management Plan](PROJECT_MANAGEMENT_PLAN.md)** - 1-3 month sprint plan
- **[Phase 1 Completion Report](PHASE1_COMPLETION.md)** - Phase 1 results and metrics
- **[KV Usage Guide](kv-usage.md)** - Library-level KV operations

### Deployment (Coming Soon)
- **Edge Device Guide** - Deploy on Raspberry Pi (Planned for Phase 2)
- **Docker Deployment** - Containerized deployment (In README.md)
- **Migration Guide** - Switching between backends (Planned for Phase 2)

---

## Quick Reference

### Health Check

Verify the service is running:

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "environment": "STANDARD",
  "message": "Commander service is running",
  "timestamp": "2026-02-03T12:34:56Z"
}
```

### Basic CRUD

#### Set a Value
```bash
curl -X POST http://localhost:8080/api/v1/kv/default/users/user1 \
  -H "Content-Type: application/json" \
  -d '{"value": {"name": "John", "age": 30}}'
```

#### Get a Value
```bash
curl http://localhost:8080/api/v1/kv/default/users/user1
```

#### Delete a Value
```bash
curl -X DELETE http://localhost:8080/api/v1/kv/default/users/user1
```

#### Check Existence
```bash
curl -I http://localhost:8080/api/v1/kv/default/users/user1
```

### Batch Operations

#### Batch Set (up to 1000 operations)
```bash
curl -X POST http://localhost:8080/api/v1/kv/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "namespace": "default",
        "collection": "users",
        "key": "user1",
        "value": {"name": "Alice"}
      },
      {
        "namespace": "default",
        "collection": "users",
        "key": "user2",
        "value": {"name": "Bob"}
      }
    ]
  }'
```

#### Batch Delete
```bash
curl -X DELETE http://localhost:8080/api/v1/kv/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {"namespace": "default", "collection": "users", "key": "user1"},
      {"namespace": "default", "collection": "users", "key": "user2"}
    ]
  }'
```

---

## API Endpoints

### Core CRUD (4 endpoints)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/kv/{namespace}/{collection}/{key}` | Retrieve a value |
| POST | `/api/v1/kv/{namespace}/{collection}/{key}` | Set a value |
| DELETE | `/api/v1/kv/{namespace}/{collection}/{key}` | Delete a value |
| HEAD | `/api/v1/kv/{namespace}/{collection}/{key}` | Check if exists |

### Batch Operations (2 endpoints)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/v1/kv/batch` | Set multiple keys (up to 1000) |
| DELETE | `/api/v1/kv/batch` | Delete multiple keys (up to 1000) |
| GET | `/api/v1/kv/{namespace}/{collection}` | List keys in collection* |

*Not implemented for all backends

### Namespace Management (5+ endpoints)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/namespaces` | List namespaces* |
| GET | `/api/v1/namespaces/{namespace}/collections` | List collections* |
| GET | `/api/v1/namespaces/{namespace}/info` | Get namespace info |
| DELETE | `/api/v1/namespaces/{namespace}` | Delete namespace* |
| DELETE | `/api/v1/namespaces/{namespace}/collections/{collection}` | Delete collection* |

*Backend-dependent implementation

### Health & Root (2 endpoints)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/` | API welcome message |
| GET | `/health` | Health check |

---

## Data Organization

### Namespaces

Top-level groupings for data organization. Defaults to `"default"` if not specified.

```bash
# Custom namespace
POST /api/v1/kv/production/users/user1

# Default namespace (equivalent)
POST /api/v1/kv/default/users/user1
POST /api/v1/kv//users/user1
```

### Collections

Second-level groupings within namespaces for organizing related data.

```bash
# Store user data
POST /api/v1/kv/default/users/user1

# Store configuration
POST /api/v1/kv/default/config/database_url

# Store settings
POST /api/v1/kv/default/settings/theme
```

### Keys

Individual identifiers for values within a collection.

---

## Error Handling

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `KEY_NOT_FOUND` | 404 | Key does not exist |
| `INVALID_PARAMS` | 400 | Missing or invalid parameters |
| `INVALID_BODY` | 400 | Invalid request body |
| `INTERNAL_ERROR` | 500 | Server error |
| `DECODE_ERROR` | 500 | Failed to decode value |
| `ENCODE_ERROR` | 400 | Failed to encode value |
| `NOT_IMPLEMENTED` | 501 | Feature not available for backend |

### Error Response Format

```json
{
  "message": "Detailed error description",
  "code": "ERROR_CODE"
}
```

---

## Environment Configuration

Configure Commander with environment variables:

```bash
# Database backend (mongodb, redis, bbolt)
export DATABASE=bbolt

# Server port (default: 8080)
export SERVER_PORT=8080

# Environment (STANDARD or PRODUCTION)
export ENVIRONMENT=STANDARD

# BBolt data path (default: /var/lib/stayforge/commander)
export DATA_PATH=/var/lib/stayforge/commander

# Redis URI (if using Redis)
export REDIS_URI=redis://localhost:6379/0

# MongoDB URI (if using MongoDB)
export MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/
```

---

## Supported Data Types

Store any JSON-compatible value:

- **Objects**: `{"key": "value", "nested": {"inner": "value"}}`
- **Strings**: `"simple text"`
- **Numbers**: `123` or `45.67`
- **Booleans**: `true` or `false`
- **Arrays**: `["item1", "item2", "item3"]`
- **Null**: `null`

---

## Limits

- **Batch Operations**: Maximum 1000 operations per request
- **Key/Collection Length**: Depends on backend
- **Value Size**: Depends on backend configuration (typically 1MB-16MB)
- **List Operations**: Maximum 10,000 items per response

---

## Common Use Cases

### Configuration Management
Store application settings and feature flags:
```bash
POST /api/v1/kv/production/config/database
POST /api/v1/kv/production/config/features
```

### Session Storage
Store user session data:
```bash
POST /api/v1/kv/default/sessions/sess_abc123
```

### Caching
Cache computation results:
```bash
POST /api/v1/kv/default/cache/report_q1_2026
```

### User Profiles
Store user information:
```bash
POST /api/v1/kv/default/users/user_123
```

---

## Languages & Clients

Supported client implementations:

- **curl**: Command-line examples in [API Quick Start](api-quickstart.md)
- **Python**: Implementation in [API Examples](api-examples.md)
- **JavaScript/Node.js**: Implementation in [API Examples](api-examples.md)
- **Go**: Use the native [KV library](kv-usage.md)

---

## Troubleshooting

### Service Not Responding

Check if the service is running:
```bash
curl http://localhost:8080/health
```

### Port Already in Use

Find and stop the service using the port:
```bash
# Linux/macOS
lsof -i :8080

# Windows
netstat -ano | findstr :8080
```

### Invalid JSON

Validate your JSON:
```bash
echo '{"value": {"key": "test"}}' | jq .
```

### Authentication Issues

The current API doesn't require authentication (planned for Phase 2).

### Database Connection Issues

Check your database configuration:
```bash
echo $MONGODB_URI
echo $REDIS_URI
echo $DATA_PATH
```

---

## Performance Tips

1. **Use batch operations** instead of individual requests
2. **Normalize data structures** to avoid deeply nested objects
3. **Use appropriate collection names** for logical grouping
4. **Monitor response times** - target <50ms for edge devices
5. **Enable caching** in your application for frequently accessed data

---

## Project Status

### Phase 1: API Foundation ✅
- Core CRUD endpoints implemented
- Batch operations functional
- 75.8% test coverage
- Comprehensive documentation

### Phase 2: Documentation & Integration (Planned)
- Edge device deployment guide
- Swagger UI generation
- Migration utilities
- Troubleshooting playbook

### Phase 3: Architecture Optimization (Planned)
- LRU caching layer
- Prometheus metrics
- Edge device optimizations
- Offline operation mode

### Phase 4: Testing & QA (Planned)
- Integration tests
- Load testing
- Performance benchmarks
- Coverage increase to 85%+

---

## Related Documentation

- **[Project Management Plan](PROJECT_MANAGEMENT_PLAN.md)** - Complete sprint plan
- **[Phase 1 Report](PHASE1_COMPLETION.md)** - Results and metrics
- **[KV Library Guide](kv-usage.md)** - Low-level KV operations
- **[Main README](../README.md)** - Project overview and setup

---

## Getting Help

1. **Check the [API Quick Start](api-quickstart.md)** for basic questions
2. **Review [API Examples](api-examples.md)** for code samples
3. **Consult the [OpenAPI Specification](api-specification.yaml)** for details
4. **See [PROJECT_MANAGEMENT_PLAN.md](PROJECT_MANAGEMENT_PLAN.md)** for roadmap

---

## API Specification

For detailed API documentation, see [OpenAPI 3.0 Specification](api-specification.yaml).

The specification includes:
- All endpoint definitions
- Request/response schemas
- Error codes and examples
- Authentication mechanisms
- Rate limits (when applicable)

---

**Last Updated**: February 3, 2026  
**Status**: Phase 1 Complete ✅  
**Version**: 1.0.0

---

## Changelog

### v1.0.0 (Phase 1 - February 3, 2026)
- Initial API release
- 12 core endpoints
- CRUD operations
- Batch operations
- Namespace management
- Comprehensive documentation
- 75.8% test coverage

---

For the latest updates and to contribute, visit the project repository.
