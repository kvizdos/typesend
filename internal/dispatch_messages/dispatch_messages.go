package dispatch_messages

import (
	"context"
	"time"

	typequeue "github.com/kvizdos/typequeue/pkg"
	"github.com/kvizdos/typesend/internal"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type DispatchOpts struct {
	Context    context.Context
	Database   typesend_db.TypeSendDatabase
	Dispatcher typequeue.TypeQueueDispatcher[*typesend_schemas.TypeSendEnvelope]
	Logger     typesend_schemas.Logger
}

func DispatchMessagesReadyToSend(opts *DispatchOpts) error {
	envelopes, err := opts.Database.GetMessagesReadyToSend(opts.Context, time.Now().UTC())

	if err != nil {
		return err
	}

	successSends := 0
	failedSends := 0
	failedUpdates := 0

	defer func() {
		if successSends > 0 || failedSends > 0 || failedUpdates > 0 {
			internal.ProtectedInfoLogger(opts.Logger, "typesend: sent %d messages (failed %d to send, %d failed to update)", successSends, failedSends, failedUpdates)
		}
	}()

	for envelope := range envelopes {
		select {
		case <-opts.Context.Done():
			return context.DeadlineExceeded
		default:
		}

		_, err := opts.Dispatcher.Dispatch(opts.Context, envelope, "email_queue")
		if err != nil {
			failedSends += 1
			internal.ProtectedErrorLogger(opts.Logger, "typesend: failed to dispatch (%s): %s", envelope.ID, err.Error())
			continue
		}

		err = opts.Database.UpdateEnvelopeStatus(opts.Context, envelope.ID, typesend_schemas.TypeSendStatus_DELIVERING)

		if err != nil {
			failedUpdates += 1
			internal.ProtectedErrorLogger(opts.Logger, "typesend: failed to update envelope status (%s): %s", envelope.ID, err.Error())
			continue
		}

		successSends += 1
	}

	return nil
}
