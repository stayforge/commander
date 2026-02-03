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
	"commander/internal/handlers"
	"commander/internal/kv"

	"github.com/gin-gonic/gin"
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

	// Create Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Set config for handlers
	handlers.Config = cfg

	// Register routes
	setupRoutes(router, kvStore)

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

func setupRoutes(router *gin.Engine, kvStore kv.KV) {
	// Health check
	router.GET("/health", handlers.HealthHandler)

	// Root
	router.GET("/", handlers.RootHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// KV CRUD operations
		// GET /api/v1/kv/{namespace}/{collection}/{key}
		v1.GET("/kv/:namespace/:collection/:key", handlers.GetKVHandler(kvStore))

		// POST /api/v1/kv/{namespace}/{collection}/{key}
		v1.POST("/kv/:namespace/:collection/:key", handlers.SetKVHandler(kvStore))

		// DELETE /api/v1/kv/{namespace}/{collection}/{key}
		v1.DELETE("/kv/:namespace/:collection/:key", handlers.DeleteKVHandler(kvStore))

		// HEAD /api/v1/kv/{namespace}/{collection}/{key}
		v1.HEAD("/kv/:namespace/:collection/:key", handlers.HeadKVHandler(kvStore))

		// Batch operations
		// POST /api/v1/kv/batch (batch set)
		v1.POST("/kv/batch", handlers.BatchSetHandler(kvStore))

		// DELETE /api/v1/kv/batch (batch delete)
		v1.DELETE("/kv/batch", handlers.BatchDeleteHandler(kvStore))

		// GET /api/v1/kv/{namespace}/{collection} (list keys)
		v1.GET("/kv/:namespace/:collection", handlers.ListKeysHandler(kvStore))
	}
}
