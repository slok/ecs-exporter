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
			Deployments:  make([]*ecs.Deployment, s.Deployments),
		}
		ss = append(ss, ds)
	}
	result := &ecs.DescribeServicesOutput{
		Services: ss,
	}
	mockMatcher.EXPECT().DescribeServices(gomock.Any()).Do(func(input interface{}) {
		i := input.(*ecs.DescribeServicesInput)
		if i.Cluster == nil || aws.StringValue(i.Cluster) == "" {
			t.Errorf("Wrong api call, needs cluster ARN")
		}
		if len(i.Services) == 0 {
			t.Errorf("Wrong api call, needs at least 1 service ARN")
		}
	}).AnyTimes().Return(result, err)
}

// MockECSListContainerInstances mocks the listing of container instance arns
func MockECSListContainerInstances(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, ids ...string) {
	log.Warnf("Mocking AWS iface: ListContainerInstances")
	var err error
	if wantError {
		err = errors.New("ListContainerInstances wrong!")
	}
	ciIds := []*string{}
	for _, id := range ids {
		tID := id
		ciIds = append(ciIds, &tID)
	}
	result := &ecs.ListContainerInstancesOutput{
		ContainerInstanceArns: ciIds,
	}
	mockMatcher.EXPECT().ListContainerInstances(gomock.Any()).Do(func(input interface{}) {
		i := input.(*ecs.ListContainerInstancesInput)
		if i.Cluster == nil || aws.StringValue(i.Cluster) == "" {
			t.Errorf("Wrong api call, needs cluster ARN")
		}
	}).AnyTimes().Return(result, err)
}

// MockECSDescribeContainerInstances mocks the description of container instances
func MockECSDescribeContainerInstances(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, cInstances ...*types.ECSContainerInstance) {
	log.Warnf("Mocking AWS iface: DescribeContainerInstances")
	var err error
	if wantError {
		err = errors.New("DescribeContainerInstances wrong!")
	}
	cis := []*ecs.ContainerInstance{}
	for _, c := range cInstances {

		status := types.ContainerInstanceStatusInactive
		if c.Active {
			status = types.ContainerInstanceStatusActive
		}

		dc := &ecs.ContainerInstance{
			ContainerInstanceArn: aws.String(c.ID),
			Ec2InstanceId:        aws.String(c.InstanceID),
			AgentConnected:       aws.Bool(c.AgentConn),
			PendingTasksCount:    aws.Int64(c.PendingT),
			Status:               aws.String(status),
		}
		cis = append(cis, dc)
	}
	result := &ecs.DescribeContainerInstancesOutput{
		ContainerInstances: cis,
	}
	mockMatcher.EXPECT().DescribeContainerInstances(gomock.Any()).Do(func(input interface{}) {
		i := input.(*ecs.DescribeContainerInstancesInput)
		if i.Cluster == nil || aws.StringValue(i.Cluster) == "" {
			t.Errorf("Wrong api call, needs cluster ARN")
		}
		if len(i.ContainerInstances) == 0 {
			t.Errorf("Wrong api call, needs at least 1 container instance ARN")
		}
	}).AnyTimes().Return(result, err)
}
