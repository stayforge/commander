package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"commander/internal/config"
	"commander/internal/database"
	"commander/internal/database/mongodb"
	"commander/internal/handlers"
	"commander/internal/kv"
	"commander/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var (
	version = "dev"     // 默認值
	commit  = "unknown" // set via ldflags during build
	date    = "unknown" // set via ldflags during build
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	cfg.Version = version
	log.Printf("Commander version: %s (commit: %s, built: %s)", version, commit, date)

	// Set Gin mode based on environment
	if cfg.Server.Environment == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize KV store
	kvStore, err := database.NewKV(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize KV store: %v", err)
	}
	defer func() {
		if closeErr := kvStore.Close(); closeErr != nil {
			log.Printf("Failed to close KV store: %v", closeErr)
		}
	}()

	// Verify KV connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := kvStore.Ping(ctx); err != nil {
		cancel()
		log.Fatalf("Failed to ping KV store: %v", err) //nolint:gocritic // Intentional exit on startup failure
	}
	cancel()

	// Initialize Card Service (only for MongoDB backend)
	var cardService *services.CardService
	if cfg.KV.BackendType == config.BackendMongoDB {
		// Type assertion to get MongoDB client
		if mongoKV, ok := kvStore.(*mongodb.MongoDBKV); ok {
			cardService = services.NewCardService(mongoKV.GetClient())
			log.Println("Card verification service initialized (MongoDB backend)")
		} else {
			log.Println("Warning: MongoDB backend expected but type assertion failed")
		}
	} else {
		log.Printf("Card verification service not available (backend: %s, requires MongoDB)", cfg.KV.BackendType)
	}

	// Create Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Set config for handlers
	handlers.Config = cfg

	// Register routes
	setupRoutes(router, kvStore, cardService)

	// Create HTTP server
	port := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRoutes(router *gin.Engine, kvStore kv.KV, cardService *services.CardService) {
	// Health check
	router.GET("/health", handlers.HealthHandler)

	// Root
	router.GET("/", handlers.RootHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// ========== KV CRUD operations (Commented for MVP) ==========
		// GET /api/v1/kv/{namespace}/{collection}/{key}
		// v1.GET("/kv/:namespace/:collection/:key", handlers.GetKVHandler(kvStore))

		// POST /api/v1/kv/{namespace}/{collection}/{key}
		// v1.POST("/kv/:namespace/:collection/:key", handlers.SetKVHandler(kvStore))

		// DELETE /api/v1/kv/{namespace}/{collection}/{key}
		// v1.DELETE("/kv/:namespace/:collection/:key", handlers.DeleteKVHandler(kvStore))

		// HEAD /api/v1/kv/{namespace}/{collection}/{key}
		// v1.HEAD("/kv/:namespace/:collection/:key", handlers.HeadKVHandler(kvStore))

		// ========== Batch operations (Commented for MVP) ==========
		// POST /api/v1/kv/batch (batch set)
		// v1.POST("/kv/batch", handlers.BatchSetHandler(kvStore))

		// DELETE /api/v1/kv/batch (batch delete)
		// v1.DELETE("/kv/batch", handlers.BatchDeleteHandler(kvStore))

		// ========== List and Management (Commented for MVP) ==========
		// GET /api/v1/kv/{namespace}/{collection} (list keys)
		// v1.GET("/kv/:namespace/:collection", handlers.ListKeysHandler(kvStore))

		// GET /api/v1/namespaces (list namespaces)
		// v1.GET("/namespaces", handlers.ListNamespacesHandler(kvStore))

		// GET /api/v1/namespaces/{namespace}/collections (list collections)
		// v1.GET("/namespaces/:namespace/collections", handlers.ListCollectionsHandler(kvStore))

		// GET /api/v1/namespaces/{namespace}/info (get namespace info)
		// v1.GET("/namespaces/:namespace/info", handlers.GetNamespaceInfoHandler(kvStore))

		// DELETE /api/v1/namespaces/{namespace} (delete namespace)
		// v1.DELETE("/namespaces/:namespace", handlers.DeleteNamespaceHandler(kvStore))

		// DELETE /api/v1/namespaces/{namespace}/collections/{collection} (delete collection)
		// v1.DELETE("/namespaces/:namespace/collections/:collection", handlers.DeleteCollectionHandler(kvStore))

		// ========== Card Verification (MVP) ==========
		if cardService != nil {
			// Standard card verification endpoint
			v1.GET("/namespaces/:namespace/device/:device_sn/card/:card_number",
				handlers.CardVerificationHandler(cardService))

			// vguang-350 model compatibility endpoint
			v1.GET("/namespaces/:namespace/device/:device_sn/card/:card_number/vguang-350",
				handlers.CardVerificationVguang350Handler(cardService))
		}
	}
}
