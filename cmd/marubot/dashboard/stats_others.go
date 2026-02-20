//go:build !linux
// +build !linux

package dashboard

func getPlatformStats() map[string]interface{} {
	return map[string]interface{}{
		"os": "Unknown",
	}
}
