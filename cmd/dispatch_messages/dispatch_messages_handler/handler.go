package dispatch_messages_handler

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	typequeue "github.com/kvizdos/typequeue/pkg"
	"github.com/kvizdos/typequeue/pkg/typequeue_helpers"
	"github.com/kvizdos/typesend/internal"
	"github.com/kvizdos/typesend/internal/dispatch_messages"
	"github.com/kvizdos/typesend/internal/sentry"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/sirupsen/logrus"
)

// Used purely for stubbing
var dispatchMessagesReadyToSendFn = dispatch_messages.DispatchMessagesReadyToSend

func TestSetDispatchMessagesReadyToSendFn(fn func(*dispatch_messages.DispatchOpts) error) {
	dispatchMessagesReadyToSendFn = fn
}

// DispatchMessagesDependencies holds all external dependencies.
type DispatchMessagesDependencies struct {
	Logger          typesend_schemas.Logger
	Dispatcher      typequeue.TypeQueueDispatcher[*typesend_schemas.TypeSendEnvelope]
	DB              typesend_db.TypeSendDatabase
	ContextDeadline time.Time
}

// DispatchMessagesLambda contains the config and dependency references.
type DispatchMessagesLambda struct {
	// Configuration
	AWSRegion string
	Project   string
	Env       string

	// Dependencies are injected here. If nil, Setup will create them.
	Deps *DispatchMessagesDependencies
}

// Setup performs all initialization tasks during the cold start.
// It creates any missing dependencies using the configuration parameters.
func (dml *DispatchMessagesLambda) Setup(ctx context.Context) error {
	// Ensure dependency container exists.
	if dml.Deps == nil {
		dml.Deps = &DispatchMessagesDependencies{
			ContextDeadline: time.Now().Truncate(time.Minute).Add(55 * time.Second),
		}
	}

	// Setup logger.
	if dml.Deps.Logger == nil {
		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{})
		dml.Deps.Logger = logger

		// Initialize Sentry.
		sentry.InitializeSentry(dml.Deps.Logger.(*logrus.Logger), "typesend_message_dispatch")
	}

	// Create the dispatcher.
	if dml.Deps.Dispatcher == nil {
		dml.Deps.Logger.Debugf("Connecting to SQS in %s", dml.AWSRegion)
		sqsClient, err := typequeue_helpers.ConnectToSQS(dml.AWSRegion)
		if err != nil {
			dml.Deps.Logger.Errorf("failed to connect to SQS: %s", err.Error())
			return fmt.Errorf("failed to connect to SQS: %w", err)
		}

		dml.Deps.Logger.Debugf("Connecting to TypeQueue SSM Helper")

		ssmHelper := &typequeue_helpers.TypeQueueSSMHelper{
			AWSRegion: dml.AWSRegion,
			Environments: map[string]string{
				"dev":  fmt.Sprintf("/typesend/%s/dev", dml.Project),
				"prod": fmt.Sprintf("/typesend/%s/prod", dml.Project),
			},
			CurrentEnv: dml.Env,
			Logger:     dml.Deps.Logger,
		}
		if err := ssmHelper.Connect(); err != nil {
			dml.Deps.Logger.Errorf("failed to connect to SSM: %s", err.Error())
			return fmt.Errorf("failed to connect to SSM: %w", err)
		}
		dml.Deps.Dispatcher = typequeue.Dispatcher[*typesend_schemas.TypeSendEnvelope]{
			SQSClient:         sqsClient,
			GetTargetQueueURL: ssmHelper.GetTargetURL,
		}
	}

	// Connect to DynamoDB.
	if dml.Deps.DB == nil {
		dynamoCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		dynamo, err := typesend_db.NewDynamoDB(dynamoCtx, &typesend_db.DynamoConfig{
			Region:      dml.AWSRegion,
			TableName:   fmt.Sprintf("typesend_%s", dml.Project),
			ForceClient: &dynamodb.DynamoDB{},
		})
		if err != nil {
			dml.Deps.Logger.Errorf("failed to connect to DynamoDB: %s", err.Error())
			return fmt.Errorf("failed to connect to DynamoDB: %w", err)
		}
		dml.Deps.DB = dynamo
	}

	return nil
}

// HandleRequest is the lambda handler method.
func (dml *DispatchMessagesLambda) HandleRequest(ctx context.Context) error {
	// Create a context with a trace ID and deadline.
	sendingCtx := context.WithValue(context.Background(), "trace-id", uuid.NewString())
	// Since this function is run every minute, it is critical to
	// confirm it finishes within the minute.
	sendingCtx, cancel := context.WithDeadline(sendingCtx, dml.Deps.ContextDeadline)
	defer cancel()

	// Dispatch messages.
	err := dispatchMessagesReadyToSendFn(&dispatch_messages.DispatchOpts{
		Context:    sendingCtx,
		Database:   dml.Deps.DB,
		Dispatcher: dml.Deps.Dispatcher,
		Logger:     dml.Deps.Logger,
	})
	if err != nil {
		if err == context.DeadlineExceeded {
			internal.ProtectedInfoLogger(dml.Deps.Logger, "typesend: deadline exceeded, finishing this check.")
			return nil
		}
		internal.ProtectedErrorLogger(dml.Deps.Logger, "typesend: failed to DispatchMessagesReadyToSend: %s", err.Error())
		return err
	}

	return nil
}
