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
	version = "dev" // 默認值
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	cfg.Version = version

	// Set Gin mode based on environment
	if cfg.Server.Environment == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize KV store
	kvStore, err := database.NewKV(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize KV store: %v", err)
	}
	defer kvStore.Close()

	// Verify KV connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := kvStore.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping KV store: %v", err)
	}

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
	// v1 := router.Group("/api/v1")
	// {
	// 	// Add your API routes here
	// 	// Example: v1.GET("/items", handlers.GetItems)
	// }
}
