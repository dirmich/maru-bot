//go:build windows
// +build windows

package dashboard

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func getPlatformStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 1. Host Info (Uptime, OS)
	hInfo, _ := host.Info()
	if hInfo != nil {
		stats["uptime"] = hInfo.Uptime
		stats["os"] = fmt.Sprintf("%s %s", hInfo.OS, hInfo.PlatformVersion)
	} else {
		stats["os"] = "Windows"
	}

	// 2. CPU
	// LoadAvg for Windows isn't direct, use CPU percent as a fallback or mock load
	cpuPercent, _ := cpu.Percent(time.Second, false)
	var load float64
	if len(cpuPercent) > 0 {
		load = cpuPercent[0]
	}
	stats["cpu"] = map[string]float64{
		"load1":  load, // Using current usage as load1 for simplicity
		"load5":  load,
		"load15": load,
	}

	// CPU Details
	cpuInfo, _ := cpu.Info()
	if len(cpuInfo) > 0 {
		stats["cpu_detail"] = map[string]interface{}{
			"model":   cpuInfo[0].ModelName,
			"cores":   runtime.NumCPU(),
			"vendor":  cpuInfo[0].VendorID,
			"mhz":     cpuInfo[0].Mhz,
			"percent": cpuPercent,
		}
	}

	// 3. Memory
	vMem, _ := mem.VirtualMemory()
	if vMem != nil {
		stats["memory"] = map[string]uint64{
			"total":     vMem.Total,
			"free":      vMem.Free,
			"available": vMem.Available,
			"used":      vMem.Used,
		}
		stats["memory_detail"] = map[string]interface{}{
			"total":     vMem.Total,
			"available": vMem.Available,
			"used":      vMem.Used,
			"free":      vMem.Free,
			"cached":    vMem.Cached,
			"percent":   vMem.UsedPercent,
		}
	}

	// 4. Disk
	usage, _ := disk.Usage("/")
	if usage != nil {
		stats["disk"] = map[string]uint64{
			"total": usage.Total,
			"free":  usage.Free,
			"used":  usage.Used,
		}
	}

	// Disk Details (All Partitions)
	partitions, _ := disk.Partitions(true)
	var diskDetails []map[string]interface{}
	for _, p := range partitions {
		u, _ := disk.Usage(p.Mountpoint)
		if u != nil {
			diskDetails = append(diskDetails, map[string]interface{}{
				"device":     p.Device,
				"mountpoint": p.Mountpoint,
				"fstype":     p.Fstype,
				"total":      u.Total,
				"free":       u.Free,
				"used":       u.Used,
				"percent":    u.UsedPercent,
			})
		}
	}
	stats["disk_detail"] = diskDetails

	return stats
}
