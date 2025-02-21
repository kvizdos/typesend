package providers_testing

import (
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

// Logging Provider can be used in development mode
// to log emails out to the console.
type LoggingProvider struct {
	Logger typesend_schemas.Logger
}

// NewLoggingProvider creates a new instance of LoggingProvider.
func NewLoggingProvider(logger typesend_schemas.Logger) *LoggingProvider {
	return &LoggingProvider{
		Logger: logger,
	}
}

func (t *LoggingProvider) Deliver(e *typesend_schemas.TypeSendEnvelope, filledTemplate *typesend_schemas.TypeSendTemplate) error {
	t.Logger.Infof("--- EMAIL ---")
	t.Logger.Infof("TO: %s (%s) ---", e.ToName, e.ToAddress)
	t.Logger.Infof("SUBJECT: %s ---", filledTemplate.Subject)
	t.Logger.Infof("%s", filledTemplate.Content)
	t.Logger.Infof("--- END EMAIL ---")
	return nil
}

// GetProviderName returns a fixed provider name.
func (t *LoggingProvider) GetProviderName() string {
	return "LoggingProvider"
}
