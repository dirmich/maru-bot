//go:build !linux && !windows
// +build !linux,!windows

package dashboard

import "runtime"

func getPlatformStats() map[string]interface{} {
	osName := "Others"
	switch runtime.GOOS {
	case "windows":
		osName = "Windows"
	case "darwin":
		osName = "macOS"
	}

	return map[string]interface{}{
		"os":              osName,
		"is_raspberry_pi": false,
	}
}
