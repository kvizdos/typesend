package typesend_livemode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	typequeue_mocks "github.com/kvizdos/typequeue/pkg/mocked"
	"github.com/kvizdos/typesend/cmd/dispatch_messages/dispatch_messages_handler"
	"github.com/kvizdos/typesend/internal/consume_messages"
	"github.com/kvizdos/typesend/internal/providers"
	providers_sendgrid "github.com/kvizdos/typesend/internal/providers/sendgrid"
	providers_testing "github.com/kvizdos/typesend/internal/providers/tester"
	"github.com/kvizdos/typesend/pkg/typesend"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	typesend_metrics_testing "github.com/kvizdos/typesend/pkg/typesend_metrics/testing"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

func StartTypeSendLive(ctx context.Context, logger typesend_schemas.Logger, appID string) (*typesend.TypeSend, *typesend_db.TestDatabase) {
	// Demo Sendgrid
	sgKey := os.Getenv("TYPESEND_SENDGRID_KEY")
	var provider providers.TypeSendProvider
	if sgKey != "" {
		logger.Infof("⚠️ TypeSend Live Mode using SendGrid")
		provider = providers_sendgrid.NewSendGridProvider(sgKey)
	} else {
		logger.Infof("✅ TypeSend Live Mode using Logger")
		provider = providers_testing.NewLoggingProvider(logger)
	}

	loggingMetrics, _ := typesend_metrics_testing.NewLoggingProvider("demo", "demo", logger)

	provider.SetMetricProvider(loggingMetrics)

	msgsChan := make(chan *typesend_schemas.TypeSendEnvelope)
	db := &typesend_db.TestDatabase{
		LiveModeChan: msgsChan,
	}
	db.Connect(nil)

	dispatchedChan := make(chan *typesend_schemas.TypeSendEnvelope)
	dispatcher := &typequeue_mocks.MockDispatcher[*typesend_schemas.TypeSendEnvelope]{
		DispatchChan: dispatchedChan,
	}

	dispatchLambda := dispatch_messages_handler.DispatchMessagesLambda{
		AWSRegion: "us-fake-1",
		Project:   appID,
		Env:       "dev",
		TraceID:   uuid.NewString(),
		Deps: &dispatch_messages_handler.DispatchMessagesDependencies{
			Logger:          logger,
			Dispatcher:      dispatcher,
			DB:              db,
			ContextDeadline: time.Now().UTC().AddDate(1, 0, 0),
		},
	}

	ts := &typesend.TypeSend{
		AppID:             appID,
		Database:          db,
		MetricProvider:    loggingMetrics,
		LiveMode_ForceNow: true,
		LiveMode_Logger:   logger,
	}

	// Listener for msgsChan in its own goroutine.
	go func() {
		defer close(msgsChan)
		for {
			select {
			case <-ctx.Done():
				logger.Infof("Shutting down TypeSend Live Mode (msgsChan listener)")
				return
			case e, ok := <-msgsChan:
				if !ok {
					// Channel has been closed.
					return
				}
				err := dispatchLambda.HandleRequest(context.Background())
				if err != nil {
					logger.Errorf("Failed to handle dispatchLambda request: %s -- %+v", err.Error(), *e)
				}
			}
		}
	}()

	// Listener for dispatchedChan in its own goroutine.
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("Shutting down TypeSend Live Mode (dispatchedChan listener)")
				return
			case e := <-dispatchedChan:
				err := consume_messages.DeliverMessage(&consume_messages.DeliverMessageOptions{
					Logger:   logger,
					Database: db,
					Provider: provider,
				}, e)
				if err != nil {
					logger.Errorf("Failed to handle consumeLambda request: %s -- %+v", err.Error(), *e)
				}
			}
		}
	}()

	return ts, db
}

func envelopeToSQSMessage(envelope *typesend_schemas.TypeSendEnvelope) events.SQSMessage {
	js, err := json.Marshal(envelope)
	if err != nil {
		panic(err)
	}

	return events.SQSMessage{
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
