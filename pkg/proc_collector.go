package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ProcessLabels             = []string{"gpu", "pid", "procName", "user"}
	getProcessStatLabelValues = func(ps ProcessStat) []string {
		return []string{
			fmt.Sprintf("%d", ps.GPUIndex),
			fmt.Sprintf("%d", ps.Pid),
			ps.ProcName,
			ps.User,
		}
	}

	// todo: configFiles
	SupportedProcessMetricsName = []string{
		PROCESS_CPU_PERCENT,
		PROCESS_CPU_MEM_USED_BYTES,
		PROCESS_GPU_SM_UTIL,
		PROCESS_GPU_MEM_UTIL,
		PROCESS_GPU_DECODE_UTIL,
		PROCESS_GPU_ENCODE_UTIL,
	}
)

type ProcessCollector struct {
	cache              *NVMLCache
	metricDescs        map[string]*prometheus.Desc
	funcGetLabelValues func(ps ProcessStat) []string
	config             *Config
}

func NewProcessCollector(config *Config, cache *NVMLCache) *ProcessCollector {
	metricsMap := make(map[string]*prometheus.Desc)
	for _, name := range SupportedProcessMetricsName {
		if !config.UseSlurm {
			metricsMap[name] = prometheus.NewDesc(
				name,
				fmt.Sprintf("nvml process exporter -- %s", name),
				ProcessLabels,
				prometheus.Labels{LabelHostName: config.HostName},
			)

		} else {
			// slurm添加的SlurmProcLabels
			metricsMap[name] = prometheus.NewDesc(
				name,
				fmt.Sprintf("nvml process exporter -- %s", name),
				SlurmProcLabels,
				prometheus.Labels{LabelHostName: config.HostName},
			)
		}
	}
	psCollector := &ProcessCollector{
		metricDescs: metricsMap,
		cache:       cache,
		config:      config,
	}
	// slurmConfig
	if config.UseSlurm {
		psCollector.funcGetLabelValues = getSlurmProcessStatLabelValues
	} else {
		psCollector.funcGetLabelValues = getProcessStatLabelValues
	}

	return psCollector

}

func (c *ProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metricDescs {
		ch <- desc
	}
}

func (c *ProcessCollector) Collect(ch chan<- prometheus.Metric) {
	processCache := c.cache.GetProcessStats()
	for metricName, desc := range c.metricDescs {
		for _, ps := range processCache {
			// todo: slurm proc
			value := ps.GetValueFromMetricName(metricName)
			metric := prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				value,
				c.funcGetLabelValues(ps)...,
			)
			if metric != nil {
				ch <- metric
			}
		}
	}
}
