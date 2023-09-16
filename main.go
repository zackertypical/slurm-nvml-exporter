package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nvml-exporter/pkg/collector"

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

	// setup config
	config := &collector.Config{
		CollectorsFile:  *metricConfigFile,
		ExporterPort:    *exporter_port,
		ServerPort:      *server_port,
		CollectInterval: *collectInterval,
		UseSlurm:        *useSlurm,
		// SupportedMetrics []string,
		HostName: hostname,
	}
	// setup signals
	stop := make(chan interface{})
	sigs := newOSWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	// run nvml cache
	nvmlCache, err := collector.NewNVMLCache(config)
	if err != nil {
		logrus.Fatalf("Failed to init nvml, err: %v", err)
		os.Exit(1)
	}

	go nvmlCache.Run(stop)

	// setup collectors
	procCollector := collector.NewProcessCollector(config, nvmlCache)
	gpuCollector := collector.NewGPUCollector(config, nvmlCache)

	registry := prometheus.NewRegistry()

	registry.MustRegister(procCollector, gpuCollector)

	// start listening exporter server
	// todo: stop chan and server
	logrus.Fatalf("ListenAndServe error: %v", http.ListenAndServe(*exporter_port, promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	for {
		select {
		case sig := <-sigs:
			close(stop)
			logrus.Infof("Receive sig: %v, Shutting down exporter...", sig)
		}
	}
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}
