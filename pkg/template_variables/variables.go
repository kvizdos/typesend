package template_variables

type TypeSendVariableInterface interface {
	GetTemplateID() string
	ToMap() map[string]interface{}
}

type TypeSendVariable struct {
	AssociatedTemplateID string `dynamodbav:"-" json:"template"`
}

func (t TypeSendVariable) GetTemplateID() string {
	return t.AssociatedTemplateID
}

func (t TypeSendVariable) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}
