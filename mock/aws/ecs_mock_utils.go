package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"

	"github.com/slok/ecs-exporter/log"
	"github.com/slok/ecs-exporter/mock/aws/sdk"
	"github.com/slok/ecs-exporter/types"
)

// MockECSListClusters mocks the listing of cluster arns
func MockECSListClusters(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, ids ...string) {
	log.Warnf("Mocking AWS iface: ListClusters")
	var err error
	if wantError {
		err = errors.New("ListClusters wrong!")
	}
	cIds := []*string{}
	for _, id := range ids {
		tID := id
		cIds = append(cIds, &tID)
	}
	result := &ecs.ListClustersOutput{
		ClusterArns: cIds,
	}
	mockMatcher.EXPECT().ListClusters(gomock.Any()).Do(func(input interface{}) {
	}).AnyTimes().Return(result, err)
}

// MockECSDescribeClusters mocks the description of service
func MockECSDescribeClusters(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, clusters ...*types.ECSCluster) {
	log.Warnf("Mocking AWS iface: DescribeClusters")
	var err error
	if wantError {
		err = errors.New("DescribeClusters wrong!")
	}
	cs := []*ecs.Cluster{}
	for _, c := range clusters {
		dc := &ecs.Cluster{
			ClusterArn:  aws.String(c.ID),
			ClusterName: aws.String(c.Name),
		}
		cs = append(cs, dc)
	}
	result := &ecs.DescribeClustersOutput{
		Clusters: cs,
	}
	mockMatcher.EXPECT().DescribeClusters(gomock.Any()).Do(func(input interface{}) {
	}).AnyTimes().Return(result, err)
}

// MockECSListServices mocks the listing of service arns
func MockECSListServices(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, ids ...string) {
	log.Warnf("Mocking AWS iface: ListServices")
	var err error
	if wantError {
		err = errors.New("ListServices wrong!")
	}
	cIds := []*string{}
	for _, id := range ids {
		tID := id
		cIds = append(cIds, &tID)
	}
	result := &ecs.ListServicesOutput{
		ServiceArns: cIds,
	}
	mockMatcher.EXPECT().ListServices(gomock.Any()).Do(func(input interface{}) {
		i := input.(*ecs.ListServicesInput)
		if i.Cluster == nil || aws.StringValue(i.Cluster) == "" {
			t.Errorf("Wrong api call, needs cluster ARN")
		}
	}).AnyTimes().Return(result, err)
}

// MockECSDescribeServices mocks the description of service
func MockECSDescribeServices(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, services ...*types.ECSService) {
	log.Warnf("Mocking AWS iface: DescribeServices")
	var err error
	if wantError {
		err = errors.New("DescribeServices wrong!")
	}
	ss := []*ecs.Service{}
	for _, s := range services {
		ds := &ecs.Service{
			ServiceArn:   aws.String(s.ID),
			ServiceName:  aws.String(s.Name),
			PendingCount: aws.Int64(s.PendingT),
			RunningCount: aws.Int64(s.RunningT),
			DesiredCount: aws.Int64(s.DesiredT),
		}
		ss = append(ss, ds)
	}
	result := &ecs.DescribeServicesOutput{
		Services: ss,
	}
	mockMatcher.EXPECT().DescribeServices(gomock.Any()).Do(func(input interface{}) {
	}).AnyTimes().Return(result, err)
}

// MockECSListServicesTimes mocks the listing of service arns, it accepts n arrays that will be mocked in order
// per call (simulating calls listing services per cluster)
func MockECSListServicesTimes(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, cIds ...[]string) {
	log.Warnf("Mocking AWS iface: ListServices")
	var err error
	if wantError {
		err = errors.New("ListServices wrong!")
	}

	calls := []*gomock.Call{}

	// Mock ids per call
	for _, ids := range cIds {

		// Create mock call
		pIds := []*string{}
		for _, id := range ids {
			tID := id
			pIds = append(pIds, &tID)
		}
		result := &ecs.ListServicesOutput{
			ServiceArns: pIds,
		}

		call := mockMatcher.EXPECT().ListServices(gomock.Any()).Do(func(input interface{}) {
			i := input.(*ecs.ListServicesInput)
			if i.Cluster == nil || aws.StringValue(i.Cluster) == "" {
				t.Errorf("Wrong api call, needs cluster ARN")
			}
		}).Return(result, err)

		calls = append(calls, call)
	}
	gomock.InOrder(calls...)
}

// MockECSDescribeServicesTimes mocks the description of service, it accepts n arrays that will be the mocked calls in order
// for each parameters (simulating calls to different services)
func MockECSDescribeServicesTimes(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, cServices ...[]*types.ECSService) {
	log.Warnf("Mocking AWS iface: DescribeServices")
	var err error
	if wantError {
		err = errors.New("DescribeServices wrong!")
	}

	calls := []*gomock.Call{}

	//mock service desciptions per call (per cluster)
	for _, css := range cServices {
		ss := []*ecs.Service{}
		for _, s := range css {
			ds := &ecs.Service{
				ServiceArn:   aws.String(s.ID),
				ServiceName:  aws.String(s.Name),
				PendingCount: aws.Int64(s.PendingT),
				RunningCount: aws.Int64(s.RunningT),
				DesiredCount: aws.Int64(s.DesiredT),
			}
			ss = append(ss, ds)
		}
		result := &ecs.DescribeServicesOutput{
			Services: ss,
		}

		call := mockMatcher.EXPECT().DescribeServices(gomock.Any()).Do(func(input interface{}) {}).Return(result, err)
		calls = append(calls, call)
	}
	gomock.InOrder(calls...)
}
