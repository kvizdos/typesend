package typesend_db

import (
	"context"
	"time"

	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TypeSendDatabase interface {
	Connect(ctx context.Context) error
	Insert(envelope *typesend_schemas.TypeSendEnvelope) error
	GetEnvelopeByID(ctx context.Context, envelopeID string) (*typesend_schemas.TypeSendEnvelope, error)
	GetMessagesReadyToSend(ctx context.Context, timestamp time.Time) (chan *typesend_schemas.TypeSendEnvelope, error)
	UpdateEnvelopeStatus(ctx context.Context, envelopeID string, toStatus typesend_schemas.TypeSendStatus) error

	GetTemplateByID(ctx context.Context, templateID string, tenantID string) (*typesend_schemas.TypeSendTemplate, error)
	InsertTemplate(context.Context, *typesend_schemas.TypeSendTemplate) error
}
