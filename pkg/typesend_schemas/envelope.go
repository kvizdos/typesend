package typesend_schemas

import (
	"time"

	typequeue "github.com/kvizdos/typequeue/pkg"
)

type TypeSendStatus int

const (
	TypeSendStatus_UNSENT     TypeSendStatus = 0
	TypeSendStatus_DELIVERING TypeSendStatus = 1
	TypeSendStatus_SENT       TypeSendStatus = 2
)

type TypeSendTo struct {
	ToAddress      string
	ToName         string
	ToInternalID   string
	MessageGroupID string
}

type TypeSendEnvelope struct {
	typequeue.SQSAble

	// Optional; can be used to send an email at a specific time.
	// Granulariy depends on your EventBridge Schedule.
	ScheduledFor time.Time `dynamodbav:"scheduledFor" json:"scheduledFor"`

	AppID string `dynamodbav:"app" json:"app"`

	ToAddress string `dynamodbav:"to" json:"to"`

	ToName string `dynamodbav:"to_name" json:"to_name"`

	ToInternalID string `dynamodbav:"toInternal" json:"toInternal"`

	TenantID string `dynamodbav:"tenant" json:"tenant"`

	Variables map[string]interface{} `dynamodbav:"variables" json:"variables"`

	TemplateID string `dynamodbav:"tid" json:"tid"`

	// This ID will be automatically set,
	// however you may predefine for testing.
	ID string `dynamodbav:"id" json:"id"`

	Status TypeSendStatus `dynamodbav:"status" json:"status"`

	// Useful if you need to send to multiple emails and combine in the frontend.
	MessageGroupID string `dynamodbav:"group" json:"group"`

	ReferenceID string `dynamodbav:"ref" json:"ref"`
}
