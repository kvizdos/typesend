package consume_messages_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kvizdos/typesend/internal/consume_messages"
	providers_testing "github.com/kvizdos/typesend/internal/providers/tester"
	"github.com/kvizdos/typesend/pkg/testutils"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/stretchr/testify/assert"
)

func TestDeliverMessageSuccess(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Error(err)
		return
	}

	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC())

	err := testDb.Insert(e)
	assert.NoError(t, err)
	assert.Len(t, testDb.Items(), 1, "Expected 1 item in the Database")

	err = testDb.InsertTemplate(nil, &typesend_schemas.TypeSendTemplate{
		TemplateID:  e.TemplateID,
		TenantID:    e.TenantID,
		Content:     "Hello world",
		Subject:     "Blahaj",
		FromAddress: "example@demo.com",
		FromName:    "Kenton Vizdos",
	})

	provider := providers_testing.NewTestingProvider()
	logger := &testutils.TestLogger{}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: provider,
	}, e)

	assert.NoError(t, err, "No error expected when delivering message")

	receivedEnvelope, err := testDb.GetEnvelopeByID(nil, e.ID)
	assert.NoError(t, err)
	assert.Equal(t, typesend_schemas.TypeSendStatus_SENT, receivedEnvelope.Status, "Status mismatch")

	sentMsg := provider.GetMessageByEnvelopeID(e.ID)
	assert.NotNil(t, sentMsg, "sent message should not be nil")

	assert.Equal(t, "Hello world", sentMsg.Content)
	assert.Equal(t, "Blahaj", sentMsg.Subject)
}

// TestDeliverMessageNotReady verifies that if the envelope's ScheduledFor is too far in the future,
// DeliverMessage returns an error indicating the message is not ready.
func TestDeliverMessageNotReady(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}

	// Create an envelope scheduled 1 minute in the future.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC().Add(1*time.Minute))
	// Insert into database.
	err := testDb.Insert(e)
	assert.NoError(t, err)

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not ready to be delivered")
}

// TestDeliverMessageEnvelopeNotFound simulates a case where GetEnvelopeByID returns nil.
func TestDeliverMessageEnvelopeNotFound(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}
	// Create an envelope but do NOT insert it into the DB so that GetEnvelopeByID returns nil.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC().Add(-1*time.Minute))

	logger := &testutils.TestLogger{Test: t}

	err := consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	// No error is returned (do not retry), but an error is logged.
	assert.NoError(t, err)
	assert.Greater(t, len(logger.ErrorLogs), 0, "Expected error log for missing envelope")
}

// TestDeliverMessageDuplicateSend verifies that if the envelope status is already SENT,
// the message is not re-sent.
func TestDeliverMessageDuplicateSend(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}
	// Create envelope with status SENT.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_SENT, time.Now().UTC().Add(-1*time.Minute))
	err := testDb.Insert(e)
	assert.NoError(t, err)

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	assert.NoError(t, err)
	// Check that a warning was logged regarding duplicate send.
	assert.Greater(t, len(logger.WarnLogs), 0, "Expected warn log for duplicate send")
}

// TestDeliverMessageIncorrectStatus checks that if the envelope status is not DELIVERING or FAILED,
// the function returns early.
func TestDeliverMessageIncorrectStatus(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}
	// Use a status that is not DELIVERING or FAILED.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, time.Now().UTC().Add(-1*time.Minute))
	err := testDb.Insert(e)
	assert.NoError(t, err)

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	assert.NoError(t, err)
	// Expect a warning log that the envelope is not in the proper state.
	assert.Greater(t, len(logger.WarnLogs), 0, "Expected warn log for incorrect envelope status")
}

// TestDeliverMessageTemplateNotFound verifies that if the associated template is not found,
// DeliverMessage returns an error.
func TestDeliverMessageTemplateNotFound(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}
	// Create envelope and insert it.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC().Add(-1*time.Minute))
	err := testDb.Insert(e)
	assert.NoError(t, err)
	// Do NOT insert a template so that GetTemplateByID returns nil.

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find associated template ID")
}

// TestDeliverMessageTemplateFillFailure simulates a failure during template.Fill.
func TestDeliverMessageTemplateFillFailure(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}
	// Create and insert an envelope.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC().Add(-1*time.Minute))
	err := testDb.Insert(e)
	assert.NoError(t, err)

	// Insert a template that will fail when Fill is called.
	// For simulation, assume that an empty Content triggers a Fill error.
	err = testDb.InsertTemplate(nil, &typesend_schemas.TypeSendTemplate{
		TemplateID:  e.TemplateID,
		TenantID:    e.TenantID,
		Content:     "<p hello</p>", // trigger fill error
		Subject:     "Subject",
		FromAddress: "noreply@example.com",
		FromName:    "Test Sender",
	})
	assert.NoError(t, err)

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "html/template:")
}

type testFailingDb struct {
	typesend_db.TestDatabase
}

func (t *testFailingDb) UpdateEnvelopeStatus(ctx context.Context, envelopeID string, toStatus typesend_schemas.TypeSendStatus) error {
	return fmt.Errorf("example failure")
}

// TestDeliverMessageDatabaseUpdateFailure simulates a failure when updating the envelope status.
func TestDeliverMessageDatabaseUpdateFailure(t *testing.T) {
	// Here we assume that TestDatabase can be made to simulate an update error.
	testDb := &testFailingDb{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}

	// Create and insert envelope.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC().Add(-1*time.Minute))
	err := testDb.Insert(e)
	assert.NoError(t, err)

	// Insert a valid template.
	err = testDb.InsertTemplate(nil, &typesend_schemas.TypeSendTemplate{
		TemplateID:  e.TemplateID,
		TenantID:    e.TenantID,
		Content:     "Hello world",
		Subject:     "Subject",
		FromAddress: "noreply@example.com",
		FromName:    "Test Sender",
	})
	assert.NoError(t, err)

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: providers_testing.NewTestingProvider(),
	}, e)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update envelope status")
}

// TestDeliverMessageProviderError simulates an error from the provider.Deliver call.
func TestDeliverMessageProviderError(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	if err := testDb.Connect(nil); err != nil {
		t.Fatal(err)
	}

	// Create and insert envelope.
	e := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC().Add(-1*time.Minute))
	err := testDb.Insert(e)
	assert.NoError(t, err)

	// Insert a valid template.
	err = testDb.InsertTemplate(nil, &typesend_schemas.TypeSendTemplate{
		TemplateID:  e.TemplateID,
		TenantID:    e.TenantID,
		Content:     "Hello world",
		Subject:     "Subject",
		FromAddress: "noreply@example.com",
		FromName:    "Test Sender",
	})
	assert.NoError(t, err)

	// Create a provider that simulates an error.
	provider := providers_testing.NewTestingProvider()
	provider.SendError = errors.New("simulated provider error")

	logger := &testutils.TestLogger{Test: t}

	err = consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
		Logger:   logger,
		Database: testDb,
		Provider: provider,
	}, e)

	// No error should be returned (so that the scheduler re-queues) but the status should be updated to FAILED.
	assert.NoError(t, err)

	// Verify that the envelope was updated to FAILED.
	updatedEnv, err := testDb.GetEnvelopeByID(context.Background(), e.ID)
	assert.NoError(t, err)
	assert.Equal(t, typesend_schemas.TypeSendStatus_FAILED, updatedEnv.Status, "Envelope status should be FAILED")

	// Verify that an error was logged.
	found := false
	for _, logMsg := range logger.ErrorLogs {
		if logMsg != nil && strings.Contains(*logMsg, "simulated provider error") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected error log for provider error")
}
