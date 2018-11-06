package collector

import (
	"context"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/slok/ecs-exporter/types"
)

type metricResult struct {
	value  float64
	labels map[string]string
}

func labels2Map(labels []*dto.LabelPair) map[string]string {
	res := map[string]string{}
	for _, l := range labels {
		res[l.GetName()] = l.GetValue()
	}
	return res
}

func readGauge(g prometheus.Metric) metricResult {
	m := &dto.Metric{}
	g.Write(m)

	return metricResult{
		value:  m.GetGauge().GetValue(),
		labels: labels2Map(m.GetLabel()),
	}
}

func TestCollectClusterMetrics(t *testing.T) {
	region := "eu-west-1"
	exp, err := New(region, "", 1, false)
	if err != nil {
		t.Errorf("Creation of exporter shoudnt error: %v", err)
	}

	ch := make(chan prometheus.Metric)
	testCs := []*types.ECSCluster{}
	for i := 0; i < 10; i++ {
		c := &types.ECSCluster{
			Name: fmt.Sprintf("cluster%d", i),
			ID:   fmt.Sprintf("c%d", i),
		}
		testCs = append(testCs, c)
	}

	// Collect mocked metrics
	go exp.collectClusterMetrics(context.TODO(), ch, testCs)

	m := (<-ch).(prometheus.Metric)
	m2 := readGauge(m)

	expectedV := 10.0
	// Check colected metrics are ok
	if m2.value != expectedV {
		t.Errorf("expected %f ecs_clusters, got %f", expectedV, m2.value)
	}

	if m2.labels["region"] != region {
		t.Errorf("expected %s region, got %s", region, m2.labels["region"])
	}

	expected := `Desc{fqName: "ecs_clusters", help: "The total number of clusters", constLabels: {}, variableLabels: [region]}`
	if expected != m.Desc().String() {
		t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
	}
}

func TestCollectClusterServiceMetrics(t *testing.T) {
	region := "eu-west-1"
	exp, err := New(region, "", 1, false)
	if err != nil {
		t.Errorf("Creation of exporter shouldnt error: %v", err)
	}

	ch := make(chan prometheus.Metric)

	testC := &types.ECSCluster{ID: "c1", Name: "cluster1"}
	testSs := []*types.ECSService{
		&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, PendingT: 5, RunningT: 5},
		&types.ECSService{ID: "s2", Name: "service2", DesiredT: 15, PendingT: 5, RunningT: 10},
		&types.ECSService{ID: "s3", Name: "service3", DesiredT: 30, PendingT: 27, RunningT: 0},
		&types.ECSService{ID: "s4", Name: "service4", DesiredT: 51, PendingT: 50, RunningT: 1},
		&types.ECSService{ID: "s5", Name: "service5", DesiredT: 109, PendingT: 99, RunningT: 2},
		&types.ECSService{ID: "s6", Name: "service6", DesiredT: 6431, PendingT: 5000, RunningT: 107},
	}
	// Collect mocked metrics
	go func() {
		exp.collectClusterServicesMetrics(context.TODO(), ch, testC, testSs)
		close(ch)
	}()

	// Check 1st received metric of services as group
	m := (<-ch).(prometheus.Metric)
	m2 := readGauge(m)
	want := float64(len(testSs))
	if m2.value != want {
		t.Errorf("expected %f ecs_services, got %f", want, m2.value)
	}
	expected := `Desc{fqName: "ecs_services", help: "The total number of services", constLabels: {}, variableLabels: [region cluster]}`
	if expected != m.Desc().String() {
		t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
	}

	for _, wantS := range testSs {
		// Check 1st received metric  per service (desired)
		m := (<-ch).(prometheus.Metric)
		m2 := readGauge(m)
		want := float64(wantS.DesiredT)
		if m2.value != want {
			t.Errorf("expected %f service_desired_tasks, got %f", want, m2.value)
		}
		expected := `Desc{fqName: "ecs_service_desired_tasks", help: "The desired number of instantiations of the task definition to keep running regarding a service", constLabels: {}, variableLabels: [region cluster service]}`
		if expected != m.Desc().String() {
			t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
		}

		// Check 1st received metric  per service (pending)
		m = (<-ch).(prometheus.Metric)
		m2 = readGauge(m)
		want = float64(wantS.PendingT)
		if m2.value != want {
			t.Errorf("expected %f service_pending_tasks, got %f", want, m2.value)
		}
		expected = `Desc{fqName: "ecs_service_pending_tasks", help: "The number of tasks in the cluster that are in the PENDING state regarding a service", constLabels: {}, variableLabels: [region cluster service]}`
		if expected != m.Desc().String() {
			t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
		}

		// Check 1st received metric  per service (running)
		m = (<-ch).(prometheus.Metric)
		m2 = readGauge(m)
		want = float64(wantS.RunningT)
		if m2.value != want {
			t.Errorf("expected %f service_running_tasks, got %f", want, m2.value)
		}
		expected = `Desc{fqName: "ecs_service_running_tasks", help: "The number of tasks in the cluster that are in the RUNNING state regarding a service", constLabels: {}, variableLabels: [region cluster service]}`
		if expected != m.Desc().String() {
			t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
		}
	}
}

