package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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
	exp, err := New(region)
	if err != nil {
		t.Errorf("Creation of exporter shoudnt error: %v", err)
	}

	ch := make(chan prometheus.Metric)
	testCs := []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9"}

	// Collect mocked metrics
	go exp.collectClusterMetrics(ch, testCs)

	m := (<-ch).(prometheus.Metric)
	m2 := readGauge(m)

	expectedV := 9.0
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
