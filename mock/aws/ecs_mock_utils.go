package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/common/log"

	"github.com/slok/ecs-exporter/mock/aws/sdk"
)

// MockECSListClusters mocks the listing of cluster arns
func MockECSListClusters(t *testing.T, mockMatcher *sdk.MockECSAPI, wantError bool, ids ...string) {
	log.Warnf("Mocking AWS iface: ListClusters")
	var err error
	if wantError {
		err = errors.New("Wrong!")
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
