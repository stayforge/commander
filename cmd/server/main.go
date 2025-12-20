package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iktahana/access-authorization-service/internal/config"
	"github.com/iktahana/access-authorization-service/internal/database"
	"github.com/iktahana/access-authorization-service/internal/handlers"
	"github.com/iktahana/access-authorization-service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Access Authorization Service...")
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Server Port: %s", cfg.ServerPort)

	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongodb, err := database.Connect(ctx, cfg.MongoDBURI, cfg.MongoDBDatabase, cfg.MongoDBCollection)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongodb.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	log.Printf("Successfully connected to MongoDB Atlas")

	// Initialize services
	cardService := service.NewCardService(mongodb.GetCollection())

	// Initialize handlers
	identifyHandler := handlers.NewIdentifyHandler(cardService)

	// Setup Gin router
	// Set mode based on environment
	if cfg.Environment == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(LoggingMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"environment": cfg.Environment,
			"timestamp":   time.Now().UTC(),
		})
	})

	// Register routes
	api := router.Group("/")
	identifyHandler.RegisterRoutes(api)

	// Setup HTTP server
	server := &http.Server{
		Addr:           ":" + cfg.ServerPort,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Device-SN, X-Environment")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs request details
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[%s] %s %s - Status: %d - Latency: %v - IP: %s",
			method, path, c.Request.Proto, statusCode, latency, clientIP)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("Error: %v", e.Err)
			}
		}
	}
}
