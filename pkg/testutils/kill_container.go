package testutils

import (
	"context"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

func KillContainer(container testcontainers.Container) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := container.Terminate(ctx)
	if err != nil {
		log.Printf("FAILED TO TERMINATE: %s", err.Error())
	}
}
