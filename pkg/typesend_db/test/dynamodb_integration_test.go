package typesend_db_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"github.com/kvizdos/typesend/pkg/template_variables"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/kvizdos/typesend/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, container, err := testutils.SetupDynamoDBLocalSession(context.Background())
	assert.NoError(t, err, "DynamoDB Setup Should Not Return Error")
	defer testutils.KillContainer(container)

	db, err := typesend_db.NewDynamoDB(context.Background(), &typesend_db.DynamoConfig{
		Region:      "us-west-2",
		TableName:   "test-typesend",
		ForceClient: client,
	})
	assert.NoError(t, err)

	err = db.Insert(&typesend_schemas.TypeSendEnvelope{
		ScheduledFor: time.Now().UTC(),
		AppID:        "test",
		ToAddress:    "test@example.com",
		ToInternalID: "123",
		Variables: testutils.DummyVariable{
			TypeSendVariable: template_variables.TypeSendVariable{
				AssociatedTemplateID: uuid.NewString(),
			},
		},
		TemplateID:     uuid.NewString(),
		ID:             uuid.NewString(),
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
		MessageGroupID: "demo-group-id",
	})

	assert.NoError(t, err, "No error expected on DynamoDB.Insert")
}

func TestIntegration_GetMessagesReadyToSend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	// Setup DynamoDB Local session using our testutils helper.
	client, container, err := testutils.SetupDynamoDBLocalSession(ctx)
	assert.NoError(t, err, "DynamoDB Setup should not return error")
	defer testutils.KillContainer(container)

	// Use ForceClient to pass in our local DynamoDB session.
	db, err := typesend_db.NewDynamoDB(ctx, &typesend_db.DynamoConfig{
		Region:      "us-west-2",
		TableName:   "test-typesend",
		ForceClient: client, // assuming your implementation allows overriding the client
	})
	assert.NoError(t, err, "NewDynamoDB should succeed")

	// Insert test envelopes.
	// We'll insert three envelopes:
	//  1. One envelope that should be returned (UNSENT, scheduledFor in the past).
	//  2. One envelope with scheduledFor in the future.
	//  3. One envelope with a different status.
	now := time.Now().UTC()
	envelope1 := &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "a@example.com",
		ToInternalID:   "1",
		MessageGroupID: "group-1",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   now.Add(-5 * time.Minute),
		Status:         typesend_schemas.TypeSendStatus_UNSENT, // expected to be returned
	}
	envelope2 := &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "b@example.com",
		ToInternalID:   "2",
		MessageGroupID: "group-2",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   now.Add(5 * time.Minute),
		Status:         typesend_schemas.TypeSendStatus_UNSENT, // scheduled in future, should not return
	}
	envelope3 := &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "c@example.com",
		ToInternalID:   "3",
		MessageGroupID: "group-3",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   now.Add(-10 * time.Minute),
		Status:         typesend_schemas.TypeSendStatus_SENT, // wrong status, should not return
	}

	// Use the Insert function for each envelope.
	err = db.Insert(envelope1)
	assert.NoError(t, err, "Insert envelope1 should succeed")
	err = db.Insert(envelope2)
	assert.NoError(t, err, "Insert envelope2 should succeed")
	err = db.Insert(envelope3)
	assert.NoError(t, err, "Insert envelope3 should succeed")

	// Give DynamoDB Local a moment to index the items.
	time.Sleep(2 * time.Second)

	// Test case 1: Basic retrieval
	// We query with a timestamp that is after envelope1's ScheduledFor.
	queryTime := now
	ch, err := db.GetMessagesReadyToSend(ctx, queryTime)
	assert.NoError(t, err, "GetMessagesReadyToSend should succeed")
	var results []*typesend_schemas.TypeSendEnvelope
	for env := range ch {
		results = append(results, env)
	}
	// We expect only envelope1 to be returned.
	assert.Len(t, results, 1, "Only one envelope should be returned")
	assert.Equal(t, envelope1.ID, results[0].ID)

	// Test case 2: Context cancellation.
	// Create a context that cancels quickly.
	cancelCtx, cancel := context.WithCancel(ctx)
	// Cancel the context immediately.
	cancel()
	ch, err = db.GetMessagesReadyToSend(cancelCtx, queryTime)
	assert.NoError(t, err, "GetMessagesReadyToSend should not error on canceled context")
	// The channel should close quickly without returning any items.
	count := 0
	for range ch {
		count++
	}
	assert.Equal(t, 0, count, "No envelopes should be returned when context is canceled")
}

func TestIntegration_UpdateEnvelopeStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	// Setup DynamoDB Local session using our testutils helper.
	client, container, err := testutils.SetupDynamoDBLocalSession(ctx)
	assert.NoError(t, err, "DynamoDB Setup should not return error")
	defer testutils.KillContainer(container)

	// Create our DynamoDB wrapper with ForceClient to use the local session.
	db, err := typesend_db.NewDynamoDB(ctx, &typesend_db.DynamoConfig{
		Region:      "us-west-2",
		TableName:   "test-typesend",
		ForceClient: client,
	})
	assert.NoError(t, err, "NewDynamoDB should succeed")

	// Insert an envelope with an initial status of UNSENT.
	envelope := &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "update@example.com",
		ToInternalID:   "upd-1",
		MessageGroupID: "group-update",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   time.Now().UTC(),
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
	}
	err = db.Insert(envelope)
	assert.NoError(t, err, "Insert envelope should succeed")

	// Allow a short delay for the item to be indexed.
	time.Sleep(1 * time.Second)

	// Update the envelope's status to SENT.
	err = db.UpdateEnvelopeStatus(ctx, envelope.ID, typesend_schemas.TypeSendStatus_SENT)
	assert.NoError(t, err, "UpdateEnvelopeStatus should succeed")

	// Retrieve the updated envelope directly using the DynamoDB client.
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String("test-typesend"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(envelope.ID)},
		},
	}

	result, err := client.GetItemWithContext(ctx, getInput)
	assert.NoError(t, err, "GetItem should succeed")
	assert.NotNil(t, result.Item, "Item should be found in DynamoDB")

	// Unmarshal the result into a TypeSendEnvelope struct.
	var updatedEnvelope typesend_schemas.TypeSendEnvelope
	err = dynamodbattribute.UnmarshalMap(result.Item, &updatedEnvelope)
	assert.NoError(t, err, "UnmarshalMap should succeed")

	// Verify that the status has been updated to SENT.
	assert.Equal(t, typesend_schemas.TypeSendStatus_SENT, updatedEnvelope.Status, "Envelope status should be updated to SENT")
}
