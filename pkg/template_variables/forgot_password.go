package template_variables

type ForgotPassword struct {
	TypeSendVariable

	ResetLink string `dynamodbav:"link" json:"link" typesend:"Direct Link to Reset the Users Password"`
}

func NewForgotPasswordVariables(base *ForgotPassword) *ForgotPassword {
	base.AssociatedTemplateID = "sdaik-forgot-password"
	return base
}
