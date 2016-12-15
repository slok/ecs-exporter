package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/slok/ecs-exporter/log"
)

const (
	defaultListenAddress    = ":9222"
	defaultAwsRegion        = ""
	defaultMetricsPath      = "/metrics"
	defaultClusterFilter    = ".*"
	defaultDebug            = false
	defaultDisableCIMetrics = false
)

// Cfg is the global configuration
var cfg *config

// Parse global config
func parse(args []string) error {
	return cfg.parse(args)
}

// Config represents an app configuration
type config struct {
	fs *flag.FlagSet

	listenAddress    string
	awsRegion        string
	metricsPath      string
	clusterFilter    string
	debug            bool
	disableCIMetrics bool
}

// init will load all the flags
func init() {
	cfg = new()
}

// New returns an initialized config
func new() *config {
	c := &config{
		fs: flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
	}

	c.fs.StringVar(
		&c.listenAddress, "web.listen-address", defaultListenAddress, "Address to listen on")

	c.fs.StringVar(
		&c.awsRegion, "aws.region", defaultAwsRegion, "The AWS region to get metrics from")

	c.fs.StringVar(
		&c.clusterFilter, "aws.cluster-filter", defaultClusterFilter, "Regex used to filter the cluster names, if doesn't match the cluster is ignored")

	c.fs.StringVar(
		&c.metricsPath, "web.telemetry-path", defaultMetricsPath, "The path where metrics will be exposed")

	c.fs.BoolVar(
		&c.debug, "debug", defaultDebug, "Run exporter in debug mode")

	c.fs.BoolVar(
		&c.disableCIMetrics, "metrics.disable-cinstances", defaultDisableCIMetrics, "Disable clusters container instances metrics gathering")

	return c
}

// parse parses the flags for configuration
func (c *config) parse(args []string) error {
	log.Debugf("Parsing flags...")

	if err := c.fs.Parse(args); err != nil {
		return err
	}

	if len(c.fs.Args()) != 0 {
		return fmt.Errorf("Invalid command line arguments. Help: %s -h", os.Args[0])
	}

	if c.awsRegion == "" {
		return fmt.Errorf("An aws region is required")
	}

	if _, err := regexp.Compile(c.clusterFilter); err != nil {
		return fmt.Errorf("Invalid cluster filtering regex: %s", c.clusterFilter)
	}

	if c.clusterFilter != defaultClusterFilter {
		log.Warnf("Filtering cluster metrics by: %s", c.clusterFilter)
	}

	return nil
}
