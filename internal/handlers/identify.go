package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/iktahana/access-authorization-service/internal/models"
	"github.com/iktahana/access-authorization-service/internal/service"
)

// IdentifyHandler handles card identification requests
type IdentifyHandler struct {
	cardService *service.CardService
}

// NewIdentifyHandler creates a new identify handler
func NewIdentifyHandler(cardService *service.CardService) *IdentifyHandler {
	return &IdentifyHandler{
		cardService: cardService,
	}
}

// RegisterRoutes registers all identify routes
func (h *IdentifyHandler) RegisterRoutes(router *gin.RouterGroup) {
	identify := router.Group("/identify")
	{
		// JSON endpoints
		identify.POST("/json", h.IdentifyJSON)
		identify.POST("/json/:device_sn", h.IdentifyJSON)

		// vguang-m350 specific endpoint
		vguang := identify.Group("/vguang-m350")
		vguang.POST("/:device_name", h.VguangIdentify)
	}
}

// IdentifyJSON handles JSON-based card identification
// @Summary Identify a device by card number
// @Description Identify a device by its serial number and card number
// @Tags Identify
// @Accept json
// @Produce json
// @Param device_sn path string false "Device serial number (can also be in header)"
// @Param X-Device-SN header string false "Device serial number (alternative to path param)"
// @Param X-Environment header string false "Environment (default: STANDARD)"
// @Param card body models.CardQuery true "Card query"
// @Success 200 {object} models.CardIdentifyResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /identify/json [post]
// @Router /identify/json/{device_sn} [post]
func (h *IdentifyHandler) IdentifyJSON(c *gin.Context) {
	// Get device SN from path parameter or header
	deviceSN := c.Param("device_sn")
	if deviceSN == "" {
		deviceSN = c.GetHeader("X-Device-SN")
	}

	if deviceSN == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Device SN is required (either in path or X-Device-SN header)",
		})
		return
	}

	// Parse request body
	var cardQuery models.CardQuery
	if err := c.ShouldBindJSON(&cardQuery); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verify the card
	card, err := h.cardService.IdentifyByDeviceAndCard(ctx, deviceSN, cardQuery.CardNumber)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, service.ErrCardNotFound) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, models.CardIdentifyResponse{
		Message:                 "Successfully",
		CardNumber:              card.CardNumber,
		Devices:                 card.Devices,
		InvalidAt:               card.InvalidAt,
		ExpiredAt:               card.ExpiredAt,
		ActivationOffsetSeconds: card.ActivationOffsetSeconds,
		OwnerClientID:           card.OwnerClientID,
		Name:                    card.Name,
	})
}

// VguangIdentify handles special vguang-m350 device identification
// This endpoint has special byte-reversal logic for hardware compatibility
// @Summary vguang-m350 specific identification endpoint
// @Description API specifically open for vguang-m350. Only runs in STANDARD environment.
// @Tags Identify:vguang
// @Accept plain
// @Produce plain
// @Param device_name path string true "Device name"
// @Success 200 {string} string "code=0000"
// @Failure 404 {object} models.ErrorResponse
// @Router /identify/vguang-m350/{device_name} [post]
func (h *IdentifyHandler) VguangIdentify(c *gin.Context) {
	deviceName := c.Param("device_name")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Device name is required",
		})
		return
	}

	// Read raw body
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Failed to read request body",
		})
		return
	}

	var cardNumber string

	// Try to decode as UTF-8 text
	textContent := strings.TrimSpace(string(rawBody))

	// Check if all characters are alphanumeric
	isAlphanumeric := true
	if textContent != "" {
		for _, ch := range textContent {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
				isAlphanumeric = false
				break
			}
		}
	} else {
		isAlphanumeric = false
	}

	if isAlphanumeric {
		// Use as card number directly (uppercase)
		cardNumber = strings.ToUpper(textContent)
	} else {
		// Reverse bytes and convert to hex
		reversed := make([]byte, len(rawBody))
		for i := 0; i < len(rawBody); i++ {
			reversed[i] = rawBody[len(rawBody)-1-i]
		}
		cardNumber = strings.ToUpper(hex.EncodeToString(reversed))
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verify the card
	_, err = h.cardService.IdentifyByDeviceAndCard(ctx, deviceName, cardNumber)
	if err != nil {
		statusCode := http.StatusNotFound
		if !errors.Is(err, service.ErrCardNotFound) {
			// Log the error for debugging
			c.Error(err)
		}

		c.JSON(statusCode, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Return plain text success response
	c.String(http.StatusOK, "code=0000")
}
