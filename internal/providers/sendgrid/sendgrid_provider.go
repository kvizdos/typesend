package providers_sendgrid

import (
	"fmt"
	"net/http"

	"github.com/kvizdos/typesend/pkg/typesend_metrics"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridEmailClient interface {
	Send(message *mail.SGMailV3) (*rest.Response, error)
}

type SendGridProvider struct {
	Client  SendGridEmailClient
	Metrics typesend_metrics.MetricsProvider
}

func (s SendGridProvider) GetProviderName() string {
	return "SendGrid"
}

func (s *SendGridProvider) SetMetricProvider(to typesend_metrics.MetricsProvider) {
	s.Metrics = to
}

func (s SendGridProvider) Deliver(e *typesend_schemas.TypeSendEnvelope, filledTemplate *typesend_schemas.TypeSendTemplate) error {
	if s.Client == nil {
		if s.Metrics != nil {
			s.Metrics.DeliverEvent(&typesend_metrics.Metric{
				AppName:    e.AppID,
				TemplateID: e.TemplateID,
				TenantID:   e.TenantID,
				Success:    false,
			})
		}
		return fmt.Errorf("requires client")
	}

	from := mail.NewEmail(filledTemplate.FromName, filledTemplate.FromAddress)
	subject := filledTemplate.Subject
	to := mail.NewEmail(e.ToName, e.ToAddress)
	htmlContent := filledTemplate.Content
	message := mail.NewSingleEmail(from, subject, to, "Please view in HTML", htmlContent)
	message.CustomArgs = make(map[string]string)
	message.CustomArgs["X-Using-TypeSend"] = "true"
	message.CustomArgs["X-TypeSend-App"] = e.AppID
	message.CustomArgs["X-TypeSend-Tenant"] = e.TenantID
	message.CustomArgs["X-TypeSend-Envelope"] = e.ID

	response, err := s.Client.Send(message)
	if err != nil {
		if s.Metrics != nil {
			s.Metrics.DeliverEvent(&typesend_metrics.Metric{
				AppName:    e.AppID,
				TemplateID: e.TemplateID,
				TenantID:   e.TenantID,
				Success:    false,
			})
		}
		return err
	}

	if response.StatusCode != http.StatusAccepted {
		if s.Metrics != nil {
			s.Metrics.DeliverEvent(&typesend_metrics.Metric{
				AppName:    e.AppID,
				TemplateID: e.TemplateID,
				TenantID:   e.TenantID,
				Success:    false,
			})
		}
		return fmt.Errorf("sendgrid status code not Accepted (%d): %s", response.StatusCode, response.Body)
	}

	if s.Metrics != nil {
		s.Metrics.DeliverEvent(&typesend_metrics.Metric{
			AppName:    e.AppID,
			TemplateID: e.TemplateID,
			TenantID:   e.TenantID,
			Success:    true,
		})
	}

	return nil
}

func NewSendGridProvider(apiKey string) *SendGridProvider {
	client := sendgrid.NewSendClient(apiKey)
	return &SendGridProvider{
		Client: client,
	}
}
