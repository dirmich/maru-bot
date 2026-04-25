//go:build windows || darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/dirmich/marubot/pkg/config"
	"github.com/getlantern/systray"
	"github.com/kardianos/service"
)

// hideConsole is now implemented in platform-specific sys_*.go files

func handleGUIMode() {
	hideConsole()

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

var (
	mDashboard  *systray.MenuItem
	mCreateLink *systray.MenuItem
	mStart      *systray.MenuItem
	mStop       *systray.MenuItem
	mUninstall  *systray.MenuItem
	mExit       *systray.MenuItem
	mUpgrade    *systray.MenuItem
	versionItem *systray.MenuItem
)

type TrayLabels struct {
	Dashboard      string
	CreateLink     string
	Start          string
	Stop           string
	Uninstall      string
	Exit           string
	Upgrade        string
	StatusInstall  string
	StatusRun      string
	ServiceStarted string
	ServiceStopped string
	NotifyTitle    string
	NotifyBgMsg    string
	LinkCreated    string
}

var trayLocales = map[string]TrayLabels{
	"en": {
		Dashboard:      "Open Dashboard",
		CreateLink:     "Create Desktop Shortcut",
		Start:          "Start Service",
		Stop:           "Stop Service",
		Uninstall:      "Uninstall MaruBot",
		Exit:           "Exit Tray (Keep Service)",
		StatusInstall:  "Installation Complete",
		StatusRun:      "MaruBot service is installed and started.",
		ServiceStarted: "MaruBot service has started.",
		ServiceStopped: "MaruBot service has stopped.",
		NotifyTitle:    "MaruBot Notice",
		NotifyBgMsg:    "Tray exited. MaruBot is still running in the background.",
		LinkCreated:    "Shortcuts created on Desktop and Start Menu.",
		Upgrade:        "Check for Upgrade",
	},
	"ko": {
		Dashboard:      "대시보드 열기",
		CreateLink:     "바탕화면에 바로가기 생성",
		Start:          "서비스 시작",
		Stop:           "서비스 중지",
		Uninstall:      "프로그램 제거",
		Exit:           "트레이 종료 (서비스 유지)",
		StatusInstall:  "설치 완료",
		StatusRun:      "MaruBot 서비스가 설치되고 시작되었습니다.",
		ServiceStarted: "MaruBot 서버가 시작되었습니다.",
		ServiceStopped: "MaruBot 서버가 중지되었습니다.",
		NotifyTitle:    "MaruBot 안내",
		NotifyBgMsg:    "트레이 아이콘이 종료되었습니다. 서비스는 백그라운드에서 계속 실행 중입니다.",
		LinkCreated:    "바탕화면과 시작 메뉴에 바로가기가 생성되었습니다.",
		Upgrade:        "업그레이드 확인",
	},
	"ja": {
		Dashboard:      "ダッシュボードを開く",
		CreateLink:     "デスクトップにショートカットを作成",
		Start:          "サービス開始",
		Stop:           "サービス停止",
		Uninstall:      "プログラムの削除",
		Exit:           "トレイを終了 (サービス維持)",
		StatusInstall:  "インストール完了",
		StatusRun:      "MaruBot서비스가インストールされ、開始されました。",
		ServiceStarted: "MaruBotサーバーが開始されました。",
		ServiceStopped: "MaruBotサーバーが停止されました。",
		NotifyTitle:    "MaruBot通知",
		NotifyBgMsg:    "トレイを終了しました。MaruBotはバックグラウンドで実行中です。",
		LinkCreated:    "デスクトップとスタートメニューにショートカットが作成されました。",
	},
}

func getLabels() TrayLabels {
	lang := "en"
	cfg, err := loadConfig()
	if err == nil && cfg.Language != "" {
		lang = cfg.Language
	}
	if l, ok := trayLocales[lang]; ok {
		return l
	}
	return trayLocales["en"]
}

func updateTrayLabels() {
	l := getLabels()
	mDashboard.SetTitle(l.Dashboard)
	mCreateLink.SetTitle(l.CreateLink)
	mStart.SetTitle(l.Start)
	mStop.SetTitle(l.Stop)
	mUninstall.SetTitle(l.Uninstall)
	mExit.SetTitle(l.Exit)
	if mUpgrade != nil {
		mUpgrade.SetTitle(l.Upgrade)
	}
}

func onTrayReady(targetExe string) {
	systray.SetTitle("MaruBot")
	systray.SetTooltip("MaruBot - AI Agent Service")
	if runtime.GOOS == "windows" {
		systray.SetIcon(windowTrayIco)
	} else {
		systray.SetIcon(macMenubarPng)
	}

	l := getLabels()
	versionItem = systray.AddMenuItem("MaruBot v"+Version, "Version information")
	versionItem.Disable()
	systray.AddSeparator()
	mDashboard = systray.AddMenuItem(l.Dashboard, "")
	mCreateLink = systray.AddMenuItem(l.CreateLink, "")
	systray.AddSeparator()
	mStart = systray.AddMenuItem(l.Start, "")
	mStop = systray.AddMenuItem(l.Stop, "")
	mUpgrade = systray.AddMenuItem(l.Upgrade, "")
	systray.AddSeparator()
	mUninstall = systray.AddMenuItem(l.Uninstall, "")
	mExit = systray.AddMenuItem(l.Exit, "")

	// Watch for language changes
	go func() {
		lastLang := ""
		for {
			cfg, err := loadConfig()
			if err == nil && cfg.Language != lastLang {
				lastLang = cfg.Language
				updateTrayLabels()
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// 3. Initial check and install
	cfg, _ := loadConfig()
	if !checkAndFixPort(cfg) {
		systray.Quit()
		return
	}

	// Automatic shortcut creation on first run
	if runtime.GOOS == "windows" {
		go createWindowsShortcuts(targetExe, false)
	}

	// Service management
	if runtime.GOOS == "windows" {
		svcConfig := &service.Config{Name: "MaruBot"}
		s, _ := service.New(&program{}, svcConfig)
		status, _ := s.Status()

		shouldInstall := false
		if status == service.StatusUnknown {
			shouldInstall = true
		} else {
			// Service exists. Check if user wants to update
			if showNativeConfirmDialog("MaruBot Installer", "MaruBot is already installed. Do you want to reinstall and update it to the latest version?") {
				fmt.Println("User opted for reinstallation.")
				serviceCmdInternalPath("uninstall", targetExe)
				time.Sleep(2 * time.Second)
				shouldInstall = true
			}
		}

		if shouldInstall {
			fmt.Println("Installing MaruBot service...")
			serviceCmdInternalPath("install", targetExe)
			time.Sleep(1 * time.Second)
			serviceCmdInternalPath("start", targetExe)
			l := getLabels()
			showWindowsNotification(l.StatusInstall, l.StatusRun)
		} else if status != service.StatusRunning {
			serviceCmdInternalPath("start", targetExe)
		}
	} else if runtime.GOOS == "darwin" {
		if !isMarubotProcessRunning() {
			go startCmd()
		}
	}

	openBrowser(fmt.Sprintf("http://localhost:%d", cfg.Gateway.Port))

	// Menu Handlers
	go func() {
		for {
			select {
			case <-mDashboard.ClickedCh:
				currCfg, _ := loadConfig()
				openBrowser(fmt.Sprintf("http://localhost:%d", currCfg.Gateway.Port))
			case <-mCreateLink.ClickedCh:
				if runtime.GOOS == "windows" {
					createWindowsShortcuts(targetExe, true)
				}
			case <-mStart.ClickedCh:
				l := getLabels()
				if runtime.GOOS == "windows" {
					serviceCmdInternalPath("start", targetExe)
					showWindowsNotification(l.Start, l.ServiceStarted)
				} else {
					go startCmd()
				}
				time.Sleep(2 * time.Second)
			case <-mStop.ClickedCh:
				l := getLabels()
				if runtime.GOOS == "windows" {
					serviceCmdInternalPath("stop", targetExe)
					showWindowsNotification(l.Stop, l.ServiceStopped)
				} else {
					stopCmd()
				}
				time.Sleep(1 * time.Second)
			case <-mUninstall.ClickedCh:
				if runtime.GOOS == "windows" {
					runAsAdminAction("uninstall --yes")
				} else if runtime.GOOS == "darwin" {
					go func() {
						uninstallCmd()
						systray.Quit()
						os.Exit(0)
					}()
				} else {
					fmt.Println("Uninstall aborted by user.")
				}
			case <-mExit.ClickedCh:
				if runtime.GOOS == "windows" {
					l := getLabels()
					showWindowsNotification(l.NotifyTitle, l.NotifyBgMsg)
					time.Sleep(1 * time.Second)
				}
				systray.Quit()
				os.Exit(0)
			case <-mUpgrade.ClickedCh:
				l := getLabels()
				showWindowsNotification(l.Upgrade, "Checking for latest version...")

				latest, err := config.CheckLatestVersion()
				if err != nil {
					showNativeMessageDialog("Error", "Failed to check for updates: "+err.Error())
				} else {
					if config.IsNewVersionAvailable(latest) {
						msg := fmt.Sprintf("New version v%s is available. Would you like to upgrade now?", latest)
						if showNativeConfirmDialog(l.Upgrade, msg) {
							go func() {
								showWindowsNotification(l.Upgrade, "Upgrade started in background...")
								exe, _ := os.Executable()
								cmd := exec.Command(exe, "upgrade", "--yes")
								// Hide console window on Windows
								cmd.SysProcAttr = getSysProcAttr()

								if err := cmd.Run(); err != nil {
									showNativeMessageDialog("Upgrade Failed", "Error during upgrade: "+err.Error())
								} else {
									showNativeMessageDialog("Upgrade Success", "MaruBot has been upgraded to v"+latest+".\nThe application will restart shortly.")
								}
							}()
						}
					} else {
						showNativeMessageDialog(l.Upgrade, fmt.Sprintf("You are already using the latest version (v%s).", config.Version))
					}
				}
			}
		}
	}()
}

func createWindowsShortcuts(exePath string, verbose bool) {
	if runtime.GOOS != "windows" {
		return
	}

	psScript := fmt.Sprintf(`
$shell = New-Object -ComObject WScript.Shell
$desktop = [System.Environment]::GetFolderPath('Desktop')
$startMenu = [System.Environment]::GetFolderPath('StartMenu')
$programs = Join-Path $startMenu "Programs"

$targets = @(Join-Path $desktop "MaruBot.lnk", Join-Path $programs "MaruBot.lnk")

foreach ($t in $targets) {
    if (-not (Test-Path $t) -or %v) {
        $s = $shell.CreateShortcut($t)
        $s.TargetPath = "%s"
        $s.WorkingDirectory = Split-Path "%s"
        $s.Description = "MaruBot AI Agent"
        $s.Save()
    }
}
`, verbose, exePath, exePath)

	exec.Command("powershell", "-NoProfile", "-Command", psScript).Run()

	if verbose {
		l := getLabels()
		showWindowsNotification(l.CreateLink, l.LinkCreated)
	}
}

// Deleted: showNativeMessageDialog (Moved to main.go as native Win32)

func showWindowsNotification(title, message string) {
	if runtime.GOOS != "windows" {
		return
	}
	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$notify = New-Object System.Windows.Forms.NotifyIcon
$notify.Icon = [System.Drawing.Icon]::ExtractAssociatedIcon("%s")
$notify.Visible = $true
$notify.ShowBalloonTip(3000, "%s", "%s", [System.Windows.Forms.ToolTipIcon]::Info)
`, os.Args[0], title, message)

	exec.Command("powershell", "-NoProfile", "-Command", psScript).Run()
}

func onTrayExit() {
	// Cleanup on exit
}

// Deleted: showConfirmDialog (Replaced by showNativeConfirmDialog in main.go)

func showAboutDialog() {
	title := "About MaruBot"
	message := fmt.Sprintf("MaruBot - AI Agent Service\n\nVersion: %s\n\n© 2026 MaruBot Contributors", Version)

	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf("display alert %q message %q as informational buttons {\"OK\"} default button \"OK\"", title, message)
		exec.Command("osascript", "-e", script).Run()
	} else if runtime.GOOS == "windows" {
		showNativeMessageDialog(title, message)
	} else {
		fmt.Printf("--- %s ---\n%s\n", title, message)
	}
}
