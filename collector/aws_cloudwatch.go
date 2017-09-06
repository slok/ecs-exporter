package collector

import (
	"fmt"
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

// CWGatherer is the interface that implements the methods required to gather cloudwath data
type CWGatherer interface {
	GetClusterContainerInstancesMetrics(instance *types.ECSContainerInstance) (*types.InstanceMetrics, error)
}

// CWClient is a wrapper for AWS ecs client that implements helpers to get ECS clusters metrics
type CWClient struct {
	client        cloudwatchiface.CloudWatchAPI
	apiMaxResults int64
}

// NewECSClient will return an initialized ECSClient
func NewCWClient(awsRegion string) (*CWClient, error) {
	// Create AWS session
	s := session.New(&aws.Config{Region: aws.String(awsRegion)})
	// s := session.Must(session.NewSession())
	if s == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	return &CWClient{
		client:        cloudwatch.New(s),
		apiMaxResults: 100,
	}, nil
}

func (cw *CWClient) GetClusterContainerInstancesMetrics(instance *types.ECSContainerInstance) (*types.InstanceMetrics, error) {

	cpu, err := cw.getInstanceMertic(instance.InstanceID, "CPUUtilization")

	metrics := &types.InstanceMetrics{
		CPUutilization: cpu,
	}

	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (cw *CWClient) getInstanceMertic(instanceID string, metricName string) (float64, error) {
	period := -20 * time.Minute
	now := time.Now()
	var result float64

	params := &cloudwatch.GetMetricStatisticsInput{
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
