package redis

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"commander/internal/kv"
	"github.com/redis/go-redis/v9"
)

// RedisKV implements KV interface using Redis
// Key format: <namespace>:<collection>:<key>
type RedisKV struct {
	client *redis.Client
}

// NewRedisKV creates a new Redis KV store from URI
// URI format: redis://[:password@]host[:port][/db]
// Examples:
//   - redis://localhost:6379
//   - redis://:password@localhost:6379
//   - redis://localhost:6379/0
//   - redis://:password@localhost:6379/1
func NewRedisKV(uri string) (*RedisKV, error) {
	if uri == "" {
		return nil, fmt.Errorf("Redis URI is required")
	}

	// Parse URI
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URI: %w", err)
	}

	// Extract components
	addr := parsedURL.Host
	if addr == "" {
		addr = "localhost:6379"
	} else if !strings.Contains(addr, ":") {
		addr = addr + ":6379"
	}

	password := ""
	if parsedURL.User != nil {
		password, _ = parsedURL.User.Password()
	}

	db := 0
	if parsedURL.Path != "" {
		dbStr := strings.TrimPrefix(parsedURL.Path, "/")
		if dbStr != "" {
			if dbNum, err := strconv.Atoi(dbStr); err == nil {
				db = dbNum
			}
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Join(kv.ErrConnectionFailed, err)
	}

	return &RedisKV{
		client: client,
	}, nil
}

// buildKey constructs the Redis key from namespace, collection, and key
// Format: <namespace>:<collection>:<key>
func (r *RedisKV) buildKey(namespace, collection, key string) string {
	namespace = kv.NormalizeNamespace(namespace)
	return fmt.Sprintf("%s:%s:%s", namespace, collection, key)
}

// Get retrieves a JSON value by key from namespace and collection
func (r *RedisKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
	redisKey := r.buildKey(namespace, collection, key)
	val, err := r.client.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, kv.ErrKeyNotFound
		}
		return nil, err
	}
	return []byte(val), nil
}

// Set stores a JSON value by key in namespace and collection
func (r *RedisKV) Set(ctx context.Context, namespace, collection, key string, value []byte) error {
	redisKey := r.buildKey(namespace, collection, key)
	return r.client.Set(ctx, redisKey, value, 0).Err()
}

// Delete removes a key-value pair from namespace and collection
func (r *RedisKV) Delete(ctx context.Context, namespace, collection, key string) error {
	redisKey := r.buildKey(namespace, collection, key)
	result := r.client.Del(ctx, redisKey)
	if result.Err() != nil {
		return result.Err()
	}
	if result.Val() == 0 {
		return kv.ErrKeyNotFound
	}
	return nil
}

// Exists checks if a key exists in namespace and collection
func (r *RedisKV) Exists(ctx context.Context, namespace, collection, key string) (bool, error) {
	redisKey := r.buildKey(namespace, collection, key)
	count, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close closes the Redis connection
func (r *RedisKV) Close() error {
	return r.client.Close()
}

// Ping checks if the connection is alive
func (r *RedisKV) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

