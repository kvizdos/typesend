package typesend

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_metrics"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TypeSend struct {
	AppID string

	Database typesend_db.TypeSendDatabase

	MetricProvider typesend_metrics.MetricsProvider

	// All envelopes will be sent NOW.
	// Used in Live Mode for testing.
	LiveMode_ForceNow bool
	LiveMode_Logger   typesend_schemas.Logger
}

func (t *TypeSend) Send(to typesend_schemas.TypeSendTo, variables typesend_schemas.TypeSendVariableInterface, sendAt time.Time) (string, error) {
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

		if t.LiveMode_ForceNow {
			t.LiveMode_Logger.Infof("Envelope to %s scheduled to be sent in %s", to.ToAddress, sendAt.Sub(time.Now()))
			sendAt = time.Now().UTC()
		}
	}

	if to.MessageGroupID == "" {
		to.MessageGroupID = uuid.NewString()
	}

	if to.ToTenantID == "" {
		to.ToTenantID = "base"
	}

	err := t.Database.Insert(&typesend_schemas.TypeSendEnvelope{
		AppID:          t.AppID,
		ScheduledFor:   sendAt,
		ToAddress:      to.ToAddress,
		ToName:         to.ToName,
		ToInternalID:   to.ToInternalID,
		MessageGroupID: to.MessageGroupID,
		TenantID:       to.ToTenantID,
		TemplateID:     variables.GetTemplateID(),
		Variables:      variables.ToMap(),
		ID:             ID,
		Status:         typesend_schemas.TypeSendStatus_UNSENT,
	})

	if t.MetricProvider != nil {
		t.MetricProvider.SendEvent(&typesend_metrics.Metric{
			AppName:    t.AppID,
			TemplateID: variables.GetTemplateID(),
			TenantID:   to.ToTenantID,
		})
	}

	return ID, err
}
