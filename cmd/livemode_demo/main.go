package main

import (
	"context"
	"time"

	"github.com/kvizdos/typesend/cmd/livemode_demo/livemode_demo_variables"
	"github.com/kvizdos/typesend/pkg/typesend_livemode"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	// Create the LiveMode TypeSend
	ts, db := typesend_livemode.StartTypeSendLive(context.Background(), logger, "demo-app")

	err := livemode_demo_variables.RegisterVariables(db)

	if err != nil {
		logger.Panic(err)
		return
	}

	_, err = ts.Send(typesend_schemas.TypeSendTo{
		ToAddress:    "kvizdos@example.com",
		ToName:       "Kenton Vizdos",
		ToInternalID: "uuid-demo",
	}, livemode_demo_variables.LiveModeDemoVariable{
		ResetURL:  "https://example.com",
		ExpiresIn: 5 * time.Minute,
		// }, time.Time{})
	}, time.Now().UTC().Add(10*time.Minute))

	if err != nil {
		logger.Panic(err)
		return
	}

	select {}
}
