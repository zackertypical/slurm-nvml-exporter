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
	exporter_port    = flag.String("exporter-port", ":9445", "exporter port")
	server_port      = flag.String("server-port", ":9446", "Address to listen on for web interface and telemetry.")
	metricConfigFile = flag.String("metric-config-file", "", "metric to export file")
	collectInterval  = flag.Int("collect-interval", 5, "interval to collect metrics")
	useSlurm         = flag.Bool("use-slurm", false, "use slurm to get process info")
)

// todo: helper
func main() {
	flag.Parse()

	// hostName
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorf("Unable to get hostname: %v", err)
		hostname = ""
	}
	// todo: add signal & nvml shutdown

	// todo: configuration file
	config := &collector.Config{
		CollectorsFile:  *metricConfigFile,
		ExporterPort:    *exporter_port,
		ServerPort:      *server_port,
		CollectInterval: *collectInterval,
		UseSlurm:        *useSlurm,
		// SupportedMetrics []string,
		HostName: hostname,
	}

	nvmlCache, err := collector.NewNVMLCache(config)
	if err != nil {
		logrus.Fatalf("Failed to init nvml, err: %v", err)
		os.Exit(1)
	}

	// go nvmlCache.Run()

	procCollector := collector.NewProcessCollector(config, nvmlCache)
	gpuCollector := collector.NewGPUCollector(config, nvmlCache)

	registry := prometheus.NewRegistry()

	registry.MustRegister(procCollector, gpuCollector)

	// Serve on all paths under addr
	logrus.Fatalf("ListenAndServe error: %v", http.ListenAndServe(*exporter_port, promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
}
