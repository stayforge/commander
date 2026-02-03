package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"commander/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCardVerificationHandler_POST_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := services.NewCardService(&mongo.Client{})

	router := gin.New()
	router.POST("/api/v1/namespaces/:namespace", CardVerificationHandler(mockService))

	// Missing X-Device-SN header
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/namespaces/org_test", bytes.NewBufferString("card001"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCardVerificationHandler_POST_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := services.NewCardService(&mongo.Client{})

	router := gin.New()
	router.POST("/api/v1/namespaces/:namespace", CardVerificationHandler(mockService))

	// Empty body
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/namespaces/org_test", bytes.NewBufferString(""))
	req.Header.Set("X-Device-SN", "SN001")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCardVerificationHandler_POST_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := services.NewCardService(&mongo.Client{})

	router := gin.New()
	router.POST("/api/v1/namespaces/:namespace", CardVerificationHandler(mockService))

	// Valid request format (will fail verification due to no mock data)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/namespaces/org_test", bytes.NewBufferString("card001"))
	req.Header.Set("X-Device-SN", "SN001")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return error status (no mock DB)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCardVerificationVguangHandler_POST_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := services.NewCardService(&mongo.Client{})

	router := gin.New()
	router.POST("/api/v1/namespaces/:namespace/device/:device_name", CardVerificationVguangHandler(mockService))

	// Empty body
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/namespaces/org_test/device/SN001", bytes.NewBufferString(""))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCardVerificationVguangHandler_POST_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := services.NewCardService(&mongo.Client{})

	router := gin.New()
	router.POST("/api/v1/namespaces/:namespace/device/:device_name", CardVerificationVguangHandler(mockService))

	// Valid request format (will fail verification due to no mock data)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/namespaces/org_test/device/SN001", bytes.NewBufferString("card001"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return error status (no mock DB)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestParseVguangCardNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "alphanumeric lowercase",
			input:    []byte("abc123"),
			expected: "ABC123",
		},
		{
			name:     "alphanumeric uppercase",
			input:    []byte("ABC123"),
			expected: "ABC123",
		},
		{
			name:     "alphanumeric mixed",
			input:    []byte("AbC123"),
			expected: "ABC123",
		},
		{
			name:     "binary data - 4 bytes",
			input:    []byte{0x01, 0x02, 0x03, 0x04},
			expected: "04030201", // reversed hex
		},
		{
			name:     "binary data - single byte",
			input:    []byte{0xFF},
			expected: "FF",
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    []byte("   "),
			expected: "202020", // After trim empty, treated as binary: 3 spaces reversed = 0x20 0x20 0x20 = "202020"
		},
		{
			name:     "mixed alphanumeric with spaces",
			input:    []byte("  ABC123  "),
			expected: "ABC123", // Spaces trimmed, then treated as alphanumeric
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVguangCardNumber(tt.input)
			assert.Equal(t, tt.expected, result, "card number parsing failed")
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "alphanumeric lowercase",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "alphanumeric uppercase",
			input:    "ABC123",
			expected: true,
		},
		{
			name:     "alphanumeric mixed",
			input:    "AbC123",
			expected: true,
		},
		{
			name:     "with special character",
			input:    "ABC123!",
			expected: false,
		},
		{
			name:     "with space",
			input:    "ABC 123",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true, // Technically all chars (none) are alphanumeric
		},
		{
			name:     "only digits",
			input:    "12345",
			expected: true,
		},
		{
			name:     "only letters",
			input:    "ABCDE",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphanumeric(tt.input)
			assert.Equal(t, tt.expected, result, "alphanumeric check failed")
		})
	}
}

func TestMapErrorToStatusCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "device not found",
			err:          services.ErrDeviceNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "card not found",
			err:          services.ErrCardNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "device not active",
			err:          services.ErrDeviceNotActive,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "card not authorized",
			err:          services.ErrCardNotAuthorized,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "card expired",
			err:          services.ErrCardExpired,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "card not yet valid",
			err:          services.ErrCardNotYetValid,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := mapErrorToStatusCode(tt.err)
			assert.Equal(t, tt.expectedCode, code, "status code mapping failed")
		})
	}
}
