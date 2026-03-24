package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home := os.Getenv("MARUBOT_HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	path := filepath.Join(home, ".marubot", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("File not found or error: %v\n", err)
		return
	}
	fmt.Println("--- CONFIG.JSON START ---")
	fmt.Println(string(data))
	fmt.Println("--- CONFIG.JSON END ---")
}
