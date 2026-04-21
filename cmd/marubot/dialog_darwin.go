//go:build darwin
package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func showNativeConfirmDialog(title, message string) bool {
	// Escape double quotes in message
	msg := strings.ReplaceAll(message, `"`, `\"`)
	script := fmt.Sprintf(`display dialog "%s" with title "%s" buttons {"Cancel", "OK"} default button "OK" with icon caution`, msg, title)
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	return err == nil
}

func showNativeMessageDialog(title, message string) {
	// Escape double quotes in message
	msg := strings.ReplaceAll(message, `"`, `\"`)
	script := fmt.Sprintf(`display alert "%s" message "%s" as informational`, title, msg)
	exec.Command("osascript", "-e", script).Run()
}
