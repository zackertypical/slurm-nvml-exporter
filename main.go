package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nvml-exporter/pkg/collector"
	"github.com/nvml-exporter/pkg/debug"
	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	server_port      = flag.String("server-port", ":9445", "Address to listen on for web interface and telemetry.")
	metricConfigFile = flag.String("metric-config-file", "", "metric to export file")
	collectInterval  = flag.Int("collect-interval", 5, "interval to collect metrics")
	useSlurm         = flag.Bool("use-slurm", false, "use slurm to get process info")
	debugLog         = flag.Bool("debug", false, "debug log level")
)

// todo: helper
func main() {
	flag.Parse()

	if *debugLog {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// hostName
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorf("Unable to get hostname: %v", err)
		hostname = ""
	}

	// setup config
	config := &collector.Config{
		ServerPort:      *server_port,
		CollectInterval: *collectInterval,
		UseSlurm:        *useSlurm,
		// SupportedMetrics []string,
		HostName: hostname,
	}
	if *metricConfigFile != "" {
		metrics, err := parseMetricsConfig(*metricConfigFile)
		if err != nil {
			logrus.Fatalf("Failed to init config file, %v", err)
			os.Exit(1)
		} else {
			config.SupportedMetrics = metrics
		}
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

	go func() {
		for {
			select {
			case sig := <-sigs:
				close(stop)
				logrus.Infof("Receive sig: %v, Shutting down exporter...", sig)
				return
			}
		}
	}()

	// setup collectors
	procCollector := collector.NewProcessCollector(config, nvmlCache)
	gpuCollector := collector.NewGPUCollector(config, nvmlCache)

	registry := prometheus.NewRegistry()

	registry.MustRegister(procCollector, gpuCollector)

	// start listening exporter server
	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	r.PathPrefix("/debug").Handler(debug.HandlerFor(nvmlCache))
	// r.Handle("/debug", debug.HandlerFor(nvmlCache))
	server := &http.Server{
		Addr:    *server_port,
		Handler: r,
	}
	go func() {
		logrus.Infof("ListenAndServe on port %v", *server_port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	// Shut down the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server shutdown error: %v", err)
	}
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}

func parseMetricsConfig(filePath string) ([]string, error) {
	// 读取配置文件内容
	configData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file, err: %v", err)
	}

	// 解析配置文件
	var config struct {
		MetricName []string `yaml:"metricName"`
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("parse config file failed：%v", err)
	}
	// check metric name in METRIC_META_MAP
	metrics := make([]string, 0)
	for _, name := range config.MetricName {
		if _, ok := collector.METRIC_META_MAP[name]; !ok {
			logrus.Errorf("metric name: %v, not supported!", name)
		} else {
			metrics = append(metrics, name)
		}
	}
	return metrics, nil
}
