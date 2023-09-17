package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	GPULabels             = []string{"gpu", "UUID", "modelName"}
	getGPUStatLabelValues = func(gpu GPUStat) []string {
		return []string{
			fmt.Sprintf("%d", gpu.GPUIndex),
			gpu.UUID,
			gpu.GPUModelName,
		}
	}
	// [x]: configFiles
	SupportedGGPUMetricsName = []string{
		GPU_SM_CLOCK,
		GPU_MEMORY_CLOCK,
		//Temperature
		GPU_TEMPERATURE,
		GPU_FAN_SPEED,
		// Power
		GPU_POWER_USAGE,
		GPU_TOTAL_ENERGY_CONSUMPTION,

		// PCIe
		GPU_PCIE_TX_BYTES,
		GPU_PCIE_RX_BYTES,

		// Utilization (the sample period varies depending on the product)
		GPU_UTILIZATION,
		GPU_MEM_COPY_UTILIZATION,
		GPU_ENC_UTILIZATION,
		GPU_DEC_UTILIZATION,

		// Memory usage
		GPU_MEMORY_FREE_BYTES,
		GPU_MEMORY_USED_BYTES,
	}
)

type GPUCollector struct {
	cache              *NVMLCache
	metricDescs        map[string]*prometheus.Desc
	funcGetLabelValues func(gpu GPUStat) []string
	config             *Config
}

func NewGPUCollector(config *Config, cache *NVMLCache) *GPUCollector {
	metricsMap := make(map[string]*prometheus.Desc)
	// 如果config是空的，用默认的SupportedGGPUMetricsName
	if len(config.SupportedMetrics) > 0 {
		SupportedGGPUMetricsName = []string{}
		for _, name := range config.SupportedMetrics {
			if ISGPUMetricName(name) {
				SupportedGGPUMetricsName = append(SupportedGGPUMetricsName, name)
			}
		}
	}
	for _, name := range SupportedGGPUMetricsName {
		metricsMap[name] = prometheus.NewDesc(
			name,
			METRIC_META_MAP[name].Help,
			GPULabels,
			prometheus.Labels{LabelHostName: config.HostName},
		)

	}
	return &GPUCollector{
		metricDescs:        metricsMap,
		cache:              cache,
		config:             config,
		funcGetLabelValues: getGPUStatLabelValues,
	}

}

func (c *GPUCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metricDescs {
		ch <- desc
	}
}

func (c *GPUCollector) Collect(ch chan<- prometheus.Metric) {
	gpuCache := c.cache.GetGPUStats()
	for metricName, desc := range c.metricDescs {
		for _, gpu := range gpuCache {
			value := gpu.GetValueFromMetricName(metricName)
			metric := prometheus.MustNewConstMetric(
				desc,
				METRIC_META_MAP[metricName].PromType, // 从METRIC_META_MAP获取指标类型
				value,
				c.funcGetLabelValues(gpu)...,
			)
			if metric != nil {
				ch <- metric
			}

		}
	}
}
