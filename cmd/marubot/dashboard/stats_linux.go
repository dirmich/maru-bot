//go:build linux
// +build linux

package dashboard

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

func getPlatformStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 1. Uptime
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err == nil {
		var up float64
		fmt.Sscanf(string(uptimeData), "%f", &up)
		stats["uptime"] = int64(up)
	}

	// 2. Memory
	memData, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		lines := strings.Split(string(memData), "\n")
		var total, free, available int64
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d", &total)
			} else if strings.HasPrefix(line, "MemFree:") {
				fmt.Sscanf(line, "MemFree: %d", &free)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d", &available)
			}
		}
		stats["memory"] = map[string]int64{
			"total":     total * 1024,
			"free":      free * 1024,
			"available": available * 1024,
			"used":      (total - available) * 1024,
		}
	}

	// 3. Disk (Root)
	var disk syscall.Statfs_t
	if err := syscall.Statfs("/", &disk); err == nil {
		stats["disk"] = map[string]uint64{
			"total": disk.Blocks * uint64(disk.Bsize),
			"free":  disk.Bfree * uint64(disk.Bsize),
			"used":  (disk.Blocks - disk.Bfree) * uint64(disk.Bsize),
		}
	}

	// 4. CPU (Load Avg)
	loadData, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		var l1, l5, l15 float64
		fmt.Sscanf(string(loadData), "%f %f %f", &l1, &l5, &l15)
		stats["cpu"] = map[string]float64{
			"load1":  l1,
			"load5":  l5,
			"load15": l15,
		}
	}

	stats["os"] = "Linux"
	return stats
}
