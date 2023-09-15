package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricName
	MetricProcessCPUPercent    MetricName = "process_cpu_precent"
	MetricProcessCPUMemoryUsed MetricName = "process_cpu_mem_used"
	MetricProcessGPUMemoryUsed MetricName = "process_gpu_mem_used"
	MetricProcessSmUtil        MetricName = "process_gpu_sm_util"
	MetricProcessGPUMemoryUtil MetricName = "process_gpu_mem_util"
	MetricProcessFrameMemUtil  MetricName = "process_gpu_frame_mem_util"
	MetricProcessDecodeUtil    MetricName = "process_gpu_decode_util"
	MetricProcessEncodeUtil    MetricName = "process_gpu_encode_util"
)

var (
	ProcessLabels               = []string{LabelGPU, LabelPID, LabelProcName, LabelUser}
	SupportedProcessMetricsName = []MetricName{
		MetricProcessCPUPercent,
		MetricProcessCPUMemoryUsed,
		MetricProcessGPUMemoryUsed,
		MetricProcessSmUtil,
		MetricProcessGPUMemoryUtil,
		MetricProcessFrameMemUtil,
		MetricProcessDecodeUtil,
		MetricProcessEncodeUtil,
	}
)

type ProcessCollector struct {
	cache       *NVMLCache
	metricDescs map[MetricName]*prometheus.Desc
}

func NewProcessCollector(config *Config) *ProcessCollector {
	metricsMap := make(map[MetricName]*prometheus.Desc)
	for _, name := range SupportedProcessMetricsName {
		if !config.UseSlurm {
			metricsMap[name] = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", string(name)),
				fmt.Sprintf("nvml process exporter -- %s", name),
				ProcessLabels,
				prometheus.Labels{LabelHostName: config.HostName},
			)
		} else {
			// slurm添加的SlurmProcLabels
			metricsMap[name] = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", string(name)),
				fmt.Sprintf("nvml process exporter -- %s", name),
				SlurmProcLabels,
				prometheus.Labels{LabelHostName: config.HostName},
			)
		}
	}
	return &ProcessCollector{
		metricDescs: metricsMap,
	}

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
			metric := convertProcStatToMetric(metricName, desc, ps)
			if metric != nil {
				ch <- metric
			}

		}
	}
}

func convertProcStatToMetric(metricName MetricName, desc *prometheus.Desc, ps ProcessStat) prometheus.Metric {
	switch metricName {
	case MetricProcessCPUPercent:
		return prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(ps.CPUPercent),
			psLabelValues(ps)...,
		)
	case MetricProcessCPUMemoryUsed:
		return prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(ps.CPUMemoryUsed),
			psLabelValues(ps)...,
		)
	
	default:
		return nil
	}

}

// 	ProcessLabels = []string{LabelGPU, LabelPID, LabelProcName, LabelUser}
func psLabelValues(ps ProcessStat) []string {
	return []string{
		fmt.Sprintf("%d", ps.GPUIndex),
		fmt.Sprintf("%d", ps.Pid),
		ps.ProcName,
		ps.UserName,
	}
}
