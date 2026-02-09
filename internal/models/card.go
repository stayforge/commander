package models

import "time"

// Device represents a device document in MongoDB
type Device struct {
	ID          string                 `bson:"_id"`
	TenantID    string                 `bson:"tenant_id"`
	DeviceID    string                 `bson:"device_id"`
	SN          string                 `bson:"sn"`
	DisplayName string                 `bson:"display_name"`
	Status      string                 `bson:"status"` // "active", "inactive", etc.
	Metadata    map[string]interface{} `bson:"metadata"`
	CreatedAt   time.Time              `bson:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at"`
}

// Card represents a card document in MongoDB
type Card struct {
	ID             string    `bson:"_id"`
	OrganizationID string    `bson:"organization_id"`
	Number         string    `bson:"number"`
	DisplayName    string    `bson:"display_name"`
	Devices        []string  `bson:"devices"` // Array of device SNs
	EffectiveAt    time.Time `bson:"effective_at"`
	InvalidAt      time.Time `bson:"invalid_at"`
	BarcodeType    string    `bson:"barcode_type"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// IsValid checks if the card is valid at the given time
// Allows ±60 seconds tolerance for NTP drift
func (c *Card) IsValid(now time.Time) bool {
	tolerance := 60 * time.Second
	effectiveWithTolerance := c.EffectiveAt.Add(-tolerance)
	invalidWithTolerance := c.InvalidAt.Add(tolerance)

	return now.After(effectiveWithTolerance) && now.Before(invalidWithTolerance)
}

// HasDevice checks if the card is authorized for the given device SN
func (c *Card) HasDevice(deviceSN string) bool {
	if len(c.Devices) == 0 {
		return false // Empty array = not authorized for any device
	}

	for _, sn := range c.Devices {
		if sn == deviceSN {
			return true
		}
	}
	return false
}
