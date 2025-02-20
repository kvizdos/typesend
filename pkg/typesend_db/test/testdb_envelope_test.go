package typesend_db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/stretchr/testify/assert"
)

// helper to create a test envelope
func createTestEnvelope(status typesend_schemas.TypeSendStatus, scheduledFor time.Time) *typesend_schemas.TypeSendEnvelope {
	return &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "test@example.com",
		ToInternalID:   "internal",
		MessageGroupID: "group",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   scheduledFor,
		Status:         status,
	}
}

func TestTestDatabase_Connect(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	err := db.Connect(context.Background())
	assert.NoError(t, err, "Connect should not return an error")

	// After connecting, Items() should return an empty slice.
	items := db.Items()
	assert.Empty(t, items, "Expected no items after connecting")
}

func TestTestDatabase_Insert(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	envelope := createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, time.Now().UTC())
	err := db.Insert(envelope)
	assert.NoError(t, err, "Insert should not return an error")

	items := db.Items()
	assert.Len(t, items, 1, "Expected one item after insertion")
	assert.Equal(t, envelope, items[0], "Inserted envelope should match the one in the database")
}

func TestTestDatabase_GetMessagesReadyToSend(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	now := time.Now().UTC()

	// envelope1: should be returned (UNSENT & scheduled in the past)
	envelope1 := createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, now.Add(-10*time.Minute))
	// envelope2: scheduled in the future; should not be returned
	envelope2 := createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, now.Add(10*time.Minute))
	// envelope3: wrong status; should not be returned
	envelope3 := createTestEnvelope(typesend_schemas.TypeSendStatus_SENT, now.Add(-10*time.Minute))

	_ = db.Insert(envelope1)
	_ = db.Insert(envelope2)
	_ = db.Insert(envelope3)

	ch, err := db.GetMessagesReadyToSend(context.Background(), now)
	assert.NoError(t, err, "GetMessagesReadyToSend should not return an error")

	var results []*typesend_schemas.TypeSendEnvelope
	for env := range ch {
		results = append(results, env)
	}

	// Only envelope1 meets the criteria.
	assert.Len(t, results, 1, "Expected only one envelope ready to send")
	assert.Equal(t, envelope1.ID, results[0].ID, "Returned envelope should be envelope1")
}

func TestTestDatabase_GetMessagesReadyToSend_ContextCancelled(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	now := time.Now().UTC()
	envelope := createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, now.Add(-10*time.Minute))
	_ = db.Insert(envelope)

	// Create a context that is cancelled immediately.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch, err := db.GetMessagesReadyToSend(ctx, now)
	assert.NoError(t, err, "GetMessagesReadyToSend should not error even if context is cancelled")

	// Expect no results since the context is cancelled.
	count := 0
	for range ch {
		count++
	}
	assert.Equal(t, 0, count, "No envelopes should be returned when the context is cancelled")
}

func TestTestDatabase_UpdateEnvelopeStatus(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	envelope := createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, time.Now().UTC())
	_ = db.Insert(envelope)

	// Update the envelope's status to SENT.
	err := db.UpdateEnvelopeStatus(context.Background(), envelope.ID, typesend_schemas.TypeSendStatus_SENT)
	assert.NoError(t, err, "UpdateEnvelopeStatus should succeed")

	// Verify that the envelope's status was updated.
	items := db.Items()
	assert.Len(t, items, 1, "Expected one envelope in the database")
	assert.Equal(t, typesend_schemas.TypeSendStatus_SENT, items[0].Status, "Envelope status should be updated to SENT")
}

func TestTestDatabase_UpdateEnvelopeStatus_NotFound(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	// Attempt to update an envelope that doesn't exist.
	err := db.UpdateEnvelopeStatus(context.Background(), "non-existent-id", typesend_schemas.TypeSendStatus_SENT)
	assert.Error(t, err, "Expected an error when updating a non-existent envelope")
	assert.Contains(t, err.Error(), "not found", "Error message should indicate envelope not found")
}

// TestTestDatabase_GetEnvelopeByID verifies that an envelope that has been inserted is returned correctly.
func TestTestDatabase_GetEnvelopeByID(t *testing.T) {
	// Create and connect the test database.
	db := &typesend_db.TestDatabase{}
	err := db.Connect(context.Background())
	assert.NoError(t, err, "Connect should not return an error")

	// Create a test envelope.
	envelope := &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "test@example.com",
		ToInternalID:   "internal",
		MessageGroupID: "group",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   time.Now().UTC(),
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
	}

	// Insert the envelope.
	err = db.Insert(envelope)
	assert.NoError(t, err, "Insert should not return an error")

	// Retrieve the envelope by its ID.
	gotEnvelope, err := db.GetEnvelopeByID(context.Background(), envelope.ID)
	assert.NoError(t, err, "GetEnvelopeByID should succeed")
	assert.NotNil(t, gotEnvelope, "Envelope should be found")

	// Verify that the envelope fields match.
	assert.Equal(t, envelope.ID, gotEnvelope.ID, "Envelope ID should match")
	assert.Equal(t, envelope.AppID, gotEnvelope.AppID, "AppID should match")
	assert.Equal(t, envelope.ToAddress, gotEnvelope.ToAddress, "ToAddress should match")
	assert.Equal(t, envelope.Status, gotEnvelope.Status, "Status should match")
}

// TestTestDatabase_GetEnvelopeByIDNotFound verifies that a lookup for a non-existent envelope returns nil.
func TestTestDatabase_GetEnvelopeByIDNotFound(t *testing.T) {
	// Create and connect the test database.
	db := &typesend_db.TestDatabase{}
	err := db.Connect(context.Background())
	assert.NoError(t, err, "Connect should not return an error")

	// Attempt to get an envelope with an ID that does not exist.
	gotEnvelope, err := db.GetEnvelopeByID(context.Background(), "non-existent-id")
	assert.NoError(t, err, "GetEnvelopeByID should not return an error")
	assert.Nil(t, gotEnvelope, "Expected nil when envelope is not found")
}
