package testutils

import "github.com/kvizdos/typesend/pkg/typesend_schemas"

// DummyVariable is a simple implementation of template_variables.TypeSendVariableInterface
type DummyVariable struct {
	typesend_schemas.TypeSendVariable
}
