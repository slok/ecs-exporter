package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/slok/ecs-exporter/log"
)

const (
	defaultListenAddress = ":9222"
	defaultAwsRegion     = ""
	defaultMetricsPath   = "/metrics"
	defaultDebug         = false
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

	listenAddress string
	awsRegion     string
	metricsPath   string
	debug         bool
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
		&c.metricsPath, "web.telemetry-path", defaultMetricsPath, "The path where metrics will be exposed")

	c.fs.BoolVar(
		&c.debug, "debug", defaultDebug, "Run exporter in debug mode")

	return c
}

// parse parses the flags for configuration
func (c *config) parse(args []string) error {
	log.Debugf("Parsing flags...")

	err := c.fs.Parse(args)
	if err != nil {
		return err
	}

	if len(c.fs.Args()) != 0 {
		err = fmt.Errorf("Invalid command line arguments. Help: %s -h", os.Args[0])
	}

	if c.awsRegion == "" {
		err = fmt.Errorf("An aws region is required")
	}

	return err
}
