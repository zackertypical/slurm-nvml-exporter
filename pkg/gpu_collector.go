package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	GPULabels = []string{"gpu", "UUID", "modelName"}

	getGPUStatLabelValues = func(gpu GPUStat) []string {
		return []string{
			fmt.Sprint("%d", gpu.GPUIndex),
			gpu.UUID,
			gpu.GPUModelName,
		}
	}
	// todo: configFiles
	SupportedGGPUMetricsName = []string{
		GPU_SM_CLOCK,
		gpu_memory_clock,

		//Temperature
		GPU_MEMORY_TEMPERATURE,
		GPU_TEMPERATURE,

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
		GPU_MEMORY_FREE,
		GPU_MEMORY_USED,
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
	for _, name := range SupportedGGPUMetricsName {
		metricsMap[name] = prometheus.NewDesc(
			name,
			fmt.Sprintf("nvml gpu exporter -- %s", name),
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
			// todo: slurm proc
			value := gpu.GetValueFromMetricName(metricName)
			metric := prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				value,
				c.funcGetLabelValues(gpu)...,
			)
			if metric != nil {
				ch <- metric
			}

		}
	}
}
