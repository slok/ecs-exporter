package collector

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	"github.com/slok/ecs-exporter/log"
	"github.com/slok/ecs-exporter/types"
)

const (
	namespace = "ecs"
	timeout   = 10 * time.Second
)

// Metrics descriptions
var (
	// exporter metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of ecs successful.",
		[]string{"region"}, nil,
	)

	// Clusters metrics
	clusterCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "clusters"),
		"The total number of clusters",
		[]string{"region"}, nil,
	)

	//  Services metrics
	serviceCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "services"),
		"The total number of services",
		[]string{"region", "cluster"}, nil,
	)

	serviceDesired = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "service_desired_tasks"),
		"The desired number of instantiations of the task definition to keep running regarding a service",
		[]string{"region", "cluster", "service"}, nil,
	)

	servicePending = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "service_pending_tasks"),
		"The number of tasks in the cluster that are in the PENDING state regarding a service",
		[]string{"region", "cluster", "service"}, nil,
	)

	serviceRunning = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "service_running_tasks"),
		"The number of tasks in the cluster that are in the RUNNING state regarding a service",
		[]string{"region", "cluster", "service"}, nil,
	)

	//  Container instances metrics
	cInstanceCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "container_instances"),
		"The total number of container instances",
		[]string{"region", "cluster"}, nil,
	)

	cInstanceAgentC = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "container_instance_agent_connected"),
		"The connected state of the container instance agent",
		[]string{"region", "cluster", "instance"}, nil,
	)

	cInstanceStatusAct = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "container_instance_active"),
		"The status of the container instance in ACTIVE state, indicates that the container instance can accept tasks.",
		[]string{"region", "cluster", "instance"}, nil,
	)

	cInstancePending = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "container_instance_pending_tasks"),
		"The number of tasks on the container instance that are in the PENDING status.",
		[]string{"region", "cluster", "instance"}, nil,
	)
)

// Exporter collects ECS clusters metrics
type Exporter struct {
	sync.Mutex                   // Our exporter object will be locakble to protect from concurrent scrapes
	client        ECSGatherer    // Custom ECS client to get information from the clusters
	region        string         // The region where the exporter will scrape
	clusterFilter *regexp.Regexp // Compiled regular expresion to filter clusters
	noCIMetrics   bool           // Don't gather container instance metrics
	timeout       time.Duration  // The timeout for the whole gathering process
}

// New returns an initialized exporter
func New(awsRegion string, clusterFilterRegexp string, maxConcurrency int, disableCIMetrics bool) (*Exporter, error) {
	c, err := NewECSClient(awsRegion, maxConcurrency)
	if err != nil {
		return nil, err
	}

	cRegexp, err := regexp.Compile(clusterFilterRegexp)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		Mutex:         sync.Mutex{},
		client:        c,
		region:        awsRegion,
		clusterFilter: cRegexp,
		noCIMetrics:   disableCIMetrics,
		timeout:       timeout,
	}, nil

}

// sendSafeMetric uses context to cancel the send over a closed channel.
// If a main function finishes (for example due to to timeout), the goroutines running in background will
// try to send metrics over a closed channel, this will panic, this way the context will check first
// if the iteraiton has been finished and dont let continue sending the metric
func sendSafeMetric(ctx context.Context, ch chan<- prometheus.Metric, metric prometheus.Metric) error {
	// Check if iteration has finished
	select {
	case <-ctx.Done():
		log.Errorf("Tried to send a metric after collection context has finished, metric: %s", metric)
		return ctx.Err()
	default: // continue
	}
	// If no then send the metric
	ch <- metric
	return nil
}

// Describe describes all the metrics ever exported by the ECS exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- clusterCount
	ch <- serviceCount
	ch <- serviceCount
	ch <- serviceDesired
	ch <- servicePending
	ch <- serviceRunning

	if e.noCIMetrics {
		return
	}

	ch <- cInstanceCount
	ch <- cInstanceAgentC
	ch <- cInstanceStatusAct
	ch <- cInstancePending
}

