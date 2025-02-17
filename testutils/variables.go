package testutils

import (
	"github.com/kvizdos/typesend/pkg/template_variables"
)

// DummyVariable is a simple implementation of template_variables.TypeSendVariableInterface
type DummyVariable struct {
	template_variables.TypeSendVariable
}
