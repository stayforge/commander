package handlers

import (
	"commander/internal/config"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	// Set test mode for gin
	gin.SetMode(gin.TestMode)

	// Initialize Config for tests
	Config = &config.Config{
		Version: "test-v1.0.0",
		Server: config.ServerConfig{
			Port:        "8080",
			Environment: "STANDARD",
		},
	}
}

func TestHealthHandler(t *testing.T) {
	// Create test router
	router := gin.New()
	router.GET("/health", HealthHandler)

	// Create request
	req, _ := http.NewRequest("GET", "/health", http.NoBody)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response fields
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}

	if response["environment"] != "STANDARD" {
		t.Errorf("Expected environment 'STANDARD', got '%v'", response["environment"])
	}

	if response["message"] != "Commander service is running" {
		t.Errorf("Expected message 'Commander service is running', got '%v'", response["message"])
	}

	// Verify timestamp exists and is valid
	if timestamp, ok := response["timestamp"].(string); ok {
		_, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Errorf("Invalid timestamp format: %v", err)
		}
	} else {
		t.Error("Timestamp field missing or not a string")
	}
}

func TestRootHandler(t *testing.T) {
	// Create test router
	router := gin.New()
	router.GET("/", RootHandler)

	// Create request
	req, _ := http.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response fields
	if response["message"] != "Welcome to Commander API" {
		t.Errorf("Expected message 'Welcome to Commander API', got '%v'", response["message"])
	}

	if response["version"] != "test-v1.0.0" {
		t.Errorf("Expected version 'test-v1.0.0', got '%v'", response["version"])
	}
}

func TestRootHandler_WithDifferentVersion(t *testing.T) {
	// Save original config
	originalConfig := Config

	// Set custom version
	Config = &config.Config{
		Version: "v2.0.0-beta",
	}

	// Restore original config after test
	defer func() {
		Config = originalConfig
	}()

	// Create test router
	router := gin.New()
	router.GET("/", RootHandler)

	// Create request
	req, _ := http.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Parse response
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify custom version
	if response["version"] != "v2.0.0-beta" {
		t.Errorf("Expected version 'v2.0.0-beta', got '%v'", response["version"])
	}
}
