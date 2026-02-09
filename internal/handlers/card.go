package handlers

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"commander/internal/services"

	"github.com/gin-gonic/gin"
)

// CardVerificationHandler handles standard card verification via POST
// POST /api/v1/namespaces/:namespace
// Header: X-Device-SN: <device_sn>
// Body: plain text card number
// Success: 204 No Content
// HTTP status codes.
func CardVerificationHandler(cardService *services.CardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		deviceSN := c.GetHeader("X-Device-SN")

		// Validate header
		if deviceSN == "" {
			log.Printf("[CardVerification] Missing X-Device-SN header: namespace=%s", namespace)
			c.Status(http.StatusBadRequest)
			return
		}

		// Read body (plain text card number)
		rawBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("[CardVerification] Failed to read body: namespace=%s, device_sn=%s, error=%v",
				namespace, deviceSN, err)
			c.Status(http.StatusBadRequest)
			return
		}

		cardNumber := strings.TrimSpace(string(rawBody))
		if cardNumber == "" {
			log.Printf("[CardVerification] Empty card number: namespace=%s, device_sn=%s",
				namespace, deviceSN)
			c.Status(http.StatusBadRequest)
			return
		}

		// Verify card
		err = cardService.VerifyCard(c.Request.Context(), namespace, deviceSN, cardNumber)
		if err != nil {
			// Error logging already done in CardService
			c.Status(mapErrorToStatusCode(err))
			return
		}

		// Success - return 204 No Content
		c.Status(http.StatusNoContent)
	}
}

// CardVerificationVguangHandler handles vguang-m350 device compatibility
// POST /api/v1/namespaces/:namespace/device/:device_name/vguang
// Body: plain text or binary card number
// Success: 200 "code=0000"
// CardVerificationVguangHandler returns a Gin handler that verifies card numbers sent by vguang-m350 devices.
// 
// The handler expects URL parameters `namespace` and `device_name`. It reads the raw request body, derives a
// normalized card number using vguang-specific rules (via parseVguangCardNumber), and calls CardService.VerifyCard
// with the request context, namespace, device name, and parsed card number. On success it responds with HTTP 200
// and the exact body "code=0000". Any failure to read the body, parse a card number, or verify the card results
// in an HTTP 404 response; failures are also logged.
func CardVerificationVguangHandler(cardService *services.CardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		deviceName := c.Param("device_name")

		// Read body
		rawBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("[CardVerification:vguang] Failed to read body: namespace=%s, device_name=%s, error=%v",
				namespace, deviceName, err)
			c.Status(http.StatusNotFound)
			return
		}

		// Parse card number (vguang special logic)
		cardNumber := parseVguangCardNumber(rawBody)
		if cardNumber == "" {
			log.Printf("[CardVerification:vguang] Empty card number: namespace=%s, device_name=%s",
				namespace, deviceName)
			c.Status(http.StatusNotFound)
			return
		}

		// Verify card
		err = cardService.VerifyCard(c.Request.Context(), namespace, deviceName, cardNumber)
		if err != nil {
			// Error logging already done in CardService
			c.Status(http.StatusNotFound)
			return
		}

		// Success - must return "code=0000" (exact match for vguang-m350)
		c.String(http.StatusOK, "code=0000")
	}
}

// parseVguangCardNumber parses card number from vguang device
// If alphanumeric: use as-is (uppercase)
// uppercase hexadecimal string.
func parseVguangCardNumber(rawBody []byte) string {
	if len(rawBody) == 0 {
		return ""
	}

	// Try to decode as UTF-8 text
	text := strings.TrimSpace(string(rawBody))

	// Check if alphanumeric (with hyphens)
	if text != "" && isAlphanumeric(text) {
		// Convert to uppercase for consistency
		return strings.ToUpper(text)
	}

	// Otherwise reverse bytes and convert to hex
	reversed := make([]byte, len(rawBody))
	for i, b := range rawBody {
		reversed[len(rawBody)-1-i] = b
	}
	return strings.ToUpper(hex.EncodeToString(reversed))
}

// isAlphanumeric reports whether s contains only ASCII letters, digits, or hyphens.
// It returns true when every rune is 0-9, A-Z, a-z, or '-', and false for any other character.
func isAlphanumeric(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '-') {
			return false
		}
	}
	return true
}

// mapErrorToStatusCode converts service-layer errors into HTTP status codes.
// Errors indicating a missing device or card return 404 Not Found.
// Errors indicating an inactive device, unauthorized card, expired card, or card not yet valid return 403 Forbidden.
// All other errors return 500 Internal Server Error.
func mapErrorToStatusCode(err error) int {
	switch {
	case errors.Is(err, services.ErrDeviceNotFound):
		return http.StatusNotFound
	case errors.Is(err, services.ErrCardNotFound):
		return http.StatusNotFound
	case errors.Is(err, services.ErrDeviceNotActive):
		return http.StatusForbidden
	case errors.Is(err, services.ErrCardNotAuthorized):
		return http.StatusForbidden
	case errors.Is(err, services.ErrCardExpired):
		return http.StatusForbidden
	case errors.Is(err, services.ErrCardNotYetValid):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}