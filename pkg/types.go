package collector

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Counter struct {
	FieldName string
	PromType  string
	Help      string
}

type ProcessStat struct {
	Pid      uint   `json:"pid"`
	ProcName string `json:"name"`
	GPUIndex int    `json:"gpu"`
	UserName string `json:"username"`

	CPUPercent    uint32 `json:"cpupercent"`
	CPUMemoryUsed uint32 `json:"cpumem"`
	GPUMemoryUsed uint64 `json:"gpumem"`

	Smutil       float64 `json:"smutil"`       // SM利用率
	Memutil      float64 `json:"memutil"`      // 显存利用率
	FrameMemUtil float64 `json:"framememutil"` // 帧缓冲区内存利用率
	Decutil      float64 `json:"decutil"`
	Encutil      float64 `json:"encutil"`
}

type GPUDevice struct {
	nvml.Device

	UUID         string
	GPUModelName string
	GPUIndex     uint
	Attributes   nvml.DeviceAttributes

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
	PCI     nvml.PciInfo
	PSUInfo nvml.PSUInfo

	GPUUtil    uint32
	MemoryUtil uint32
	MemoryFree uint64
	MemoryUsed uint64
}

type Config struct {
	CollectorsFile  string
	Address         string
	CollectInterval int
	UseSlurm        bool
	ConfigMapData   string
	HostName        string
}
