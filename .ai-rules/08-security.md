# Security Rules

## Input Validation

### Always Validate User Input

**Parameters**
```go
func GetKVHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        namespace := c.Param("namespace")
        collection := c.Param("collection")
        key := c.Param("key")
        
        // Validate required parameters
        if namespace == "" || collection == "" || key == "" {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "namespace, collection, and key are required",
                Code:    "INVALID_PARAMS",
            })
            return
        }
        
        // Validate parameter format
        if !isValidNamespace(namespace) {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "invalid namespace format",
                Code:    "INVALID_NAMESPACE",
            })
            return
        }
        
        // Continue processing...
    }
}

func isValidNamespace(ns string) bool {
    // Alphanumeric and hyphens only
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, ns)
    return matched && len(ns) <= 255
}
```

**Request Body**
```go
type KVRequestBody struct {
    Value interface{} `json:"value" binding:"required"`
}

func SetKVHandler(kvStore kv.KV) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req KVRequestBody
        if err := c.BindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "invalid request body: " + err.Error(),
                Code:    "INVALID_BODY",
            })
            return
        }
        
        // Validate value size
        valueJSON, err := json.Marshal(req.Value)
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "failed to encode value",
                Code:    "ENCODE_ERROR",
            })
            return
        }
        
        // Limit value size (e.g., 1MB)
        if len(valueJSON) > 1024*1024 {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Message: "value size exceeds 1MB limit",
                Code:    "VALUE_TOO_LARGE",
            })
            return
        }
        
        // Continue processing...
    }
}
```

### Query Parameters

```go
// Sanitize and validate query parameters
limit := 1000
if limitParam := c.Query("limit"); limitParam != "" {
    parsedLimit, err := strconv.Atoi(limitParam)
    if err != nil || parsedLimit < 1 || parsedLimit > 10000 {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Message: "invalid limit parameter (must be 1-10000)",
            Code:    "INVALID_LIMIT",
        })
        return
    }
    limit = parsedLimit
}
```

## Error Messages

### Don't Leak Sensitive Information

**Bad Examples**
```go
// DON'T - Exposes database path
return fmt.Errorf("failed to open database at /var/lib/stayforge/commander/secret.db")

// DON'T - Exposes internal structure
return fmt.Errorf("mongodb connection failed: mongodb://admin:password123@...")

// DON'T - Stack traces to users
panic(err)  // Never panic in production handlers
```

**Good Examples**
```go
// DO - Generic error message
c.JSON(http.StatusInternalServerError, ErrorResponse{
    Message: "internal server error",
    Code:    "INTERNAL_ERROR",
})

// DO - Log details server-side
log.Printf("Database error: %v", err)
c.JSON(http.StatusInternalServerError, ErrorResponse{
    Message: "failed to process request",
    Code:    "INTERNAL_ERROR",
})

// DO - Helpful but not revealing
if errors.Is(err, kv.ErrKeyNotFound) {
    c.JSON(http.StatusNotFound, ErrorResponse{
        Message: "key not found",
        Code:    "KEY_NOT_FOUND",
    })
}
```

### Error Logging

```go
// Log errors with context, but don't expose to users
func handleError(c *gin.Context, err error, operation string) {
    // Log detailed error server-side
    log.Printf("[ERROR] %s failed: %v, IP: %s, User-Agent: %s",
        operation, err,
        c.ClientIP(),
        c.Request.UserAgent())
    
    // Return generic error to user
    c.JSON(http.StatusInternalServerError, ErrorResponse{
        Message: "an error occurred processing your request",
        Code:    "INTERNAL_ERROR",
    })
}
```

## Secrets Management

### Environment Variables

**Never Commit Secrets**
```bash
# .gitignore must include
.env
*.key
*.pem
credentials.json
```

**Load from Environment**
```go
// Good - from environment
mongoURI := os.Getenv("MONGODB_URI")
redisURI := os.Getenv("REDIS_URI")

// Bad - hardcoded
mongoURI := "mongodb://admin:password123@..."
```

### Configuration Validation

```go
func LoadConfig() (*Config, error) {
    cfg := &Config{
        MongoURI: os.Getenv("MONGODB_URI"),
        RedisURI: os.Getenv("REDIS_URI"),
    }
    
    // Validate required secrets are present
    if cfg.MongoURI == "" {
        return nil, errors.New("MONGODB_URI is required")
    }
    
    // Redact secrets in logs
    log.Printf("Loaded config with MongoDB URI: %s", redactURI(cfg.MongoURI))
    
    return cfg, nil
}

func redactURI(uri string) string {
    // mongodb://user:password@host -> mongodb://user:***@host
    re := regexp.MustCompile(`:([^@]+)@`)
    return re.ReplaceAllString(uri, ":***@")
}
```

## Rate Limiting (Future)

### Middleware

