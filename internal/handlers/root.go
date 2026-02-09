package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RootHandler writes a 200 OK JSON response containing a welcome message and the application's version.
// The response body is a JSON object with "message" set to "Welcome to Commander API" and "version" set from Config.Version.
func RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Commander API",
		"version": Config.Version,
	})
}