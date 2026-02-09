package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"environment": "STANDARD",
		"message":     "Commander service is running",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}
