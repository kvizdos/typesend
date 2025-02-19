package typesend_templates_test

import (
	"context"
	"testing"

	"github.com/kvizdos/typesend/pkg/template_variables"
	"github.com/kvizdos/typesend/pkg/testutils"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/kvizdos/typesend/pkg/typesend_templates"
	"github.com/stretchr/testify/assert"
)

// This test should ensure that templates get added to the database
// upon registering.
func TestRegisterTemplateDoesNotExist(t *testing.T) {
	defer typesend_templates.Dangerous_ResetRegisteredTemplates()
	db := &typesend_db.TestDatabase{}
	err := db.Connect(context.Background())
	assert.NoError(t, err)

	expectTemplate := &typesend_schemas.TypeSendTemplate{
		TemplateID:  "test-template",
		TenantID:    "base",
		Content:     "This is a test body.",
		Subject:     "This is a test subject.",
		FromAddress: "bob@example.com",
		FromName:    "Bobby",
	}

	testTemplate := &typesend_templates.RegisteredTemplate{
		FromAddress:   expectTemplate.FromAddress,
		FromName:      expectTemplate.FromName,
		BootstrapBody: expectTemplate.Content,
		Variables: testutils.DummyVariable{
			TypeSendVariable: template_variables.TypeSendVariable{
				AssociatedTemplateID: expectTemplate.TemplateID,
			},
		},
		BootstrapSubject: expectTemplate.Subject,
	}

	err = typesend_templates.RegisterTemplate(db, "Demo UI Group", testTemplate)
	assert.NoError(t, err, "Registering a template should not cause an error")

	assert.Len(t, db.Templates(), 1, "expected 1 template")
	assert.Equal(t, expectTemplate, db.Templates()[0], "mismatch!")
}

func TestRegisterTemplateDoesNotRecreate(t *testing.T) {
	defer typesend_templates.Dangerous_ResetRegisteredTemplates()
	db := &typesend_db.TestDatabase{}
	err := db.Connect(context.Background())
	assert.NoError(t, err)

	expectTemplate := &typesend_schemas.TypeSendTemplate{
		TemplateID:  "test-template",
		TenantID:    "base",
		Content:     "This is a test body.",
		Subject:     "This is a test subject.",
		FromAddress: "bob@example.com",
		FromName:    "Bobby",
	}

	err = db.InsertTemplate(nil, expectTemplate)
	assert.NoError(t, err)
	assert.Len(t, db.Templates(), 1, "expected 1 template")

	testTemplate := &typesend_templates.RegisteredTemplate{
		Variables: testutils.DummyVariable{
			TypeSendVariable: template_variables.TypeSendVariable{
				AssociatedTemplateID: expectTemplate.TemplateID,
			},
		},
		FromAddress:      expectTemplate.FromAddress,
		FromName:         expectTemplate.FromName,
		BootstrapBody:    expectTemplate.Content,
		BootstrapSubject: expectTemplate.Subject,
	}

	err = typesend_templates.RegisterTemplate(db, "Demo UI Group", testTemplate)
	assert.NoError(t, err, "Registering a template should not cause an error")

	assert.Len(t, db.Templates(), 1, "expected 1 template")
	assert.Equal(t, expectTemplate, db.Templates()[0], "mismatch!")
}
