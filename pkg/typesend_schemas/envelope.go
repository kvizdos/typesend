package typesend_schemas

import (
	"time"

	typequeue "github.com/kvizdos/typequeue/pkg"
	"github.com/kvizdos/typesend/pkg/template_variables"
)

type TypeSendStatus int

const (
	TypeSendStatus_UNSENT     TypeSendStatus = 0
	TypeSendStatus_DELIVERING TypeSendStatus = 1
	TypeSendStatus_SENT       TypeSendStatus = 2
)

type TypeSendEnvelope struct {
	typequeue.SQSAble

	// Optional; can be used to send an email at a specific time.
	// Granulariy depends on your EventBridge Schedule.
	ScheduledFor time.Time `dynamodbav:"scheduledFor" json:"scheduledFor"`

	AppID string `dynamodbav:"app" json:"app"`

	ToAddress string `dynamodbav:"to" json:"to"`

	ToInternalID string `dynamodbav:"toInternal" json:"toInternal"`

	Variables template_variables.TypeSendVariableInterface `dynamodbav:"variables" json:"variables"`

	TemplateID string `dynamodbav:"tid" json:"tid"`

	// This ID will be automatically set,
	// however you may predefine for testing.
	ID string `dynamodbav:"id" json:"id"`

	Status TypeSendStatus `dynamodbav:"status" json:"status"`

	// Useful if you need to send to multiple emails and combine in the frontend.
	MessageGroupID string `dynamodbav:"group" json:"group"`

	ReferenceID string `dynamodbav:"ref" json:"ref"`
}
