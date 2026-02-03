package services

import (
	"testing"
	"time"

	"commander/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCardIsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		card      *models.Card
		checkTime time.Time
		expected  bool
	}{
		{
			name: "card is within valid time range",
			card: &models.Card{
				EffectiveAt: now.Add(-1 * time.Hour),
				InvalidAt:   now.Add(1 * time.Hour),
			},
			checkTime: now,
			expected:  true,
		},
		{
			name: "card effective_at in future, but within tolerance",
			card: &models.Card{
				EffectiveAt: now.Add(30 * time.Second),
				InvalidAt:   now.Add(1 * time.Hour),
			},
			checkTime: now,
			expected:  true, // now is after (effective - 60s) = now - 30s, which is true
		},
		{
			name: "card invalid_at in past, but within tolerance",
			card: &models.Card{
				EffectiveAt: now.Add(-1 * time.Hour),
				InvalidAt:   now.Add(-30 * time.Second),
			},
			checkTime: now,
			expected:  true, // now is before (invalid + 60s) = now + 30s, which is true
		},
		{
			name: "card is before effective_at (beyond tolerance)",
			card: &models.Card{
				EffectiveAt: now.Add(120 * time.Second),
				InvalidAt:   now.Add(1 * time.Hour),
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "card has expired (beyond tolerance)",
			card: &models.Card{
				EffectiveAt: now.Add(-1 * time.Hour),
				InvalidAt:   now.Add(-120 * time.Second),
			},
			checkTime: now,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.card.IsValid(tt.checkTime)
			assert.Equal(t, tt.expected, result, "card validity check failed")
		})
	}
}

func TestCardHasDevice(t *testing.T) {
	tests := []struct {
		name         string
		devices      []string
		searchDevice string
		expected     bool
	}{
		{
			name:         "device found in list",
			devices:      []string{"device-001", "device-002", "device-003"},
			searchDevice: "device-002",
			expected:     true,
		},
		{
			name:         "device not found",
			devices:      []string{"device-001", "device-002"},
			searchDevice: "device-999",
			expected:     false,
		},
		{
			name:         "empty devices array",
			devices:      []string{},
			searchDevice: "device-001",
			expected:     false,
		},
		{
			name:         "nil devices array",
			devices:      nil,
			searchDevice: "device-001",
			expected:     false,
		},
		{
			name:         "single device match",
			devices:      []string{"device-001"},
			searchDevice: "device-001",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &models.Card{Devices: tt.devices}
			result := card.HasDevice(tt.searchDevice)
			assert.Equal(t, tt.expected, result, "device authorization check failed")
		})
	}
}

// Note: Full CardService.VerifyCard tests require MongoDB integration tests
// These tests focus on the data model validation logic which is testable without MongoDB
