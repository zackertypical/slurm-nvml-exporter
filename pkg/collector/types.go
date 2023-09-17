/*
** API Reference Doc: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html
**
 */

package collector

import (
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
)

const (
	LabelHostName = "Hostname"
)

type Config struct {
	ServerPort     string
	CollectInterval  int
	UseSlurm         bool
	SupportedMetrics []string
	HostName         string
}

type GPUDevice struct {
	nvml.Device

	GPUInfo
}

// todo: add GPUInfo
type GPUInfo struct {
	UUID             string                `json:"UUID"`
	GPUModelName     string                `json:"modelName"`
	GPUIndex         uint                  `json:"gpuIndex"`
	Attributes       nvml.DeviceAttributes `json:"attributes"`
	PcieLinkMaxSpeed uint32                `json:"pcieLinkMaxSpeed"`
}

type ProcessStat struct {
	Pid      uint32 `json:"pid"`
	GPUIndex int    `json:"gpu"`
	ProcName string `json:"procName"`
	User     string `json:"user"`

	// CPU Metrics
	CPUPercent         float64 `json:"cpu_percent"`
	CPUMemoryUsedBytes uint64  `json:"cpu_mem_used_bytes"`
	NumThreads         int32   `json:"num_threads"`

	// GPU Metrics
	Smutil  uint32 `json:"smutil"`  // SM利用率
	Memutil uint32 `json:"memutil"` // 显存利用率
	Decutil uint32 `json:"decutil"`
	Encutil uint32 `json:"encutil"`

	// Slurm Lables
	SlurmProcInfo
}

// type DeviceAttributes struct {
// 	MultiprocessorCount       uint32
// 	SharedCopyEngineCount     uint32
// 	SharedDecoderCount        uint32
// 	SharedEncoderCount        uint32
// 	SharedJpegCount           uint32
// 	SharedOfaCount            uint32
// 	GpuInstanceSliceCount     uint32
// 	ComputeInstanceSliceCount uint32
// 	MemorySizeMB              uint64
// }

type GPUStat struct {
	GPUIndex     uint
	UUID         string
	GPUModelName string

	SMClock  uint32 `json:"sm_clock"`  //gauge, SM clock frequency (in MHz).
	MemClock uint32 `json:"mem_clock"` //gauge, Memory clock frequency (in MHz).

	PowerUsage             uint32 `json:"power_usage"`        // gauge, Power draw (in W).
	TotalEnergyConsumption uint64 `json:"energy_consumption"` //  counter, Total energy consumption since boot (in mJ).

	FanSpeed    uint32 `json:"fan_speed"`
	Temperature uint32 `json:"temperature"`

	GPUUtil     uint32 `json:"gpu_util"`
	EncoderUtil uint32 `json:"encoder_util"`
	DecoderUtil uint32 `json:"dncoder_util"`
	MemCopyUtil uint32 `json:"memcpy_util"`

	PCIETXBytes uint32 `json:"pcie_tx_bytes"` // gauge, The rate of data transmitted over the PCIe bus - including both protocol headers and data payloads - in bytes per second.
	PCIERXBytes uint32 `json:"pcie_rx_bytes"` // gauge, The rate of data received over the PCIe bus - including both protocol headers and data payloads - in bytes per second.

	// DCGM_FI_PROF_DRAM_ACTIVE,        gauge, Ratio of cycles the device memory interface is active sending or receiving data (in %).
	// DCGM_FI_PROF_GR_ENGINE_ACTIVE,   gauge, Ratio of time the graphics engine is active (in %).
	// DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL,            counter, Total number of NVLink bandwidth counters for all lanes.

	MemoryUtil uint32 `json:"mem_util"`
	MemoryFree uint64 `json:"mem_free"`
	MemoryUsed uint64 `json:"mem_used"`
}

// [x]: configuration
// DeviceGetGPUStat Only gets the metric from arg metrics
func (g *GPUDevice) DeviceGetGPUStat(metrics []string) GPUStat {
	gpuStat := GPUStat{
		GPUIndex:     g.GPUIndex,
		UUID:         g.UUID,
		GPUModelName: g.GPUModelName,
	}
	utilizationRates, _ := g.GetUtilizationRates()
	memoryInfo, _ := g.GetMemoryInfo()
	for _, metric := range metrics {
		if !ISGPUMetricName(metric) {
			continue
		}
		switch metric {
		case GPU_SM_CLOCK:
			gpuStat.SMClock, _ = g.GetClockInfo(nvml.CLOCK_SM)
		case GPU_MEMORY_CLOCK:
			gpuStat.MemClock, _ = g.GetClockInfo(nvml.CLOCK_MEM)
		case GPU_TEMPERATURE:
			gpuStat.Temperature, _ = g.GetTemperature(nvml.TEMPERATURE_GPU)
		case GPU_POWER_USAGE:
			power, _ := g.GetPowerUsage()
			gpuStat.PowerUsage = power / 1000 // 转换为W
		case GPU_TOTAL_ENERGY_CONSUMPTION:
			energy, _ := g.GetTotalEnergyConsumption()
			gpuStat.TotalEnergyConsumption = energy * 1000 // 转换为mJ
		case GPU_PCIE_TX_BYTES:
			kb, _ := g.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES)
			gpuStat.PCIETXBytes = kb * 1024 // KB/s 转换为bytes per second
		case GPU_PCIE_RX_BYTES:
			kb, _ := g.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES)
			gpuStat.PCIERXBytes = kb * 1024 // KB/s 转换为bytes per second
		case GPU_UTILIZATION:
			gpuStat.GPUUtil = utilizationRates.Gpu
		case GPU_MEM_COPY_UTILIZATION:
			gpuStat.MemCopyUtil = utilizationRates.Memory
		case GPU_ENC_UTILIZATION:
			gpuStat.EncoderUtil, _, _ = g.GetEncoderUtilization()
		case GPU_DEC_UTILIZATION:
			gpuStat.DecoderUtil, _, _ = g.GetEncoderUtilization()
		case GPU_MEMORY_FREE:
			gpuStat.MemoryFree = memoryInfo.Free
		case GPU_MEMORY_USED:
			gpuStat.MemoryUsed = memoryInfo.Used

		}
	}
	return gpuStat
}

