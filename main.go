package main

import (
	"flag"
	"net/http"
	"os"

	collector "github.com/nvml-exporter/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	exporter_port = flag.String("exporter-port", ":9445", "Address to listen on for web interface and telemetry.")
	configFile    = flag.String("config", ":9445", "metric to export file")
)

func main() {
	flag.Parse()

	// todo: configuration file
	config := &collector.Config{}

	nvmlCache, err := collector.NewNVMLCache(config)
	if err != nil {
		logrus.Fatalf("Failed to init nvml, err: %v", err)
		os.Exit(1)
	}
	procCollector := collector.NewProcessCollector(config, nvmlCache)
	gpuCollector := collector.NewGPUCollector(config, nvmlCache)

	registry := prometheus.NewRegistry()

	registry.MustRegister(procCollector, gpuCollector)

	// Serve on all paths under addr
	logrus.Fatalf("ListenAndServe error: %v", http.ListenAndServe(*exporter_port, promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
}
