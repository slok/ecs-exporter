// +build integration

package collector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	awsMock "github.com/slok/ecs-exporter/mock/aws"
	"github.com/slok/ecs-exporter/mock/aws/sdk"
	"github.com/slok/ecs-exporter/types"
)

func TestCollectError(t *testing.T) {

	tests := []struct {
		errorListClusters     bool
		errorDescribeClusters bool
		errorListServices     bool
		errorDescribeServices bool
	}{
		{true, false, false, false},
		{false, true, false, false},
		{false, false, true, false},
		{false, false, false, true},
	}

	for _, test := range tests {

		// Mock the AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSListClusters(t, mockECS, test.errorListClusters, "test")
		awsMock.MockECSDescribeClusters(t, mockECS, test.errorDescribeClusters, &types.ECSCluster{ID: "t1", Name: "test1"})
		awsMock.MockECSListServices(t, mockECS, test.errorListServices, "test")
		awsMock.MockECSDescribeServices(t, mockECS, test.errorDescribeServices, &types.ECSService{ID: "t1", Name: "test1"})
		e := &ECSClient{client: mockECS}

		exp, err := New("eu-west-1")
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

func TestCollect(t *testing.T) {
	tests := []struct {
		cServices map[string][]*types.ECSService
		want      []string
	}{
		{
			cServices: map[string][]*types.ECSService{
				"c1uster0": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 3, RunningT: 2, PendingT: 1}},
				"c1uster1": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 10, RunningT: 5, PendingT: 5}},
				"c1uster2": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 15, RunningT: 7, PendingT: 8}},
				"c1uster3": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 30, RunningT: 15, PendingT: 15}},
				"c1uster4": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 100, RunningT: 10, PendingT: 90}},
				"c1uster5": {&types.ECSService{ID: "s0", Name: "service0", DesiredT: 75, RunningT: 50, PendingT: 25}},
			},
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_cluster_total{region="eu-west-1"} 6`,

				`ecs_service_desired_tasks{cluster="c1uster0",region="eu-west-1",service="service0"} 3`,
				`ecs_service_running_tasks{cluster="c1uster0",region="eu-west-1",service="service0"} 2`,
				`ecs_service_pending_tasks{cluster="c1uster0",region="eu-west-1",service="service0"} 1`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service0"} 10`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service0"} 5`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service0"} 5`,

				`ecs_service_desired_tasks{cluster="c1uster2",region="eu-west-1",service="service0"} 15`,
				`ecs_service_running_tasks{cluster="c1uster2",region="eu-west-1",service="service0"} 7`,
				`ecs_service_pending_tasks{cluster="c1uster2",region="eu-west-1",service="service0"} 8`,

				`ecs_service_desired_tasks{cluster="c1uster3",region="eu-west-1",service="service0"} 30`,
				`ecs_service_running_tasks{cluster="c1uster3",region="eu-west-1",service="service0"} 15`,
				`ecs_service_pending_tasks{cluster="c1uster3",region="eu-west-1",service="service0"} 15`,

				`ecs_service_desired_tasks{cluster="c1uster4",region="eu-west-1",service="service0"} 100`,
				`ecs_service_running_tasks{cluster="c1uster4",region="eu-west-1",service="service0"} 10`,
				`ecs_service_pending_tasks{cluster="c1uster4",region="eu-west-1",service="service0"} 90`,

				`ecs_service_desired_tasks{cluster="c1uster5",region="eu-west-1",service="service0"} 75`,
				`ecs_service_running_tasks{cluster="c1uster5",region="eu-west-1",service="service0"} 50`,
				`ecs_service_pending_tasks{cluster="c1uster5",region="eu-west-1",service="service0"} 25`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"c1uster1": {
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6},
					&types.ECSService{ID: "s2", Name: "service2", DesiredT: 987, RunningT: 67, PendingT: 62},
					&types.ECSService{ID: "s3", Name: "service3", DesiredT: 43, RunningT: 20, PendingT: 0},
				},
			},
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_cluster_total{region="eu-west-1"} 1`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service1"} 10`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service1"} 4`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service1"} 6`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service2"} 987`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service2"} 67`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service2"} 62`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service3"} 43`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service3"} 20`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service3"} 0`,
			},
		},
		{
			cServices: map[string][]*types.ECSService{
				"c1uster1": {
					&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, RunningT: 4, PendingT: 6},
					&types.ECSService{ID: "s2", Name: "service2", DesiredT: 987, RunningT: 67, PendingT: 62},
					&types.ECSService{ID: "s3", Name: "service3", DesiredT: 43, RunningT: 20, PendingT: 0},
					&types.ECSService{ID: "s4", Name: "service4", DesiredT: 88, RunningT: 77, PendingT: 11},
					&types.ECSService{ID: "s5", Name: "service5", DesiredT: 3, RunningT: 2, PendingT: 1},
				},

				"c1uster2": {
					&types.ECSService{ID: "s98", Name: "service98", DesiredT: 100, RunningT: 50, PendingT: 23},
				},

				"c1uster3": {
					&types.ECSService{ID: "s1000", Name: "service1000", DesiredT: 1000, RunningT: 500, PendingT: 500},
					&types.ECSService{ID: "s2000", Name: "service2000", DesiredT: 2000, RunningT: 1997, PendingT: 3},
					&types.ECSService{ID: "s3000", Name: "service3000", DesiredT: 3000, RunningT: 2000, PendingT: 1000},
				},
			},
			want: []string{
				`ecs_up{region="eu-west-1"} 1`,
				`ecs_cluster_total{region="eu-west-1"} 3`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service1"} 10`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service1"} 4`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service1"} 6`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service2"} 987`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service2"} 67`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service2"} 62`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service3"} 43`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service3"} 20`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service3"} 0`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service4"} 88`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service4"} 77`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service4"} 11`,

				`ecs_service_desired_tasks{cluster="c1uster1",region="eu-west-1",service="service5"} 3`,
				`ecs_service_running_tasks{cluster="c1uster1",region="eu-west-1",service="service5"} 2`,
				`ecs_service_pending_tasks{cluster="c1uster1",region="eu-west-1",service="service5"} 1`,

				`ecs_service_desired_tasks{cluster="c1uster2",region="eu-west-1",service="service98"} 100`,
				`ecs_service_running_tasks{cluster="c1uster2",region="eu-west-1",service="service98"} 50`,
				`ecs_service_pending_tasks{cluster="c1uster2",region="eu-west-1",service="service98"} 23`,

				`ecs_service_desired_tasks{cluster="c1uster3",region="eu-west-1",service="service1000"} 1000`,
				`ecs_service_running_tasks{cluster="c1uster3",region="eu-west-1",service="service1000"} 500`,
				`ecs_service_pending_tasks{cluster="c1uster3",region="eu-west-1",service="service1000"} 500`,

				`ecs_service_desired_tasks{cluster="c1uster3",region="eu-west-1",service="service2000"} 2000`,
				`ecs_service_running_tasks{cluster="c1uster3",region="eu-west-1",service="service2000"} 1997`,
				`ecs_service_pending_tasks{cluster="c1uster3",region="eu-west-1",service="service2000"} 3`,

				`ecs_service_desired_tasks{cluster="c1uster3",region="eu-west-1",service="service3000"} 3000`,
				`ecs_service_running_tasks{cluster="c1uster3",region="eu-west-1",service="service3000"} 2000`,
				`ecs_service_pending_tasks{cluster="c1uster3",region="eu-west-1",service="service3000"} 1000`,
			},
		},
	}

	for _, test := range tests {

		csl := []string{}
		csd := []*types.ECSCluster{}
		servsl := [][]string{}
		servsd := [][]*types.ECSService{}

		for c, css := range test.cServices {
			// Cluster mocks
			csl = append(csl, c)
			csd = append(csd, &types.ECSCluster{ID: c, Name: c})

			// Services mocks
			sl := make([]string, len(css))
			sd := make([]*types.ECSService, len(css))

			for i, s := range css {
				sl[i] = s.ID
				sd[i] = s
			}
			servsl = append(servsl, sl)
			servsd = append(servsd, sd)

		}

		// Mock the AWS API
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSListClusters(t, mockECS, false, csl...)
		awsMock.MockECSDescribeClusters(t, mockECS, false, csd...)
		awsMock.MockECSListServicesTimes(t, mockECS, false, servsl...)
		awsMock.MockECSDescribeServicesTimes(t, mockECS, false, servsd...)
		e := &ECSClient{client: mockECS}

		exp, err := New("eu-west-1")
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

		// Unregister the exporter
		prometheus.Unregister(exp)
	}
}
