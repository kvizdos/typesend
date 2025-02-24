package use_provider

import (
	"os"

	"github.com/kvizdos/typesend/internal/providers"
	providers_sendgrid "github.com/kvizdos/typesend/internal/providers/sendgrid"
	"github.com/sendgrid/sendgrid-go"
)

func GetProvider() providers.TypeSendProvider {
	sendgridAPIKey := os.Getenv("TYPESEND_SENDGRID_KEY")
	if sendgridAPIKey == "" {
		panic("Missing TYPESEND_SENDGRID_KEY env")
	}
	client := sendgrid.NewSendClient(sendgridAPIKey)
	return &providers_sendgrid.SendGridProvider{
		Client: client,
	}
}
