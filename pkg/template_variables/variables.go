package template_variables

type TypeSendVariableInterface interface {
	GetTemplateID() string
}

type TypeSendVariable struct {
	AssociatedTemplateID string `dynamodbav:"-" json:"template"`
}

func (t TypeSendVariable) GetTemplateID() string {
	return t.AssociatedTemplateID
}
