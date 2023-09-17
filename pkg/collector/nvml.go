/*
 * Copyright (c) 2021, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package collector

import (
	"log"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/sirupsen/logrus"
)

type NVMLCache struct {
	sync.RWMutex
	DeviceCount  uint
	DeviceInfos  []GPUDevice
	GPUStats     []GPUStat
	ProcessStats map[uint]ProcessStat
	Hostname     string
	config       *Config
}

func NewNVMLCache(config *Config) (*NVMLCache, error) {
	logrus.Infof("NVML metrics collection enabled!")

	ret := nvml.Init()

	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to init NVML: %v", nvml.ErrorString(ret))
	}

	// 获取GPU数量
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}

	// 初始化GPU设备信息
	deviceInfos := make([]GPUDevice, count)

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		deviceInfos[i].Device = device
		deviceInfos[i].UUID, _ = device.GetUUID()
		deviceInfos[i].GPUIndex = uint(i)
		deviceInfos[i].GPUModelName, _ = device.GetName()
		deviceInfos[i].Attributes, _ = device.GetAttributes()
		// deviceInfos[i].Attributes = DeviceAttributes{
		// 	MultiprocessorCount:       attr.MultiprocessorCount,
		// 	SharedCopyEngineCount:     attr.SharedCopyEngineCount,
		// 	SharedDecoderCount:        attr.SharedDecoderCount,
		// 	SharedEncoderCount:        attr.SharedEncoderCount,
		// 	SharedJpegCount:           attr.SharedJpegCount,
		// 	SharedOfaCount:            attr.SharedOfaCount,
		// 	GpuInstanceSliceCount:     attr.GpuInstanceSliceCount,
		// 	ComputeInstanceSliceCount: attr.ComputeInstanceSliceCount,
		// 	MemorySizeMB:              attr.MemorySizeMB,
		// }
		deviceInfos[i].PcieLinkMaxSpeed, _ = device.GetPcieLinkMaxSpeed()
	}

	cache := &NVMLCache{
		DeviceInfos:  deviceInfos,
		DeviceCount:  uint(count),
		GPUStats:     make([]GPUStat, count),
		ProcessStats: make(map[uint]ProcessStat),
		Hostname:     config.HostName,
		config:       config,
	}

	return cache, nil
}

func (c *NVMLCache) Run(stop chan interface{}) {
	t := time.NewTicker(time.Second * time.Duration(c.config.CollectInterval))
	defer nvml.Shutdown()
	defer t.Stop()
	c.udpateCache()
	for {
		select {
		case <-stop:
			logrus.Infof("Shutdown nvml cache...")
			return
		case <-t.C:
			err := c.udpateCache()
			// logrus.Infof("Updating nvml cache...")
			if err != nil {
				logrus.Errorf("Failed to collect metrics with error: %v", err)
				/* flush output rather than output stale data */
				continue
			}
		}
	}
}

func (c *NVMLCache) udpateCache() error {

	start := time.Now()
	newProcStat := make(map[uint]ProcessStat)
	newGPUStat := make([]GPUStat, c.DeviceCount)
	for i, devcie := range c.DeviceInfos {
		// fixme: pcie带宽获取速度很慢
		// 更新GPUStat
		// s := time.Now()
		newGPUStat[i] = devcie.DeviceGetGPUStat(SupportedGGPUMetricsName)
		// logrus.Infof("get gpu stat time: %v", time.Since(s))
		// 更新ProcStat
		// s = time.Now()
		psStats := devcie.GetProcessStat(c.config.UseSlurm)
		for _, ps := range psStats {
			newProcStat[uint(ps.Pid)] = ps
		}
		// logrus.Infof("get proc stat time: %v", time.Since(s))
	}
	// var wg sync.WaitGroup
	// gpuStatCh := make(chan GPUStat)
	// procStatCh := make(chan map[uint]ProcessStat)

	// for i, device := range c.DeviceInfos {
	// 	wg.Add(1)
	// 	go func(i int, device GPUDevice) {
	// 		defer wg.Done()

	// 		// 更新 GPUStat
	// 		newGPUStat[i] = device.DeviceGetGPUStat(SupportedGGPUMetricsName)

	// 		// 更新 ProcStat
	// 		psStats := device.GetProcessStat(c.config.UseSlurm)
	// 		procStatCh <- psStats
	// 	}(i, device)
	// }

	// go func() {
	// 	wg.Wait()
	// 	close(gpuStatCh)
	// 	close(procStatCh)
	// }()

	// // 从通道接收并处理 ProcStat
	// for psStats := range procStatCh {
	// 	for _, ps := range psStats {
	// 		newProcStat[uint(ps.Pid)] = ps
	// 	}
	// }

	c.Lock()
	c.GPUStats = newGPUStat
	c.ProcessStats = newProcStat
	c.Unlock()
	logrus.Debugf("udpate nvml cache time: %v", time.Since(start))
	return nil
}

// get cache snapshot
func (c *NVMLCache) GetProcessStats() map[uint]ProcessStat {
	snapshot := make(map[uint]ProcessStat)
	c.Lock()
	for pid, stat := range c.ProcessStats {
		snapshot[pid] = stat
	}
	c.Unlock()
	return snapshot
}

// get cache snapshot
func (c *NVMLCache) GetGPUStats() []GPUStat {
	snapshot := make([]GPUStat, c.DeviceCount)
	c.Lock()
	for i, stat := range c.GPUStats {
		snapshot[i] = stat
	}
	c.Unlock()
	return snapshot
}

func (c *NVMLCache) GetGPUInfos() []GPUInfo {
	snapshot := make([]GPUInfo, c.DeviceCount)

	for i, d := range c.DeviceInfos {
		snapshot[i] = d.GPUInfo
	}
	return snapshot
}
