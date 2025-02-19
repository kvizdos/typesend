package typesend_schemas

import (
	"bytes"
	"html/template"
)

type TypeSendTemplate struct {
	TemplateID  string `dynamodbav:"id" json:"id"`
	Name        string `dynamodbav:"-" json:"name"`
	Description string `dynamodbav:"-" json:"description"`
	TenantID    string `dynamodbav:"tenant" json:"-"`
	Content     string `dynamodbav:"content" json:"-"`
	Subject     string `dynamodbav:"subject" json:"subject"`
	FromAddress string `dynamodbav:"from" json:"from"`
	FromName    string `dynamodbav:"from_name" json:"from_name"`
}

func (t TypeSendTemplate) Fill(vars map[string]interface{}) (*string, error) {
	tmpl, err := template.New("content").Parse(t.Content)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, err
	}

	out := buf.String()
	return &out, nil
}
