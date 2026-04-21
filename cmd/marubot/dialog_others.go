//go:build !windows && !darwin
package main

import (
	"fmt"
)

func showNativeConfirmDialog(title, message string) bool {
	// For Linux server/CLI environments, we default to true 
	// because GUI dialogs are often not available.
	// You might want to implement zenity or kdialog support here in the future.
	fmt.Printf("[%s] %s (Auto-confirmed as true in CLI mode)\n", title, message)
	return true
}

func showNativeMessageDialog(title, message string) {
	fmt.Printf("[%s] %s\n", title, message)
}
