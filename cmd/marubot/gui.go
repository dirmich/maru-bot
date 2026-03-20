//go:build windows || darwin
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/kardianos/service"
)

func handleGUIMode() {
	// Check for elevation flag (Windows specific mostly)
	isElevated := false
	for _, arg := range os.Args {
		if arg == "--elevated" {
			isElevated = true
			break
		}
	}

	// 1. Install/Setup check
	if runtime.GOOS == "windows" {
		if !isAdmin() && !isElevated {
			targetExe, _ := getTargetBinaryPath()
			if _, err := os.Stat(targetExe); err != nil {
				fmt.Println("Elevation required for installation. Requesting administrator privileges...")
				runAsAdmin()
				os.Exit(0)
			}
		}
	}

	// 0. Initialization
	setupWorkspace()

	targetExe, _ := installBinary()

	// 2. Start Tray Icon
	systray.Run(func() { onTrayReady(targetExe) }, onTrayExit)
}

func onTrayReady(targetExe string) {
	systray.SetTitle("MaruBot")
	systray.SetTooltip("MaruBot - AI Agent Service")
	if runtime.GOOS == "windows" {
		systray.SetIcon(windowTrayIco)
	} else {
		systray.SetIcon(macMenubarPng)
	}

	systray.AddMenuItem("MaruBot v"+Version, "Version information").Disable()
	systray.AddSeparator()
	mDashboard := systray.AddMenuItem("Dashboard", "Open Web-Admin")
	systray.AddSeparator()
	mStart := systray.AddMenuItem("Start MaruBot", "Start the service")
	mStop := systray.AddMenuItem("Stop MaruBot", "Stop the service")
	systray.AddSeparator()
	mUninstall := systray.AddMenuItem("Uninstall MaruBot", "Remove service and cleanup")
	mExit := systray.AddMenuItem("Exit", "Exit tray only")

	// 3. Initial check and install
	cfg, _ := loadConfig()
	if !checkAndFixPort(cfg) {
		systray.Quit()
		return
	}

	// Service management (Currently prioritized for Windows/Linux)
	if runtime.GOOS == "windows" {
		svcConfig := &service.Config{Name: "MaruBot"}
		s, _ := service.New(&program{}, svcConfig)
		status, _ := s.Status()

		if status == service.StatusUnknown {
			serviceCmdInternalPath("install", targetExe)
			time.Sleep(1 * time.Second)
			serviceCmdInternalPath("start", targetExe)
		} else if status != service.StatusRunning {
			serviceCmdInternalPath("start", targetExe)
		}
	} else if runtime.GOOS == "darwin" {
		// On macOS, if it's not running, start it
		if !isMarubotProcessRunning() {
			fmt.Println("Starting MaruBot backend in background...")
			go startCmd()
		}
	}

	// Auto-open browser on first start
	openBrowser(fmt.Sprintf("http://localhost:%d", cfg.Gateway.Port))

	// Menu Handlers
	go func() {
		for {
			select {
			case <-mDashboard.ClickedCh:
				currCfg, _ := loadConfig()
				openBrowser(fmt.Sprintf("http://localhost:%d", currCfg.Gateway.Port))
			case <-mStart.ClickedCh:
				fmt.Println("Starting MaruBot service...")
				if runtime.GOOS == "windows" {
					serviceCmdInternalPath("start", targetExe)
				} else {
					go startCmd()
				}
				time.Sleep(2 * time.Second) // Give time to start
			case <-mStop.ClickedCh:
				fmt.Println("Stopping MaruBot service...")
				if runtime.GOOS == "windows" {
					serviceCmdInternalPath("stop", targetExe)
				} else {
					stopCmd()
				}
				time.Sleep(1 * time.Second)
			case <-mUninstall.ClickedCh:
				fmt.Println("Uninstalling MaruBot...")
				if runtime.GOOS == "windows" {
					runAsAdminAction("uninstall")
				} else {
					uninstallCmd()
				}
				systray.Quit()
				os.Exit(0)
			case <-mExit.ClickedCh:
				fmt.Println("Exiting tray...")
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

func onTrayExit() {
	// Cleanup on exit
}
