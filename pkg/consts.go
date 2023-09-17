package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricMeta struct {
	FieldName string
	PromType  prometheus.ValueType
	Help      string
}

func ISGPUMetricName(name string) bool {
	return strings.HasPrefix(name, "gpu_")
}

func ISProcessMetricName(name string) bool {
	return strings.HasPrefix(name, "process_")
}

const (
	// MetricName

	// GPU
	// Clocks
	GPU_SM_CLOCK     = "gpu_sm_clock"     //     gauge, SM clock frequency (in MHz).
	GPU_MEMORY_CLOCK = "gpu_memory_clock" //gauge, Memory clock frequency (in MHz).

	// Temperature
	GPU_TEMPERATURE = "gpu_temperature" //   gauge, GPU temperature (in C).

	// FAN
	GPU_FAN_SPEED = "gpu_fan_speed" // gauge, Fan speed (in %).

	// Power
	GPU_POWER_USAGE              = "gpu_power_usage"              //              gauge, Power draw (in W).
	GPU_TOTAL_ENERGY_CONSUMPTION = "gpu_total_energy_consumption" // counter, Total energy consumption since boot (in mJ).

	// PCIe
	GPU_PCIE_TX_BYTES = "gpu_pcie_tx_bytes" //  counter, Total number of bytes transmitted through PCIe TX (in KB) via NVML.
	GPU_PCIE_RX_BYTES = "gpu_pcie_rx_bytes" //counter, Total number of bytes received through PCIe RX (in KB) via NVML.

	// Utilization (the sample period varies depending on the product)
	GPU_UTILIZATION          = "gpu_utilization"          //  gauge, GPU utilization (in %).
	GPU_MEM_COPY_UTILIZATION = "gpu_mem_copy_utilization" // gauge, Memory utilization (in %).
	GPU_ENC_UTILIZATION      = "gpu_enc_utilization"      //    gauge, Encoder utilization (in %).
	GPU_DEC_UTILIZATION      = "gpu_dec_utilization"      // gauge, Decoder utilization (in %).

	// Memory usage
	GPU_MEMORY_FREE = "gpu_memory_free" // gauge, Framebuffer memory free (in MiB).
	GPU_MEMORY_USED = "gpu_memory_used" // gauge, Framebuffer memory used (in MiB).

	// Process
	PROCESS_CPU_PERCENT        = "process_cpu_precent"
	PROCESS_CPU_MEM_USED_BYTES = "process_cpu_mem_used_bytes"
	PROCESS_NUM_THREADS        = "process_num_threads"
	PROCESS_GPU_SM_UTIL        = "process_gpu_sm_util"
	PROCESS_GPU_MEM_UTIL       = "process_gpu_mem_util"
	PROCESS_GPU_DECODE_UTIL    = "process_gpu_decode_util"
	PROCESS_GPU_ENCODE_UTIL    = "process_gpu_encode_util"

	// PROCESS_GPU_FRAME_MEM_UTIL = "process_gpu_frame_mem_util"
	// PROCESS_GPU_MEM_USED       = "process_gpu_mem_used"
)

var (
	// todo: add specific help info of process info
	METRIC_META_MAP = map[string]MetricMeta{
		GPU_SM_CLOCK:                 {GPU_SM_CLOCK, prometheus.GaugeValue, "SM clock frequency (in MHz)."},
		GPU_MEMORY_CLOCK:             {GPU_MEMORY_CLOCK, prometheus.GaugeValue, "Memory clock frequency (in MHz)."},
		GPU_TEMPERATURE:              {GPU_TEMPERATURE, prometheus.GaugeValue, "GPU temperature (in C)."},
		GPU_FAN_SPEED:                {GPU_FAN_SPEED, prometheus.GaugeValue, "Fan speed (in %)."},
		GPU_POWER_USAGE:              {GPU_POWER_USAGE, prometheus.GaugeValue, "Power draw (in W)."},
		GPU_TOTAL_ENERGY_CONSUMPTION: {GPU_TOTAL_ENERGY_CONSUMPTION, prometheus.CounterValue, "Total energy consumption since boot (in mJ)."},
		GPU_PCIE_TX_BYTES:            {GPU_PCIE_TX_BYTES, prometheus.GaugeValue, "Total number of bytes transmitted through PCIe TX (in KB) via NVML."},
		GPU_PCIE_RX_BYTES:            {GPU_PCIE_RX_BYTES, prometheus.GaugeValue, "Total number of bytes received through PCIe RX (in KB) via NVML."},
		GPU_UTILIZATION:              {GPU_UTILIZATION, prometheus.GaugeValue, "GPU utilization (in %)."},
		GPU_MEM_COPY_UTILIZATION:     {GPU_MEM_COPY_UTILIZATION, prometheus.GaugeValue, "Memory utilization (in %)."},
		GPU_ENC_UTILIZATION:          {GPU_ENC_UTILIZATION, prometheus.GaugeValue, "Encoder utilization (in %)."},
		GPU_DEC_UTILIZATION:          {GPU_DEC_UTILIZATION, prometheus.GaugeValue, "Decoder utilization (in %)."},
		GPU_MEMORY_FREE:              {GPU_MEMORY_FREE, prometheus.GaugeValue, "Framebuffer memory free (in MiB)."},
		GPU_MEMORY_USED:              {GPU_MEMORY_USED, prometheus.GaugeValue, "Framebuffer memory used (in MiB)."},
		PROCESS_CPU_PERCENT:          {PROCESS_CPU_PERCENT, prometheus.GaugeValue, "Process CPU percent."},
		PROCESS_CPU_MEM_USED_BYTES:   {PROCESS_CPU_MEM_USED_BYTES, prometheus.GaugeValue, "Process CPU memory used bytes."},
		PROCESS_NUM_THREADS:          {PROCESS_NUM_THREADS, prometheus.GaugeValue, "Process num threads."},
		PROCESS_GPU_SM_UTIL:          {PROCESS_GPU_SM_UTIL, prometheus.GaugeValue, "Process GPU SM util."},
		PROCESS_GPU_MEM_UTIL:         {PROCESS_GPU_MEM_UTIL, prometheus.GaugeValue, "Process GPU memory util."},
		PROCESS_GPU_DECODE_UTIL:      {PROCESS_GPU_DECODE_UTIL, prometheus.GaugeValue, "Process GPU decode util."},
		PROCESS_GPU_ENCODE_UTIL:      {PROCESS_GPU_ENCODE_UTIL, prometheus.GaugeValue, "Process GPU encode util."},
	}
)
