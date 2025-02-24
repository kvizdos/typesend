package typesend_metrics_cloudwatch

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/kvizdos/typesend/pkg/typesend_metrics"
)

// CloudWatchProvider implements the MetricsProvider interface using AWS CloudWatch.
type CloudWatchProvider struct {
	client     *cloudwatch.CloudWatch
	namespace  string
	metricName string
}

// NewCloudWatchProvider creates a new instance of CloudWatchProvider.
func NewCloudWatchProvider(namespace, metricName, region string) (*CloudWatchProvider, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return &CloudWatchProvider{
		client:     cloudwatch.New(sess),
		namespace:  namespace,
		metricName: metricName,
	}, nil
}

func (p *CloudWatchProvider) SendEvent(metric *typesend_metrics.Metric) error {
	input := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(p.namespace),
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String(fmt.Sprintf("%s-sends", p.metricName)),
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  aws.String("AppName"),
						Value: aws.String(metric.AppName),
					},
					{
						Name:  aws.String("TemplateID"),
						Value: aws.String(metric.TemplateID),
					},
					{
						Name:  aws.String("TenantID"),
						Value: aws.String(metric.TenantID),
					},
				},
				Value: aws.Float64(1.0),
				Unit:  aws.String("Count"),
			},
		},
	}

	_, err := p.client.PutMetricData(input)

	return err
}

func (p *CloudWatchProvider) DeliverEvent(metric *typesend_metrics.Metric) error {
	status := "Failure"
	if metric.Success {
		status = "Success"
	}

	input := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(p.namespace),
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String(fmt.Sprintf("%s-deliveries", p.metricName)),
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  aws.String("AppName"),
						Value: aws.String(metric.AppName),
					},
					{
						Name:  aws.String("TemplateID"),
						Value: aws.String(metric.TemplateID),
					},
					{
						Name:  aws.String("TenantID"),
						Value: aws.String(metric.TenantID),
					},
					{
						Name:  aws.String("Status"),
						Value: aws.String(status),
					},
				},
				Value: aws.Float64(1.0),
				Unit:  aws.String("Count"),
			},
		},
	}

	_, err := p.client.PutMetricData(input)

	return err
}
