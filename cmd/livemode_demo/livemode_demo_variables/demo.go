package livemode_demo_variables

import "time"

type LiveModeDemoVariable struct {
	ResetURL  string
	ExpiresIn time.Duration
}

func (t LiveModeDemoVariable) GetTemplateID() string {
	return "livemode-demo-template"
}

func (t LiveModeDemoVariable) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ResetURL":  t.ResetURL,
		"ExpiresIn": t.ExpiresIn.Minutes(),
	}
}
