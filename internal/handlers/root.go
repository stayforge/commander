package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RootHandler handles root requests
func RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Commander API",
		"version": "1.0.0",
	})
}

