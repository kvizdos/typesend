package consume_messages_handler

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	typequeue "github.com/kvizdos/typequeue/pkg"
	typequeue_lambda "github.com/kvizdos/typequeue/pkg/lambda"
	"github.com/kvizdos/typesend/internal/consume_messages"
	"github.com/kvizdos/typesend/internal/providers"
	"github.com/kvizdos/typesend/internal/sentry"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/sirupsen/logrus"
)

// ConsumeMessageHandlerDependencies holds all external dependencies.
type ConsumeMessageHandlerDependencies struct {
	Logger   typesend_schemas.Logger
	DB       typesend_db.TypeSendDatabase
	Provider providers.TypeSendProvider
}

// ConsumeMessageHandler contains the config and dependency references.
type ConsumeMessageHandler struct {
	// Configuration
	AWSRegion string
	Project   string
	Env       string

	// Dependencies are injected here. If nil, Setup will create them.
	Deps *ConsumeMessageHandlerDependencies
}

func (cmh *ConsumeMessageHandler) Setup() error {
	// Ensure dependency container exists.
	if cmh.Deps == nil {
		cmh.Deps = &ConsumeMessageHandlerDependencies{}
	}

	// Setup logger.
	if cmh.Deps.Logger == nil {
		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{})
		cmh.Deps.Logger = logger

		// Initialize Sentry.
		sentry.InitializeSentry(cmh.Deps.Logger.(*logrus.Logger), "typesend_message_dispatch")
	}

	// Connect to DynamoDB.
	if cmh.Deps.DB == nil {
		dynamoCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dynamo, err := typesend_db.NewDynamoDB(dynamoCtx, &typesend_db.DynamoConfig{
			Region:         cmh.AWSRegion,
			EnvelopesTable: fmt.Sprintf("%s_typesend_envelopes", cmh.Project),
			TemplatesTable: fmt.Sprintf("%s_typesend_templates", cmh.Project),
			ForceClient:    &dynamodb.DynamoDB{},
		})
		if err != nil {
			cmh.Deps.Logger.Errorf("failed to connect to DynamoDB: %s", err.Error())
			return fmt.Errorf("failed to connect to DynamoDB: %w", err)
		}
		cmh.Deps.DB = dynamo
	}

	return nil
}

func (cmh *ConsumeMessageHandler) Handle(ctx context.Context, sqsEvent events.SQSEvent) (map[string]interface{}, error) {
	consumer := typequeue_lambda.LambdaConsumer[*typesend_schemas.TypeSendEnvelope]{
		Logger:    cmh.Deps.Logger,
		SQSEvents: sqsEvent,
	}
	consumer.Consume(context.Background(), typequeue.ConsumerSQSOptions{}, func(envelope *typesend_schemas.TypeSendEnvelope) error {
		l := cmh.Deps.Logger.(*logrus.Logger)
		logger := l.WithFields(logrus.Fields{
			"trace-id":    *envelope.GetTraceID(),
			"envelope-id": envelope.ID,
		})
		err := consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
			Logger:   logger,
			Database: cmh.Deps.DB,
			Provider: cmh.Deps.Provider,
		}, envelope)

		if err != nil {
			logger.Errorf("Failed to deliver message: %s", err.Error())
			return err
		}

		return nil
	})
	return consumer.GetBatchItemFailures(), nil
}
