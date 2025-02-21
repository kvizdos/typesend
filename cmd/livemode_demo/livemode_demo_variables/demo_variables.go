package livemode_demo_variables

import (
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_templates"
)

func RegisterVariables(db typesend_db.TypeSendDatabase) error {
	// Create a Test Template
	err := typesend_templates.RegisterTemplate(db, "Demo", &typesend_templates.RegisteredTemplate{
		Variables:        LiveModeDemoVariable{},
		FromAddress:      "example@demo.org",
		FromName:         "Test From Name",
		BootstrapBody:    "<p>Hello, the reset link is {{.ResetURL}}.</p>",
		BootstrapSubject: "Your Reset Link: expires in {{.ExpiresIn}} minutes!",
	})
	if err != nil {
		return err
	}
	return nil
}
