//go:build !linux && !windows
// +build !linux,!windows

package dashboard

import (
	"runtime"
	"github.com/shirou/gopsutil/v3/host"
)

func getPlatformStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	hInfo, _ := host.Info()
	if hInfo != nil {
		stats["os"] = hInfo.OS + " " + hInfo.PlatformVersion
		stats["hw_model"] = hInfo.Hostname + " (" + hInfo.KernelArch + ")"
		stats["uptime"] = hInfo.Uptime
	} else {
		osName := "Others"
		switch runtime.GOOS {
		case "windows":
			osName = "Windows"
		case "darwin":
			osName = "macOS"
		}
		stats["os"] = osName
		stats["hw_model"] = runtime.GOARCH + " " + osName + " Device"
	}
	
	stats["is_raspberry_pi"] = false
	return stats
}
