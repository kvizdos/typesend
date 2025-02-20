package consume_messages

import (
	"context"
	"time"

	"github.com/kvizdos/typesend/internal"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type DeliverMessageOptions struct {
	Logger   typesend_schemas.Logger
	Database typesend_db.TypeSendDatabase
}

func DeliverMessage(opts *DeliverMessageOptions, queuedEnvelope *typesend_schemas.TypeSendEnvelope) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	envelope, err := opts.Database.GetEnvelopeByID(ctx, queuedEnvelope.ID)
	if err != nil {
		return err
	}

	if envelope == nil {
		internal.ProtectedErrorLogger(opts.Logger, "typesend: envelope ID is not found: %+v", *queuedEnvelope)
		return nil // don't retry..
	}

	if envelope.Status == typesend_schemas.TypeSendStatus_SENT {
		internal.ProtectedWarnLogger(opts.Logger, "typesend: rejecting duplicate send request for envelope %s", envelope.ID)
		return nil // don't resend!
	}

	if envelope.Status != typesend_schemas.TypeSendStatus_DELIVERING {
		internal.ProtectedWarnLogger(opts.Logger, "typesend: envelope (%s) was not marked as DELIVERING prior to receive. Will not process right now.", envelope.ID)
		return nil // don't retry; the scheduler is going to try and send it anyways.
	}

	template, err := opts.Database.GetTemplateByID(ctx, queuedEnvelope.TemplateID, queuedEnvelope.TenantID)

	err = template.Fill(queuedEnvelope.Variables)

	if err != nil {
		return err
	}

	/*
		- [x] Confirm envelope.ScheduledFor is before right now
		- [x] Confirm that the envelope is set to "DELIVERING" in Database
			-- [x] If its not, just return `nil` here. It will be retried by the Scheduler.
		- [x] Build the Template (e.g. fill variables)
		- Send Email
		... Easy right??
	*/

	return nil
}
