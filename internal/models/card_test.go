package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeviceModel(t *testing.T) {
	device := &Device{
		ID:          "dev-001",
		TenantID:    "tenant-001",
		DeviceID:    "device-001",
		SN:          "SN20250101001",
		DisplayName: "Front Door Lock",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, "dev-001", device.ID)
	assert.Equal(t, "active", device.Status)
	assert.Equal(t, "SN20250101001", device.SN)
}

func TestCardModel(t *testing.T) {
	now := time.Now()
	card := &Card{
		ID:             "card-001",
		OrganizationID: "org-001",
		Number:         "11110011",
		DisplayName:    "Room 101 Card",
		Devices:        []string{"SN20250101001", "SN20250101002"},
		EffectiveAt:    now.Add(-1 * time.Hour),
		InvalidAt:      now.Add(1 * time.Hour),
		BarcodeType:    "qrcode",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	assert.Equal(t, "card-001", card.ID)
	assert.Equal(t, "11110011", card.Number)
	assert.Len(t, card.Devices, 2)
}

func TestCardIsValid_WithinRange(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		effectiveAt  time.Time
		invalidAt    time.Time
		checkTime    time.Time
		expectedBool bool
	}{
		{
			name:         "time in middle of range",
			effectiveAt:  now.Add(-2 * time.Hour),
			invalidAt:    now.Add(2 * time.Hour),
			checkTime:    now,
			expectedBool: true,
		},
		{
			name:         "time at effective boundary with tolerance",
			effectiveAt:  now.Add(30 * time.Second),
			invalidAt:    now.Add(2 * time.Hour),
			checkTime:    now,
			expectedBool: true, // within 60s tolerance
		},
		{
			name:         "time at invalid boundary with tolerance",
			effectiveAt:  now.Add(-2 * time.Hour),
			invalidAt:    now.Add(-30 * time.Second),
			checkTime:    now,
			expectedBool: true, // within 60s tolerance
		},
		{
			name:         "time before effective with tolerance",
			effectiveAt:  now.Add(120 * time.Second),
			invalidAt:    now.Add(2 * time.Hour),
			checkTime:    now,
			expectedBool: false, // beyond 60s tolerance
		},
		{
			name:         "time after invalid with tolerance",
			effectiveAt:  now.Add(-2 * time.Hour),
			invalidAt:    now.Add(-120 * time.Second),
			checkTime:    now,
			expectedBool: false, // beyond 60s tolerance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{
				EffectiveAt: tt.effectiveAt,
				InvalidAt:   tt.invalidAt,
			}
			result := card.IsValid(tt.checkTime)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestCardHasDevice(t *testing.T) {
	tests := []struct {
		name         string
		devices      []string
		searchDevice string
		expectedBool bool
	}{
		{
			name:         "device exists",
			devices:      []string{"SN001", "SN002", "SN003"},
			searchDevice: "SN002",
			expectedBool: true,
		},
		{
			name:         "device not in list",
			devices:      []string{"SN001", "SN002"},
			searchDevice: "SN999",
			expectedBool: false,
		},
		{
			name:         "empty devices array",
			devices:      []string{},
			searchDevice: "SN001",
			expectedBool: false,
		},
		{
			name:         "nil devices array",
			devices:      nil,
			searchDevice: "SN001",
			expectedBool: false,
		},
		{
			name:         "single device match",
			devices:      []string{"SN001"},
			searchDevice: "SN001",
			expectedBool: true,
		},
		{
			name:         "first device in list",
			devices:      []string{"SN001", "SN002", "SN003"},
			searchDevice: "SN001",
			expectedBool: true,
		},
		{
			name:         "last device in list",
			devices:      []string{"SN001", "SN002", "SN003"},
			searchDevice: "SN003",
			expectedBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{Devices: tt.devices}
			result := card.HasDevice(tt.searchDevice)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestCardIsValidEdgeCases(t *testing.T) {
	now := time.Now()

	// Test exact boundaries
	t.Run("exactly at effective time", func(t *testing.T) {
		card := &Card{
			EffectiveAt: now,
			InvalidAt:   now.Add(1 * time.Hour),
		}
		// Should be valid: now > (now - 60s) is true, now < (now + 1h + 60s) is true
		result := card.IsValid(now)
		assert.True(t, result)
	})

	t.Run("exactly at invalid time", func(t *testing.T) {
		card := &Card{
			EffectiveAt: now.Add(-1 * time.Hour),
			InvalidAt:   now,
		}
		// Should be valid: now > (now - 1h - 60s) is true, now < (now + 60s) is true
		result := card.IsValid(now)
		assert.True(t, result)
	})

	t.Run("59 seconds before effective", func(t *testing.T) {
		card := &Card{
			EffectiveAt: now.Add(59 * time.Second),
			InvalidAt:   now.Add(1 * time.Hour),
		}
		// now > (now + 59s - 60s) = now > (now - 1s) = true
		// now < (now + 1h + 60s) = true
		result := card.IsValid(now)
		assert.True(t, result)
	})

	t.Run("61 seconds before effective", func(t *testing.T) {
		card := &Card{
			EffectiveAt: now.Add(61 * time.Second),
			InvalidAt:   now.Add(1 * time.Hour),
		}
		// now > (now + 61s - 60s) = now > (now + 1s) = false
		result := card.IsValid(now)
		assert.False(t, result)
	})

	t.Run("59 seconds after invalid", func(t *testing.T) {
		card := &Card{
			EffectiveAt: now.Add(-1 * time.Hour),
			InvalidAt:   now.Add(-59 * time.Second),
		}
		// now > (now - 1h - 60s) = true
		// now < (now - 59s + 60s) = now < (now + 1s) = true
		result := card.IsValid(now)
		assert.True(t, result)
	})

	t.Run("61 seconds after invalid", func(t *testing.T) {
		card := &Card{
			EffectiveAt: now.Add(-1 * time.Hour),
			InvalidAt:   now.Add(-61 * time.Second),
		}
		// now > (now - 1h - 60s) = true
		// now < (now - 61s + 60s) = now < (now - 1s) = false
		result := card.IsValid(now)
		assert.False(t, result)
	})
}

func TestCardHasDeviceCaseSensitive(t *testing.T) {
	// Device lookup should be case-sensitive
	card := &Card{
		Devices: []string{"SN001", "SN002"},
	}

	t.Run("exact case match", func(t *testing.T) {
		assert.True(t, card.HasDevice("SN001"))
	})

	t.Run("different case no match", func(t *testing.T) {
		assert.False(t, card.HasDevice("sn001"))
	})

	t.Run("different case no match uppercase", func(t *testing.T) {
		assert.False(t, card.HasDevice("sn001"))
	})
}
