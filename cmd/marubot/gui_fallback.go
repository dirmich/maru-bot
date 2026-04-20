//go:build !windows && !darwin
package main

import (
	"fmt"
	"os"
)

func handleGUIMode() {
	fmt.Println("GUI mode is not supported on this platform.")
	printHelp()
	os.Exit(1)
}
