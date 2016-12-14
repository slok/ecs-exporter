package collector

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	awsMock "github.com/slok/ecs-exporter/mock/aws"
	"github.com/slok/ecs-exporter/mock/aws/sdk"
	"github.com/slok/ecs-exporter/types"
)

func TestGetClusters(t *testing.T) {
	tests := []struct {
		clusters          []*types.ECSCluster
		wantErrorList     bool
		wantErrorDescribe bool
		expectError       bool
	}{
		{
			[]*types.ECSCluster{
				&types.ECSCluster{ID: "c1", Name: "cluster1"},
				&types.ECSCluster{ID: "c2", Name: "cluster2"},
				&types.ECSCluster{ID: "c3", Name: "cluster3"},
				&types.ECSCluster{ID: "c4", Name: "cluster4"},
			},
			false, false, false,
		},
		{
			[]*types.ECSCluster{
				&types.ECSCluster{ID: "c1", Name: "cluster1"},
			},
			false, false, false,
		},
		{
			[]*types.ECSCluster{},
			false, false, false,
		},
		{
			[]*types.ECSCluster{},
			true, false, true,
		},
		{
			[]*types.ECSCluster{},
			false, true, true,
		},
	}

	for _, test := range tests {
		cIDs := []string{}

		for _, c := range test.clusters {
			cIDs = append(cIDs, c.ID)
		}

		// Mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSListClusters(t, mockECS, test.wantErrorList, cIDs...)
		awsMock.MockECSDescribeClusters(t, mockECS, test.wantErrorDescribe, test.clusters...)

		e := &ECSClient{
			client: mockECS,
		}

		cs, err := e.GetClusters()
		if !test.expectError {
			if err != nil {
				t.Errorf("\n- %v\n-  Shouldn't return an error, it did: %v", test, err)
			}

			if len(cs) != len(test.clusters) {
				t.Errorf("\n- %v\n-  Length in returned clusters differ, want: %d; got: %d", test, len(test.clusters), len(cs))
			}

			for i, got := range cs {
				want := test.clusters[i]
				if !reflect.DeepEqual(want, got) {
					t.Errorf("\n- %v\n-  Received cluster from API is wrong, want: %v; got: %v", test, want, got)
				}
			}

		} else {
			if err == nil {
				t.Errorf("\n- %v\n-  Should return an error, it didn't", test)
			}
		}

	}
}

func TestGetClusterServices(t *testing.T) {
	tests := []struct {
		services          []*types.ECSService
		wantErrorList     bool
		wantErrorDescribe bool
		expectError       bool
	}{
		{
			[]*types.ECSService{
				&types.ECSService{ID: "s1", Name: "service1", PendingT: 1, RunningT: 9, DesiredT: 10},
				&types.ECSService{ID: "s2", Name: "service2", PendingT: 5, RunningT: 5, DesiredT: 10},
				&types.ECSService{ID: "s3", Name: "service3", PendingT: 7, RunningT: 3, DesiredT: 10},
			},
			false, false, false,
		},
		{
			[]*types.ECSService{
				&types.ECSService{ID: "s1", Name: "service1", PendingT: 1, RunningT: 9, DesiredT: 10},
			},
			true, false, true,
		},
		{
			[]*types.ECSService{
				&types.ECSService{ID: "s1", Name: "service1", PendingT: 1, RunningT: 9, DesiredT: 10},
			},
			true, true, true,
		},
		{
			[]*types.ECSService{},
			false, false, false,
		},
	}

	for _, test := range tests {
		sIDs := []string{}

		for _, s := range test.services {
			sIDs = append(sIDs, s.ID)
		}

		// Mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSListServices(t, mockECS, test.wantErrorList, sIDs...)
		awsMock.MockECSDescribeServices(t, mockECS, test.wantErrorDescribe, test.services...)

		e := &ECSClient{
			client: mockECS,
		}

		services, err := e.GetClusterServices(&types.ECSCluster{ID: "t1", Name: "test1"})

		if !test.expectError {
			if err != nil {
				t.Errorf("\n- %v\n-  Shouldn't return an error, it did: %v", test, err)
			}

			if len(services) != len(test.services) {
				t.Errorf("\n- %v\n-  Length in returned services differ, want: %d; got: %d", test, len(test.services), len(services))
			}

			for i, got := range services {
				want := test.services[i]
				if !reflect.DeepEqual(want, got) {
					t.Errorf("\n- %v\n-  Received service from API is wrong, want: %v; got: %v", test, want, got)
				}
			}

		} else {
			if err == nil {
				t.Errorf("\n- %v\n-  Should return an error, it didn't", test)
			}
		}

	}
}

func TestGetClusterContainerInstances(t *testing.T) {
	tests := []struct {
		cis               []*types.ECSContainerInstance
		wantErrorList     bool
		wantErrorDescribe bool
		expectError       bool
	}{
		{
			[]*types.ECSContainerInstance{},
			false, false, false,
		},
		{
			[]*types.ECSContainerInstance{
				&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0},
				&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: true, Active: false, PendingT: 5},
				&types.ECSContainerInstance{ID: "ci2", InstanceID: "i-00000000000000002", AgentConn: false, Active: true, PendingT: 0},
			},
			false, false, false,
		},
		{
			[]*types.ECSContainerInstance{
				&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0},
			},
			true, false, true,
		},
		{
			[]*types.ECSContainerInstance{
				&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0},
			},
			false, true, true,
		},
	}

	for _, test := range tests {
		ciIDs := []string{}

		for _, ci := range test.cis {
			ciIDs = append(ciIDs, ci.ID)
		}

		// Mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSListContainerInstances(t, mockECS, test.wantErrorList, ciIDs...)
		awsMock.MockECSDescribeContainerInstances(t, mockECS, test.wantErrorDescribe, test.cis...)

		e := &ECSClient{
			client: mockECS,
		}

		cis, err := e.GetClusterContainerInstances(&types.ECSCluster{ID: "t1", Name: "test1"})

		if !test.expectError {
			if err != nil {
				t.Errorf("\n- %v\n-  Shouldn't return an error, it did: %v", test, err)
			}

			if len(cis) != len(test.cis) {
				t.Errorf("\n- %v\n-  Length in returned container instances differ, want: %d; got: %d", test, len(test.cis), len(cis))
			}

			for i, got := range cis {
				want := test.cis[i]
				if !reflect.DeepEqual(want, got) {
					t.Errorf("\n- %v\n-  Received container instnace from API is wrong, want: %v; got: %v", test, want, got)
				}
			}

		} else {
			if err == nil {
				t.Errorf("\n- %v\n-  Should return an error, it didn't", test)
			}
		}

	}
}
