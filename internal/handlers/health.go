package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler responds to health check requests with HTTP 200 and a JSON
// payload containing "status", "environment", "message", and a UTC "timestamp"
// formatted in RFC3339.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"environment": "STANDARD",
		"message":     "Commander service is running",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}