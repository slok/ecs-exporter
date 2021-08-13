// +build integration

package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/slok/ecs-exporter/types"
)

type ECSMockClient struct {
	cdError  bool                                     // Should error on cluster descriptions
	sdError  bool                                     // Should error on service descriptions
	cidError bool                                     // Should error on container instance descriptions
	sleepFor time.Duration                            // Should sleep before returning?
	sd       map[string][]*types.ECSService           // Cluster service descriptions
	cid      map[string][]*types.ECSContainerInstance // container instance descriptions
}

func (e *ECSMockClient) GetClusters() ([]*types.ECSCluster, error) {
	if e.sleepFor > 0 {
		time.Sleep(e.sleepFor)
	}

	if e.cdError {
		return nil, fmt.Errorf("GetClusters Error: wanted")
	}

	// Group clusters
	cd := []*types.ECSCluster{}
	for k, _ := range e.sd {
		cd = append(cd, &types.ECSCluster{ID: k, Name: k})
	}

	return cd, nil
}

func (e *ECSMockClient) GetClusterServices(cluster *types.ECSCluster) ([]*types.ECSService, error) {
	if e.sleepFor > 0 {
		time.Sleep(e.sleepFor)
	}

	if e.sdError {
		return nil, fmt.Errorf("GetClusterServices Error: wanted")
	}

	// return the correct services
	ss, ok := e.sd[cluster.ID]

	if !ok {
		return nil, fmt.Errorf("GetClusterServices Error: not valid cluster %s", cluster.ID)
	}
	return ss, nil
}

func (e *ECSMockClient) GetClusterContainerInstances(cluster *types.ECSCluster) ([]*types.ECSContainerInstance, error) {
	if e.sleepFor > 0 {
		time.Sleep(e.sleepFor)
	}

	if e.cidError {
		return nil, fmt.Errorf("GetClusterContainerInstances Error: wanted")
	}

	// return the correct container instances
	cis, ok := e.cid[cluster.ID]

	if !ok {
		return nil, fmt.Errorf("GetClusterContainerInstances Error: not valid cluster %s", cluster.ID)
	}

	return cis, nil
}

func TestCollectError(t *testing.T) {

	tests := []struct {
		errorDescribeClusters           bool
		errorDescribeServices           bool
		errorDescribeContainerInstances bool
	}{
		{true, false, false},
		{false, true, false},
		{false, false, true},
	}

	for _, test := range tests {

		e := &ECSMockClient{
			cdError:  test.errorDescribeClusters,
			sdError:  test.errorDescribeServices,
			cidError: test.errorDescribeContainerInstances,
			sd: map[string][]*types.ECSService{ // At least 1 to check the service description call
				"cluster0": []*types.ECSService{
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6},
				},
			},
			cid: map[string][]*types.ECSContainerInstance{ // At least 1 to check the container description call
				"cluster0": []*types.ECSContainerInstance{
					&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0},
				},
			},
		}

		exp, err := New("eu-west-1", "", false, defaultCollectTimeout, defaultMaxConcurrency)
		if err != nil {
			t.Errorf("Creation of exporter shouldn't error: %v", err)
		}
		exp.client = e

		// Register the exporter
		prometheus.MustRegister(exp)

		// Make the request
		req, _ := http.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		prometheus.Handler().ServeHTTP(w, req)

		// Check the result
		if w.Code != http.StatusOK {
			t.Errorf("%+v\n -Metrics endpoing status code is wrong, got: %d; want: %d", test, w.Code, http.StatusOK)
		}

		expectedMs := []string{
			`# HELP ecs_up Was the last query of ecs successful.`,
			`# TYPE ecs_up gauge`,
			`ecs_up{region="eu-west-1"} 0`,
		}
		got := w.Body.String()
		for _, m := range expectedMs {
			if !strings.Contains(got, m) {
				t.Errorf("%+v\n -Expected metric data but missing: %s", test, m)
			}
		}

		// Unregister the exporter
		prometheus.Unregister(exp)
	}
}