func TestCollectClusterContainerInstanceMetrics(t *testing.T) {
	region := "eu-west-1"
	exp, err := New(region, "", 1, false)
	if err != nil {
		t.Errorf("Creation of exporter shouldnt error: %v", err)
	}

	ch := make(chan prometheus.Metric)

	testC := &types.ECSCluster{ID: "c1", Name: "cluster1"}
	testCIs := []*types.ECSContainerInstance{
		&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 12},
		&types.ECSContainerInstance{ID: "ci1", InstanceID: "i-00000000000000001", AgentConn: false, Active: true, PendingT: 7},
		&types.ECSContainerInstance{ID: "ci2", InstanceID: "i-00000000000000002", AgentConn: true, Active: false, PendingT: 24},
		&types.ECSContainerInstance{ID: "ci3", InstanceID: "i-00000000000000003", AgentConn: false, Active: false, PendingT: 197},
	}
	// Collect mocked metrics
	go func() {
		exp.collectClusterContainerInstancesMetrics(context.TODO(), ch, testC, testCIs)
		close(ch)
	}()

	// Check 1st received metric of container instances as group
	m := (<-ch).(prometheus.Metric)
	m2 := readGauge(m)
	want := float64(len(testCIs))
	if m2.value != want {
		t.Errorf("expected %f container_instances, got %f", want, m2.value)
	}
	expected := `Desc{fqName: "ecs_container_instances", help: "The total number of container instances", constLabels: {}, variableLabels: [region cluster]}`
	if expected != m.Desc().String() {
		t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
	}

	for _, wantCi := range testCIs {
		// Check 1st received metric per container instance (agent connected)
		m := (<-ch).(prometheus.Metric)
		m2 := readGauge(m)
		var want float64
		if wantCi.AgentConn {
			want = 1
		}
		if m2.value != want {
			t.Errorf("expected %f container_instance_agent_connected, got %f", want, m2.value)
		}
		expected := `Desc{fqName: "ecs_container_instance_agent_connected", help: "The connected state of the container instance agent", constLabels: {}, variableLabels: [region cluster instance]}`
		if expected != m.Desc().String() {
			t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
		}

		// Check 1st received metric per container instance (status active)
		m = (<-ch).(prometheus.Metric)
		m2 = readGauge(m)
		want = 0
		if wantCi.Active {
			want = 1
		}
		if m2.value != want {
			t.Errorf("expected %f container_instance_active, got %f", want, m2.value)
		}
		expected = `Desc{fqName: "ecs_container_instance_active", help: "The status of the container instance in ACTIVE state, indicates that the container instance can accept tasks.", constLabels: {}, variableLabels: [region cluster instance]}`
		if expected != m.Desc().String() {
			t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
		}

		// Check 1st received metric  per service (running)
		m = (<-ch).(prometheus.Metric)
		m2 = readGauge(m)
		want = float64(wantCi.PendingT)
		if m2.value != want {
			t.Errorf("expected %f container_instance_pending_tasks, got %f", want, m2.value)
		}
		expected = `Desc{fqName: "ecs_container_instance_pending_tasks", help: "The number of tasks on the container instance that are in the PENDING status.", constLabels: {}, variableLabels: [region cluster instance]}`
		if expected != m.Desc().String() {
			t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
		}
	}
}

