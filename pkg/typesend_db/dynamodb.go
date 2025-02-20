package typesend_db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

func NewDynamoDB(ctx context.Context, conf *DynamoConfig) (*DynamoTypeSendDB, error) {
	db := &DynamoTypeSendDB{
		Config: conf,
	}

	if conf.ForceClient != nil {
		db.client = conf.ForceClient
		return db, nil
	}

	err := db.Connect(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}

type DynamoConfig struct {
	Region         string
	EnvelopesTable string
	TemplatesTable string

	ForceClient *dynamodb.DynamoDB
}

type DynamoTypeSendDB struct {
	Config *DynamoConfig
	logger typesend_schemas.Logger
	client *dynamodb.DynamoDB
}

func (db *DynamoTypeSendDB) Connect(ctx context.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(db.Config.Region),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	db.client = dynamodb.New(sess)

	return nil
}

func (db *DynamoTypeSendDB) Insert(envelope *typesend_schemas.TypeSendEnvelope) error {
	if db.client == nil {
		return fmt.Errorf("typesend: Insert requires a connection")
	}
	// Marshal the envelope struct into a DynamoDB attribute map.
	item, err := dynamodbattribute.MarshalMap(envelope)
	if err != nil {
		return fmt.Errorf("typesend: failed to marshal envelope: %w", err)
	}

	// Build the PutItem input.
	input := &dynamodb.PutItemInput{
		TableName: aws.String(db.Config.EnvelopesTable),
		Item:      item,
	}

	// Put the item into DynamoDB.
	_, err = db.client.PutItem(input)
	if err != nil {
		return fmt.Errorf("typesend: failed to put item: %w", err)
	}

	return nil
}

func (db *DynamoTypeSendDB) GetEnvelopeByID(ctx context.Context, envelopeID string) (*typesend_schemas.TypeSendEnvelope, error) {
	if db.client == nil {
		return nil, fmt.Errorf("typesend: GetEnvelopeByID requires a connection")
	}

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(envelopeID)},
		},
		TableName: &db.Config.EnvelopesTable,
	}

	rawItem, err := db.client.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("typesend: failed to get template: %w", err)
	}

	if rawItem.Item == nil {
		return nil, nil
	}

	var envelope *typesend_schemas.TypeSendEnvelope
	if err := dynamodbattribute.UnmarshalMap(rawItem.Item, &envelope); err != nil {
		return nil, fmt.Errorf("typesend: failed to unmarshal template: %w", err)
	}
	return envelope, nil
}

func (db *DynamoTypeSendDB) GetMessagesReadyToSend(ctx context.Context, timestamp time.Time) (chan *typesend_schemas.TypeSendEnvelope, error) {
	if db.client == nil {
		return nil, fmt.Errorf("typesend: GetMessagesReadyToSend requires a connection")
	}
	ch := make(chan *typesend_schemas.TypeSendEnvelope)

	// Format the timestamp to match how it was stored.
	tsStr := timestamp.Format(time.RFC3339)

	// Build the query input for the "status-scheduledFor-index" index.
	// Here we query for items where status = 0 (i.e. UNSENT)
	// and scheduledFor is less than or equal to our timestamp.
	input := &dynamodb.QueryInput{
		TableName:              aws.String(db.Config.EnvelopesTable),
		IndexName:              aws.String("status-scheduledFor-index"),
		KeyConditionExpression: aws.String("#status = :unsent and scheduledFor <= :ts"),
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":unsent": {N: aws.String("0")}, // assuming UNSENT is represented as 0
			":ts":     {S: aws.String(tsStr)},
		},
	}

	go func() {
		defer close(ch)
		// QueryPagesWithContext iterates over the results page by page.
		err := db.client.QueryPagesWithContext(ctx, input, func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				// Always check if the context has been canceled.
				select {
				case <-ctx.Done():
					return false
				default:
				}

				var envelope typesend_schemas.TypeSendEnvelope
				if err := dynamodbattribute.UnmarshalMap(item, &envelope); err != nil {
					// Log the error and skip this item if unmarshaling fails.
					// In production code, consider using a proper logging library.
					if db.logger != nil {
						db.logger.Errorf("typesend: failed to unmarshal item: %v", err)
					} else {
						log.Printf("typesend: failed to unmarshal item: %v", err)
					}
					continue
				}

				// Attempt to send the envelope to the channel.
				select {
				case <-ctx.Done():
					return false
				case ch <- &envelope:
				}
			}
			return !lastPage
		})
		if err != nil {
			if db.logger != nil {
				db.logger.Errorf("typesend: error during query: %v", err)
			} else {
				log.Printf("typesend: error during query: %v", err)
			}
		}
	}()

	return ch, nil
}

func (db *DynamoTypeSendDB) UpdateEnvelopeStatus(ctx context.Context, envelopeID string, toStatus typesend_schemas.TypeSendStatus) error {
	if db.client == nil {
		return fmt.Errorf("typesend: UpdateEnvelopeStatus requires a connection")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.Config.EnvelopesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(envelopeID)},
		},
		UpdateExpression: aws.String("SET #status = :newStatus"),
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":newStatus": {N: aws.String(fmt.Sprintf("%d", toStatus))},
		},
	}

	_, err := db.client.UpdateItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("typesend: failed to update envelope status: %w", err)
	}
	return nil
}

func (db *DynamoTypeSendDB) GetTemplateByID(ctx context.Context, templateID string, tenantID string) (*typesend_schemas.TypeSendTemplate, error) {
	if db.client == nil {
		return nil, fmt.Errorf("typesend: GetTemplateByID requires a connection")
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(db.Config.TemplatesTable),
		KeyConditionExpression: aws.String("#id = :id AND #tenant = :tenant"),
		ExpressionAttributeNames: map[string]*string{
			"#id":     aws.String("id"),
			"#tenant": aws.String("tenant"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id":     {S: aws.String(templateID)},
			":tenant": {S: aws.String(tenantID)},
		},
	}
	result, err := db.client.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("typesend: failed to get template: %w", err)
	}

	if len(result.Items) == 0 {
		if tenantID != "base" {
			return db.GetTemplateByID(ctx, templateID, "base")
		}
		return nil, nil
	}

	var template *typesend_schemas.TypeSendTemplate
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &template); err != nil {
		return nil, fmt.Errorf("typesend: failed to unmarshal template: %w", err)
	}
	return template, nil
}

func (db *DynamoTypeSendDB) InsertTemplate(ctx context.Context, template *typesend_schemas.TypeSendTemplate) error {
	if db.client == nil {
		return fmt.Errorf("typesend: InsertTemplate requires a connection")
	}
	// Marshal the envelope struct into a DynamoDB attribute map.
	item, err := dynamodbattribute.MarshalMap(template)
	if err != nil {
		return fmt.Errorf("typesend: failed to marshal envelope: %w", err)
	}

	// Build the PutItem input.
	input := &dynamodb.PutItemInput{
		TableName: aws.String(db.Config.TemplatesTable),
		Item:      item,
	}

	// Put the item into DynamoDB.
	_, err = db.client.PutItemWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("typesend: failed to put item: %w", err)
	}

	return nil
}
