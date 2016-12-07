package collector

import (
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
	exp, err := New(region, "")
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
	go exp.collectClusterMetrics(ch, testCs)

	m := (<-ch).(prometheus.Metric)
	m2 := readGauge(m)

	expectedV := 10.0
	// Check colected metrics are ok
	if m2.value != expectedV {
		t.Errorf("expected %f cluster_total, got %f", expectedV, m2.value)
	}

	if m2.labels["region"] != region {
		t.Errorf("expected %s region, got %s", region, m2.labels["region"])
	}

	expected := `Desc{fqName: "ecs_cluster_total", help: "The total number of clusters", constLabels: {}, variableLabels: [region]}`
	if expected != m.Desc().String() {
		t.Errorf("expected '%s', \ngot '%s'", expected, m.Desc().String())
	}
}

func TestCollectClusterServiceMetrics(t *testing.T) {
	region := "eu-west-1"
	exp, err := New(region, "")
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
		exp.collectClusterServicesMetrics(ch, testC, testSs)
		close(ch)
	}()

	// Check 1st received metric of services as group
	m := (<-ch).(prometheus.Metric)
	m2 := readGauge(m)
	want := float64(len(testSs))
	if m2.value != want {
		t.Errorf("expected %f service_total, got %f", want, m2.value)
	}

	for _, wantS := range testSs {
		// Check 1st received metric  per service (desired)
		m := (<-ch).(prometheus.Metric)
		m2 := readGauge(m)
		want := float64(wantS.DesiredT)
		if m2.value != want {
			t.Errorf("expected %f service_desired_tasks, got %f", want, m2.value)
		}

		// Check 1st received metric  per service (pending)
		m = (<-ch).(prometheus.Metric)
		m2 = readGauge(m)
		want = float64(wantS.PendingT)
		if m2.value != want {
			t.Errorf("expected %f service_pending_tasks, got %f", want, m2.value)
		}

		// Check 1st received metric  per service (running)
		m = (<-ch).(prometheus.Metric)
		m2 = readGauge(m)
		want = float64(wantS.RunningT)
		if m2.value != want {
			t.Errorf("expected %f service_running_tasks, got %f", want, m2.value)
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
		e, err := New("eu-west-1", test.filter)
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
