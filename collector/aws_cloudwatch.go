package collector

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"

	"github.com/slok/ecs-exporter/log"
	"github.com/slok/ecs-exporter/types"
)

const (
	maxServicesCWAPI = 10
)

var metricsToGet = []string{
	"CPUUtilization",
}

// CWGatherer is the interface that implements the methods required to gather cloudwath data
type CWGatherer interface {
	GetClusterContainerInstancesMetrics(instance *types.ECSContainerInstance) (*types.InstanceMetrics, error)
}

// CWClient is a wrapper for AWS ecs client that implements helpers to get ECS clusters metrics
type CWClient struct {
	client        cloudwatchiface.CloudWatchAPI
	apiMaxResults int64
}

// NewCWClient create a Cloudwatch API client
func NewCWClient(awsRegion string) (*CWClient, error) {
	// Create AWS session
	s := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))

	if s == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	return &CWClient{
		client:        cloudwatch.New(s),
		apiMaxResults: 100,
	}, nil
}

// GetClusterContainerInstancesMetrics return metric for an instance
func (cw *CWClient) GetClusterContainerInstancesMetrics(instance *types.ECSContainerInstance) (*types.InstanceMetrics, error) {
	metrics := &types.InstanceMetrics{}

	for _, m := range metricsToGet {
		result, err := cw.getInstanceMertic(instance.InstanceID, m)
		if err != nil {
			return nil, err
		}
		v := reflect.ValueOf(metrics).Elem().FieldByName(m)

		if v.IsValid() {
			v.SetFloat(result)
		}
	}

	return metrics, nil
}

func (cw *CWClient) getInstanceMertic(instanceID string, metricName string) (float64, error) {
	var result float64

	params := generateStatInput(instanceID, metricName)

	log.Debugf("Getting metric  '%s'  for : %s", metricName, instanceID)
	resp, err := cw.client.GetMetricStatistics(params)

	if err != nil {
		return result, err
	}

	datapointLen := len(resp.Datapoints)
	if datapointLen > 0 {
		result = *resp.Datapoints[datapointLen-1].Maximum
	} else {
		result = 0
	}

	return result, nil
}

func generateStatInput(instanceID string, metricName string) *cloudwatch.GetMetricStatisticsInput {
	period := -20 * time.Minute
	now := time.Now()

	return &cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(now.Add(period)), // Required
		EndTime:    aws.Time(now),             // Required
		MetricName: aws.String(metricName),    // Required
		Namespace:  aws.String("AWS/EC2"),     // Required
		Period:     aws.Int64(60),             // Required
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("InstanceId"), // Required
				Value: aws.String(instanceID),   // Required
			},
		},
		Statistics: []*string{
			aws.String("Maximum"),
		},
	}
}
