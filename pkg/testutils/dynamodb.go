package testutils

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupDynamoDBLocalSession spins up a DynamoDB Local container and creates an AWS session
// that connects to the container. It returns the session, the container (for cleanup), and any error.
func SetupDynamoDBLocalSession(ctx context.Context) (*dynamodb.DynamoDB, testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local", // Official DynamoDB Local image.
		ExposedPorts: []string{"8000/tcp"},
		WaitingFor:   wait.ForListeningPort("8000"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start DynamoDB Local container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, container, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "8000")
	if err != nil {
		return nil, container, fmt.Errorf("failed to get container port: %w", err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Endpoint:    aws.String(endpoint), // Connect to DynamoDB Local.
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
	})
	if err != nil {
		return nil, container, fmt.Errorf("failed to create AWS session: %w", err)
	}

	dynamoClient := dynamodb.New(sess)

	var setupWg sync.WaitGroup
	setupWg.Add(2)

	go func() {
		defer setupWg.Done()
		// (Optional) Create the table if needed.
		_, err = dynamoClient.CreateTable(&dynamodb.CreateTableInput{
			TableName: aws.String("test-typesend-envelopes"),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: aws.String("S"),
				},
				{
					AttributeName: aws.String("status"),
					AttributeType: aws.String("N"),
				},
				{
					AttributeName: aws.String("scheduledFor"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       aws.String("HASH"),
				},
			},
			BillingMode: aws.String("PAY_PER_REQUEST"),
			GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
				{
					IndexName: aws.String("status-scheduledFor-index"),
					KeySchema: []*dynamodb.KeySchemaElement{
						{
							AttributeName: aws.String("status"),
							KeyType:       aws.String("HASH"),
						},
						{
							AttributeName: aws.String("scheduledFor"),
							KeyType:       aws.String("RANGE"),
						},
					},
					Projection: &dynamodb.Projection{
						ProjectionType: aws.String("ALL"),
					},
				},
			},
		})
		if err != nil {
			panic(fmt.Errorf("failed to create table: %w", err))
		}
	}()

	go func() {
		defer setupWg.Done()
		_, err = dynamoClient.CreateTable(&dynamodb.CreateTableInput{
			TableName: aws.String("test-typesend-templates"),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("tenant"),
					AttributeType: aws.String("S"),
				},
				{
					AttributeName: aws.String("id"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("tenant"),
					KeyType:       aws.String("HASH"), // Partition key
				},
				{
					AttributeName: aws.String("id"),
					KeyType:       aws.String("RANGE"), // Sort key
				},
			},
			BillingMode: aws.String("PAY_PER_REQUEST"),
		})
		if err != nil {
			panic(fmt.Errorf("failed to create table: %w", err))
		}
	}()

	setupWg.Wait()

	var readyWg sync.WaitGroup
	readyWg.Add(2)
	go func() {
		defer readyWg.Done()
		// Wait until the table exists.
		err = dynamoClient.WaitUntilTableExists(&dynamodb.DescribeTableInput{
			TableName: aws.String("test-typesend-envelopes"),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to wait for table creation: %s", err.Error()))
		}
	}()
	go func() {
		defer readyWg.Done()
		err = dynamoClient.WaitUntilTableExists(&dynamodb.DescribeTableInput{
			TableName: aws.String("test-typesend-templates"),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to wait for table creation: %s", err.Error()))
		}
	}()
	readyWg.Wait()
	return dynamoClient, container, nil
}
