# Documentation Rules

## Code Documentation

### Package Documentation

Every package must have a package-level comment:

```go
// Package handlers provides HTTP request handlers for the Commander API.
// It includes CRUD operations, batch operations, and namespace management.
//
// All handlers follow a consistent pattern:
//   1. Extract and validate parameters
//   2. Call KV store operations
//   3. Return standardized JSON responses
//
// Example usage:
//   router.GET("/api/v1/kv/:ns/:col/:key", handlers.GetKVHandler(kvStore))
package handlers
```

### Function Documentation

**Exported Functions** (Required)
```go
// GetKVHandler handles GET /api/v1/kv/{namespace}/{collection}/{key}
// It retrieves a value from the KV store by namespace, collection, and key.
//
// Parameters:
//   - kvStore: The KV storage backend
//
// Returns:
//   - gin.HandlerFunc: HTTP handler function
//
// Response:
//   - 200: Value retrieved successfully
//   - 400: Invalid parameters
//   - 404: Key not found
//   - 500: Internal server error
func GetKVHandler(kvStore kv.KV) gin.HandlerFunc {
```

**Unexported Functions** (Optional)
```go
// marshalJSON converts a value to JSON bytes.
// If the value is already a string, it's returned as-is.
func marshalJSON(value interface{}) ([]byte, error) {
```

### Type Documentation

**Structs**
```go
// KVResponse represents a standard KV API response.
// It includes the namespace, collection, key, value, and timestamp.
type KVResponse struct {
    Message   string      `json:"message"`   // Status message
    Namespace string      `json:"namespace"` // Namespace name
    Key       string      `json:"key"`       // Key identifier
    Value     interface{} `json:"value"`     // Retrieved value
    Timestamp string      `json:"timestamp"` // RFC3339 timestamp
}
```

**Interfaces**
```go
// KV is the interface for key-value storage backends.
// All database implementations must satisfy this interface.
//
// Implementations:
//   - BBolt: Embedded database (internal/database/bbolt)
//   - Redis: In-memory database (internal/database/redis)
//   - MongoDB: Cloud database (internal/database/mongodb)
type KV interface {
    // Get retrieves a value by key from namespace and collection.
    // Returns ErrKeyNotFound if the key doesn't exist.
    Get(ctx context.Context, namespace, collection, key string) ([]byte, error)
    
    // Set stores a value by key in namespace and collection.
    Set(ctx context.Context, namespace, collection, key string, value []byte) error
}
```

### Constants and Variables

```go
var (
    // ErrKeyNotFound is returned when a key does not exist
    ErrKeyNotFound = errors.New("key not found")
    
    // DefaultNamespace is the default namespace used when namespace is empty
    DefaultNamespace = "default"
)

const (
    // MaxBatchSize is the maximum number of operations per batch request
    MaxBatchSize = 1000
    
    // DefaultTimeout is the default operation timeout
    DefaultTimeout = 5 * time.Second
)
```

## API Documentation

### OpenAPI Specification

All endpoints must be documented in `docs/api-specification.yaml`:

```yaml
/api/v1/kv/{namespace}/{collection}/{key}:
  get:
    tags:
      - KV Operations
    summary: Get a value
    description: Retrieve a value from the KV store by namespace, collection, and key
    operationId: getKV
    parameters:
      - name: namespace
        in: path
        description: Namespace (defaults to 'default')
        required: true
        schema:
          type: string
          example: "default"
      - name: collection
        in: path
        description: Collection within the namespace
        required: true
        schema:
          type: string
          example: "users"
      - name: key
        in: path
        description: Key to retrieve
        required: true
        schema:
          type: string
          example: "user1"
    responses:
      '200':
        description: Value retrieved successfully
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/KVResponse'
            example:
              message: "Successfully"
              namespace: "default"
              collection: "users"
              key: "user1"
              value:
                name: "John Doe"
                email: "john@example.com"
              timestamp: "2026-02-03T12:34:56Z"
```

### Code Examples

Provide examples in `docs/api-examples.md`:

**curl**
```bash
curl -X POST http://localhost:8080/api/v1/kv/default/users/user1 \
  -H "Content-Type: application/json" \
  -d '{"value": {"name": "John", "age": 30}}'
```

**Python**
```python
import requests

response = requests.post(
    "http://localhost:8080/api/v1/kv/default/users/user1",
    json={"value": {"name": "John", "age": 30}}
)
print(response.json())
```

**JavaScript**
```javascript
const response = await fetch(
  'http://localhost:8080/api/v1/kv/default/users/user1',
  {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value: { name: 'John', age: 30 } })
  }
);
const data = await response.json();
```

## README Documentation

### Project README Structure

```markdown
# Project Title

Brief description (1-2 sentences)

## Features
- Feature 1
- Feature 2

## Quick Start
5-minute setup guide

## Installation
Step-by-step installation

## Configuration
Environment variables

## API Documentation
Link to API docs

## Development
How to contribute

## License
```

### Section Guidelines

**Quick Start** (Essential)
- Copy-paste commands
- Minimal explanation
- Get user running in 5 minutes