func (g *GPUDevice) GetProcessStat(useSlurm bool) []ProcessStat {
	psInfos, _ := g.GetProcessUtilization(0)
	ret := make([]ProcessStat, 0)
	for _, psInfo := range psInfos {
		// gpu related
		ps := ProcessStat{
			Pid:      psInfo.Pid,
			GPUIndex: int(g.GPUIndex),
			Smutil:   psInfo.SmUtil,
			Memutil:  psInfo.MemUtil,
			Decutil:  psInfo.DecUtil,
			Encutil:  psInfo.EncUtil,
		}
		ps.UpdateProcessInfoCPU(useSlurm)
		ret = append(ret, ps)
	}
	return ret
}

func (ps *ProcessStat) UpdateProcessInfoCPU(useSlurm bool) {
	proc, err := process.NewProcess(int32(ps.Pid))
	if err != nil {
		logrus.Errorf("unable to get process, pid:%d", ps.Pid)
		return
	}

	procName, _ := proc.Name()
	userName, _ := proc.Username()
	cpuPercent, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()
	memUsedBytes := memInfo.RSS
	numThreads, _ := proc.NumThreads()

	ps.ProcName = procName
	ps.User = userName
	ps.CPUPercent = cpuPercent
	ps.CPUMemoryUsedBytes = memUsedBytes
	ps.NumThreads = numThreads

	// slurm realted
	if useSlurm {
		envs, _ := proc.Environ()
		for _, env := range envs {
			kvPair := strings.Split(env, "=")
			key, value := kvPair[0], kvPair[1]
			switch key {
			case SLURM_ENV_JOBID:
				ps.SlurmJobID = value
			case SLURM_ENV_STEP_ID:
				ps.SlurmStepID = value
			case SLURM_ENV_USER:
				ps.SlurmUser = value
			case SLURM_ENV_ACCOUNT:
				ps.SlurmAccount = value
			case SLURM_ENV_JOBNAME:
				ps.SlurmJobName = value
			}
		}
	}
}

func (ps *ProcessStat) GetValueFromMetricName(metricName string) float64 {
	switch metricName {
	case PROCESS_CPU_PERCENT:
		return float64(ps.CPUPercent)
	case PROCESS_CPU_MEM_USED_BYTES:
		return float64(ps.CPUMemoryUsedBytes)
	case PROCESS_GPU_SM_UTIL:
		return float64(ps.Smutil)
	case PROCESS_GPU_MEM_UTIL:
		return float64(ps.Memutil)
	case PROCESS_GPU_DECODE_UTIL:
		return float64(ps.Decutil)
	case PROCESS_GPU_ENCODE_UTIL:
		return float64(ps.Encutil)
	default:
		return 0
	}
}

func (gpu *GPUStat) GetValueFromMetricName(metricName string) float64 {
	// [x]: add value conversion from consts.go metricName
	switch metricName {
	case GPU_SM_CLOCK:
		return float64(gpu.SMClock)
	case GPU_MEMORY_CLOCK:
		return float64(gpu.MemClock)
	case GPU_TEMPERATURE:
		return float64(gpu.Temperature)
	case GPU_FAN_SPEED:
		return float64(gpu.FanSpeed)
	case GPU_POWER_USAGE:
		return float64(gpu.PowerUsage)
	case GPU_TOTAL_ENERGY_CONSUMPTION:
		return float64(gpu.TotalEnergyConsumption)
	case GPU_PCIE_TX_BYTES:
		return float64(gpu.PCIETXBytes)
	case GPU_PCIE_RX_BYTES:
		return float64(gpu.PCIERXBytes)
	case GPU_UTILIZATION:
		return float64(gpu.GPUUtil)
	case GPU_MEM_COPY_UTILIZATION:
		return float64(gpu.MemCopyUtil)
	case GPU_ENC_UTILIZATION:
		return float64(gpu.EncoderUtil)
	case GPU_DEC_UTILIZATION:
		return float64(gpu.DecoderUtil)
	case GPU_MEMORY_FREE:
		return float64(gpu.MemoryFree)
	case GPU_MEMORY_USED:
		return float64(gpu.MemoryUsed)
	default:
		return 0
	}
}
