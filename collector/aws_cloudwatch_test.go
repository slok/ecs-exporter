package collector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/golang/mock/gomock"
	"github.com/slok/ecs-exporter/mock/aws/sdk"
	"github.com/slok/ecs-exporter/types"
)

const (
	instanceID = "i-000000002"
	metricName = "CPUUtilization"
)

func createMockCW(t *testing.T) *CWClient {
	// Mock
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cwAPI := sdk.NewMockCloudWatchAPI(mockCtrl)

	// build CW answare
	result := &cloudwatch.GetMetricStatisticsOutput{}
	result.Label = aws.String("FakeMetric")
	result.Datapoints = []*cloudwatch.Datapoint{&cloudwatch.Datapoint{
		Maximum: aws.Float64(95),
	},
	}

	// Mock the answare
	cwAPI.EXPECT().GetMetricStatistics(gomock.Any()).Return(result, nil).AnyTimes()
	return &CWClient{client: cwAPI}
}

func TestCWClient_getInstanceMertic(t *testing.T) {
	tests := []struct {
		name        string
		metricValue []*cloudwatch.Datapoint
		want        float64
		wantErr     bool
	}{
		{
			name: "MT01",
			metricValue: []*cloudwatch.Datapoint{&cloudwatch.Datapoint{
				Maximum: aws.Float64(99.6),
			},
			},
			want:    99.6,
			wantErr: false,
		},
		{
			name:        "MT02",
			metricValue: []*cloudwatch.Datapoint{},
			want:        0,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			cwAPI := sdk.NewMockCloudWatchAPI(mockCtrl)

			// build CW answare
			result := &cloudwatch.GetMetricStatisticsOutput{}
			result.Label = aws.String("FakeMetric")
			result.Datapoints = tt.metricValue

			// Mock the answare
			cwAPI.EXPECT().GetMetricStatistics(gomock.Any()).Return(result, nil)

			// Create a fake client
			cw := &CWClient{
				client:        cwAPI,
				apiMaxResults: 100,
			}

			got, err := cw.getInstanceMertic(instanceID, metricName)

			if (err != nil) != tt.wantErr {
				t.Errorf("CWClient.getInstanceMertic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CWClient.getInstanceMertic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCWClient_GetClusterContainerInstancesMetrics(t *testing.T) {

	tests := []struct {
		name    string
		want    *types.InstanceMetrics
		wantErr bool
	}{
		{
			name:    "CIM01",
			want:    &types.InstanceMetrics{CPUUtilization: 96},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			cwAPI := sdk.NewMockCloudWatchAPI(mockCtrl)

			// build CW answare
			result := &cloudwatch.GetMetricStatisticsOutput{}
			result.Label = aws.String("FakeMetric")
			result.Datapoints = []*cloudwatch.Datapoint{&cloudwatch.Datapoint{
				Maximum: aws.Float64(tt.want.CPUUtilization),
			},
			}

			// Mock the answare
			cwAPI.EXPECT().GetMetricStatistics(gomock.Any()).Return(result, nil)

			// Create a fake client
			cw := &CWClient{
				client:        cwAPI,
				apiMaxResults: 100,
			}

			got, err := cw.GetClusterContainerInstancesMetrics(&types.ECSContainerInstance{InstanceID: instanceID})
			if (err != nil) != tt.wantErr {
				t.Errorf("CWClient.GetClusterContainerInstancesMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CWClient.GetClusterContainerInstancesMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
