package typesend_db_test

import (
	"context"
	"testing"

	"github.com/kvizdos/typesend/pkg/testutils"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_InsertTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, container, err := testutils.SetupDynamoDBLocalSession(context.Background())
	if ok := assert.NoError(t, err, "DynamoDB Setup Should Not Return Error"); !ok {
		return
	}
	defer testutils.KillContainer(container)

	db, err := typesend_db.NewDynamoDB(context.Background(), &typesend_db.DynamoConfig{
		Region:         "us-west-2",
		TemplatesTable: "test-typesend",
		ForceClient:    client,
	})
	assert.NoError(t, err)

	err = db.InsertTemplate(context.Background(), createTestTemplate("test-template"))

	assert.NoError(t, err, "No error expected on DynamoDB.InsertTemplate")
}

func TestIntegration_GetTemplateByExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, container, err := testutils.SetupDynamoDBLocalSession(context.Background())
	if ok := assert.NoError(t, err, "DynamoDB Setup Should Not Return Error"); !ok {
		return
	}
	defer testutils.KillContainer(container)

	db, err := typesend_db.NewDynamoDB(context.Background(), &typesend_db.DynamoConfig{
		Region:         "us-west-2",
		TemplatesTable: "test-typesend",
		ForceClient:    client,
	})
	assert.NoError(t, err)

	insertedTemplate := createTestTemplate("test-template")
	err = db.InsertTemplate(context.Background(), insertedTemplate)
	assert.NoError(t, err, "No error expected on DynamoDB.InsertTemplate")

	template, err := db.GetTemplateByID(context.Background(), "test-template", "base")
	assert.NoError(t, err, "No error expected on DynamoDB.GetTemplateByID")
	assert.NotNil(t, template, "Template should not be nil")
	assert.Equal(t, insertedTemplate, template)
}

func TestIntegration_GetTemplateByTenantDoesNotExistButBaseDoes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, container, err := testutils.SetupDynamoDBLocalSession(context.Background())
	if ok := assert.NoError(t, err, "DynamoDB Setup Should Not Return Error"); !ok {
		return
	}
	defer testutils.KillContainer(container)

	db, err := typesend_db.NewDynamoDB(context.Background(), &typesend_db.DynamoConfig{
		Region:         "us-west-2",
		TemplatesTable: "test-typesend",
		ForceClient:    client,
	})
	assert.NoError(t, err)

	insertedTemplate := createTestTemplate("test-template")
	err = db.InsertTemplate(context.Background(), insertedTemplate)
	assert.NoError(t, err, "No error expected on DynamoDB.InsertTemplate")

	template, err := db.GetTemplateByID(context.Background(), "test-template", "test-tenant")
	assert.NoError(t, err, "No error expected on DynamoDB.GetTemplateByID")
	assert.NotNil(t, template, "Template should not be nil")
	assert.Equal(t, insertedTemplate, template)
}

func TestIntegration_GetTemplateByTenantDoesExist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, container, err := testutils.SetupDynamoDBLocalSession(context.Background())
	if ok := assert.NoError(t, err, "DynamoDB Setup Should Not Return Error"); !ok {
		return
	}
	defer testutils.KillContainer(container)

	db, err := typesend_db.NewDynamoDB(context.Background(), &typesend_db.DynamoConfig{
		Region:         "us-west-2",
		TemplatesTable: "test-typesend",
		ForceClient:    client,
	})
	assert.NoError(t, err)

	baseTemplate := createTestTemplate("test-template")
	err = db.InsertTemplate(context.Background(), baseTemplate)
	assert.NoError(t, err, "No error expected on DynamoDB.InsertTemplate")

	tenantTemplate := createTestTemplate("test-template")
	tenantTemplate.TenantID = "test-tenant"
	err = db.InsertTemplate(context.Background(), tenantTemplate)
	assert.NoError(t, err, "No error expected on DynamoDB.InsertTemplate")

	template, err := db.GetTemplateByID(context.Background(), "test-template", "test-tenant")
	assert.NoError(t, err, "No error expected on DynamoDB.GetTemplateByID")
	assert.NotNil(t, template, "Template should not be nil")
	assert.Equal(t, tenantTemplate, template)
}

func TestIntegration_GetTemplateByDoesNotExist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, container, err := testutils.SetupDynamoDBLocalSession(context.Background())
	if ok := assert.NoError(t, err, "DynamoDB Setup Should Not Return Error"); !ok {
		return
	}
	defer testutils.KillContainer(container)

	db, err := typesend_db.NewDynamoDB(context.Background(), &typesend_db.DynamoConfig{
		Region:         "us-west-2",
		TemplatesTable: "test-typesend",
		ForceClient:    client,
	})
	assert.NoError(t, err)

	template, err := db.GetTemplateByID(context.Background(), "test-template", "base")
	assert.NoError(t, err, "No error expected on DynamoDB.GetTemplateByID")
	assert.Nil(t, template, "Template should not be nil")
}
