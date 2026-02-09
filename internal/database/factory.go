package database

import (
	"commander/internal/config"
	"commander/internal/database/bbolt"
	"commander/internal/database/mongodb"
	"commander/internal/database/redis"
	"commander/internal/kv"
	"fmt"
)

// NewKV creates a kv.KV implementation configured according to cfg.KV.BackendType.
// It validates that MongoURI or RedisURI are provided when those backends are selected and returns an error for unsupported backend types.
func NewKV(cfg *config.Config) (kv.KV, error) {
	switch cfg.KV.BackendType {
	case config.BackendMongoDB:
		if cfg.KV.MongoURI == "" {
			return nil, fmt.Errorf("MongoDB URI is required (set MONGODB_URI)")
		}
		return mongodb.NewMongoDBKV(cfg.KV.MongoURI)
	case config.BackendRedis:
		if cfg.KV.RedisURI == "" {
			return nil, fmt.Errorf("Redis URI is required (set REDIS_URI)")
		}
		return redis.NewRedisKV(cfg.KV.RedisURI)
	case config.BackendBBolt:
		return bbolt.NewBBoltKV(cfg.KV.BBoltPath)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", cfg.KV.BackendType)
	}
}