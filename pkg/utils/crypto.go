package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"
)

// GetMachineID returns a unique machine identifier for Windows
func GetMachineID() string {
	// Get MachineGuid from registry
	cmd := exec.Command("powershell", "-NoProfile", "Get-ItemProperty -Path HKLM:\\SOFTWARE\\Microsoft\\Cryptography -Name MachineGuid | Select-Object -ExpandProperty MachineGuid")
	out, err := cmd.Output()
	if err != nil {
		return "static-marubot-salt" // Very basic fallback
	}
	return strings.TrimSpace(string(out))
}

// HashPassword hashes a password with a machine-specific salt
func HashPassword(password string) string {
	machineID := GetMachineID()
	hash := sha256.New()
	// Using a combination of password and machineID as salt
	hash.Write([]byte(password + ":" + machineID))
	return hex.EncodeToString(hash.Sum(nil))
}

// IsPasswordHashed checks if the password string is likely a SHA256 hash
func IsPasswordHashed(password string) bool {
	if len(password) != 64 {
		return false
	}
	for _, c := range password {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
