package mocks

import (
	"context"
	"errors"

	"commander/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockCollection implements mongo.Collection interface for testing
type MockCollection struct {
	Documents      map[string]interface{} // key -> document
	CreatedIndexes []string               // track created indexes
	FindError      error
	InsertErr      error
	UpdateErr      error
	DeleteErr      error
	CountErr       error
}

// FindOne mocks mongo.Collection.FindOne
func (m *MockCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	return &mongo.SingleResult{}
}

// InsertOne mocks mongo.Collection.InsertOne
func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	if m.InsertErr != nil {
		return nil, m.InsertErr
	}
	return &mongo.InsertOneResult{InsertedID: "mock-id"}, nil
}

// UpdateOne mocks mongo.Collection.UpdateOne
func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	if m.UpdateErr != nil {
		return nil, m.UpdateErr
	}
	return &mongo.UpdateResult{ModifiedCount: 1}, nil
}

// DeleteOne mocks mongo.Collection.DeleteOne
func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	if m.DeleteErr != nil {
		return nil, m.DeleteErr
	}
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

// CountDocuments mocks mongo.Collection.CountDocuments
func (m *MockCollection) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	if m.CountErr != nil {
		return 0, m.CountErr
	}
	return int64(len(m.Documents)), nil
}

// GetIndexes returns the created indexes (helper method, not part of mongo.Collection interface)
func (m *MockCollection) GetIndexes() []string {
	return m.CreatedIndexes
}

// MockDatabase implements mongo.Database interface for testing
type MockDatabase struct {
	Collections map[string]*MockCollection
}

// Collection returns a mock collection
func (m *MockDatabase) Collection(name string) *mongo.Collection {
	if _, exists := m.Collections[name]; !exists {
		m.Collections[name] = &MockCollection{
			Documents: make(map[string]interface{}),
		}
	}
	return nil // Return nil since we're mocking at a higher level
}

// MockClient implements a mock MongoDB client for testing
type MockClient struct {
	Databases          map[string]*MockDatabase
	PingError          error
	ConnectError       error
	DisconnectError    error
	Collections        map[string]map[string]*MockCollection // namespace -> collection -> data
	FindOneFunc        func(ctx context.Context, namespace, collection, key string) (interface{}, error)
	UpdateOneFunc      func(ctx context.Context, namespace, collection, key string, value interface{}) error
	DeleteOneFunc      func(ctx context.Context, namespace, collection, key string) error
	CountDocumentsFunc func(ctx context.Context, namespace, collection, filter map[string]interface{}) (int64, error)
}

// NewMockClient creates a new mock MongoDB client
func NewMockClient() *MockClient {
	return &MockClient{
		Databases:   make(map[string]*MockDatabase),
		Collections: make(map[string]map[string]*MockCollection),
	}
}

// Database returns a mock database
func (m *MockClient) Database(name string) *mongo.Database {
	if _, exists := m.Databases[name]; !exists {
		m.Databases[name] = &MockDatabase{
			Collections: make(map[string]*MockCollection),
		}
	}
	return nil // Return nil since we're mocking at a higher level
}

// Ping mocks mongo.Client.Ping
func (m *MockClient) Ping(ctx context.Context) error {
	return m.PingError
}

// Connect mocks mongo.Client.Connect
func (m *MockClient) Connect(ctx context.Context) error {
	return m.ConnectError
}

// Disconnect mocks mongo.Client.Disconnect
func (m *MockClient) Disconnect(ctx context.Context) error {
	return m.DisconnectError
}

// ===== Mock Data Management Methods =====

// SetupDevice adds a device to the mock database
func (m *MockClient) SetupDevice(namespace string, device *models.Device) {
	if _, exists := m.Collections[namespace]; !exists {
		m.Collections[namespace] = make(map[string]*MockCollection)
	}
	if _, exists := m.Collections[namespace]["devices"]; !exists {
		m.Collections[namespace]["devices"] = &MockCollection{
			Documents: make(map[string]interface{}),
		}
	}
	m.Collections[namespace]["devices"].Documents[device.SN] = device
}

