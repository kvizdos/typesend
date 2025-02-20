package providers_sendgrid_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	providers_sendgrid "github.com/kvizdos/typesend/internal/providers/sendgrid"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/assert"
)

// mockEmailClient implements SendGridEmailClient for testing.
type mockEmailClient struct {
	SentMessage *mail.SGMailV3
	Response    *rest.Response
	Err         error
}

func (m *mockEmailClient) Send(message *mail.SGMailV3) (*rest.Response, error) {
	m.SentMessage = message
	return m.Response, m.Err
}

// Test that Deliver returns an error if the client is nil.
func TestDeliver_NilClient(t *testing.T) {
	provider := providers_sendgrid.SendGridProvider{
		Client: nil,
	}
	envelope := &typesend_schemas.TypeSendEnvelope{
		ToName:    "Recipient",
		ToAddress: "recipient@example.com",
		Variables: map[string]interface{}{},
	}
	template := &typesend_schemas.TypeSendTemplate{
		FromName:    "Sender",
		FromAddress: "sender@example.com",
		Subject:     "Test Subject",
		Content:     "<p>Hello World</p>",
	}
	err := provider.Deliver(envelope, template)
	assert.Error(t, err, "expected error due to nil client")
}

// Test that Deliver successfully sends an email and sets custom args.
func TestDeliver_Success(t *testing.T) {
	mockClient := &mockEmailClient{
		Response: &rest.Response{
			StatusCode: http.StatusAccepted,
			Body:       "Accepted",
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
		},
	}

	provider := providers_sendgrid.SendGridProvider{
		Client: mockClient,
	}

	envelope := &typesend_schemas.TypeSendEnvelope{
		ToName:    "Recipient",
		ToAddress: "recipient@example.com",
		AppID:     "TestApp",
		TenantID:  "TestTenant",
	}
	template := &typesend_schemas.TypeSendTemplate{
		FromName:    "Sender",
		FromAddress: "sender@example.com",
		Subject:     "Test Subject",
		Content:     "<p>Hello World</p>",
	}

	err := provider.Deliver(envelope, template)
	assert.NoError(t, err, "expected no error during successful delivery")
	assert.NotNil(t, mockClient.SentMessage, "expected a sent message to be set in mockClient")

	// Check that custom arguments were set.
	assert.Equal(t, "true", mockClient.SentMessage.CustomArgs["X-Using-TypeSend"], "expected custom arg X-Using-TypeSend to be 'true'")
	assert.Equal(t, "TestApp", mockClient.SentMessage.CustomArgs["X-TypeSend-App"], "expected custom arg X-TypeSend-App to be 'TestApp'")
	assert.Equal(t, "TestTenant", mockClient.SentMessage.CustomArgs["X-TypeSend-Tenant"], "expected custom arg X-TypeSend-App to be 'TestTenant'")
}

// Test that Deliver propagates the error when Send fails.
func TestDeliver_SendError(t *testing.T) {
	sendErr := errors.New("send error")
	mockClient := &mockEmailClient{
		Err: sendErr,
	}

	provider := providers_sendgrid.SendGridProvider{
		Client: mockClient,
	}

	envelope := &typesend_schemas.TypeSendEnvelope{
		ToName:    "Recipient",
		ToAddress: "recipient@example.com",
	}
	template := &typesend_schemas.TypeSendTemplate{
		FromName:    "Sender",
		FromAddress: "sender@example.com",
		Subject:     "Test Subject",
		Content:     "<p>Hello World</p>",
	}

	err := provider.Deliver(envelope, template)
	assert.EqualError(t, err, "send error", "expected the send error to be returned")
	assert.NotNil(t, mockClient.SentMessage, "expected a sent message attempt even if sending fails")
}

// Test that Deliver returns an error when response status code is not Accepted.
func TestDeliver_NonAcceptedStatusCode(t *testing.T) {
	nonAcceptedStatus := http.StatusBadRequest
	errorBody := "Bad Request"
	mockClient := &mockEmailClient{
		Response: &rest.Response{
			StatusCode: nonAcceptedStatus,
			Body:       errorBody,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
		},
	}

	provider := providers_sendgrid.SendGridProvider{
		Client: mockClient,
	}

	envelope := &typesend_schemas.TypeSendEnvelope{
		ToName:    "Recipient",
		ToAddress: "recipient@example.com",
	}
	template := &typesend_schemas.TypeSendTemplate{
		FromName:    "Sender",
		FromAddress: "sender@example.com",
		Subject:     "Test Subject",
		Content:     "<p>Hello World</p>",
	}

	err := provider.Deliver(envelope, template)
	expectedErr := fmt.Sprintf("sendgrid status code not Accepted (%d): %s", nonAcceptedStatus, errorBody)
	assert.EqualError(t, err, expectedErr, "expected error due to non-Accepted status code")
}