```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }
    
    return limiter
}

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        limiter := rl.getLimiter(c.ClientIP())
        
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, ErrorResponse{
                Message: "rate limit exceeded",
                Code:    "RATE_LIMIT_EXCEEDED",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## Authentication (Future)

### Basic Auth Example

```go
func BasicAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        username, password, ok := c.Request.BasicAuth()
        if !ok {
            c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Message: "authentication required",
                Code:    "AUTH_REQUIRED",
            })
            c.Abort()
            return
        }
        
        // Validate credentials (use constant-time comparison)
        if !validateCredentials(username, password) {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Message: "invalid credentials",
                Code:    "AUTH_FAILED",
            })
            c.Abort()
            return
        }
        
        // Store user info in context
        c.Set("username", username)
        c.Next()
    }
}

func validateCredentials(username, password string) bool {
    // Use constant-time comparison to prevent timing attacks
    expectedUser := os.Getenv("API_USERNAME")
    expectedPass := os.Getenv("API_PASSWORD")
    
    return subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) == 1 &&
           subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) == 1
}
```

## CORS (If Needed)

```go
import "github.com/gin-contrib/cors"

func setupCORS(router *gin.Engine) {
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"https://example.com"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
    
    router.Use(cors.New(config))
}
```

## HTTPS/TLS

### Production Deployment

```go
// TLS configuration for production
func main() {
    router := gin.Default()
    setupRoutes(router)
    
    // Use TLS in production
    if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
        certFile := os.Getenv("TLS_CERT_FILE")
        keyFile := os.Getenv("TLS_KEY_FILE")
        
        log.Fatal(http.ListenAndServeTLS(":8443", certFile, keyFile, router))
    } else {
        log.Fatal(http.ListenAndServe(":8080", router))
    }
}
```

## Database Security

### Connection Security

**MongoDB**
```go
// Use TLS for MongoDB connections
clientOpts := options.Client().
    ApplyURI(uri).
    SetTLSConfig(&tls.Config{
        MinVersion: tls.VersionTLS12,
    })
```

**Redis**
```go
// Use TLS for Redis connections
client := redis.NewClient(&redis.Options{
    Addr:      uri,
    Password:  password,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
})
```

### BBolt File Permissions

```go
// Restrict file permissions
db, err := bolt.Open(path, 0600, nil)  // Owner read/write only
```

## Logging Security

### Sanitize Logs

```go
// Don't log sensitive data
log.Printf("User authenticated: %s", username)  // OK
log.Printf("User logged in with password: %s", password)  // NEVER

// Redact sensitive fields
type User struct {
    Username string
    Password string `json:"-"`  // Don't serialize
    Email    string
}

func (u *User) String() string {
    return fmt.Sprintf("User{username=%s, email=%s}", u.Username, u.Email)
}
```

### Log Levels

```go
// Use appropriate log levels
log.Printf("[INFO] Server started on port %s", port)
log.Printf("[WARN] High memory usage: %d MB", memUsage)
log.Printf("[ERROR] Failed to connect to database: %v", err)

// Don't log at debug level in production
if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
    log.Printf("[DEBUG] Request body: %s", body)
}
```

## Dependency Security

### Regular Updates

```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Audit
go mod verify
```

### Minimal Dependencies

```go
// Prefer standard library
import "encoding/json"  // Good
// import "github.com/heavy/json-lib"  // Avoid if possible
```

## Request Limits

### Size Limits

```go
// Limit request body size
func LimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1MB limit
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
        c.Next()
    }
}
```

### Timeout Limits

```go
// Set timeout on all operations
ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
defer cancel()

value, err := kvStore.Get(ctx, namespace, collection, key)
```

## Best Practices

### DO
- ✅ Validate all user input
- ✅ Use environment variables for secrets
- ✅ Return generic error messages
- ✅ Log detailed errors server-side
- ✅ Use HTTPS in production
- ✅ Implement rate limiting
- ✅ Set timeouts on operations
- ✅ Use secure file permissions
- ✅ Update dependencies regularly
- ✅ Sanitize logs

### DON'T
- ❌ Trust user input
- ❌ Hardcode secrets
- ❌ Expose internal errors
- ❌ Log passwords or tokens
- ❌ Skip input validation
- ❌ Use HTTP in production
- ❌ Ignore rate limiting
- ❌ Leave debug logs in production
- ❌ Use weak TLS versions
- ❌ Commit secrets to git

## Security Checklist

Before deploying:
- [ ] All secrets in environment variables
- [ ] No secrets in git history
- [ ] Input validation on all endpoints
- [ ] Rate limiting enabled
- [ ] HTTPS/TLS configured
- [ ] Error messages sanitized
- [ ] Logs don't contain secrets
- [ ] Dependencies updated
- [ ] File permissions secure (0600)
- [ ] Timeouts on all operations

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Guide](https://github.com/OWASP/Go-SCP)
- [CWE Top 25](https://cwe.mitre.org/top25/)
