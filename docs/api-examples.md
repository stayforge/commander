# Commander API Examples

Practical examples for common use cases with the Commander API.

## Using curl

### Basic CRUD Operations

#### Create (Set)
```bash
curl -X POST http://localhost:8080/api/v1/kv/default/users/user123 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "id": 123,
      "name": "John Doe",
      "email": "john@example.com",
      "active": true
    }
  }'
```

#### Read (Get)
```bash
curl -s http://localhost:8080/api/v1/kv/default/users/user123 | jq .
```

#### Update (Replace)
```bash
curl -X POST http://localhost:8080/api/v1/kv/default/users/user123 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "id": 123,
      "name": "Jane Doe",
      "email": "jane@example.com",
      "active": true
    }
  }'
```

#### Delete
```bash
curl -X DELETE http://localhost:8080/api/v1/kv/default/users/user123
```

### Checking Existence

```bash
curl -I http://localhost:8080/api/v1/kv/default/users/user123
# HTTP 200 = exists, HTTP 404 = not found
```

### Batch Operations

#### Create Multiple Records
```bash
curl -X POST http://localhost:8080/api/v1/kv/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "namespace": "default",
        "collection": "products",
        "key": "prod_001",
        "value": {
          "name": "Laptop",
          "price": 999.99,
          "in_stock": true
        }
      },
      {
        "namespace": "default",
        "collection": "products",
        "key": "prod_002",
        "value": {
          "name": "Mouse",
          "price": 29.99,
          "in_stock": true
        }
      },
      {
        "namespace": "default",
        "collection": "products",
        "key": "prod_003",
        "value": {
          "name": "Keyboard",
          "price": 79.99,
          "in_stock": false
        }
      }
    ]
  }'
```

#### Delete Multiple Records
```bash
curl -X DELETE http://localhost:8080/api/v1/kv/batch \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {
        "namespace": "default",
        "collection": "products",
        "key": "prod_001"
      },
      {
        "namespace": "default",
        "collection": "products",
        "key": "prod_002"
      }
    ]
  }'
```

## Using Python

### Basic Setup

```python
import requests
import json

BASE_URL = "http://localhost:8080/api/v1"
HEADERS = {"Content-Type": "application/json"}

def set_value(namespace, collection, key, value):
    """Set a value in KV store"""
    url = f"{BASE_URL}/kv/{namespace}/{collection}/{key}"
    payload = {"value": value}
    response = requests.post(url, json=payload, headers=HEADERS)
    return response.json()

def get_value(namespace, collection, key):
    """Get a value from KV store"""
    url = f"{BASE_URL}/kv/{namespace}/{collection}/{key}"
    response = requests.get(url, headers=HEADERS)
    return response.json()

def delete_value(namespace, collection, key):
    """Delete a value from KV store"""
    url = f"{BASE_URL}/kv/{namespace}/{collection}/{key}"
    response = requests.delete(url, headers=HEADERS)
    return response.json()

def key_exists(namespace, collection, key):
    """Check if key exists"""
    url = f"{BASE_URL}/kv/{namespace}/{collection}/{key}"
    response = requests.head(url, headers=HEADERS)
    return response.status_code == 200
```

### Example Usage

```python
# Set a user
user = {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com"
}
result = set_value("default", "users", "alice_001", user)
print(f"Set user: {result['message']}")

# Get the user
user_data = get_value("default", "users", "alice_001")
print(f"Retrieved: {user_data['value']}")

# Check existence
exists = key_exists("default", "users", "alice_001")
print(f"User exists: {exists}")

# Delete the user
delete_value("default", "users", "alice_001")
print("User deleted")
```

### Batch Operations

```python
def batch_set(operations):
    """Set multiple values"""
    url = f"{BASE_URL}/kv/batch"
    payload = {"operations": operations}
    response = requests.post(url, json=payload, headers=HEADERS)
    return response.json()

# Example
operations = [
    {
        "namespace": "default",
        "collection": "users",
        "key": "user_001",
        "value": {"name": "Alice", "age": 30}
    },
    {
        "namespace": "default",
        "collection": "users",
        "key": "user_002",
        "value": {"name": "Bob", "age": 25}
    }
]

result = batch_set(operations)
print(f"Success: {result['success_count']}, Failed: {result['failure_count']}")
```

## Using JavaScript/Node.js

### Basic Setup

