package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iktahana/access-authorization-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCardNotFound        = errors.New("card not found")
	ErrCardNotActive       = errors.New("card is not active yet (before start time)")
	ErrCardExpired         = errors.New("card has expired")
	ErrDeviceNotAuthorized = errors.New("device is not authorized for this card")
)

// CardService handles all card-related business logic
type CardService struct {
	collection *mongo.Collection
}

// NewCardService creates a new card service
func NewCardService(collection *mongo.Collection) *CardService {
	return &CardService{
		collection: collection,
	}
}

// GetCard retrieves a card by card number from the database
func (s *CardService) GetCard(ctx context.Context, cardNumber string) (*models.Card, error) {
	// Convert card number to uppercase for consistency
	cardNumber = strings.ToUpper(cardNumber)

	var card models.Card
	filter := bson.M{"card_number": cardNumber}

	err := s.collection.FindOne(ctx, filter).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}

	return &card, nil
}

// IsCardActive checks if the card is within its valid time range
func (s *CardService) IsCardActive(card *models.Card) bool {
	now := time.Now().UTC()

	// Calculate activation time with offset
	// The offset allows cards to be active slightly before the invalid_at time
	// to compensate for NTP clock drift
	activationTime := card.InvalidAt.Add(-time.Duration(card.ActivationOffsetSeconds) * time.Second)

	// Card is active if current time is after activation time and before expiration
	return now.After(activationTime) || now.Equal(activationTime) && (now.Before(card.ExpiredAt) || now.Equal(card.ExpiredAt))
}

// IsDeviceAuthorized checks if the device is in the card's authorized devices list
func (s *CardService) IsDeviceAuthorized(card *models.Card, deviceSN string) bool {
	for _, device := range card.Devices {
		if device == deviceSN {
			return true
		}
	}
	return false
}

// IdentifyByDeviceAndCard performs the complete verification:
// 1. Get card from database
// 2. Check if card is active
// 3. Check if device is authorized
func (s *CardService) IdentifyByDeviceAndCard(ctx context.Context, deviceSN, cardNumber string) (*models.Card, error) {
	// Get the card
	card, err := s.GetCard(ctx, cardNumber)
	if err != nil {
		return nil, err
	}

	// Check if card is active
	if !s.IsCardActive(card) {
		now := time.Now().UTC()
		activationTime := card.InvalidAt.Add(-time.Duration(card.ActivationOffsetSeconds) * time.Second)

		if now.Before(activationTime) {
			return nil, ErrCardNotActive
		}
		return nil, ErrCardExpired
	}

	// Check if device is authorized
	if !s.IsDeviceAuthorized(card, deviceSN) {
		return nil, ErrDeviceNotAuthorized
	}

	return card, nil
}
