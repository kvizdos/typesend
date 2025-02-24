package providers

import (
	"github.com/kvizdos/typesend/pkg/typesend_metrics"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type TypeSendProvider interface {
	Deliver(e *typesend_schemas.TypeSendEnvelope, filledTemplate *typesend_schemas.TypeSendTemplate) error
	GetProviderName() string
	SetMetricProvider(typesend_metrics.MetricsProvider)
}
