package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"commander/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCardVerificationHandler_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock CardService (will not be called)
	mockService := services.NewCardService(&mongo.Client{})

	tests := []struct {
		name         string
		namespace    string
		deviceSN     string
		cardNumber   string
		expectBadReq bool
	}{
		{
			name:         "all parameters present",
			namespace:    "org_test",
			deviceSN:     "SN001",
			cardNumber:   "card001",
			expectBadReq: false, // Will fail during verification (no mock data), but params are valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test/:namespace/device/:device_sn/card/:card_number",
				CardVerificationHandler(mockService))

			url := fmt.Sprintf("/test/%s/device/%s/card/%s", tt.namespace, tt.deviceSN, tt.cardNumber)
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tt.expectBadReq {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Contains(t, w.Body.String(), "invalid_parameters")
			}
		})
	}
}

func TestCardVerificationVguang350Handler_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock CardService
	mockService := services.NewCardService(&mongo.Client{})

	router := gin.New()
	router.GET("/test/:namespace/device/:device_sn/card/:card_number/vguang-350",
		CardVerificationVguang350Handler(mockService))

	// Test with missing parameters
	req, _ := http.NewRequest(http.MethodGet, "/test//device//card//vguang-350", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_parameters")
}

func TestErrorResponseFormats(t *testing.T) {
	tests := []struct {
		name         string
		errorCode    string
		errorMessage string
		statusCode   int
	}{
		{
			name:         "device_not_found",
			errorCode:    "device_not_found",
			errorMessage: "Device not found",
			statusCode:   http.StatusNotFound,
		},
		{
			name:         "card_not_found",
			errorCode:    "card_not_found",
			errorMessage: "Card not found",
			statusCode:   http.StatusNotFound,
		},
		{
			name:         "device_not_active",
			errorCode:    "device_not_active",
			errorMessage: "Device is not active",
			statusCode:   http.StatusForbidden,
		},
		{
			name:         "card_not_authorized",
			errorCode:    "card_not_authorized",
			errorMessage: "Card is not authorized for this device",
			statusCode:   http.StatusForbidden,
		},
		{
			name:         "card_expired",
			errorCode:    "card_expired",
			errorMessage: "Card has expired",
			statusCode:   http.StatusForbidden,
		},
		{
			name:         "card_not_yet_valid",
			errorCode:    "card_not_yet_valid",
			errorMessage: "Card is not yet valid",
			statusCode:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Map error code to service error
			var err error
			switch tt.errorCode {
			case "device_not_found":
				err = services.ErrDeviceNotFound
			case "device_not_active":
				err = services.ErrDeviceNotActive
			case "card_not_found":
				err = services.ErrCardNotFound
			case "card_not_authorized":
				err = services.ErrCardNotAuthorized
			case "card_expired":
				err = services.ErrCardExpired
			case "card_not_yet_valid":
				err = services.ErrCardNotYetValid
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{
				{Key: "namespace", Value: "org_test"},
				{Key: "device_sn", Value: "SN001"},
				{Key: "card_number", Value: "card001"},
			}

			handleVerificationError(c, err, "org_test", "SN001", "card001")

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.errorCode)
			assert.Contains(t, w.Body.String(), "org_test")
			assert.Contains(t, w.Body.String(), "SN001")
			assert.Contains(t, w.Body.String(), "card001")
			assert.Contains(t, w.Body.String(), "timestamp")
		})
	}
}
