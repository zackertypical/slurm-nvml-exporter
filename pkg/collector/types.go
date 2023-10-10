/*
** API Reference Doc: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html
**
 */

package collector

import (
	"fmt"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
)

const (
	LabelHostName = "Hostname"
)

type Config struct {
	ServerPort       string
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

// type DeviceAttributes struct {
// 	MultiprocessorCount       uint32 `json:"multiprocessorCount"`
// 	SharedCopyEngineCount     uint32 `json:"sharedCopyEngineCount"`
// 	SharedDecoderCount        uint32 `json:"sharedDecoderCount"`
// 	SharedEncoderCount        uint32 `json:"sharedEncoderCount"`
// 	SharedJpegCount           uint32 `json:"sharedJpegCount"`
// 	SharedOfaCount            uint32 `json:"sharedOfaCount"`
// 	GpuInstanceSliceCount     uint32 `json:"gpuInstanceSliceCount"`
// 	ComputeInstanceSliceCount uint32 `json:"computeInstanceSliceCount"`
// 	MemorySizeMB              uint64 `json:"memorySizeMB"`
// }

type ProcessStat struct {
	Pid         uint32 `json:"pid"`
	GPUIndex    int    `json:"gpu"`
	ProcName    string `json:"procName"`
	User        string `json:"user"`
	Status      string `json:"status"`
	PPid        uint32 `json:"ppid"`
	WorkingDir  string `json:"workingDir"`
	CommandLine string `json:"commandLine"`

	// CPU Metrics
	CPUPercent         float64 `json:"cpu_percent"`
	CPUMemoryUsedBytes uint64  `json:"cpu_mem_used_bytes"`
	NumThreads         int32   `json:"num_threads"`

	// TODO:
	// IOCounters
	// NetIOCounters
	// Status

	// GPU Metrics
	Smutil             uint32 `json:"smutil"`  // SM利用率
	Memutil            uint32 `json:"memutil"` // 显存利用率
	Decutil            uint32 `json:"decutil"`
	Encutil            uint32 `json:"encutil"`
	GPUUsedMemoryBytes uint64 `json:"gpu_used_memory_bytes"`

	// Slurm Lables
	SlurmProcInfo
}

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

	// MemoryUtil uint32 `json:"mem_util"`
	MemoryFreeBytes uint64 `json:"mem_free_bytes"`
	MemoryUsedBytes uint64 `json:"mem_used_bytes"`
}

// [x]: configuration
// DeviceGetGPUStat Only gets the metric from arg metrics
func (g *GPUDevice) DeviceGetGPUStat(metrics []string) GPUStat {
	gpuStat := GPUStat{
		GPUIndex:     g.GPUIndex,
		UUID:         g.UUID,
		GPUModelName: g.GPUModelName,
	}
	utilizationRates, ret := g.GetUtilizationRates()
	if ret != nvml.SUCCESS {
		logrus.Errorf("cannot get utilizationRates of gpu:%v", g.GPUIndex)
	}
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
		case GPU_MEMORY_FREE_BYTES:
			gpuStat.MemoryFreeBytes = memoryInfo.Free
		case GPU_MEMORY_USED_BYTES:
			gpuStat.MemoryUsedBytes = memoryInfo.Used
			// case GPU_NVLINK_RX_BYTES:
			// 	rxCounter, _, _ := g.GetNvLinkUtilizationCounter(0,0)
			// 	_, txCounter, _ := g.GetNvLinkUtilizationCounter(0,1)

		}
	}
	return gpuStat
}

func (g *GPUDevice) GetProcessStat(useSlurm bool) map[uint]ProcessStat {

	retMap := make(map[uint]ProcessStat)
	computeProcs, ret := g.GetComputeRunningProcesses()
	if ret != nvml.SUCCESS {
		return retMap
	}
	// fixme: process infos 不全？？
	utilProcs, ret := g.GetProcessUtilization(0)
	// logrus.Infof("gpu:%d, psInfos:%+v", g.GPUIndex, psInfos)
	if ret != nvml.SUCCESS {
		return retMap
	}

	// update gpu mem
	for _, proc := range computeProcs {
		if proc.Pid < 1 {
			continue
		}
		ps := ProcessStat{
			Pid:                proc.Pid,
			GPUIndex:           int(g.GPUIndex),
			GPUUsedMemoryBytes: proc.UsedGpuMemory,
		}
		err := ps.UpdateProcessInfoCPU(useSlurm)
		if err != nil {
			continue
		}
		retMap[uint(proc.Pid)] = ps
		// logrus.Infof("gpu:%d, psInfo:%+v", g.GPUIndex, ps)
	}

	// update util
	for _, proc := range utilProcs {
		if proc.Pid < 1 {
			continue
		}
		logrus.Debugf("gpu:%d, psInfo:%+v", g.GPUIndex, proc)
		if p, ok := retMap[uint(proc.Pid)]; ok {
			p.Smutil = proc.SmUtil
			p.Memutil = proc.MemUtil
			p.Decutil = proc.DecUtil
			p.Encutil = proc.EncUtil
			retMap[uint(proc.Pid)] = p
		}
	}
	return retMap
}

func (ps *ProcessStat) UpdateProcessInfoCPU(useSlurm bool) error {
	proc, err := process.NewProcess(int32(ps.Pid))
	if err != nil {
		// logrus.Errorf("unable to get process, pid:%d", ps.Pid)
		return fmt.Errorf("unable to get process, pid:%d, err:%v", ps.Pid, err)
	}
	running, _ := proc.IsRunning()
	if !running {
		return fmt.Errorf("process is not running, pid:%d", ps.Pid)
	}
	ps.Status, _ = proc.Status()
	ppid, _ := proc.Ppid()
	ps.PPid = uint32(ppid)
	ps.ProcName, _ = proc.Name()
	ps.CommandLine, _ = proc.Cmdline()
	ps.WorkingDir, _ = proc.Cwd()
	ps.User, _ = proc.Username()
	ps.CPUPercent, _ = proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()
	ps.CPUMemoryUsedBytes = memInfo.RSS
	ps.NumThreads, _ = proc.NumThreads()

	// slurm realted
	if useSlurm {
		envs, _ := proc.Environ()
		for _, env := range envs {
			kvPair := strings.Split(env, "=")
			if len(kvPair) != 2 {
				continue
			}
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
	return nil
}

func (ps *ProcessStat) GetValueFromMetricName(metricName string) float64 {
	switch metricName {
	case PROCESS_INFO:
		return 1
	case PROCESS_CPU_PERCENT:
		return float64(ps.CPUPercent)
	case PROCESS_CPU_MEM_USED_BYTES:
		return float64(ps.CPUMemoryUsedBytes)
	case PROCESS_NUM_THREADS:
		return float64(ps.NumThreads)
	case PROCESS_GPU_SM_UTIL:
		return float64(ps.Smutil)
	case PROCESS_GPU_MEM_UTIL:
		return float64(ps.Memutil)
	case PROCESS_GPU_DECODE_UTIL:
		return float64(ps.Decutil)
	case PROCESS_GPU_ENCODE_UTIL:
		return float64(ps.Encutil)
	case PROCESS_GPU_MEM_USED_BYTES:
		return float64(ps.GPUUsedMemoryBytes)
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
	case GPU_MEMORY_FREE_BYTES:
		return float64(gpu.MemoryFreeBytes)
	case GPU_MEMORY_USED_BYTES:
		return float64(gpu.MemoryUsedBytes)
	default:
		return 0
	}
}
