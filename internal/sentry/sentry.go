package sentry

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
)

var connectedToSentry = false

func InitializeSentry(logger *logrus.Logger, subsystem string) {
	if connectedToSentry {
		return
	}

	dsn := os.Getenv("SENTRY_DSN")

	if dsn == "" {
		connectedToSentry = false
		return
	}

	sentryEnv := "development"

	if os.Getenv("ENV") == "production" {
		sentryEnv = "production"
	}

	logrus.Infof("Configuring Sentry for %s", sentryEnv)

	hook, err := sentrylogrus.New([]logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}, sentry.ClientOptions{
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			return event
		},
		Dsn:              dsn,
		AttachStacktrace: true,
		SampleRate:       1,
		EnableTracing:    true,
		TracesSampleRate: 0.1,
		SendDefaultPII:   false,
		ServerName:       subsystem,
		Environment:      sentryEnv,
	})

	if err != nil {
		logrus.Fatalf("sentrylogrus.NewHook: %s", err)
	}

	defer hook.Flush(5 * time.Second)

	logger.AddHook(hook)

	logrus.RegisterExitHandler(func() { hook.Flush(5 * time.Second) })

	connectedToSentry = true
}
