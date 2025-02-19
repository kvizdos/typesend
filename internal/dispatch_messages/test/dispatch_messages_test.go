package dispatch_messages_test

import (
	"context"
	"errors"
	"testing"
	"time"

	typequeue "github.com/kvizdos/typequeue/pkg/mocked"
	"github.com/kvizdos/typesend/internal/dispatch_messages"
	"github.com/kvizdos/typesend/pkg/testutils"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/stretchr/testify/assert"
)

// failingDatabase simulates a failure in GetMessagesReadyToSend.
type failingDatabase struct {
	typesend_db.TestDatabase
}

func (f *failingDatabase) Connect(ctx context.Context) error { return nil }
func (f *failingDatabase) GetMessagesReadyToSend(ctx context.Context, now time.Time) (chan *typesend_schemas.TypeSendEnvelope, error) {
	return nil, errors.New("get messages error")
}
func (f *failingDatabase) UpdateEnvelopeStatus(ctx context.Context, envelopeID string, status typesend_schemas.TypeSendStatus) error {
	return nil
}
func (f *failingDatabase) Insert(envelope *typesend_schemas.TypeSendEnvelope) error { return nil }

// errorDispatcher always returns an error when dispatch is attempted.
type errorDispatcher struct{}

func (ed *errorDispatcher) Dispatch(ctx context.Context, envelope *typesend_schemas.TypeSendEnvelope, queueName string, delay ...int64) (*string, error) {
	return nil, errors.New("dispatch error")
}

// updateFailingDatabase wraps a TestDatabase but returns an error when UpdateEnvelopeStatus is called.
type updateFailingDatabase struct {
	*typesend_db.TestDatabase
}

func (db *updateFailingDatabase) UpdateEnvelopeStatus(ctx context.Context, envelopeID string, status typesend_schemas.TypeSendStatus) error {
	return errors.New("update status error")
}

func TestDispatchMessagesSuccessful(t *testing.T) {
	// Set up a test database with one envelope that is ready to send.
	db := &typesend_db.TestDatabase{}
	db.Connect(context.Background())
	env := &typesend_schemas.TypeSendEnvelope{
		AppID:          "demo",
		ToAddress:      "example@gmail.com",
		ToInternalID:   "rand-uid",
		Variables:      testutils.DummyVariable{}.ToMap(),
		TemplateID:     "demo-template",
		ID:             "random-uuid",
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
		MessageGroupID: "ajdsofksdof",
		ReferenceID:    "okoekowkfowkco",
		ScheduledFor:   time.Now().UTC().Add(-10 * time.Second),
	}
	db.Insert(env)

	testDispatcher := &typequeue.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	ctx := context.WithValue(context.Background(), "trace-id", "demo-trace")
	err := dispatch_messages.DispatchMessagesReadyToSend(&dispatch_messages.DispatchOpts{
		Context:    ctx,
		Database:   db,
		Dispatcher: testDispatcher,
		Logger:     nil,
	})

	assert.NoError(t, err, "Did not expect an error back")
	assert.NotNil(t, testDispatcher.Messages["email_queue"], "email_queue should not be nil")
	assert.Len(t, testDispatcher.Messages["email_queue"], 1, "email_queue should have 1 item")
	assert.Equal(t, testDispatcher.Messages["email_queue"][0], env, "envelope does not match")

	assert.Equal(t, typesend_schemas.TypeSendStatus_DELIVERING, db.GetEnvelopeByID(env.ID).Status, "wrong status")
}

func TestDispatchMessagesGetMessagesError(t *testing.T) {
	// Use a failing database that returns an error on GetMessagesReadyToSend.
	failingDB := &failingDatabase{}
	testDispatcher := &typequeue.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}
	ctx := context.WithValue(context.Background(), "trace-id", "demo")

	err := dispatch_messages.DispatchMessagesReadyToSend(&dispatch_messages.DispatchOpts{
		Context:    ctx,
		Database:   failingDB,
		Dispatcher: testDispatcher,
		Logger:     nil,
	})

	assert.Error(t, err)
	assert.Equal(t, "get messages error", err.Error())
}