func TestValidClusters(t *testing.T) {
	tests := []struct {
		filter    string
		correct   []*types.ECSCluster
		incorrect []*types.ECSCluster
	}{
		{
			filter: ".*",
			correct: []*types.ECSCluster{
				&types.ECSCluster{Name: "cluster-good-1"},
				&types.ECSCluster{Name: "mycluster-good-2"},
				&types.ECSCluster{Name: "cluster33"},
				&types.ECSCluster{Name: "clustergood-0"},
				&types.ECSCluster{Name: "cluster.also-good-5"},
			},
			incorrect: []*types.ECSCluster{},
		},
		{
			filter: "cluster[02468]",
			correct: []*types.ECSCluster{
				&types.ECSCluster{Name: "cluster0"},
				&types.ECSCluster{Name: "cluster2"},
				&types.ECSCluster{Name: "cluster4"},
				&types.ECSCluster{Name: "cluster6"},
				&types.ECSCluster{Name: "cluster8"},
			},
			incorrect: []*types.ECSCluster{
				&types.ECSCluster{Name: "cluster1"},
				&types.ECSCluster{Name: "cluster3"},
				&types.ECSCluster{Name: "cluster5"},
				&types.ECSCluster{Name: "cluster7"},
				&types.ECSCluster{Name: "cluster9"},
			},
		},
		{
			filter: "prod-cluster-.*",
			correct: []*types.ECSCluster{
				&types.ECSCluster{Name: "prod-cluster-big"},
				&types.ECSCluster{Name: "prod-cluster-small"},
				&types.ECSCluster{Name: "prod-cluster-main"},
				&types.ECSCluster{Name: "prod-cluster-infra"},
				&types.ECSCluster{Name: "prod-cluster-monitoring"},
			},
			incorrect: []*types.ECSCluster{
				&types.ECSCluster{Name: "staging-cluster-big"},
				&types.ECSCluster{Name: "staging-cluster-small"},
				&types.ECSCluster{Name: "staging-cluster-main"},
				&types.ECSCluster{Name: "staging-cluster-infra"},
				&types.ECSCluster{Name: "staging-cluster-monitoring"},
			},
		},
	}

	for _, test := range tests {
		e, err := New("eu-west-1", test.filter, 1, false)
		if err != nil {
			t.Errorf("Creation of exporter shoudn't error: %v", err)
		}

		// Check correct ones
		for _, c := range test.correct {
			if !e.validCluster(c) {
				t.Errorf("Expeceted valid cluster, got incorrect for regexp: %s; cluster: %s", test.filter, c.Name)
			}
		}

		// Check incorrect ones
		for _, c := range test.incorrect {
			if e.validCluster(c) {
				t.Errorf("Expeceted invalid cluster, got valid for regexp: '%s' ; cluster: %s", test.filter, c.Name)
			}
		}

	}
}

func TestCollectClusterMetricsTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test shouldn't panic, it did: %v", r)
		}
	}()

	exp, _ := New("eu-west-1", "", 1, false)
	ch := make(chan prometheus.Metric)
	close(ch)

	testCs := []*types.ECSCluster{&types.ECSCluster{ID: "c1", Name: "cluster1"}}

	// Cancel the context to mock as a finished main function
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	exp.collectClusterMetrics(ctx, ch, testCs)
}

func TestCollectClusterServiceMetricsTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test shouldn't panic, it did: %v", r)
		}
	}()

	exp, _ := New("eu-west-1", "", 1, false)
	ch := make(chan prometheus.Metric)
	close(ch)

	testC := &types.ECSCluster{ID: "c1", Name: "cluster1"}
	testSs := []*types.ECSService{&types.ECSService{ID: "s1", Name: "service1", DesiredT: 10, PendingT: 5, RunningT: 5}}

	// Cancel the context to mock as a finished main function
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	exp.collectClusterServicesMetrics(ctx, ch, testC, testSs)
}

func TestCollectContainerInstanceMetricsTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test shouldn't panic, it did: %v", r)
		}
	}()

	exp, _ := New("eu-west-1", "", 1, false)
	ch := make(chan prometheus.Metric)
	close(ch)

	testC := &types.ECSCluster{ID: "c1", Name: "cluster1"}
	testCIs := []*types.ECSContainerInstance{&types.ECSContainerInstance{ID: "ci0", InstanceID: "i-00000000000000000", AgentConn: true, Active: true, PendingT: 12}}

	// Cancel the context to mock as a finished main function
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	exp.collectClusterContainerInstancesMetrics(ctx, ch, testC, testCIs)
}
