package models

import "time"

// Card represents a card document in MongoDB
type Card struct {
	CardNumber              string    `bson:"card_number" json:"card_number"`
	Devices                 []string  `bson:"devices" json:"devices"`
	InvalidAt               time.Time `bson:"invalid_at" json:"invalid_at"`
	ExpiredAt               time.Time `bson:"expired_at" json:"expired_at"`
	ActivationOffsetSeconds int       `bson:"activation_offset_seconds" json:"activation_offset_seconds"`
	OwnerClientID           string    `bson:"owner_client_id,omitempty" json:"owner_client_id,omitempty"`
	Name                    string    `bson:"name,omitempty" json:"name,omitempty"`
}

// CardQuery represents the request body for card identification
type CardQuery struct {
	CardNumber string `json:"card_number" binding:"required"`
}

// CardIdentifyResponse represents the successful identification response
type CardIdentifyResponse struct {
	Message                 string    `json:"message"`
	CardNumber              string    `json:"card_number"`
	Devices                 []string  `json:"devices"`
	InvalidAt               time.Time `json:"invalid_at"`
	ExpiredAt               time.Time `json:"expired_at"`
	ActivationOffsetSeconds int       `json:"activation_offset_seconds"`
	OwnerClientID           string    `json:"owner_client_id,omitempty"`
	Name                    string    `json:"name,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
}
