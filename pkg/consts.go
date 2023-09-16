package collector

const (
	// MetricName

	// GPU
	// Clocks
	GPU_SM_CLOCK     = "gpu_sm_clock"     //     gauge, SM clock frequency (in MHz).
	gpu_memory_clock = "gpu_memory_clock" //gauge, Memory clock frequency (in MHz).

	//Temperature
	GPU_MEMORY_TEMPERATURE = "gpu_memory_temperature" // gauge, Memory temperature (in C).
	GPU_TEMPERATURE        = "gpu_temperature"        //   gauge, GPU temperature (in C).

	// Power
	GPU_POWER_USAGE              = "gpu_power_usage"              //              gauge, Power draw (in W).
	GPU_TOTAL_ENERGY_CONSUMPTION = "gpu_total_energy_consumption" // counter, Total energy consumption since boot (in mJ).

	// PCIe
	GPU_PCIE_TX_BYTES = "gpu_pcie_tx_bytes" //  counter, Total number of bytes transmitted through PCIe TX (in KB) via NVML.
	GPU_PCIE_RX_BYTES = "gpu_pcie_rx_bytes" //counter, Total number of bytes received through PCIe RX (in KB) via NVML.

	// Utilization (the sample period varies depending on the product)
	GPU_UTILIZATION          = "gpu_utilization"          //  gauge, GPU utilization (in %).
	GPU_MEM_COPY_UTILIZATION = "gpu_mem_copy_utilization" // gauge, Memory utilization (in %).
	GPU_ENC_UTILIZATION      = "gpu_mem_copy_utilization" //    gauge, Encoder utilization (in %).
	GPU_DEC_UTILIZATION      = "gpu_mem_copy_utilization" // gauge, Decoder utilization (in %).

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

	// todo: help desc
)

var (
	METRICHELP = map[string]string{
		PROCESS_CPU_PERCENT:        "process_cpu_precent",
		PROCESS_CPU_MEM_USED_BYTES: "process_cpu_mem_used_bytes",
		PROCESS_NUM_THREADS:        "process_num_threads",
		PROCESS_GPU_SM_UTIL:        "process_gpu_sm_util",
		PROCESS_GPU_MEM_UTIL:       "process_gpu_mem_util",
		PROCESS_GPU_DECODE_UTIL:    "process_gpu_decode_util",
		PROCESS_GPU_ENCODE_UTIL:    "process_gpu_encode_util",
	}
)
