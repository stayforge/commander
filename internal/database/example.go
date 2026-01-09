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
	defer kv.Close()

	ctx := context.Background()

	// Ping to check connection
	if err := kv.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping KV store: %v", err)
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
	valueBytes, _ := json.Marshal(value)

	// Set a value with namespace and collection
	if err := kv.Set(ctx, namespace, collection, key, valueBytes); err != nil {
		log.Fatalf("Failed to set value: %v", err)
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

