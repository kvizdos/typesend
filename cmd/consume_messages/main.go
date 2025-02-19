package main

import (
	"log"
	"os"

	"github.com/kvizdos/typesend/cmd/consume_messages/consume_messages_handler"
)

func main() {
	handler := &consume_messages_handler.ConsumeMessageHandler{
		AWSRegion: os.Getenv("AWS_REGION"),
		Project:   os.Getenv("TYPESEND_PROJECT"),
		Env:       os.Getenv("ENV"),
	}
	err := handler.Setup()
	if err != nil {
		log.Fatalf("Failed to set up handler: %v", err)
	}
}
