package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"commander/internal/config"
)

// ExampleUsage demonstrates how to use the KV abstraction layer with namespace and collection
func ExampleUsage() {
	// Load configuration from environment variables
	cfg := config.LoadConfig()

	// Create KV store based on configuration
	kv, err := NewKV(cfg)
	if err != nil {
		log.Fatalf("Failed to create KV store: %v", err)
	}
	defer func() {
		if closeErr := kv.Close(); closeErr != nil {
			log.Printf("Failed to close KV store: %v", closeErr)
		}
	}()

	ctx := context.Background()

	// Ping to check connection
	if pingErr := kv.Ping(ctx); pingErr != nil {
		_ = kv.Close()                                     //nolint:errcheck // Best effort close before exit
		log.Fatalf("Failed to ping KV store: %v", pingErr) //nolint:gocritic // Example code intentionally exits
	}

	// Define namespace and collection
	// If namespace is empty, it will default to "default"
	namespace := "commander" // Use "" to test default namespace
	collection := "cards"
	key := "card_001"

	// Create a JSON value (example: card data)
	value := map[string]interface{}{
		"name":        "Fire Dragon",
		"card_number": "ABC123DEF456",
		"devices":     []string{"device-001", "device-002"},
		"status":      "active",
	}
	valueBytes, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		log.Fatalf("Failed to marshal value: %v", marshalErr)
	}

	// Set a value with namespace and collection
	if setErr := kv.Set(ctx, namespace, collection, key, valueBytes); setErr != nil {
		log.Fatalf("Failed to set value: %v", setErr)
	}
	fmt.Printf("Set key: %s in namespace: %s, collection: %s\n", key, namespace, collection)

	// Check if key exists
	exists, err := kv.Exists(ctx, namespace, collection, key)
	if err != nil {
		log.Fatalf("Failed to check existence: %v", err)
	}
	fmt.Printf("Key exists: %v\n", exists)

	// Get the value
	retrieved, err := kv.Get(ctx, namespace, collection, key)
	if err != nil {
		log.Fatalf("Failed to get value: %v", err)
	}
	fmt.Printf("Retrieved value: %s\n", string(retrieved))

	// Delete the key
	if err := kv.Delete(ctx, namespace, collection, key); err != nil {
		log.Fatalf("Failed to delete key: %v", err)
	}
	fmt.Printf("Deleted key: %s\n", key)
}
