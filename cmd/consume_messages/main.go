package main

import (
	"log"
	"os"

	"github.com/kvizdos/typesend/cmd/consume_messages/consume_messages_handler"
	"github.com/kvizdos/typesend/cmd/consume_messages/use_provider"
)

func main() {
	provider := use_provider.GetProvider()
	handler := &consume_messages_handler.ConsumeMessageHandler{
		AWSRegion: os.Getenv("AWS_REGION"),
		Project:   os.Getenv("TYPESEND_PROJECT"),
		Env:       os.Getenv("ENV"),
		Deps: &consume_messages_handler.ConsumeMessageHandlerDependencies{
			Provider: provider,
		},
	}
	err := handler.Setup()
	if err != nil {
		log.Fatalf("Failed to set up handler: %v", err)
	}
}
