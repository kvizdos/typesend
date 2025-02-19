package typesend_db_test

import (
	"context"
	"testing"

	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/stretchr/testify/assert"
)

func createTestTemplate(tid string) *typesend_schemas.TypeSendTemplate {
	return &typesend_schemas.TypeSendTemplate{
		TemplateID:  tid,
		TenantID:    "base",
		Content:     "This is a test body.",
		Subject:     "This is a test subject.",
		FromAddress: "bob@example.com",
		FromName:    "Bobby",
	}
}

func TestTestDatabase_InsertTemplate(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	temp := createTestTemplate("test-template")
	err := db.InsertTemplate(context.Background(), temp)
	assert.NoError(t, err, "should not error when inserting template")

	assert.Len(t, db.Templates(), 1, "expected 1 template")
	assert.Equal(t, temp, db.Templates()[0])
}

func TestTestDatabase_GetTemplateByIDMissing(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	template, err := db.GetTemplateByID(context.Background(), "test-template", "base")
	assert.NoError(t, err, "should not error when getting template")
	assert.Nil(t, template)
}

func TestTestDatabase_GetTemplateByIDSuccess(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	temp := createTestTemplate("test-template")
	db.InsertTemplate(nil, temp)

	template, err := db.GetTemplateByID(context.Background(), "test-template", "base")
	assert.NoError(t, err, "should not error when getting template")
	assert.NotNil(t, template, "template should be found here..")
	assert.Equal(t, temp, db.Templates()[0])
}

func TestTestDatabase_GetTemplateByIDTenantSuccess(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	temp := createTestTemplate("test-template")
	temp.TenantID = "test-tenant"
	db.InsertTemplate(nil, temp)

	template, err := db.GetTemplateByID(context.Background(), "test-template", "test-tenant")
	assert.NoError(t, err, "should not error when getting template")
	assert.NotNil(t, template, "template should be found here..")
	assert.Equal(t, temp, db.Templates()[0])
}

func TestTestDatabase_GetTemplateByIDTenantSuccessRecursive(t *testing.T) {
	db := &typesend_db.TestDatabase{}
	_ = db.Connect(context.Background())

	temp := createTestTemplate("test-template")
	db.InsertTemplate(nil, temp)

	template, err := db.GetTemplateByID(context.Background(), "test-template", "test-tenant")
	assert.NoError(t, err, "should not error when getting template")
	assert.NotNil(t, template, "template should be found here..")
	assert.Equal(t, temp, db.Templates()[0])
}