func TestDispatchMessagesDispatchError(t *testing.T) {
	// Set up a test database with one envelope.
	db := &typesend_db.TestDatabase{}
	db.Connect(context.Background())
	env := &typesend_schemas.TypeSendEnvelope{
		AppID:          "demo",
		ToAddress:      "error@example.com",
		ToInternalID:   "error-uid",
		Variables:      testutils.DummyVariable{}.ToMap(),
		TemplateID:     "demo-template",
		ID:             "error-uuid",
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
		MessageGroupID: "group-error",
		ReferenceID:    "ref-error",
		ScheduledFor:   time.Now().UTC().Add(-10 * time.Second),
	}
	db.Insert(env)

	// Use an errorDispatcher that always returns an error.
	errDispatcher := &errorDispatcher{}

	ctx := context.WithValue(context.Background(), "trace-id", "demo")

	// Capture logs if needed; here Logger is nil.
	err := dispatch_messages.DispatchMessagesReadyToSend(&dispatch_messages.DispatchOpts{
		Context:    ctx,
		Database:   db,
		Dispatcher: errDispatcher,
		Logger:     nil,
	})
	// The function should not return an error (it logs and continues).
	assert.NoError(t, err, "DispatchMessagesReadyToSend should not return an error even if dispatch fails")
	// Since dispatch failed, the envelope should not have been updated or queued.
}

func TestDispatchMessagesUpdateStatusError(t *testing.T) {
	// Set up a test database with one envelope.
	originalDB := &typesend_db.TestDatabase{}
	originalDB.Connect(context.Background())
	env := &typesend_schemas.TypeSendEnvelope{
		AppID:          "demo",
		ToAddress:      "updateerror@example.com",
		ToInternalID:   "updateerror-uid",
		Variables:      testutils.DummyVariable{}.ToMap(),
		TemplateID:     "demo-template",
		ID:             "updateerror-uuid",
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
		MessageGroupID: "group-update-error",
		ReferenceID:    "ref-update-error",
		ScheduledFor:   time.Now().UTC().Add(-10 * time.Second),
	}
	originalDB.Insert(env)

	// Wrap the TestDatabase so that UpdateEnvelopeStatus returns an error.
	db := &updateFailingDatabase{TestDatabase: originalDB}

	// Use a dispatcher that successfully dispatches.
	testDispatcher := &typequeue.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	ctx := context.WithValue(context.Background(), "trace-id", "demo")
	err := dispatch_messages.DispatchMessagesReadyToSend(&dispatch_messages.DispatchOpts{
		Context:    ctx,
		Database:   db,
		Dispatcher: testDispatcher,
		Logger:     nil,
	})
	// Function should complete without returning an error even if the update fails.
	assert.NoError(t, err, "DispatchMessagesReadyToSend should not return an error even if update status fails")
	// Verify that the envelope was dispatched even though status update failed.
	assert.NotNil(t, testDispatcher.Messages["email_queue"], "email_queue should not be nil")
	assert.Len(t, testDispatcher.Messages["email_queue"], 1, "email_queue should have 1 item")
}

func TestDispatchMessagesNoMessages(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	db.Connect(context.Background())

	// Use a dispatcher that successfully dispatches.
	testDispatcher := &typequeue.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	ctx := context.WithValue(context.Background(), "trace-id", "demo")
	err := dispatch_messages.DispatchMessagesReadyToSend(&dispatch_messages.DispatchOpts{
		Context:    ctx,
		Database:   db,
		Dispatcher: testDispatcher,
		Logger:     nil,
	})

	assert.NoError(t, err, "DispatchMessagesReadyToSend should not return an error even if update status fails")
	assert.Nil(t, testDispatcher.Messages["email_queue"], "email_queue should be nil")
}
