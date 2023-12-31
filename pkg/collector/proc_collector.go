package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ProcessLabels             = []string{"gpu", "pid", "procName", "user", "status", "ppid"}
	ProcessInfoLables         = []string{"gpu", "pid", "procName", "user", "status", "ppid", "workDir", "cmdLine"}
	getProcessStatLabelValues = func(ps ProcessStat) []string {
		return []string{
			fmt.Sprintf("%d", ps.GPUIndex),
			fmt.Sprintf("%d", ps.Pid),
			ps.ProcName,
			ps.User,
			ps.Status,
			fmt.Sprintf("%d", ps.PPid),
		}
	}

	SupportedProcessMetricsName = []string{
		PROCESS_INFO,
		PROCESS_CPU_PERCENT,
		PROCESS_CPU_MEM_USED_BYTES,
		PROCESS_NUM_THREADS,
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
	if len(config.SupportedMetrics) > 0 {
		SupportedProcessMetricsName = []string{}
		for _, name := range config.SupportedMetrics {
			if ISProcessMetricName(name) {
				SupportedProcessMetricsName = append(SupportedProcessMetricsName, name)
			}
		}
	}

	for _, name := range SupportedProcessMetricsName {
		if !config.UseSlurm {
			if name == PROCESS_INFO {
				metricsMap[name] = prometheus.NewDesc(
					name,
					METRIC_META_MAP[name].Help,
					ProcessInfoLables,
					prometheus.Labels{LabelHostName: config.HostName},
				)
			} else {
				metricsMap[name] = prometheus.NewDesc(
					name,
					METRIC_META_MAP[name].Help,
					ProcessLabels,
					prometheus.Labels{LabelHostName: config.HostName},
				)
			}

		} else {
			// slurm添加的SlurmProcLabels
			if name == PROCESS_INFO {
				metricsMap[name] = prometheus.NewDesc(
					name,
					METRIC_META_MAP[name].Help,
					SlurmProcInfoLabels,
					prometheus.Labels{LabelHostName: config.HostName},
				)
			} else {
				metricsMap[name] = prometheus.NewDesc(
					name,
					METRIC_META_MAP[name].Help,
					SlurmProcLabels,
					prometheus.Labels{LabelHostName: config.HostName},
				)
			}

		}
	}
	psCollector := &ProcessCollector{
		metricDescs: metricsMap,
		cache:       cache,
		config:      config,
	}
	// slurm相关添加的labels
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
			value := ps.GetValueFromMetricName(metricName)
			var metric prometheus.Metric
			labelValues := c.funcGetLabelValues(ps)

			if metricName == PROCESS_INFO {
				labelValues = append(labelValues, ps.WorkingDir, ps.CommandLine)
				metric = prometheus.MustNewConstMetric(
					desc,
					METRIC_META_MAP[metricName].PromType,
					value,
					labelValues...,
				)
			} else {
				metric = prometheus.MustNewConstMetric(
					desc,
					METRIC_META_MAP[metricName].PromType,
					value,
					labelValues...,
				)
			}

			if metric != nil {
				ch <- metric
			}
		}
	}
}