**Installation** (Detailed)
- Prerequisites
- Step-by-step instructions
- Common issues

**Configuration** (Complete)
- All environment variables
- Default values
- Examples

**API Documentation** (Reference)
- Link to OpenAPI spec
- Link to examples
- Common endpoints

## Inline Comments

### When to Comment

**DO Comment**
- Complex algorithms
- Non-obvious logic
- Business rules
- Workarounds
- TODOs

```go
// Normalize namespace to "default" if empty to maintain consistency
// across all database backends
namespace = kv.NormalizeNamespace(namespace)

// TODO: Add rate limiting (see issue #123)

// Workaround for BBolt file locking issue on Windows
// See: https://github.com/etcd-io/bbolt/issues/456
```

**DON'T Comment**
- Obvious code
- What the code does (use function names)
- Commented-out code

```go
// Bad - obvious
// Set the user name
user.Name = "John"

// Bad - explains what (function name should explain)
// This function gets the user by ID
func getUserByID(id int) (*User, error) {

// Bad - commented-out code (delete it)
// oldValue := getValue()
newValue := getNewValue()
```

### Comment Style

**Single Line**
```go
// This is a single-line comment
x := 1
```

**Multiple Lines**
```go
// This is a longer comment that spans multiple lines.
// Each line should be self-contained and end with proper punctuation.
// Use proper grammar and capitalization.
```

**Block Comments** (rare)
```go
/*
Block comments are rarely needed in Go.
Use them only for:
  - Package documentation
  - Long explanations
  - Disabling large code blocks (temporarily)
*/
```

## Documentation Files

### Required Files

```
docs/
├── README.md                    # Documentation index
├── api-specification.yaml       # OpenAPI 3.0 spec
├── api-quickstart.md            # 5-minute tutorial
├── api-examples.md              # Code examples
├── PROJECT_MANAGEMENT_PLAN.md   # Project plan
├── PHASE1_COMPLETION.md         # Phase status
└── kv-usage.md                  # Library usage
```

### File Guidelines

**README.md**
- Quick reference
- Links to other docs
- Common operations
- Getting started

**api-specification.yaml**
- Complete OpenAPI 3.0 spec
- All endpoints
- All schemas
- Examples

**api-quickstart.md**
- 5-minute setup
- Basic operations
- Common use cases

**api-examples.md**
- Multiple languages
- Real-world scenarios
- Error handling

## Changelog

### Format

```markdown
# Changelog

## [Unreleased]
### Added
- New feature X

### Changed
- Updated feature Y

### Fixed
- Bug fix Z

## [1.0.0] - 2026-02-03
### Added
- Initial release
- 12 API endpoints
- Three database backends
```

### Guidelines

**Categories**
- `Added`: New features
- `Changed`: Changes to existing functionality
- `Deprecated`: Soon-to-be removed features
- `Removed`: Removed features
- `Fixed`: Bug fixes
- `Security`: Security fixes

## TODO Comments

### Format

```go
// TODO: Description of what needs to be done
// TODO(username): Assigned task
// TODO(username, 2026-02-15): Task with deadline
// FIXME: Known issue that needs fixing
// HACK: Temporary workaround
// NOTE: Important information
```

### Examples

```go
// TODO: Implement Redis connection pooling
// TODO(john): Add rate limiting middleware
// FIXME: BBolt file locking issue on Windows
// HACK: Temporary fix for NTP drift, proper solution in #123
// NOTE: This function is called by both API and CLI
```

## Documentation Standards

### Language

- Use American English
- Be concise and clear
- Use active voice
- Use present tense

### Format

- Use Markdown for documentation
- Use code blocks with syntax highlighting
- Use tables for structured data
- Use bullet points for lists

### Examples

Always provide:
- Working code examples
- Expected output
- Error cases
- Multiple languages (API docs)

### Updates

- Update docs with code changes
- Keep examples current
- Test examples before committing
- Version documentation

## Best Practices

### DO
- ✅ Document exported functions
- ✅ Provide code examples
- ✅ Keep docs up-to-date
- ✅ Use clear, concise language
- ✅ Include error cases
- ✅ Test documentation examples
- ✅ Link related documentation

### DON'T
- ❌ Document obvious code
- ❌ Leave outdated docs
- ❌ Use technical jargon excessively
- ❌ Skip examples
- ❌ Forget to update OpenAPI spec
- ❌ Leave TODO comments forever
- ❌ Comment out code instead of deleting

## Commander-Specific Guidelines

### Documentation Priority

1. **API Specification** (OpenAPI) - Must be complete
2. **Quick Start Guide** - For new users
3. **Code Examples** - Multiple languages
4. **Code Comments** - For exported functions
5. **Architecture Docs** - For contributors

### Tone

- Professional but friendly
- Clear and direct
- Helpful and encouraging
- No unnecessary jargon

### Target Audience

- **API Docs**: API consumers (developers)
- **Code Comments**: Contributors (developers)
- **README**: Everyone (users and developers)
- **Architecture Docs**: Contributors (advanced)

## References

- [Go Documentation Guide](https://go.dev/doc/comment)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Keep a Changelog](https://keepachangelog.com/)
