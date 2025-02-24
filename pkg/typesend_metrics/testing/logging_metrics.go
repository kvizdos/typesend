package typesend_metrics_testing

import (
	"github.com/kvizdos/typesend/pkg/typesend_metrics"
	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

type LoggingProvider struct {
	namespace  string
	metricName string
	logger     typesend_schemas.Logger
}

func NewLoggingProvider(namespace, metricName string, logger typesend_schemas.Logger) (*LoggingProvider, error) {
	return &LoggingProvider{
		namespace:  namespace,
		metricName: metricName,
		logger:     logger,
	}, nil
}

func (p *LoggingProvider) SendEvent(metric *typesend_metrics.Metric) error {
	p.logger.Infof("SendEvent = appID=%s templateID=%s tenantID=%s", metric.AppName, metric.TemplateID, metric.TenantID)

	return nil
}

func (p *LoggingProvider) DeliverEvent(metric *typesend_metrics.Metric) error {
	status := "Failure"
	if metric.Success {
		status = "Success"
	}

	p.logger.Infof("DeliverEvent = appID=%s templateID=%s tenantID=%s status=%s", metric.AppName, metric.TemplateID, metric.TenantID, status)

	return nil
}