func TestCollectOk(t *testing.T) {
	tests := []struct {
		cServices   map[string][]*types.ECSService
		cCInstances map[string][]*types.ECSContainerInstance
		cFilter     string
		disableCIM  bool
		want        []string
		dontWant    []string
	}{
		{
			cServices: map[string][]*types.ECSService{
				"cluster1": {
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6}},
			},
			cCInstances: map[string][]*types.ECSContainerInstance{
				"cluster1": {
					&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 12},
					&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: false, Active: true, PendingT: 7},
					&types.ECSContainerInstance{ID: "ci2", InstanceID: "i-00000000000000002", AgentConn: true, Active: false, PendingT: 24},
					&types.ECSContainerInstance{ID: "ci3", InstanceID: "i-00000000000000003", AgentConn: false, Active: false, PendingT: 50},
				},
			},
			cFilter:    ".*",
			disableCIM: false,
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_clusters{region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster1",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster1",region="eu-west-1"} 4`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 10`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 4`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 6`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 12`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 7`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000002",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000002",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000002",region="eu-west-1"} 24`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000003",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000003",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000003",region="eu-west-1"} 50`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"cluster1": {
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6}},
			},
			cCInstances: map[string][]*types.ECSContainerInstance{
				"cluster1": {
					&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 12},
					&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: false, Active: true, PendingT: 7},
					&types.ECSContainerInstance{ID: "ci2", InstanceID: "i-00000000000000002", AgentConn: true, Active: false, PendingT: 24},
					&types.ECSContainerInstance{ID: "ci3", InstanceID: "i-00000000000000003", AgentConn: false, Active: false, PendingT: 50},
				},
			},
			cFilter:    ".*",
			disableCIM: true,
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_clusters{region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster1",region="eu-west-1"} 1`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 10`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 4`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 6`,
			},
			dontWant: []string{
				`ecs_container_instances{cluster="cluster1",region="eu-west-1"} 4`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 12`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 7`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000002",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000002",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000002",region="eu-west-1"} 24`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000003",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000003",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000003",region="eu-west-1"} 50`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"cluster1": {
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6},
					&types.ECSService{ID: "s2", Name: "service2", DesiredT: 987, RunningT: 67, PendingT: 62},
					&types.ECSService{ID: "s3", Name: "service3", DesiredT: 43, RunningT: 20, PendingT: 0},
				},
				"cluster2": {
					&types.ECSService{ID: "s4", Name: "service4", DesiredT: 11, RunningT: 11, PendingT: 11},
				},
			},
			cCInstances: map[string][]*types.ECSContainerInstance{
				"cluster1": {
					&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 12}},

				"cluster2": {
					&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: false, Active: true, PendingT: 7},
					&types.ECSContainerInstance{ID: "ci2", InstanceID: "i-00000000000000002", AgentConn: true, Active: false, PendingT: 24},
					&types.ECSContainerInstance{ID: "ci3", InstanceID: "i-00000000000000003", AgentConn: false, Active: false, PendingT: 50},
				},
			},
			cFilter:    ".*",
			disableCIM: false,
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_clusters{region="eu-west-1"} 2`,
				`ecs_services{cluster="cluster1",region="eu-west-1"} 3`,
				`ecs_services{cluster="cluster2",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster1",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster2",region="eu-west-1"} 3`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 10`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 4`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 6`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service2"} 987`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service2"} 67`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service2"} 62`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service3"} 43`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service3"} 20`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service3"} 0`,

				`ecs_service_desired_tasks{cluster="cluster2",region="eu-west-1",service="service4"} 11`,
				`ecs_service_running_tasks{cluster="cluster2",region="eu-west-1",service="service4"} 11`,
				`ecs_service_pending_tasks{cluster="cluster2",region="eu-west-1",service="service4"} 11`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 12`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000001",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000001",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000001",region="eu-west-1"} 7`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000002",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000002",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000002",region="eu-west-1"} 24`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000003",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000003",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000003",region="eu-west-1"} 50`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"cluster0": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 3, RunningT: 2, PendingT: 1}},
				"cluster1": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 10, RunningT: 5, PendingT: 5}},
				"cluster2": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 15, RunningT: 7, PendingT: 8}},
				"cluster3": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 30, RunningT: 15, PendingT: 15}},
				"cluster4": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 100, RunningT: 10, PendingT: 90}},
				"cluster5": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 75, RunningT: 50, PendingT: 25}},
			},
			cCInstances: map[string][]*types.ECSContainerInstance{
				"cluster0": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0}},
				"cluster1": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: false, Active: true, PendingT: 10}},
				"cluster2": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: false, PendingT: 20}},
				"cluster3": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: false, Active: false, PendingT: 30}},
				"cluster4": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 40}},
				"cluster5": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: false, Active: true, PendingT: 50}},
			},
			cFilter:    ".*",
			disableCIM: false,
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_clusters{region="eu-west-1"} 6`,
				`ecs_services{cluster="cluster0",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster1",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster2",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster3",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster4",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster5",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster0",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster1",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster2",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster3",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster4",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster5",region="eu-west-1"} 1`,

				`ecs_service_desired_tasks{cluster="cluster0",region="eu-west-1",service="service0"} 3`,
				`ecs_service_running_tasks{cluster="cluster0",region="eu-west-1",service="service0"} 2`,
				`ecs_service_pending_tasks{cluster="cluster0",region="eu-west-1",service="service0"} 1`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service0"} 10`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service0"} 5`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service0"} 5`,

				`ecs_service_desired_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 15`,
				`ecs_service_running_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 7`,
				`ecs_service_pending_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 8`,

				`ecs_service_desired_tasks{cluster="cluster3",region="eu-west-1",service="service0"} 30`,
				`ecs_service_running_tasks{cluster="cluster3",region="eu-west-1",service="service0"} 15`,
				`ecs_service_pending_tasks{cluster="cluster3",region="eu-west-1",service="service0"} 15`,

				`ecs_service_desired_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 100`,
				`ecs_service_running_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 10`,
				`ecs_service_pending_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 90`,

				`ecs_service_desired_tasks{cluster="cluster5",region="eu-west-1",service="service0"} 75`,
				`ecs_service_running_tasks{cluster="cluster5",region="eu-west-1",service="service0"} 50`,
				`ecs_service_pending_tasks{cluster="cluster5",region="eu-west-1",service="service0"} 25`,

				`ecs_container_instance_agent_connected{cluster="cluster0",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster0",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster0",instance="i-00000000000000000",region="eu-west-1"} 0`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 10`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000000",region="eu-west-1"} 20`,

				`ecs_container_instance_agent_connected{cluster="cluster3",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster3",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster3",instance="i-00000000000000000",region="eu-west-1"} 30`,

				`ecs_container_instance_agent_connected{cluster="cluster4",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster4",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster4",instance="i-00000000000000000",region="eu-west-1"} 40`,

				`ecs_container_instance_agent_connected{cluster="cluster5",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster5",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster5",instance="i-00000000000000000",region="eu-west-1"} 50`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"cluster0": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 3, RunningT: 2, PendingT: 1}},
				"cluster1": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 10, RunningT: 5, PendingT: 5}},
				"cluster2": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 15, RunningT: 7, PendingT: 8}},
				"cluster3": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 30, RunningT: 15, PendingT: 15}},
				"cluster4": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 100, RunningT: 10, PendingT: 90}},
				"cluster5": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 75, RunningT: 50, PendingT: 25}},
			},
			cCInstances: map[string][]*types.ECSContainerInstance{
				"cluster0": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0}},
				"cluster1": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: false, Active: true, PendingT: 10}},
				"cluster2": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: false, PendingT: 20}},
				"cluster3": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: false, Active: false, PendingT: 30}},
				"cluster4": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 40}},
				"cluster5": {&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: false, Active: true, PendingT: 50}},
			},
			cFilter:    "cluster[024]",
			disableCIM: false,
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_clusters{region="eu-west-1"} 6`,
				`ecs_services{cluster="cluster0",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster2",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster4",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster0",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster2",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster4",region="eu-west-1"} 1`,

				`ecs_service_desired_tasks{cluster="cluster0",region="eu-west-1",service="service0"} 3`,
				`ecs_service_running_tasks{cluster="cluster0",region="eu-west-1",service="service0"} 2`,
				`ecs_service_pending_tasks{cluster="cluster0",region="eu-west-1",service="service0"} 1`,

				`ecs_service_desired_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 15`,
				`ecs_service_running_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 7`,
				`ecs_service_pending_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 8`,

				`ecs_service_desired_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 100`,
				`ecs_service_running_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 10`,
				`ecs_service_pending_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 90`,

				`ecs_container_instance_agent_connected{cluster="cluster0",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster0",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster0",instance="i-00000000000000000",region="eu-west-1"} 0`,

				`ecs_service_desired_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 15`,
				`ecs_service_running_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 7`,
				`ecs_service_pending_tasks{cluster="cluster2",region="eu-west-1",service="service0"} 8`,

				`ecs_service_desired_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 100`,
				`ecs_service_running_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 10`,
				`ecs_service_pending_tasks{cluster="cluster4",region="eu-west-1",service="service0"} 90`,
			},
			dontWant: []string{
				`ecs_services{cluster="cluster1",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster3",region="eu-west-1"} 1`,
				`ecs_services{cluster="cluster5",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster1",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster3",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster5",region="eu-west-1"} 1`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service0"} 10`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service0"} 5`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service0"} 5`,

				`ecs_service_desired_tasks{cluster="cluster3",region="eu-west-1",service="service0"} 30`,
				`ecs_service_running_tasks{cluster="cluster3",region="eu-west-1",service="service0"} 15`,
				`ecs_service_pending_tasks{cluster="cluster3",region="eu-west-1",service="service0"} 15`,

				`ecs_service_desired_tasks{cluster="cluster5",region="eu-west-1",service="service0"} 75`,
				`ecs_service_running_tasks{cluster="cluster5",region="eu-west-1",service="service0"} 50`,
				`ecs_service_pending_tasks{cluster="cluster5",region="eu-west-1",service="service0"} 25`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 10`,

				`ecs_container_instance_agent_connected{cluster="cluster3",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster3",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster3",instance="i-00000000000000000",region="eu-west-1"} 30`,

				`ecs_container_instance_agent_connected{cluster="cluster5",instance="i-00000000000000000",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster5",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster5",instance="i-00000000000000000",region="eu-west-1"} 50`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"cluster1": {
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6},
					&types.ECSService{ID: "s2", Name: "service2", DesiredT: 987, RunningT: 67, PendingT: 62},
					&types.ECSService{ID: "s3", Name: "service3", DesiredT: 43, RunningT: 20, PendingT: 0},
					&types.ECSService{ID: "s4", Name: "service4", DesiredT: 88, RunningT: 77, PendingT: 11},
					&types.ECSService{ID: "s5", Name: "service5", DesiredT: 3, RunningT: 2, PendingT: 1},
				},

				"cluster2": {
					&types.ECSService{ID: "s98", Name: "service98", DesiredT: 100, RunningT: 50, PendingT: 23},
				},

				"cluster3": {
					&types.ECSService{ID: "s1000", Name: "service1000", DesiredT: 1000, RunningT: 500, PendingT: 500},
					&types.ECSService{ID: "s2000", Name: "service2000", DesiredT: 2000, RunningT: 1997, PendingT: 3},
					&types.ECSService{ID: "s3000", Name: "service3000", DesiredT: 3000, RunningT: 2000, PendingT: 1000},
				},
			},
			cCInstances: map[string][]*types.ECSContainerInstance{
				"cluster1": {
					&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 0},
					&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: true, Active: false, PendingT: 99},
				},

				"cluster2": {
					&types.ECSContainerInstance{ID: "ci80", InstanceID: "i-00000000000000080", AgentConn: true, Active: false, PendingT: 14},
					&types.ECSContainerInstance{ID: "ci81", InstanceID: "i-00000000000000081", AgentConn: false, Active: false, PendingT: 67},
					&types.ECSContainerInstance{ID: "ci82", InstanceID: "i-00000000000000082", AgentConn: true, Active: false, PendingT: 89},
					&types.ECSContainerInstance{ID: "ci83", InstanceID: "i-00000000000000083", AgentConn: true, Active: true, PendingT: 2},
				},

				"cluster3": {
					&types.ECSContainerInstance{ID: "ci1234", InstanceID: "i-00000000000001234", AgentConn: false, Active: true, PendingT: 98},
					&types.ECSContainerInstance{ID: "ci5678", InstanceID: "i-00000000000005678", AgentConn: true, Active: true, PendingT: 63},
					&types.ECSContainerInstance{ID: "ci9876", InstanceID: "i-00000000000009876", AgentConn: true, Active: false, PendingT: 13},
				},
			},
			cFilter:    "^cluster[^2]$",
			disableCIM: false,
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_clusters{region="eu-west-1"} 3`,
				`ecs_services{cluster="cluster1",region="eu-west-1"} 5`,
				`ecs_services{cluster="cluster3",region="eu-west-1"} 3`,
				`ecs_container_instances{cluster="cluster1",region="eu-west-1"} 2`,
				`ecs_container_instances{cluster="cluster3",region="eu-west-1"} 3`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 10`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 4`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service1"} 6`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service2"} 987`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service2"} 67`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service2"} 62`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service3"} 43`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service3"} 20`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service3"} 0`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service4"} 88`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service4"} 77`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service4"} 11`,

				`ecs_service_desired_tasks{cluster="cluster1",region="eu-west-1",service="service5"} 3`,
				`ecs_service_running_tasks{cluster="cluster1",region="eu-west-1",service="service5"} 2`,
				`ecs_service_pending_tasks{cluster="cluster1",region="eu-west-1",service="service5"} 1`,

				`ecs_service_desired_tasks{cluster="cluster3",region="eu-west-1",service="service1000"} 1000`,
				`ecs_service_running_tasks{cluster="cluster3",region="eu-west-1",service="service1000"} 500`,
				`ecs_service_pending_tasks{cluster="cluster3",region="eu-west-1",service="service1000"} 500`,

				`ecs_service_desired_tasks{cluster="cluster3",region="eu-west-1",service="service2000"} 2000`,
				`ecs_service_running_tasks{cluster="cluster3",region="eu-west-1",service="service2000"} 1997`,
				`ecs_service_pending_tasks{cluster="cluster3",region="eu-west-1",service="service2000"} 3`,

				`ecs_service_desired_tasks{cluster="cluster3",region="eu-west-1",service="service3000"} 3000`,
				`ecs_service_running_tasks{cluster="cluster3",region="eu-west-1",service="service3000"} 2000`,
				`ecs_service_pending_tasks{cluster="cluster3",region="eu-west-1",service="service3000"} 1000`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000000",region="eu-west-1"} 0`,

				`ecs_container_instance_agent_connected{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster1",instance="i-00000000000000001",region="eu-west-1"} 99`,

				`ecs_container_instance_agent_connected{cluster="cluster3",instance="i-00000000000001234",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster3",instance="i-00000000000001234",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster3",instance="i-00000000000001234",region="eu-west-1"} 98`,

				`ecs_container_instance_agent_connected{cluster="cluster3",instance="i-00000000000005678",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster3",instance="i-00000000000005678",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster3",instance="i-00000000000005678",region="eu-west-1"} 63`,
			},
			dontWant: []string{
				`ecs_services{cluster="cluster2",region="eu-west-1"} 1`,
				`ecs_container_instances{cluster="cluster2",region="eu-west-1"} 4`,

				`ecs_service_desired_tasks{cluster="cluster2",region="eu-west-1",service="service98"} 100`,
				`ecs_service_running_tasks{cluster="cluster2",region="eu-west-1",service="service98"} 50`,
				`ecs_service_pending_tasks{cluster="cluster2",region="eu-west-1",service="service98"} 23`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000080",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000080",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000080",region="eu-west-1"} 13`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000081",region="eu-west-1"} 0`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000081",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000081",region="eu-west-1"} 67`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000082",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000082",region="eu-west-1"} 0`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000082",region="eu-west-1"} 89`,

				`ecs_container_instance_agent_connected{cluster="cluster2",instance="i-00000000000000083",region="eu-west-1"} 1`,
				`ecs_container_instance_active{cluster="cluster2",instance="i-00000000000000083",region="eu-west-1"} 1`,
				`ecs_container_instance_pending_tasks{cluster="cluster2",instance="i-00000000000000083",region="eu-west-1"} 2`,
			},
		},
	}

	for _, test := range tests {

		e := &ECSMockClient{
			sd:  test.cServices,
			cid: test.cCInstances,
		}

		exp, err := New("eu-west-1", test.cFilter, test.disableCIM, defaultCollectTimeout, defaultMaxConcurrency)
		if err != nil {
			t.Errorf("Creation of exporter shouldn't error: %v", err)
		}
		exp.client = e

		// Register the exporter
		prometheus.MustRegister(exp)

		// Make the request
		req, _ := http.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		prometheus.Handler().ServeHTTP(w, req)

		// Check the result
		if w.Code != http.StatusOK {
			t.Errorf("%+v\n -Metrics endpoing status code is wrong, got: %d; want: %d", test, w.Code, http.StatusOK)
		}
		got := w.Body.String()
		for _, m := range test.want {
			if !strings.Contains(got, m) {
				t.Errorf("%+v\n -Expected metric data but missing: %s", test, m)
			}
		}

		for _, m := range test.dontWant {
			if strings.Contains(got, m) {
				t.Errorf("%+v\n -Didn't expected metric data but found: %s", test, m)
			}
		}

		// Unregister the exporter
		prometheus.Unregister(exp)
	}
}

