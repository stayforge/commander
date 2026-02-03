package handlers

import (
	"errors"
	"net/http"
	"time"

	"commander/internal/services"

	"github.com/gin-gonic/gin"
)

// CardVerificationHandler handles standard card verification
// GET /api/v1/namespaces/:namespace/device/:device_sn/card/:card_number
// Returns: 204 No Content (success) or error JSON
func CardVerificationHandler(cardService *services.CardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		deviceSN := c.Param("device_sn")
		cardNumber := c.Param("card_number")

		// Validate parameters
		if namespace == "" || deviceSN == "" || cardNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     "invalid_parameters",
				"message":   "namespace, device_sn, and card_number are required",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		// Verify card
		err := cardService.VerifyCard(c.Request.Context(), namespace, deviceSN, cardNumber)
		if err != nil {
			handleVerificationError(c, err, namespace, deviceSN, cardNumber)
			return
		}

		// Success - return 204 No Content
		c.Status(http.StatusNoContent)
	}
}

// CardVerificationVguang350Handler handles vguang-350 model compatibility
// GET /api/v1/namespaces/:namespace/device/:device_sn/card/:card_number/vguang-350
// Returns: 200 + "0000" (success) or error JSON
func CardVerificationVguang350Handler(cardService *services.CardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		deviceSN := c.Param("device_sn")
		cardNumber := c.Param("card_number")

		// Validate parameters
		if namespace == "" || deviceSN == "" || cardNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     "invalid_parameters",
				"message":   "namespace, device_sn, and card_number are required",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		// Verify card (same logic as standard endpoint)
		err := cardService.VerifyCard(c.Request.Context(), namespace, deviceSN, cardNumber)
		if err != nil {
			handleVerificationError(c, err, namespace, deviceSN, cardNumber)
			return
		}

		// Success - return 200 + plain text "0000" for vguang-350 compatibility
		c.String(http.StatusOK, "0000")
	}
}

// handleVerificationError handles verification errors and returns appropriate HTTP response
func handleVerificationError(c *gin.Context, err error, namespace, deviceSN, cardNumber string) {
	var statusCode int
	var errorCode string
	var message string

	switch {
	case errors.Is(err, services.ErrDeviceNotFound):
		statusCode = http.StatusNotFound
		errorCode = "device_not_found"
		message = "Device not found"

	case errors.Is(err, services.ErrDeviceNotActive):
		statusCode = http.StatusForbidden
		errorCode = "device_not_active"
		message = "Device is not active"

	case errors.Is(err, services.ErrCardNotFound):
		statusCode = http.StatusNotFound
		errorCode = "card_not_found"
		message = "Card not found"

	case errors.Is(err, services.ErrCardNotAuthorized):
		statusCode = http.StatusForbidden
		errorCode = "card_not_authorized"
		message = "Card is not authorized for this device"

	case errors.Is(err, services.ErrCardExpired):
		statusCode = http.StatusForbidden
		errorCode = "card_expired"
		message = "Card has expired"

	case errors.Is(err, services.ErrCardNotYetValid):
		statusCode = http.StatusForbidden
		errorCode = "card_not_yet_valid"
		message = "Card is not yet valid"

	default:
		statusCode = http.StatusInternalServerError
		errorCode = "internal_error"
		message = "Internal server error"
	}

	c.JSON(statusCode, gin.H{
		"error":       errorCode,
		"message":     message,
		"namespace":   namespace,
		"device_sn":   deviceSN,
		"card_number": cardNumber,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}
