package consume_messages

import (
	"context"
	"fmt"
	"time"

	"github.com/kvizdos/typesend/internal"
	"github.com/kvizdos/typesend/internal/providers"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type DeliverMessageOptions struct {
	Logger   typesend_schemas.Logger
	Database typesend_db.TypeSendDatabase
	Provider providers.TypeSendProvider
}

func DeliverMessage(opts *DeliverMessageOptions, queuedEnvelope *typesend_schemas.TypeSendEnvelope) error {
	/*
		- [x] Confirm envelope.ScheduledFor is before right now
		- [x] Confirm the envelope hasn't been sent yet.
		- [x] Confirm that the envelope is set to "DELIVERING" in Database
			-- [x] If its not, just return `nil` here. It will be retried by the Scheduler.
		- [x] Build the Template (e.g. fill variables)
		- [x] Update Envelope status to "SENT"
		- Send Email
		... Easy right??
	*/
	if queuedEnvelope.ScheduledFor.After(time.Now().UTC().Add(30 + time.Second)) {
		return fmt.Errorf("message not ready to be delivered")
	}

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

	if envelope.Status != typesend_schemas.TypeSendStatus_DELIVERING && envelope.Status != typesend_schemas.TypeSendStatus_FAILED {
		internal.ProtectedWarnLogger(opts.Logger, "typesend: envelope (%s) was not marked as DELIVERING/FAILED prior to receive. Will not process right now.", envelope.ID)
		return nil // don't retry; the scheduler is going to try and send it anyways.
	}

	template, err := opts.Database.GetTemplateByID(ctx, queuedEnvelope.TemplateID, queuedEnvelope.TenantID)

	if err != nil {
		return err
	}

	if template == nil {
		return fmt.Errorf("could not find associated template ID")
	}

	err = template.Fill(queuedEnvelope.Variables)

	if err != nil {
		return err
	}

	err = opts.Database.UpdateEnvelopeStatus(context.Background(), envelope.ID, typesend_schemas.TypeSendStatus_SENT)

	if err != nil {
		// Should retry here; we're before the "moment of no return"
		// Worst case scenario, an email that was meant to be sent
		// wasn't sent.
		// IMO, this is better than updating Post-Delivery where emails
		// could be sent multiple times.
		return fmt.Errorf("failed to update envelope status to DELIVERING: %w", err)
	}

	err = opts.Provider.Deliver(envelope, template)

	if err != nil {
		opts.Database.UpdateEnvelopeStatus(context.Background(), envelope.ID, typesend_schemas.TypeSendStatus_FAILED)

		internal.ProtectedErrorLogger(opts.Logger, "Failed to deliver envelope via %s (%s): %s", opts.Provider.GetProviderName(), envelope.ID, err.Error())
		return nil // Causes the scheduler to re-queue this message.
	}

	return nil
}
