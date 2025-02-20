package providers

import (
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TypeSendProvider interface {
	Deliver(e *typesend_schemas.TypeSendEnvelope, filledTemplate *typesend_schemas.TypeSendTemplate) error
	GetProviderName() string
}
