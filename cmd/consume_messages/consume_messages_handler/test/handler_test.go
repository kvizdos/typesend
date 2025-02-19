package consume_messages_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/kvizdos/typesend/cmd/consume_messages/consume_messages_handler"
	"github.com/kvizdos/typesend/pkg/testutils"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/stretchr/testify/assert"
)

func createTestEvent() (typesend_schemas.TypeSendEnvelope, events.SQSMessage) {
	testEnvelope := testutils.CreateTestEnvelope(typesend_schemas.TypeSendStatus_DELIVERING, time.Now().UTC())

	js, err := json.Marshal(testEnvelope)
	if err != nil {
		panic(err)
	}

	return *testEnvelope, events.SQSMessage{
		MessageId:     fmt.Sprintf("mid-%s", uuid.NewString()),
		ReceiptHandle: fmt.Sprintf("rid-%s", uuid.NewString()),
		Body:          string(js),
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"X-Trace-ID": {
				StringValue: aws.String("rand-id"),
				DataType:    "STRING",
			},
		},
	}
}

func TestConsumerSuccess(t *testing.T) {
	testDb := &typesend_db.TestDatabase{}
	err := testDb.Connect(context.Background())
	assert.NoError(t, err)

	testLogger := &testutils.TestLogger{
		Test:  t,
		DoLog: true,
	}

	handler := &consume_messages_handler.ConsumeMessageHandler{
		AWSRegion: "us-east-1",
		Project:   "test",
		Env:       "testing",
		Deps: &consume_messages_handler.ConsumeMessageHandlerDependencies{
			DB:     testDb,
			Logger: testLogger,
		},
	}

	_, tev1 := createTestEvent()
	failedMsgs, err := handler.Handle(context.Background(), events.SQSEvent{
		Records: []events.SQSMessage{
			tev1,
		},
	})

	assert.NoError(t, err)
	assert.Len(t, failedMsgs["batchItemFailures"], 0)
}
