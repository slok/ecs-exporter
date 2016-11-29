package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/common/log"
)

const (
	defaultListenAddress = ":9222"
	defaultAwsRegion     = ""
	defaultMetricsPath   = "/metrics"
)

// Cfg is the global configuration
var Cfg *Config

// Parse global config
func Parse(args []string) error {
	return Cfg.Parse(args)
}

// Config represents an app configuration
type Config struct {
	fs *flag.FlagSet

	listenAddress string
	awsRegion     string
	metricsPath   string
}

// init will load all the flags
func init() {
	Cfg = New()
}

// New returns an initialized config
func New() *Config {
	c := &Config{
		fs: flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
	}

	c.fs.StringVar(
		&c.listenAddress, "web.listen-address", defaultListenAddress, "Address to listen on")

	c.fs.StringVar(
		&c.awsRegion, "aws.region", defaultAwsRegion, "The AWS region to get metrics from")

	c.fs.StringVar(
		&c.metricsPath, "web.telemetry-path", defaultMetricsPath, "The path where metrics will be exposed")

	return c
}

// Parse parses the flags for configuration
func (c *Config) Parse(args []string) error {
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
