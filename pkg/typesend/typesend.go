package typesend

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/kvizdos/typesend/pkg/template_variables"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TypeSendTo struct {
	ToAddress      string
	ToInternalID   string
	MessageGroupID string
}

type TypeSend struct {
	AppID string

	Database typesend_db.TypeSendDatabase
}

func (t *TypeSend) Send(to TypeSendTo, variables template_variables.TypeSendVariableInterface, sendAt time.Time) (string, error) {
	if _, err := mail.ParseAddress(to.ToAddress); err != nil {
		return "", TypeSendError_INVALID_EMAIL
	}

	ID := uuid.NewString()

	if sendAt.IsZero() {
		sendAt = time.Now().UTC()
	} else {
		if sendAt.Location() != time.UTC {
			return "", TypeSendError_UTC_MISMATCH
		}
	}

	if to.MessageGroupID == "" {
		to.MessageGroupID = uuid.NewString()
	}

	err := t.Database.Insert(&typesend_schemas.TypeSendEnvelope{
		AppID:          t.AppID,
		ScheduledFor:   sendAt,
		ToAddress:      to.ToAddress,
		ToInternalID:   to.ToInternalID,
		MessageGroupID: to.MessageGroupID,
		TemplateID:     variables.GetTemplateID(),
		Variables:      variables.ToMap(),
		ID:             ID,
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
	})

	return ID, err
}
