//go:build windows || darwin
package main

import (
	"fmt"
	"os"
	"os/exec"
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
		systray.SetIcon(trayIconIco)
	} else if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(trayIconMacPng, trayIconMacPng)
	} else {
		systray.SetIcon(trayIconPng)
	}

	mDashboard := systray.AddMenuItem("Dashboard", "Open Web-Admin")
	systray.AddSeparator()
	mStart := systray.AddMenuItem("Start MaruBot", "Start the service")
	mStop := systray.AddMenuItem("Stop MaruBot", "Stop the service")
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "Show version info")
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
			fmt.Println("Starting MaruBot backend...")
			startCmd()
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
			case <-mAbout.ClickedCh:
				showAboutDialog()
			case <-mStart.ClickedCh:
				fmt.Println("Starting MaruBot service...")
				if runtime.GOOS == "windows" {
					serviceCmdInternalPath("start", targetExe)
				} else {
					startCmd()
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
				if showConfirmDialog("Uninstall MaruBot", "Are you sure you want to completely remove MaruBot and its resources?") {
					fmt.Println("Uninstalling MaruBot...")
					if runtime.GOOS == "windows" {
						runAsAdminAction("uninstall --yes")
					} else {
						// Pass --yes flag to skip CLI prompt
						os.Args = append(os.Args, "--yes")
						uninstallCmd()
					}
					systray.Quit()
					os.Exit(0)
				} else {
					fmt.Println("Uninstall aborted by user.")
				}
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

func showConfirmDialog(title, message string) bool {
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf("display dialog %q with title %q buttons {\"Cancel\", \"OK\"} default button \"Cancel\" with icon caution", message, title)
		cmd := exec.Command("osascript", "-e", script)
		err := cmd.Run()
		return err == nil
	} else if runtime.GOOS == "windows" {
		// Use PowerShell for a simple message box
		vbs := fmt.Sprintf("res = MsgBox(\"%s\", 1+48, \"%s\"): If res = 1 Then WScript.Quit(0) Else WScript.Quit(1)", message, title)
		cmd := exec.Command("wscript", "-e", "vbscript", "-e", vbs)
		// Or simpler with PowerShell:
		ps := fmt.Sprintf("[System.Windows.Forms.MessageBox]::Show('%s', '%s', 'OKCancel', 'Warning')", message, title)
		cmd = exec.Command("powershell", "-Command", "Add-Type -AssemblyName System.Windows.Forms;", "if ("+ps+" -eq 'OK') { exit 0 } else { exit 1 }")
		err := cmd.Run()
		return err == nil
	}
	return true // Fallback for other OS
}

func showAboutDialog() {
	title := "About MaruBot"
	message := fmt.Sprintf("MaruBot - AI Agent Service\n\nVersion: %s\n\n© 2026 MaruBot Contributors", Version)

	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf("display alert %q message %q as informational buttons {\"OK\"} default button \"OK\"", title, message)
		exec.Command("osascript", "-e", script).Run()
	} else if runtime.GOOS == "windows" {
		ps := fmt.Sprintf("[System.Windows.Forms.MessageBox]::Show('%s', '%s', 'OK', 'Information')", message, title)
		exec.Command("powershell", "-Command", "Add-Type -AssemblyName System.Windows.Forms;", ps).Run()
	} else {
		fmt.Printf("--- %s ---\n%s\n", title, message)
	}
}
