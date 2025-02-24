package typesend_metrics

type Metric struct {
	AppName    string
	TemplateID string
	TenantID   string
	Success    bool
}

type MetricsProvider interface {
	SendEvent(metric *Metric) error
	DeliverEvent(metric *Metric) error
}