func TestCollectTimeoutNoPanic(t *testing.T) {
	// If fails should panic!
	cServices := map[string][]*types.ECSService{
		"cluster1": {
			&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6}},
	}
	cCInstances := map[string][]*types.ECSContainerInstance{
		"cluster1": {
			&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 12},
			&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: false, Active: true, PendingT: 7},
			&types.ECSContainerInstance{ID: "ci2", InstanceID: "i-00000000000000002", AgentConn: true, Active: false, PendingT: 24},
			&types.ECSContainerInstance{ID: "ci3", InstanceID: "i-00000000000000003", AgentConn: false, Active: false, PendingT: 50},
		},
	}

	e := &ECSMockClient{
		sd:       cServices,
		cid:      cCInstances,
		sleepFor: 10 * time.Millisecond,
	}

	exp, err := New("eu-west-1", ".*", false, defaultCollectTimeout, defaultMaxConcurrency)
	if err != nil {
		t.Errorf("Creation of exporter shouldn't error: %v", err)
	}
	exp.client = e
	exp.timeout = 0

	// Register the exporter
	prometheus.MustRegister(exp)

	// Make the request
	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	prometheus.Handler().ServeHTTP(w, req)

	// Check the result
	if w.Code != http.StatusOK {
		t.Errorf("Metrics endpoing status code is wrong, got: %d; want: %d", w.Code, http.StatusOK)
	}

	// Wait for the panic
	time.Sleep(100 * time.Millisecond)

	// Unregister the exporter
	prometheus.Unregister(exp)
}
