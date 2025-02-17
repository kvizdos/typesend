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

	for envelope := range envelopes {
		_, err := opts.Dispatcher.Dispatch(opts.Context, envelope, "email_queue")
		if err != nil {
			internal.ProtectedLogger(opts.Logger, "typesend: failed to dispatch (%s): %s", envelope.ID, err.Error())
			continue
		}

		err = opts.Database.UpdateEnvelopeStatus(opts.Context, envelope.ID, typesend_schemas.TypeSendStatus_DELIVERING)

		if err != nil {
			internal.ProtectedLogger(opts.Logger, "typesend: failed to update envelope status (%s): %s", envelope.ID, err.Error())
			continue
		}
	}

	return nil
}