```javascript
const BASE_URL = 'http://localhost:8080/api/v1';

async function setValue(namespace, collection, key, value) {
  const url = `${BASE_URL}/kv/${namespace}/${collection}/${key}`;
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value })
  });
  return response.json();
}

async function getValue(namespace, collection, key) {
  const url = `${BASE_URL}/kv/${namespace}/${collection}/${key}`;
  const response = await fetch(url);
  return response.json();
}

async function deleteValue(namespace, collection, key) {
  const url = `${BASE_URL}/kv/${namespace}/${collection}/${key}`;
  const response = await fetch(url, { method: 'DELETE' });
  return response.json();
}

async function keyExists(namespace, collection, key) {
  const url = `${BASE_URL}/kv/${namespace}/${collection}/${key}`;
  const response = await fetch(url, { method: 'HEAD' });
  return response.ok;
}
```

### Example Usage

```javascript
(async () => {
  // Set a value
  const user = {
    id: 1,
    name: 'John',
    email: 'john@example.com'
  };
  const setResult = await setValue('default', 'users', 'john_001', user);
  console.log(`Set: ${setResult.message}`);

  // Get a value
  const getResult = await getValue('default', 'users', 'john_001');
  console.log(`Retrieved:`, getResult.value);

  // Check existence
  const exists = await keyExists('default', 'users', 'john_001');
  console.log(`Key exists: ${exists}`);

  // Delete a value
  const deleteResult = await deleteValue('default', 'users', 'john_001');
  console.log(`Deleted: ${deleteResult.message}`);
})();
```

### Batch Operations

```javascript
async function batchSet(operations) {
  const url = `${BASE_URL}/kv/batch`;
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ operations })
  });
  return response.json();
}

// Example
const operations = [
  {
    namespace: 'default',
    collection: 'products',
    key: 'prod_001',
    value: { name: 'Laptop', price: 999.99 }
  },
  {
    namespace: 'default',
    collection: 'products',
    key: 'prod_002',
    value: { name: 'Mouse', price: 29.99 }
  }
];

const result = await batchSet(operations);
console.log(`Success: ${result.success_count}, Failed: ${result.failure_count}`);
```

## Real-World Scenarios

### Session Management

Store user session data:

```bash
# Create session
curl -X POST http://localhost:8080/api/v1/kv/default/sessions/sess_abc123xyz \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "user_id": 42,
      "username": "alice",
      "login_time": "2026-02-03T10:00:00Z",
      "last_activity": "2026-02-03T10:15:00Z",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "permissions": ["read", "write"]
    }
  }'

# Update last activity
curl -X POST http://localhost:8080/api/v1/kv/default/sessions/sess_abc123xyz \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "user_id": 42,
      "username": "alice",
      "login_time": "2026-02-03T10:00:00Z",
      "last_activity": "2026-02-03T10:20:00Z",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "permissions": ["read", "write"]
    }
  }'
```

### Configuration Storage

Store application configuration:

```bash
# Store database config
curl -X POST http://localhost:8080/api/v1/kv/production/config/database \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "host": "db.production.example.com",
      "port": 5432,
      "database": "app_prod",
      "pool_size": 20,
      "timeout_ms": 5000
    }
  }'

# Store feature flags
curl -X POST http://localhost:8080/api/v1/kv/production/config/features \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "dark_mode": true,
      "new_dashboard": true,
      "beta_api": false,
      "maintenance_mode": false
    }
  }'
```

### Caching

Store cached data with metadata:

```bash
curl -X POST http://localhost:8080/api/v1/kv/default/cache/user_count_2026_02 \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "count": 5000,
      "cached_at": "2026-02-03T10:00:00Z",
      "expires_at": "2026-02-04T10:00:00Z",
      "source": "database_query"
    }
  }'
```

## Error Handling

```python
def safe_get_value(namespace, collection, key):
    """Get value with error handling"""
    try:
        url = f"{BASE_URL}/kv/{namespace}/{collection}/{key}"
        response = requests.get(url)
        
        if response.status_code == 200:
            return response.json()['value']
        elif response.status_code == 404:
            print(f"Key not found: {key}")
            return None
        elif response.status_code == 400:
            print(f"Invalid parameters")
            return None
        else:
            print(f"Error: {response.status_code}")
            return None
    except Exception as e:
        print(f"Request failed: {e}")
        return None
```

For more details, see the [API specification](api-specification.yaml).
