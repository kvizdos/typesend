package providers

import (
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TypeSendProvider interface {
	Send(to typesend_schemas.TypeSendTo, withVariables typesend_schemas.TypeSendVariable) error
}
