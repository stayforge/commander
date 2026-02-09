package mongodb

import (
	"context"
	"errors"
	"time"

	"commander/internal/kv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBKV implements KV interface using MongoDB
// namespace = database, collection = collection
//
//nolint:revive // MongoDBKV name is intentional to match package name
type MongoDBKV struct {
	client *mongo.Client
	uri    string
}

// NewMongoDBKV creates a MongoDB-backed key-value store and verifies connectivity.
// It connects using the provided URI with a 10-second timeout and pings the server.
// On connection or ping failure it returns an error wrapped with kv.ErrConnectionFailed.
func NewMongoDBKV(uri string) (*MongoDBKV, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Join(kv.ErrConnectionFailed, err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, errors.Join(kv.ErrConnectionFailed, err)
	}

	return &MongoDBKV{
		client: client,
		uri:    uri,
	}, nil
}

// getCollection returns the collection for the given namespace and collection
// namespace is used as database name, collection is used as collection name
func (m *MongoDBKV) getCollection(namespace, collection string) *mongo.Collection {
	db := m.client.Database(namespace)
	return db.Collection(collection)
}

// ensureIndex ensures unique index on key for the collection
func (m *MongoDBKV) ensureIndex(ctx context.Context, coll *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	// Ignore errors if index already exists
	return err
}

// Get retrieves a JSON value by key from namespace and collection
func (m *MongoDBKV) Get(ctx context.Context, namespace, collection, key string) ([]byte, error) {
	namespace = kv.NormalizeNamespace(namespace)
	coll := m.getCollection(namespace, collection)
	_ = m.ensureIndex(ctx, coll) //nolint:errcheck // Best effort index creation

	var doc struct {
		Key   string `bson:"key"`
		Value string `bson:"value"`
	}

	err := coll.FindOne(ctx, bson.M{"key": key}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, kv.ErrKeyNotFound
		}
		return nil, err
	}

	return []byte(doc.Value), nil
}

// Set stores a JSON value by key in namespace and collection
func (m *MongoDBKV) Set(ctx context.Context, namespace, collection, key string, value []byte) error {
	namespace = kv.NormalizeNamespace(namespace)
	coll := m.getCollection(namespace, collection)
	_ = m.ensureIndex(ctx, coll) //nolint:errcheck // Best effort index creation

	doc := bson.M{
		"key":   key,
		"value": string(value),
	}

	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{"$set": doc},
		opts,
	)

	return err
}

// Delete removes a key-value pair from namespace and collection
func (m *MongoDBKV) Delete(ctx context.Context, namespace, collection, key string) error {
	namespace = kv.NormalizeNamespace(namespace)
	coll := m.getCollection(namespace, collection)

	result, err := coll.DeleteOne(ctx, bson.M{"key": key})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return kv.ErrKeyNotFound
	}
	return nil
}

// Exists checks if a key exists in namespace and collection
func (m *MongoDBKV) Exists(ctx context.Context, namespace, collection, key string) (bool, error) {
	namespace = kv.NormalizeNamespace(namespace)
	coll := m.getCollection(namespace, collection)

	count, err := coll.CountDocuments(ctx, bson.M{"key": key})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close closes the MongoDB connection
func (m *MongoDBKV) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

// Ping checks if the connection is alive
func (m *MongoDBKV) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// GetClient returns the underlying MongoDB client for advanced operations
// This is used by business services that need MongoDB-specific features
func (m *MongoDBKV) GetClient() *mongo.Client {
	return m.client
}