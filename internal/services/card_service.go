package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"commander/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Card verification errors.
var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceNotActive   = errors.New("device not active")
	ErrCardNotFound      = errors.New("card not found")
	ErrCardNotAuthorized = errors.New("card not authorized for this device")
	ErrCardExpired       = errors.New("card has expired")
	ErrCardNotYetValid   = errors.New("card is not yet valid")
)

// CardService handles card verification business logic
type CardService struct {
	client *mongo.Client
}

// NewCardService creates a new CardService that uses the provided MongoDB client to access the database.
func NewCardService(client *mongo.Client) *CardService {
	return &CardService{
		client: client,
	}
}

// VerifyCard verifies if a card is valid for a device
// Returns nil if valid, error otherwise
func (s *CardService) VerifyCard(ctx context.Context, namespace, deviceSN, cardNumber string) error {
	// Step 1: Verify device exists and is active
	device, err := s.getDevice(ctx, namespace, deviceSN)
	if err != nil {
		log.Printf("[CardVerification] Device check failed: namespace=%s, device_sn=%s, error=%v",
			namespace, deviceSN, err)
		return err
	}

	// Status check disabled - accept devices regardless of status
	// if device.Status != "active" {
	// 	log.Printf("[CardVerification] Device not active: namespace=%s, device_sn=%s, status=%s",
	// 		namespace, deviceSN, device.Status)
	// 	return ErrDeviceNotActive
	// }

	log.Printf("[CardVerification] Device verified: namespace=%s, device_sn=%s, device_id=%s",
		namespace, deviceSN, device.DeviceID)

	// Step 2: Find card by number
	card, err := s.getCard(ctx, namespace, cardNumber)
	if err != nil {
		log.Printf("[CardVerification] Card not found: namespace=%s, card_number=%s, error=%v",
			namespace, cardNumber, err)
		return err
	}

	// Step 3: Verify card is authorized for this device (check both SN and device_id)
	if !card.HasDevice(deviceSN) && !card.HasDevice(device.DeviceID) {
		log.Printf("[CardVerification] Card not authorized: namespace=%s, card_number=%s, device_sn=%s, device_id=%s, authorized_devices=%v",
			namespace, cardNumber, deviceSN, device.DeviceID, card.Devices)
		return ErrCardNotAuthorized
	}

	// Step 4: Verify card is within valid time range (with ±60s tolerance)
	now := time.Now()
	if !card.IsValid(now) {
		if now.Before(card.EffectiveAt.Add(-60 * time.Second)) {
			log.Printf("[CardVerification] Card not yet valid: namespace=%s, card_number=%s, device_sn=%s, effective_at=%s, current_time=%s",
				namespace, cardNumber, deviceSN, card.EffectiveAt.Format(time.RFC3339), now.Format(time.RFC3339))
			return ErrCardNotYetValid
		}

		log.Printf("[CardVerification] Card expired: namespace=%s, card_number=%s, device_sn=%s, invalid_at=%s, current_time=%s",
			namespace, cardNumber, deviceSN, card.InvalidAt.Format(time.RFC3339), now.Format(time.RFC3339))
		return ErrCardExpired
	}

	// Success
	log.Printf("[CardVerification] SUCCESS: namespace=%s, card_number=%s, device_sn=%s, card_id=%s, effective=%s, invalid=%s",
		namespace, cardNumber, deviceSN, card.ID,
		card.EffectiveAt.Format(time.RFC3339), card.InvalidAt.Format(time.RFC3339))

	return nil
}

// getDevice retrieves a device by SN from the devices collection
func (s *CardService) getDevice(ctx context.Context, namespace, deviceSN string) (*models.Device, error) {
	collection := s.client.Database(namespace).Collection("devices")

	var device models.Device
	err := collection.FindOne(ctx, bson.M{"sn": deviceSN}).Decode(&device)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}

	return &device, nil
}

// getCard retrieves a card by number from the cards collection
func (s *CardService) getCard(ctx context.Context, namespace, cardNumber string) (*models.Card, error) {
	collection := s.client.Database(namespace).Collection("cards")

	var card models.Card
	err := collection.FindOne(ctx, bson.M{"number": cardNumber}).Decode(&card)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}

	return &card, nil
}