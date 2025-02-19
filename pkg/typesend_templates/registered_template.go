package typesend_templates

import (
	"context"
	"time"

	"github.com/kvizdos/typesend/pkg/template_variables"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type RegisteredTemplate struct {
	Variables   template_variables.TypeSendVariableInterface
	FromAddress string
	FromName    string

	BootstrapBody    string
	BootstrapSubject string
}

var registeredTemplates = make(map[string]*RegisteredTemplate)

// Use for tests
func Dangerous_ResetRegisteredTemplates() {
	registeredTemplates = make(map[string]*RegisteredTemplate)
}

func RegisterTemplate(db typesend_db.TypeSendDatabase, UIGroup string, t *RegisteredTemplate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	template, err := db.GetTemplateByID(ctx, t.Variables.GetTemplateID(), "base")

	if err != nil {
		return err
	}

	if template == nil {
		baseTemplate := &typesend_schemas.TypeSendTemplate{
			TemplateID:  t.Variables.GetTemplateID(),
			TenantID:    "base",
			Content:     t.BootstrapBody,
			Subject:     t.BootstrapSubject,
			FromAddress: t.FromAddress,
			FromName:    t.FromName,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := db.InsertTemplate(ctx, baseTemplate)
		if err != nil {
			return err
		}
	}

	registeredTemplates[UIGroup] = t

	return nil
}
