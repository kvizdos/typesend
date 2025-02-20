package testutils

import (
	"time"

	"github.com/google/uuid"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

// helper to create a test envelope
func CreateTestEnvelope(status typesend_schemas.TypeSendStatus, scheduledFor time.Time) *typesend_schemas.TypeSendEnvelope {
	return &typesend_schemas.TypeSendEnvelope{
		ID:             uuid.NewString(),
		AppID:          "test",
		ToAddress:      "test@example.com",
		ToInternalID:   "internal",
		MessageGroupID: "group",
		TemplateID:     uuid.NewString(),
		Variables:      nil,
		ScheduledFor:   scheduledFor,
		Status:         status,
		TenantID:       "base",
	}
}
