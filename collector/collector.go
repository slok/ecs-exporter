package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

const (
	namespace = "ecs"
)

// Metrics descriptions
var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of ecs successful.",
		[]string{"region"}, nil,
	)

	clusterCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_total"),
		"The total number of clusters",
		[]string{"region"}, nil,
	)
)

// Exporter collects ECS clusters metrics
type Exporter struct {
	sync.Mutex            // Our exporter object will be locakble to protect from concurrent scrapes
	client     *ECSClient // Custom ECS client to get informationfrom the clusters
	region     string     // The region where the exporter will scrape
}

// New returns an initialized exporter
func New(awsRegion string) (*Exporter, error) {
	c, err := NewECSClient(awsRegion)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		Mutex:  sync.Mutex{},
		client: c,
		region: awsRegion,
	}, nil

}

// Describe describes all the metrics ever exported by the ECS exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- clusterCount
}

// Collect fetches the stats from configured ECS and delivers them
// as Prometheus metrics. It implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Debugf("Start collecting...")

	e.Lock()
	defer e.Unlock()

	// Get clusters
	cIDs, err := e.client.GetClusterIDs()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0, e.region,
		)

		log.Errorf("Error collecting metrics: %v", err)
		return
	}
	e.collectClusterMetrics(ch, cIDs)

	// Get services

	// Seems everything went ok
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1, e.region,
	)
}

func (e *Exporter) collectClusterMetrics(ch chan<- prometheus.Metric, clusterIDs []string) {
	ch <- prometheus.MustNewConstMetric(
		clusterCount, prometheus.GaugeValue, float64(len(clusterIDs)), e.region,
	)
}

func init() {
	prometheus.MustRegister(version.NewCollector("ecs_exporter"))
}
