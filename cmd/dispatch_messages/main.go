package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kvizdos/typesend/cmd/dispatch_messages/dispatch_messages_handler"
)

func main() {
	handler := &dispatch_messages_handler.DispatchMessagesLambda{
		AWSRegion: os.Getenv("AWS_REGION"),
		Project:   os.Getenv("TYPESEND_PROJECT"),
		Env:       os.Getenv("ENV"),
	}
	lambda.Start(handler.HandleRequest)
}
