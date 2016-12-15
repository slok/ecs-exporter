package main

//import _ "net/http/pprof"

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/slok/ecs-exporter/collector"
	"github.com/slok/ecs-exporter/log"
)

// Main is the application entry point
func Main() int {
	log.Infof("Starting ECS exporter...")

	// Parse command line flags
	if err := parse(os.Args[1:]); err != nil {
		log.Error(err)
		return 1
	}

	if cfg.debug {
		log.SetLevel(log.DebugLevel)
	}

	if cfg.disableCIMetrics {
		log.Warnf("Cluster container instance metrics have been disabled")
	}

	// Create the exporter and register it
	exporter, err := collector.New(cfg.awsRegion, cfg.clusterFilter, cfg.disableCIMetrics)
	if err != nil {
		log.Error(err)
		return 1
	}
	prometheus.MustRegister(exporter)

	// Serve metrics
	http.Handle(cfg.metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>ECS Exporter</title></head>
             <body>
             <h1>ECS Exporter</h1>
             <p><a href='` + cfg.metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Infoln("Listening on", cfg.listenAddress)
	log.Fatal(http.ListenAndServe(cfg.listenAddress, nil))

	return 0
}

func main() {
	// Run main program
	exCode := Main()
	os.Exit(exCode)
}
