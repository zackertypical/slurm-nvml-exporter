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

type GPUDevice struct {
	nvml.Device

	UUID             string
	GPUModelName     string
	GPUIndex         uint
	Attributes       nvml.DeviceAttributes
	PcieLinkMaxSpeed uint32

	// Model                 *string
	// Power                 *uint
	// Memory                *uint64
	// CPUAffinity           *uint
	// PCI                   PCIInfo
	// Clocks                ClockInfo
	// Topology              []P2PLink
	// CudaComputeCapability CudaComputeCapabilityInfo
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

type Config struct {
	CollectorsFile  string
	Address         string
	CollectInterval int
	UseSlurm        bool
	ConfigMapData   string
	HostName        string
}

// todo: configuration
func (g *GPUDevice) DeviceGetGPUStat() GPUStat {
	// 获取SM clock frequency
	smClock, _ := g.GetClockInfo(nvml.CLOCK_SM)

	// 获取Memory clock frequency
	memClock, _ := g.GetClockInfo(nvml.CLOCK_MEM)

	// 获取Power draw
	powerUsage, _ := g.GetPowerUsage()

	// 获取Total energy consumption since boot
	energyConsumption, _ := g.GetTotalEnergyConsumption()

	// 获取Fan speed
	fanSpeed, _ := g.GetFanSpeed()

	// 获取Temperature
	temperature, _ := g.GetTemperature(nvml.TEMPERATURE_GPU)

	// 获取GPU utilization rate
	utilizationRates, _ := g.GetUtilizationRates()

	// 获取GPU encoderUtil, decoderUtil
	encoderUtil, _, _ := g.GetEncoderUtilization()
	decoderUtil, _, _ := g.GetEncoderUtilization()

	// 获取Memory utilization rate
	memoryInfo, _ := g.GetMemoryInfo()

	// 获取PCIe TX and RX bytes per second
	// https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1gd86f1c74f81b5ddfaa6cb81b51030c72
	pcieThroughputTX, _ := g.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES)
	pcieThroughputRX, _ := g.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES)
	return GPUStat{
		GPUIndex:               g.GPUIndex,
		UUID:                   g.UUID,
		GPUModelName:           g.GPUModelName,
		SMClock:                smClock,
		MemClock:               memClock,
		PowerUsage:             powerUsage / 1000,        // 转换为W
		TotalEnergyConsumption: energyConsumption * 1000, // 转换为mJ
		FanSpeed:               fanSpeed,
		Temperature:            temperature,
		GPUUtil:                utilizationRates.Gpu,
		EncoderUtil:            encoderUtil,
		DecoderUtil:            decoderUtil,
		MemoryUtil:             utilizationRates.Memory,
		MemoryFree:             memoryInfo.Free,
		MemoryUsed:             memoryInfo.Used,
		PCIETXBytes:            pcieThroughputTX * 1024, // KB/s 转换为bytes per second
		PCIERXBytes:            pcieThroughputRX * 1024, // KB/s 转换为bytes per second
	}

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
	// todo: add value conversion from consts.go metricName
	switch metricName {

	default:
		return 0
	}
}
