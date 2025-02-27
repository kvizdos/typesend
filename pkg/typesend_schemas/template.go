package typesend_schemas

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig/v3"
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

func (t *TypeSendTemplate) Fill(vars map[string]interface{}) error {
	if err := t.fillContent(vars); err != nil {
		return err
	}
	if err := t.fillSubject(vars); err != nil {
		return err
	}
	return nil
}

func (t *TypeSendTemplate) fillContent(vars map[string]interface{}) error {
	tmpl, err := template.New("content").Funcs(sprig.FuncMap()).Parse(t.Content)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return err
	}

	t.Content = buf.String()
	return nil
}

func (t *TypeSendTemplate) fillSubject(vars map[string]interface{}) error {
	tmpl, err := template.New("content").Funcs(sprig.FuncMap()).Parse(t.Subject)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return err
	}

	t.Subject = buf.String()
	return nil
}
