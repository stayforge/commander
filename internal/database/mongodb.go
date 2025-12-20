package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the database client and configuration
type MongoDB struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

// Connect establishes a connection to MongoDB Atlas
func Connect(ctx context.Context, uri, database, collection string) (*MongoDB, error) {
	// Set client options with timeout
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(database)
	coll := db.Collection(collection)

	return &MongoDB{
		Client:     client,
		Database:   db,
		Collection: coll,
	}, nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// GetCollection returns the MongoDB collection
func (m *MongoDB) GetCollection() *mongo.Collection {
	return m.Collection
}