// SetupCard adds a card to the mock database
func (m *MockClient) SetupCard(namespace string, card *models.Card) {
	if _, exists := m.Collections[namespace]; !exists {
		m.Collections[namespace] = make(map[string]*MockCollection)
	}
	if _, exists := m.Collections[namespace]["cards"]; !exists {
		m.Collections[namespace]["cards"] = &MockCollection{
			Documents: make(map[string]interface{}),
		}
	}
	m.Collections[namespace]["cards"].Documents[card.Number] = card
}

// GetDevice retrieves a device from the mock database
func (m *MockClient) GetDevice(namespace string, sn string) (*models.Device, error) {
	if _, exists := m.Collections[namespace]; !exists {
		return nil, mongo.ErrNoDocuments
	}
	if _, exists := m.Collections[namespace]["devices"]; !exists {
		return nil, mongo.ErrNoDocuments
	}
	if doc, exists := m.Collections[namespace]["devices"].Documents[sn]; exists {
		if device, ok := doc.(*models.Device); ok {
			return device, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

// GetCard retrieves a card from the mock database
func (m *MockClient) GetCard(namespace string, cardNumber string) (*models.Card, error) {
	if _, exists := m.Collections[namespace]; !exists {
		return nil, mongo.ErrNoDocuments
	}
	if _, exists := m.Collections[namespace]["cards"]; !exists {
		return nil, mongo.ErrNoDocuments
	}
	if doc, exists := m.Collections[namespace]["cards"].Documents[cardNumber]; exists {
		if card, ok := doc.(*models.Card); ok {
			return card, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

// SetError sets the error for FindOne operations
func (m *MockClient) SetError(err error) {
	for _, namespace := range m.Collections {
		for _, collection := range namespace {
			collection.FindError = err
		}
	}
}

// ===== Test Helper Methods =====

// GetAllDevices returns all devices in a namespace
func (m *MockClient) GetAllDevices(namespace string) []*models.Device {
	var devices []*models.Device
	if _, exists := m.Collections[namespace]; !exists {
		return devices
	}
	if collection, exists := m.Collections[namespace]["devices"]; exists {
		for _, doc := range collection.Documents {
			if device, ok := doc.(*models.Device); ok {
				devices = append(devices, device)
			}
		}
	}
	return devices
}

// GetAllCards returns all cards in a namespace
func (m *MockClient) GetAllCards(namespace string) []*models.Card {
	var cards []*models.Card
	if _, exists := m.Collections[namespace]; !exists {
		return cards
	}
	if collection, exists := m.Collections[namespace]["cards"]; exists {
		for _, doc := range collection.Documents {
			if card, ok := doc.(*models.Card); ok {
				cards = append(cards, card)
			}
		}
	}
	return cards
}

// Clear clears all data from the mock client
func (m *MockClient) Clear() {
	m.Collections = make(map[string]map[string]*MockCollection)
	m.Databases = make(map[string]*MockDatabase)
}

// ===== Document Finder Helper =====

// DocumentFinder helps simulate mongo.SingleResult.Decode behavior
type DocumentFinder struct {
	Document interface{}
	Error    error
}

// Decode decodes the document (simulates mongo.SingleResult.Decode)
func (d *DocumentFinder) Decode(v interface{}) error {
	if d.Error != nil {
		return d.Error
	}
	if d.Document == nil {
		return mongo.ErrNoDocuments
	}
	// Simple type assertion (in real scenario, would use BSON marshaling)
	switch target := v.(type) {
	case *models.Device:
		if doc, ok := d.Document.(*models.Device); ok {
			*target = *doc
			return nil
		}
	case *models.Card:
		if doc, ok := d.Document.(*models.Card); ok {
			*target = *doc
			return nil
		}
	}
	return errors.New("type mismatch")
}

// ===== Mock Query Filter Helper =====

// ExtractKeyFromFilter extracts the key value from a BSON filter
func ExtractKeyFromFilter(filter interface{}) string {
	if filterMap, ok := filter.(bson.M); ok {
		if keyVal, exists := filterMap["key"]; exists {
			return keyVal.(string)
		}
		if snVal, exists := filterMap["sn"]; exists {
			return snVal.(string)
		}
		if numberVal, exists := filterMap["number"]; exists {
			return numberVal.(string)
		}
	}
	return ""
}
