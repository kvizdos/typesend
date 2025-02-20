package providers_email

import (
	"fmt"
	"os"

	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridProvider struct {
}

func (s SendGridProvider) Deliver(e *typesend_schemas.TypeSendEnvelope, filledTemplate *typesend_schemas.TypeSendTemplate) error {
	from := mail.NewEmail(filledTemplate.FromName, filledTemplate.FromAddress)
	subject := filledTemplate.Subject
	to := mail.NewEmail("Example Recipient", e.ToAddress)
	htmlContent := filledTemplate.Content
	message := mail.NewSingleEmail(from, subject, to, "Please view in HTML", htmlContent)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return nil
	}

	fmt.Printf("Status Code: %d\nBody: %s\nHeaders: %v\n", response.StatusCode, response.Body, response.Headers)
	return nil
}
