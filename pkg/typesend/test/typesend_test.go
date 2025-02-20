package typesend_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/kvizdos/typesend/pkg/testutils"
	"github.com/kvizdos/typesend/pkg/typesend"
	"github.com/kvizdos/typesend/pkg/typesend_db"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

func TestStubbed_Send(t *testing.T) {
	// Create and connect the in-memory test database.
	ctx := context.Background()
	db := &typesend_db.TestDatabase{}
	err := db.Connect(ctx)
	assert.NoError(t, err)

	// Initialize TypeSend with the test database.
	ts := &typesend.TypeSend{
		AppID:    "test-app",
		Database: db,
	}

	// Build a TypeSendTo instance.
	to := typesend_schemas.TypeSendTo{
		ToAddress:    "test.dsad+test@example.com",
		ToInternalID: "internal-123",
		// Leave MessageGroupID empty so that Send() generates one.
	}

	// Set a UTC timestamp for sending.
	sendAt := time.Now().UTC()

	// Create dummy template variables.
	vars := testutils.DummyVariable{
		TypeSendVariable: typesend_schemas.TypeSendVariable{
			AssociatedTemplateID: uuid.NewString(),
		},
	}

	// Call Send, which should insert an envelope into our test database.
	id, err := ts.Send(to, vars, sendAt)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)

	// Verify that the envelope was inserted.
	items := db.Items()
	assert.Len(t, items, 1)

	envelope := items[0]
	assert.Equal(t, "test-app", envelope.AppID)
	assert.Equal(t, to.ToAddress, envelope.ToAddress)
	assert.Equal(t, to.ToInternalID, envelope.ToInternalID)
	assert.Equal(t, sendAt, envelope.ScheduledFor)
	assert.Equal(t, typesend_schemas.TypeSendStatus_UNSENT, envelope.Status)
	assert.Equal(t, id, envelope.ID)
	assert.Equal(t, vars.GetTemplateID(), envelope.TemplateID)
	assert.NotEmpty(t, envelope.MessageGroupID)
}

func TestStubbed_Send_InvalidEmailFormat(t *testing.T) {
	// Create and connect the in-memory test database.
	ctx := context.Background()
	db := &typesend_db.TestDatabase{}
	err := db.Connect(ctx)
	assert.NoError(t, err)

	// Initialize TypeSend with the test database.
	ts := &typesend.TypeSend{
		AppID:    "test-app",
		Database: db,
	}

	// Build a TypeSendTo instance.
	to := typesend_schemas.TypeSendTo{
		ToAddress:    "bademail",
		ToInternalID: "internal-123",
		// Leave MessageGroupID empty so that Send() generates one.
	}

	// Set a UTC timestamp for sending.
	sendAt := time.Now().UTC()

	// Create dummy template variables.
	vars := testutils.DummyVariable{}

	// Call Send, which should insert an envelope into our test database.
	_, err = ts.Send(to, vars, sendAt)
	assert.ErrorIs(t, err, typesend.TypeSendError_INVALID_EMAIL)
}
