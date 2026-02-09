package bbolt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"commander/internal/kv"

	"go.etcd.io/bbolt"
)

// BBoltKV implements KV interface using bbolt
// namespace = different files, collection = bucket
// Key: card_001 / Value: {"name": "Fire Dragon", ...}
//
//nolint:revive // BBoltKV name is intentional to match package name
type BBoltKV struct {
	baseDir string
	dbs     map[string]*bbolt.DB
	mu      sync.RWMutex
}

// NewBBoltKV creates a new bbolt KV store
func NewBBoltKV(baseDir string) (*BBoltKV, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &BBoltKV{
		baseDir: baseDir,
		dbs:     make(map[string]*bbolt.DB),
	}, nil
}

// getDB returns the database for the given namespace (file)
// Each namespace corresponds to a different .db file
func (b *BBoltKV) getDB(namespace string) (*bbolt.DB, error) {
	b.mu.RLock()
	db, exists := b.dbs[namespace]
	b.mu.RUnlock()

	if exists {
		return db, nil
	}

	// Create new database connection
	b.mu.Lock()
	defer b.mu.Unlock()

	// Double check after acquiring write lock
	if existingDB, exists := b.dbs[namespace]; exists {
		return existingDB, nil
	}

	// Create database file path: <baseDir>/<namespace>.db
	dbPath := filepath.Join(b.baseDir, fmt.Sprintf("%s.db", namespace))

	db, err := bbolt.Open(dbPath, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", dbPath, err)
	}

	// Store the database connection
	b.dbs[namespace] = db

	return db, nil
}

// Get retrieves a JSON value by key from namespace and collection
func (b *BBoltKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
	namespace = kv.NormalizeNamespace(namespace)
	db, err := b.getDB(namespace)
	if err != nil {
		return nil, err
	}

	var value []byte
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(collection))
		if bucket == nil {
			return kv.ErrKeyNotFound
		}

		value = bucket.Get([]byte(key))
		if value == nil {
			return kv.ErrKeyNotFound
		}

		// Copy the value since it's only valid within the transaction
		value = append([]byte(nil), value...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Set stores a JSON value by key in namespace and collection
func (b *BBoltKV) Set(ctx context.Context, namespace, collection, key string, value []byte) error {
	namespace = kv.NormalizeNamespace(namespace)
	db, err := b.getDB(namespace)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", collection, err)
		}

		return bucket.Put([]byte(key), value)
	})
}

// Delete removes a key-value pair from namespace and collection
func (b *BBoltKV) Delete(ctx context.Context, namespace, collection, key string) error {
	namespace = kv.NormalizeNamespace(namespace)
	db, err := b.getDB(namespace)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(collection))
		if bucket == nil {
			return kv.ErrKeyNotFound
		}

		value := bucket.Get([]byte(key))
		if value == nil {
			return kv.ErrKeyNotFound
		}

		return bucket.Delete([]byte(key))
	})
}

// Exists checks if a key exists in namespace and collection
func (b *BBoltKV) Exists(ctx context.Context, namespace, collection, key string) (bool, error) {
	namespace = kv.NormalizeNamespace(namespace)
	db, err := b.getDB(namespace)
	if err != nil {
		return false, err
	}

	exists := false
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(collection))
		if bucket == nil {
			return nil
		}

		value := bucket.Get([]byte(key))
		exists = value != nil
		return nil
	})

	return exists, err
}

// Close closes all database connections
func (b *BBoltKV) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var lastErr error
	for namespace, db := range b.dbs {
		if err := db.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close database %s: %w", namespace, err)
		}
		delete(b.dbs, namespace)
	}

	return lastErr
}

// Ping checks if the connection is alive
func (b *BBoltKV) Ping(ctx context.Context) error {
	// Try to open a test database to verify the base directory is accessible
	testDB, err := bbolt.Open(filepath.Join(b.baseDir, ".ping.db"), 0o600, nil)
	if err != nil {
		return errors.Join(kv.ErrConnectionFailed, err)
	}
	defer func() {
		if closeErr := testDB.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	return testDB.View(func(tx *bbolt.Tx) error {
		return nil
	})
}