// Collect fetches the stats from configured ECS and delivers them
// as Prometheus metrics. It implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Debugf("Start collecting...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.Lock()
	defer e.Unlock()

	// Get clusters
	cs, err := e.client.GetClusters()
	if err != nil {
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0, e.region))
		log.Errorf("Error collecting metrics: %v", err)
		return
	}

	e.collectClusterMetrics(ctx, ch, cs)

	// Start getting metrics per cluster on its own goroutine
	errC := make(chan bool)
	totalCs := 0 // total cluster metrics gorotine ran

	for _, c := range cs {
		// Filter not desired clusters
		if !e.validCluster(c) {
			log.Debugf("Cluster '%s' filtered", c.Name)
			continue
		}
		totalCs++
		go func(c types.ECSCluster) {
			// Get services
			ss, err := e.client.GetClusterServices(&c)
			if err != nil {
				errC <- true
				log.Errorf("Error collecting cluster service metrics: %v", err)
				return
			}
			e.collectClusterServicesMetrics(ctx, ch, &c, ss)

			// Get container instance metrics (if enabled)
			if e.noCIMetrics {
				log.Debug("Container instance metrics disabled, no gathering these metrics...")
				errC <- false
				return
			}

			cs, err := e.client.GetClusterContainerInstances(&c)
			if err != nil {
				errC <- true
				log.Errorf("Error collecting cluster container instance metrics: %v", err)
				return
			}
			e.collectClusterContainerInstancesMetrics(ctx, ch, &c, cs)

			errC <- false
		}(*c)
	}

	// Grab result or not result error for each goroutine, on first error exit
	result := float64(1)

ServiceCollector:
	for i := 0; i < totalCs; i++ {
		select {
		case err := <-errC:
			if err {
				result = 0
				break ServiceCollector
			}
		case <-time.After(e.timeout):
			log.Errorf("Error collecting metrics: Timeout making calls, waited for %v  without response", e.timeout)
			result = 0
			break ServiceCollector
		}

	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, result, e.region,
	)
}

// validCluster will return true if the cluster is valid for the exporter cluster filtering regexp, otherwise false
func (e *Exporter) validCluster(cluster *types.ECSCluster) bool {
	return e.clusterFilter.MatchString(cluster.Name)
}

func (e *Exporter) collectClusterMetrics(ctx context.Context, ch chan<- prometheus.Metric, clusters []*types.ECSCluster) {
	// Total cluster count
	sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(clusterCount, prometheus.GaugeValue, float64(len(clusters)), e.region))
}

func (e *Exporter) collectClusterServicesMetrics(ctx context.Context, ch chan<- prometheus.Metric, cluster *types.ECSCluster, services []*types.ECSService) {

	// Total services
	sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(serviceCount, prometheus.GaugeValue, float64(len(services)), e.region, cluster.Name))

	for _, s := range services {
		// Desired task count
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(serviceDesired, prometheus.GaugeValue, float64(s.DesiredT), e.region, cluster.Name, s.Name))

		// Pending task count
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(servicePending, prometheus.GaugeValue, float64(s.PendingT), e.region, cluster.Name, s.Name))

		// Running task count
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(serviceRunning, prometheus.GaugeValue, float64(s.RunningT), e.region, cluster.Name, s.Name))
	}
}

func (e *Exporter) collectClusterContainerInstancesMetrics(ctx context.Context, ch chan<- prometheus.Metric, cluster *types.ECSCluster, cInstances []*types.ECSContainerInstance) {
	// Total container instances
	sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(cInstanceCount, prometheus.GaugeValue, float64(len(cInstances)), e.region, cluster.Name))

	for _, c := range cInstances {
		// Agent connected
		var conn float64
		if c.AgentConn {
			conn = 1
		}
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(cInstanceAgentC, prometheus.GaugeValue, conn, e.region, cluster.Name, c.InstanceID))

		// Instance status
		var active float64
		if c.Active {
			active = 1
		}
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(cInstanceStatusAct, prometheus.GaugeValue, active, e.region, cluster.Name, c.InstanceID))

		// Pending tasks
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(cInstancePending, prometheus.GaugeValue, float64(c.PendingT), e.region, cluster.Name, c.InstanceID))
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("ecs_exporter"))
}
