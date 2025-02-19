package dispatch_messages_handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	typequeue_mocks "github.com/kvizdos/typequeue/pkg/mocked"
	"github.com/kvizdos/typesend/cmd/dispatch_messages/dispatch_messages_handler"
	"github.com/kvizdos/typesend/internal/dispatch_messages"
	"github.com/kvizdos/typesend/internal/testutils"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/stretchr/testify/assert"
)

/*
You may notice that these tests don't test the dependency setups,
this is because they are all fully tested on their own.

--

Also not covered here is testing to confirm that DispatchMessagesReadyToSend
works properly, as its fully tested on its own.
*/

type stubbedDb struct {
	typesend_db.TestDatabase // fulfill requirements
	MessagesReadyToSend      []*typesend_schemas.TypeSendEnvelope
}

func (s *stubbedDb) UpdateEnvelopeStatus(ctx context.Context, envelopeID string, toStatus typesend_schemas.TypeSendStatus) error {
	return nil
}

func (s *stubbedDb) GetMessagesReadyToSend(ctx context.Context, timestamp time.Time) (chan *typesend_schemas.TypeSendEnvelope, error) {
	ch := make(chan *typesend_schemas.TypeSendEnvelope)

	go func() {
		defer close(ch)
		for _, env := range s.MessagesReadyToSend {
			ch <- env
		}
	}()

	return ch, nil
}

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

func TestHandlerSuccess(t *testing.T) {
	dispatcher := &typequeue_mocks.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	db := &stubbedDb{
		MessagesReadyToSend: []*typesend_schemas.TypeSendEnvelope{
			createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, time.Now().UTC()),
		},
	}
	testLogger := &testutils.TestLogger{}

	handler := &dispatch_messages_handler.DispatchMessagesLambda{
		AWSRegion: "us-east-1",
		Project:   "test",
		Env:       "testing",
		Deps: &dispatch_messages_handler.DispatchMessagesDependencies{
			Dispatcher:      dispatcher,
			DB:              db,
			ContextDeadline: time.Now().Add(30 * time.Second),
			Logger:          testLogger,
		},
	}

	err := handler.HandleRequest(context.Background())

	assert.NoError(t, err, "Handle Request should not have errored")

	assert.Nil(t, testLogger.ErrorLogs, "No errors should be logged.")
	assert.NotNil(t, testLogger.InfoLogs, "Expected info logs to not be nil.")
	assert.Len(t, testLogger.InfoLogs, 1, "Expected info logs to 1.")
	assert.Contains(t, *testLogger.InfoLogs[0], "sent 1 message", "Expected logs to show that 1 message was sent.")
}

func TestHandlerDeadlineExceeded(t *testing.T) {
	dispatcher := &typequeue_mocks.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	db := &stubbedDb{
		MessagesReadyToSend: []*typesend_schemas.TypeSendEnvelope{
			createTestEnvelope(typesend_schemas.TypeSendStatus_UNSENT, time.Now().UTC()),
		},
	}
	testLogger := &testutils.TestLogger{}
	handler := &dispatch_messages_handler.DispatchMessagesLambda{
		AWSRegion: "us-east-1",
		Project:   "test",
		Env:       "testing",
		Deps: &dispatch_messages_handler.DispatchMessagesDependencies{
			Dispatcher:      dispatcher,
			DB:              db,
			ContextDeadline: time.Now().Add(-30 * time.Second),
			Logger:          testLogger,
		},
	}

	err := handler.HandleRequest(context.Background())
	assert.NoError(t, err, "Handle Request should not have errored")

	assert.Nil(t, testLogger.ErrorLogs, "No errors should be logged.")
	assert.NotNil(t, testLogger.InfoLogs, "Info logs should be nil.")
	assert.Len(t, testLogger.InfoLogs, 1, "Expected an info log about deadline being exceeded")
	assert.Contains(t, *testLogger.InfoLogs[0], "deadline exceeded", "Expected an info log about deadline being exceeded")
}

func TestHandlerNoMessages(t *testing.T) {
	dispatcher := &typequeue_mocks.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	db := &stubbedDb{
		MessagesReadyToSend: []*typesend_schemas.TypeSendEnvelope{},
	}
	dbErr := db.Connect(context.Background())
	assert.NoError(t, dbErr, "connecting to testdb should not have failed!")

	testLogger := &testutils.TestLogger{}
	handler := &dispatch_messages_handler.DispatchMessagesLambda{
		AWSRegion: "us-east-1",
		Project:   "test",
		Env:       "testing",
		Deps: &dispatch_messages_handler.DispatchMessagesDependencies{
			Dispatcher:      dispatcher,
			DB:              db,
			ContextDeadline: time.Now().Add(-30 * time.Second),
			Logger:          testLogger,
		},
	}

	err := handler.HandleRequest(context.Background())
	assert.NoError(t, err, "Handle Request should not have errored")

	assert.Nil(t, testLogger.ErrorLogs, "No errors should be logged.")
	assert.Nil(t, testLogger.InfoLogs, "Info logs should be nil.")
}

func TestHandlerDispatchThrowsError(t *testing.T) {
	defer func() {
		dispatch_messages_handler.TestSetDispatchMessagesReadyToSendFn(dispatch_messages.DispatchMessagesReadyToSend)
	}()

	dispatch_messages_handler.TestSetDispatchMessagesReadyToSendFn(func(opts *dispatch_messages.DispatchOpts) error {
		return errors.New("stubbed error")
	})

	dispatcher := &typequeue_mocks.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		Messages: make(map[string][]*typesend_schemas.TypeSendEnvelope),
	}

	db := &stubbedDb{
		MessagesReadyToSend: []*typesend_schemas.TypeSendEnvelope{},
	}
	dbErr := db.Connect(context.Background())
	assert.NoError(t, dbErr, "connecting to testdb should not have failed!")

	testLogger := &testutils.TestLogger{}
	handler := &dispatch_messages_handler.DispatchMessagesLambda{
		AWSRegion: "us-east-1",
		Project:   "test",
		Env:       "testing",
		Deps: &dispatch_messages_handler.DispatchMessagesDependencies{
			Dispatcher:      dispatcher,
			DB:              db,
			ContextDeadline: time.Now().Add(-30 * time.Second),
			Logger:          testLogger,
		},
	}

	err := handler.HandleRequest(context.Background())
	assert.Contains(t, err.Error(), "stubbed error", "Unexpected error returned")

	assert.NotNil(t, testLogger.ErrorLogs, "An error should be logged.")
	assert.Len(t, testLogger.ErrorLogs, 1, "Only 1 error should be logged.")
	assert.Contains(t, *testLogger.ErrorLogs[0], "stubbed error", "Unexpected error logged.")
}
