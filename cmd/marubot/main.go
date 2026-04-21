// MaruBot - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 MaruBot contributors

package main

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/dirmich/marubot/cmd/marubot/dashboard"
	"github.com/dirmich/marubot/pkg/admin"
	"github.com/dirmich/marubot/pkg/agent"
	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/channels"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/cron"
	"github.com/dirmich/marubot/pkg/hardware/gpio"
	"github.com/dirmich/marubot/pkg/heartbeat"
	"github.com/dirmich/marubot/pkg/logger"
	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/skills"
	"github.com/dirmich/marubot/pkg/utils"
	"github.com/dirmich/marubot/pkg/voice"

	"github.com/chzyer/readline"
	"github.com/kardianos/service"
	"sync"
)

// 0.4.59: Fix tray icon visibility with ICO and elevated uninstall via RunAs
// 0.4.58: Windows GUI mode optimization (hide console)

const logo = "[MaruBot]"

//go:embed assets/app_icon.png
var appIconPng []byte

//go:embed assets/mac_menubar.png
var macMenubarPng []byte

//go:embed assets/window_tray.ico
var windowTrayIco []byte

var Version = config.Version
var overrideResourceDir = ""

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func main() {
	// Early flag parsing for --home to initialize paths
	newArgs := []string{}
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--home" && i+1 < len(os.Args) {
			overrideResourceDir = os.Args[i+1]
			os.Setenv("MARUBOT_HOME", overrideResourceDir)
			i++
		} else {
			newArgs = append(newArgs, os.Args[i])
		}
	}
	os.Args = newArgs

	if len(os.Args) < 2 || (len(os.Args) == 2 && os.Args[1] == "--elevated") {
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			handleGUIMode()
			return
		}
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "--version-only" {
		fmt.Print(Version)
		return
	}

	switch command {
	case "onboard":
		onboard()
	case "agent":
		agentCmd()
	case "gateway":
		gatewayCmd()
	case "status":
		statusCmd()
	case "service":
		serviceCmd()
	case "config":
		configCmd()
	case "cron":
		cronCmd()
	case "migrate-paths":
		migratePathsCmd()
	case "start":
		startCmd()
	case "reload":
		reloadCmd()
	case "skills":
		skillsCmd()
	case "voice":
		// Reserved for future use
		fmt.Println("Voice command is not yet implemented.")
	case "stop":
		stopCmd()
	case "upgrade":
		upgradeCmd()
	case "uninstall":
		uninstallCmd()
	case "version", "--version", "-v":
		fmt.Printf("%s marubot v%s\n", logo, strings.TrimPrefix(Version, "v"))
	case "admin-user":
		// Reserved for future use
		fmt.Println("Admin-user command is not yet implemented.")
	default:
		printHelp()
		os.Exit(1)
	}
}

// Global state for background services to support Hot Reload
var (
	backgroundCtx      context.Context
	backgroundCancel   context.CancelFunc
	currentAgentLoop   *agent.AgentLoop
	currentChanManager *channels.Manager
	currentCron        *cron.CronService
	currentHeartbeat   *heartbeat.HeartbeatService
	backgroundMu       sync.Mutex
)

func execHidden(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = getSysProcAttr()
	return cmd
}

func logUninstall(message string) {
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".marubot", "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "uninstall-"+time.Now().Format("2006-01-02")+".log")
	
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		f.WriteString(fmt.Sprintf("[%s] [UNINSTALL] %s\n", timestamp, message))
	}
	fmt.Println(message)
}

func uninstallCmd() {
	autoConfirm := false
	for _, arg := range os.Args {
		if arg == "--yes" || arg == "-y" {
			autoConfirm = true
			break
		}
	}

	if runtime.GOOS == "windows" {
		if !isAdmin() {
			fmt.Println("Elevation required for uninstallation. Requesting administrator privileges...")
			if autoConfirm {
				runAsAdminAction("uninstall --yes")
			} else {
				runAsAdminAction("uninstall")
			}
			os.Exit(0)
		}
	}

	// 0. Initial confirmation (Skip if autoConfirm is true)
	if !autoConfirm {
		if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			if !showNativeConfirmDialog("Uninstall MaruBot", "Are you sure you want to uninstall MaruBot and all its services?") {
				fmt.Println("Aborted by user.")
				return
			}
		} else {
			fmt.Printf("%s MaruBot Uninstaller\n", logo)
			fmt.Println("WARNING: This will remove MaruBot and its resources from your system.")
			fmt.Print("Are you sure you want to continue? (y/N): ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) != "y" {
				fmt.Println("Aborted.")
				return
			}
		}
	}

	// 0. Remove Services and Kill Processes (Cross-platform)
	if runtime.GOOS == "windows" {
		logUninstall("--- Windows ?∏Ïù∏?§ÌÜ® ?úÏûë ---")
		// Clean up shortcuts
		removeWindowsShortcuts()

		// 1. Force kill all processes with 'marubot' in their path, EXCLUDING current PID
		currentPid := os.Getpid()
		logUninstall(fmt.Sprintf("Í¥Ä???ÑÎ°ú?∏Ïä§ Ï¢ÖÎ£å Ï§?(?ÑÏû¨ PID ?úÏô∏: %d)...", currentPid))
		killScript := fmt.Sprintf(`Get-Process | Where-Object { ($_.Path -like "*marubot*" -or $_.Name -like "*marubot*") -and ($_.Id -ne %d) } | ForEach-Object { try { Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue; "Killed: $($_.Name) ($($_.Id))" } catch { "Failed to kill: $($_.Name) ($($_.Id))" } }`, currentPid)
		killOut, _ := exec.Command("powershell", "-NoProfile", "-Command", killScript).CombinedOutput()
		logUninstall(string(killOut))

		logUninstall("Step: Discovering services...")
		findSvcScript := `Get-Service *marubot* | Select-Object -ExpandProperty Name`
		svcOut, _ := exec.Command("powershell", "-NoProfile", "-Command", findSvcScript).Output()
		svcNamesStr := strings.TrimSpace(string(svcOut))
		var svcNames []string
		if svcNamesStr != "" {
			svcNames = strings.Split(svcNamesStr, "\r\n")
			// handle both line endings
			if len(svcNames) == 1 {
				svcNames = strings.Split(svcNamesStr, "\n")
			}
		}
		// Fallback to defaults if discovery fails or is empty
		if len(svcNames) == 0 {
			svcNames = []string{"MaruBot", "marubot"}
		}
		logUninstall(fmt.Sprintf("Target services: %v", svcNames))

		scPath := `C:\Windows\System32\sc.exe`
		for _, svcName := range svcNames {
			svcName = strings.TrimSpace(svcName)
			if svcName == "" {
				continue
			}
			logUninstall(fmt.Sprintf("Deep cleaning service '%s'...", svcName))

			// Force stop
			stopOut, _ := exec.Command(scPath, "stop", svcName).CombinedOutput()
			logUninstall(fmt.Sprintf("Stop result: %s", string(stopOut)))

			// Wait and delete
			time.Sleep(1 * time.Second)
			delOut, _ := exec.Command(scPath, "delete", svcName).CombinedOutput()
			logUninstall(fmt.Sprintf("Delete result: %s", string(delOut)))

			// 4. Robust wait for deletion
			isRemoved := false
			for i := 0; i < 15; i++ {
				out, _ := exec.Command(scPath, "query", svcName).Output()
				if strings.Contains(string(out), "1060") {
					isRemoved = true
					break
				}
				logUninstall(fmt.Sprintf("  Waiting for service '%s' to be removed... (%d/15)", svcName, i+1))
				time.Sleep(1 * time.Second)
				if i == 5 {
					hint := "Service deletion is taking longer than expected. Please close 'Service Manager (services.msc)' or Task Manager."
					if autoConfirm {
						go showNativeMessageDialog("Action Required", hint)
					}
				}
			}

			if isRemoved {
				logUninstall(fmt.Sprintf("???úÎπÑ??'%s'Í∞Ä ?±Í≥µ?ÅÏúºÎ°??úÍ±∞?òÏóà?µÎãà??", svcName))
			} else {
				logUninstall(fmt.Sprintf("! ?úÎπÑ??'%s' ?úÍ±∞ ?úÍ∞Ñ Ï¥àÍ≥º.", svcName))
			}
		}

		// Always ensure bin folder is cleaned on Windows uninstallation
		resourceDir := getResourceDir()
		binDir := filepath.Join(resourceDir, "bin")
		if _, err := os.Stat(binDir); err == nil {
			logUninstall("Î∞îÏù¥?àÎ¶¨ ?¥Îçî ?ïÎ¶¨ Ï§? " + binDir)

			// Get current executable path to avoid deleting self here
			currentExe, _ := os.Executable()

			files, _ := os.ReadDir(binDir)
			for _, file := range files {
				filePath := filepath.Join(binDir, file.Name())
				if strings.EqualFold(filePath, currentExe) {
					continue // Skip self, will be deleted by post-exit script
				}
				err := os.RemoveAll(filePath)
				if err != nil {
					logUninstall(fmt.Sprintf("! ?åÏùº ??†ú ?§Ìå®: %s (%v)", file.Name(), err))
				}
			}
			logUninstall("??Î∞îÏù¥?àÎ¶¨ ?¥Îçî ?ïÎ¶¨ ?ÑÎ£å (?ÑÏû¨ ?§Ìñâ ?åÏùº ?úÏô∏)")
		}
	} else if runtime.GOOS == "linux" {
		// Stop and disable systemd service if exists
		u, _ := user.Current()
		if u != nil {
			serviceDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
			servicePath := filepath.Join(serviceDir, "marubot.service")
			if _, err := os.Stat(servicePath); err == nil {
				fmt.Println("Removing Linux systemd user service...")
				exec.Command("systemctl", "--user", "stop", "marubot.service").Run()
				exec.Command("systemctl", "--user", "disable", "marubot.service").Run()
				os.Remove(servicePath)
				exec.Command("systemctl", "--user", "daemon-reload").Run()
			}
		}
		// Kill any remaining marubot processes
		exec.Command("pkill", "-9", "marubot").Run()
	} else if runtime.GOOS == "darwin" {
		// For macOS, just pkill for now (can expand to launchctl if needed)
		exec.Command("pkill", "-9", "marubot").Run()
	}

	// Ask for data deletion using native dialog on GUI-capable platforms
	var deleteAll bool
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		deleteAll = showNativeConfirmDialog("Cleanup MaruBot Data?", "Do you want to PERMANENTLY DELETE all your MaruBot data (configuration, logs, memory, and workspace)?\n\nSelect 'No' if you want to keep them for a future reinstallation.")
	} else {
		fmt.Print("Do you want to PERMANENTLY DELETE all user data (config, memory, workspace)? (y/N): ")
		var deleteChoice string
		fmt.Scanln(&deleteChoice)
		deleteAll = strings.ToLower(deleteChoice) == "y"
	}
	keep := !deleteAll

	// 1. Remove resources
	resourceDir := getResourceDir()
	if _, err := os.Stat(resourceDir); err == nil {
		if keep {
			fmt.Println("Cleaning system resources (keeping user data)...")
			// Carefully only remove binary and system assets
			os.RemoveAll(filepath.Join(resourceDir, "skills"))
			os.RemoveAll(filepath.Join(resourceDir, "tools"))
			os.RemoveAll(filepath.Join(resourceDir, "web-admin"))
			os.RemoveAll(filepath.Join(resourceDir, "bin"))
			fmt.Printf("??System binaries and assets removed. User data (config, workspace, memory) preserved in: %s\n", resourceDir)
		} else {
			fmt.Println("Removing ALL MaruBot data including configuration and workspace...")
			// Make sure we kill processes again just in case a lock is holding the folder
			if runtime.GOOS == "windows" {
				exec.Command("taskkill", "/F", "/T", "/IM", "marubot.exe").Run()
				exec.Command("taskkill", "/F", "/T", "/IM", "marubot-*.exe").Run()
				time.Sleep(1 * time.Second)
			}
			if err := os.RemoveAll(resourceDir); err != nil {
				fmt.Printf("Error removing %s: %v\n", resourceDir, err)
				fmt.Println("  Please manually delete the folder after closing any related applications.")
			} else {
				fmt.Printf("??Entire MaruBot home directory (%s) removed successfully.\n", resourceDir)
			}
		}
	}

	// 2. Remove binary
	exePath, err := os.Executable()
	if err == nil {
		home, _ := os.UserHomeDir()
		installBinDir := filepath.Join(home, ".marubot", "bin")
		// Clean paths for reliable comparison
		cleanExePath := filepath.Clean(exePath)
		cleanInstallDir := filepath.Clean(installBinDir)
		isInstalledBin := strings.HasPrefix(cleanExePath, cleanInstallDir)

		if !isInstalledBin {
			logUninstall("?åÎ¶º: ?ÑÏû¨ ?§Ìñâ ?åÏùº???§Ïπò Í≤ΩÎ°ú???àÏ? ?äÏúºÎØÄÎ°???†úÎ•?Í±¥ÎÑà?ÅÎãà??(?§Ïπò ?åÏùº Î≥¥Ï°¥).")
		} else {
			fmt.Printf("?§Ïπò ?òÍ≤Ω???§Ìñâ ?åÏùº???úÍ±∞?©Îãà?? %s\n", exePath)
			if runtime.GOOS == "windows" {
				// Windows cannot delete a running executable.
				// Use a PowerShell trick to delete after exit with better path handling and retry.
				destDir := filepath.Dir(exePath)
				// Revised PowerShell script for robust deletion with retry
				script := fmt.Sprintf(`
Start-Sleep -Seconds 3
for ($i=1; $i -le 10; $i++) {
    if (Test-Path '%s') {
        try {
            Remove-Item -Path '%s' -Force -ErrorAction Stop
            break
        } catch {
            Start-Sleep -Seconds 2
        }
    } else {
        break
    }
}
`, exePath, exePath)
				cmd := exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Start-Process powershell -ArgumentList \"-NoProfile -WindowStyle Hidden -Command \\\"%s\\\"\" -WindowStyle Hidden -WorkingDirectory '%s'", script, destDir))
				if err := cmd.Start(); err != nil {
					logUninstall(fmt.Sprintf("! ?êÍ? ??†ú ?àÏïΩ ?§Ìå®: %v", err))
				} else {
					logUninstall("??Ï¢ÖÎ£å ???§Ìñâ ?åÏùº ??†úÍ∞Ä ?àÏïΩ?òÏóà?µÎãà??")
				}
			} else {
				if err := os.Remove(exePath); err != nil {
					logUninstall(fmt.Sprintf("! ?§Ìñâ ?åÏùº ??†ú ?§Ìå®: %v", err))
				} else {
					fmt.Println("??Executable removed")
				}
			}
		}
	} else {
		fmt.Println("Could not determine executable path. Please remove it manually.")
	}

	fmt.Println("\nMaruBot has been uninstalled.")
	if keep {
		fmt.Printf("To remove user data later, delete: %s\n", resourceDir)
	}
}

func printHelp() {
	fmt.Printf("%s marubot - Personal AI Assistant v%s\n", logo, Version)
	fmt.Println("Usage: marubot <command>")
	fmt.Println("Commands:")
	fmt.Println("  agent       Interact with the agent directly")
	fmt.Println("  config      Manage hardware/system configuration")
	fmt.Println("  cron        Manage scheduled tasks")
	fmt.Println("  gateway     Start marubot gateway")
	fmt.Println("  onboard     Initialize marubot configuration and workspace")
	fmt.Println("  reload      Reload marubot configuration")
	fmt.Println("  skills      Manage skills (install, list, remove)")
	fmt.Println("  start       Start both gateway and web UI dashboard in background")
	fmt.Println("  status      Show marubot status")
	fmt.Println("  stop        Stop background dashboard process")
	fmt.Println("  uninstall   Remove marubot from system")
	fmt.Println("  upgrade     Upgrade marubot to the latest version")
	fmt.Println("  version     Show version information")
}

func onboard() {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s\n", configPath)
		fmt.Print("Overwrite? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" {
			fmt.Println("Aborted.")
			return
		}
	}

	// Load existing config if available to preserve some settings like password
	cfg := config.DefaultConfig()
	existingCfg, err := config.LoadConfig(configPath)
	if err == nil {
		cfg.AdminPassword = existingCfg.AdminPassword
	}

	fmt.Printf("Set Admin Password for Web Dashboard [%s]: ", func() string {
		if cfg.AdminPassword != "" {
			return cfg.AdminPassword
		}
		return "admin"
	}())

	var password string
	fmt.Scanln(&password)

	if password == "" {
		if cfg.AdminPassword == "" {
			cfg.AdminPassword = "admin"
			fmt.Println("No password entered. Defaulting to 'admin'.")
		} else {
			fmt.Printf("No password entered. Keeping existing password.\n")
		}
	} else {
		cfg.AdminPassword = password
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	workspace := cfg.WorkspacePath()
	os.MkdirAll(workspace, 0755)
	os.MkdirAll(filepath.Join(workspace, "memory"), 0755)
	os.MkdirAll(filepath.Join(workspace, "skills"), 0755)

	createWorkspaceTemplates(workspace)

	fmt.Printf("%s marubot is ready!\n", logo)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to", configPath)
	fmt.Println("     Get one at: https://openrouter.ai/keys")
	fmt.Println("  2. Chat: marubot agent -m \"Hello!\"")
}

func ensureAdminPassword(cfg *config.Config) {
	fmt.Printf("Set Admin Password for Web Dashboard: ")
	var password string
	fmt.Scanln(&password)

	if password == "" {
		fmt.Println("Password cannot be empty. Defaulting to 'admin'.")
		password = "admin"
	}

	cfg.AdminPassword = utils.HashPassword(password)
	if err := config.SaveConfig(getConfigPath(), cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
	} else {
		fmt.Println("Password saved successfully.")
	}
}

func createWorkspaceTemplates(workspace string) {
	templates := map[string]string{
		"AGENTS.md": `# Agent Instructions

You are a helpful AI assistant. Be concise, accurate, and friendly.

## Guidelines

- Always explain what you're doing before taking actions
- Ask for clarification when request is ambiguous
- Use tools to help accomplish tasks
- Remember important information in your memory files
- Be proactive and helpful
- Learn from user feedback
`,
		"SOUL.md": `# Soul

I am marubot, a lightweight AI assistant powered by AI.

## Personality

- Helpful and friendly
- Concise and to the point
- Curious and eager to learn
- Honest and transparent

## Values

- Accuracy over speed
- User privacy and safety
- Transparency in actions
- Continuous improvement
`,
		"USER.md": `# User

Information about user goes here.

## Preferences

- Communication style: (casual/formal)
- Timezone: (your timezone)
- Language: (your preferred language)

## Personal Information

- Name: (optional)
- Location: (optional)
- Occupation: (optional)

## Learning Goals

- What the user wants to learn from AI
- Preferred interaction style
- Areas of interest
`,
		"TOOLS.md": `# Available Tools

This document describes the tools available to marubot.

## File Operations

### Read Files
- Read file contents
- Supports text, markdown, code files

### Write Files
- Create new files
- Overwrite existing files
- Supports various formats

### List Directories
- List directory contents
- Recursive listing support

### Edit Files
- Make specific edits to files
- Line-by-line editing
- String replacement

## Web Tools

### Web Search
- Search the internet using search API
- Returns titles, URLs, snippets
- Optional: Requires API key for best results

### Web Fetch
- Fetch specific URLs
- Extract readable content
- Supports HTML, JSON, plain text
- Automatic content extraction

## Command Execution

### Shell Commands
- Execute any shell command
- Run in workspace directory
- Full shell access with timeout protection

### SSH Connections
- Execute commands on remote hosts via SSH
- Supports password and key-based authentication
- Handles Windows-specific limitations for remote access

## Messaging

### Send Messages
- Send messages to chat channels
- Supports Telegram, WhatsApp, Feishu
- Used for notifications and responses
- Supports rich markdown (tables, bold, italic) but avoid HTML tags like <br>

## AI Capabilities

### Context Building
- Load system instructions from files
- Load skills dynamically
- Build conversation history
- Include timezone and other context

### Memory Management
- Long-term memory via MEMORY.md
- Daily notes via dated files
- Persistent across sessions
`,
		"IDENTITY.md": `# Identity

## Name
MaruBot ?¶û

## Description
Ultra-lightweight personal AI assistant written in Go, inspired by nanobot.

## Version
0.7.3

## Purpose
- Provide intelligent AI assistance with minimal resource usage
- Support multiple LLM providers (OpenAI, Anthropic, Zhipu, etc.)
- Enable easy customization through skills system
- Run on minimal hardware ($10 boards, <10MB RAM)

## Capabilities

- Web search and content fetching
- File system operations (read, write, edit)
- Shell command execution
- Multi-channel messaging (Telegram, WhatsApp, Feishu)
- SSH & Remote System Access (Support for password and key authentication)
- Skill-based extensibility
- Memory and context management
- GPIO/Hardware control and monitoring

## Philosophy

- Simplicity over complexity
- Performance over features
- User control and privacy
- Transparent operation
- Community-driven development

## Goals

- Provide a fast, lightweight AI assistant
- Support offline-first operation where possible
- Enable easy customization and extension
- Maintain high quality responses
- Run efficiently on constrained hardware

## License
MIT License - Free and open source

## Repository
https://github.com/dirmich/maru-bot

## Contact
Issues: https://github.com/dirmich/maru-bot/issues
Discussions: https://github.com/dirmich/maru-bot/discussions

---

"Simple, Powerful, Personal."
`,
	}

	for filename, content := range templates {
		filePath := filepath.Join(workspace, filename)
		// Always overwrite core identity and instruction files to ensure the latest prompt is used.
		// These files act as the system-managed identity.
		if filename == "IDENTITY.md" || filename == "AGENTS.md" || filename == "TOOLS.md" || filename == "SOUL.md" {
			os.WriteFile(filePath, []byte(content), 0644)
			if filename == "IDENTITY.md" {
				fmt.Printf("  Updated %s (current Version: %s)\n", filename, config.Version)
			}
		} else if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.WriteFile(filePath, []byte(content), 0644)
			fmt.Printf("  Created %s\n", filename)
		}
	}

	memoryDir := filepath.Join(workspace, "memory")
	os.MkdirAll(memoryDir, 0755)
	memoryFile := filepath.Join(memoryDir, "MEMORY.md")
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		memoryContent := `# Long-term Memory

This file stores important information that should persist across sessions.

## User Information

(Important facts about user)

## Preferences

(User preferences learned over time)

## Important Notes

(Things to remember)

## Configuration

- Model preferences
- Channel settings
- Skills enabled
`
		os.WriteFile(memoryFile, []byte(memoryContent), 0644)
		fmt.Println("  Created memory/MEMORY.md")

		skillsDir := filepath.Join(workspace, "skills")
		if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
			os.MkdirAll(skillsDir, 0755)
			fmt.Println("  Created skills/")
		}
	}

	for filename, content := range templates {
		filePath := filepath.Join(workspace, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.WriteFile(filePath, []byte(content), 0644)
			fmt.Printf("  Created %s\n", filename)
		}
	}
}

func agentCmd() {
	message := ""
	sessionKey := "cli:default"

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "-s", "--session":
			if i+1 < len(args) {
				sessionKey = args[i+1]
				i++
			}
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	workspace := cfg.WorkspacePath()
	createWorkspaceTemplates(workspace)

	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		os.Exit(1)
	}

	bus := bus.NewMessageBus()
	agentLoop := agent.NewAgentLoop(cfg, bus, provider, Version)

	if runtime.GOOS == "linux" {
		gpioService := gpio.NewGPIOService(cfg, bus)
		gpioService.Start(context.Background())
		defer gpioService.Stop()
	}

	if message != "" {
		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, message, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n%s %s\n", logo, response)
	} else {
		fmt.Printf("%s Interactive mode (Ctrl+C to exit)\n\n", logo)
		interactiveMode(agentLoop, sessionKey)
	}
}

func interactiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	prompt := fmt.Sprintf("%s You: ", logo)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     "/tmp/.marubot_history",
		HistoryLimit:    100,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		fmt.Println("Falling back to simple input mode...")
		simpleInteractiveMode(agentLoop, sessionKey)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s %s\n\n", logo, response)
	}
}

func simpleInteractiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s You: ", logo)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s %s\n\n", logo, response)
	}
}

func gatewayCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	workspace := cfg.WorkspacePath()
	createWorkspaceTemplates(workspace)

	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		os.Exit(1)
	}

	bus := bus.NewMessageBus()
	agentLoop := agent.NewAgentLoop(cfg, bus, provider, Version)

	gpioService := gpio.NewGPIOService(cfg, bus)
	gpioService.Start(context.Background())
	defer gpioService.Stop()

	if cfg.Admin.BackendURL != "" && cfg.Admin.UserID != "" {
		adminClient := admin.NewAdminClient(cfg.Admin.BackendURL, cfg.Admin.UserID, Version, cfg.Language)
		adminClient.StartReporting()
		fmt.Println("??Admin reporting service started")
	}

	cronStorePath := filepath.Join(filepath.Dir(getConfigPath()), "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, func(job *cron.CronJob) (string, error) {
		return agentLoop.ProcessDirect(context.Background(), job.Payload.Message, "cron:"+job.ID)
	})

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		func(msg string) (string, error) {
			return agentLoop.ProcessDirect(context.Background(), msg, "heartbeat")
		},
		30 * 60,
		true,
	)

	channelManager, err := channels.NewManager(cfg, bus)
	if err != nil {
		fmt.Printf("Error creating channel manager: %v\n", err)
		os.Exit(1)
	}

	var transcriber *voice.GroqTranscriber
	for _, m := range cfg.Providers.Groq.Models {
		if m.APIKey != "" {
			transcriber = voice.NewGroqTranscriber(m.APIKey)
			logger.InfoC("voice", "Groq voice transcription enabled")
			break
		}
	}

	if transcriber != nil {
		if telegramChannel, ok := channelManager.GetChannel("telegram"); ok {
			if tc, ok := telegramChannel.(*channels.TelegramChannel); ok {
				tc.SetTranscriber(transcriber)
				logger.InfoC("voice", "Groq transcription attached to Telegram channel")
			}
		}
	}

	enabledChannels := channelManager.GetEnabledChannels()
	if len(enabledChannels) > 0 {
		fmt.Printf("??Channels enabled: %s\n", enabledChannels)
	} else {
		fmt.Println("??Warning: No channels enabled")
	}

	fmt.Printf("??Gateway started on %s:%d\n", cfg.Gateway.Host, cfg.Gateway.Port)
	fmt.Println("Press Ctrl+C to stop")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := cronService.Start(); err != nil {
		fmt.Printf("Error starting cron service: %v\n", err)
	}
	fmt.Println("??Cron service started")

	if err := heartbeatService.Start(); err != nil {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	fmt.Println("??Heartbeat service started")

	if err := channelManager.StartAll(ctx); err != nil {
		fmt.Printf("Error starting channels: %v\n", err)
	}

	go agentLoop.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\nShutting down...")
	cancel()
	heartbeatService.Stop()
	cronService.Stop()
	agentLoop.Stop()
	channelManager.StopAll(ctx)
	fmt.Println("??Gateway stopped")
}

func loadConfig() (*config.Config, error) {
	return config.LoadConfig(getConfigPath())
}

func cronCmd() {
	if len(os.Args) < 3 {
		cronHelp()
		return
	}

	subcommand := os.Args[2]

	dataDir := filepath.Join(filepath.Dir(getConfigPath()), "cron")
	cronStorePath := filepath.Join(dataDir, "jobs.json")

	switch subcommand {
	case "list":
		cronListCmd(cronStorePath)
	case "add":
		cronAddCmd(cronStorePath)
	case "remove":
		if len(os.Args) < 4 {
			fmt.Println("Usage: marubot cron remove <job_id>")
			return
		}
		cronRemoveCmd(cronStorePath, os.Args[3])
	case "enable":
		cronEnableCmd(cronStorePath, false)
	case "disable":
		cronEnableCmd(cronStorePath, true)
	default:
		fmt.Printf("Unknown cron command: %s\n", subcommand)
		cronHelp()
	}
}

func cronHelp() {
	fmt.Println("\nCron commands:")
	fmt.Println("  add              Add a new scheduled job")
	fmt.Println("  disable <id>     Disable a job")
	fmt.Println("  enable <id>      Enable a job")
	fmt.Println("  list              List all scheduled jobs")
	fmt.Println("  remove <id>       Remove a job by ID")
	fmt.Println()
	fmt.Println("Add options:")
	fmt.Println("  -n, --name       Job name")
	fmt.Println("  -m, --message    Message for agent")
	fmt.Println("  -e, --every      Run every N seconds")
	fmt.Println("  -c, --cron       Cron expression (e.g. '0 9 * * *')")
	fmt.Println("  -d, --deliver     Deliver response to channel")
	fmt.Println("  --to             Recipient for delivery")
	fmt.Println("  --channel        Channel for delivery")
}

func cronListCmd(storePath string) {
	cs := cron.NewCronService(storePath, nil)
	jobs := cs.ListJobs(false)

	if len(jobs) == 0 {
		fmt.Println("No scheduled jobs.")
		return
	}

	fmt.Println("\nScheduled Jobs:")
	fmt.Println("----------------")
	for _, job := range jobs {
		var schedule string
		if job.Schedule.Kind == "every" && job.Schedule.EveryMS != nil {
			schedule = fmt.Sprintf("every %ds", *job.Schedule.EveryMS/1000)
		} else if job.Schedule.Kind == "cron" {
			schedule = job.Schedule.Expr
		} else {
			schedule = "one-time"
		}

		nextRun := "scheduled"
		if job.State.NextRunAtMS != nil {
			nextTime := time.UnixMilli(*job.State.NextRunAtMS)
			nextRun = nextTime.Format("2006-01-02 15:04")
		}

		status := "enabled"
		if !job.Enabled {
			status = "disabled"
		}

		fmt.Printf("  %s (%s)\n", job.Name, job.ID)
		fmt.Printf("    Schedule: %s\n", schedule)
		fmt.Printf("    Status: %s\n", status)
		fmt.Printf("    Next run: %s\n", nextRun)
	}
}

func cronAddCmd(storePath string) {
	name := ""
	message := ""
	var everySec *int64
	cronExpr := ""
	deliver := false
	channel := ""
	to := ""

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "-e", "--every":
			if i+1 < len(args) {
				var sec int64
				fmt.Sscanf(args[i+1], "%d", &sec)
				everySec = &sec
				i++
			}
		case "-c", "--cron":
			if i+1 < len(args) {
				cronExpr = args[i+1]
				i++
			}
		case "-d", "--deliver":
			deliver = true
		case "--to":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		case "--channel":
			if i+1 < len(args) {
				channel = args[i+1]
				i++
			}
		}
	}

	if name == "" {
		fmt.Println("Error: --name is required")
		return
	}

	if message == "" {
		fmt.Println("Error: --message is required")
		return
	}

	if everySec == nil && cronExpr == "" {
		fmt.Println("Error: Either --every or --cron must be specified")
		return
	}

	var schedule cron.CronSchedule
	if everySec != nil {
		everyMS := *everySec * 1000
		schedule = cron.CronSchedule{
			Kind:    "every",
			EveryMS: &everyMS,
		}
	} else {
		schedule = cron.CronSchedule{
			Kind: "cron",
			Expr: cronExpr,
		}
	}

	cs := cron.NewCronService(storePath, nil)
	job, err := cs.AddJob(name, schedule, message, deliver, channel, to)
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	fmt.Printf("??Added job '%s' (%s)\n", job.Name, job.ID)
}

func cronRemoveCmd(storePath, jobID string) {
	cs := cron.NewCronService(storePath, nil)
	if cs.RemoveJob(jobID) {
		fmt.Printf("??Removed job %s\n", jobID)
	} else {
		fmt.Printf("??Job %s not found\n", jobID)
	}
}

func cronEnableCmd(storePath string, disable bool) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: marubot cron enable/disable <job_id>")
		return
	}

	jobID := os.Args[3]
	cs := cron.NewCronService(storePath, nil)
	enabled := !disable

	job := cs.EnableJob(jobID, enabled)
	if job != nil {
		status := "enabled"
		if disable {
			status = "disabled"
		}
		fmt.Printf("??Job '%s' %s\n", job.Name, status)
	} else {
		fmt.Printf("??Job %s not found\n", jobID)
	}
}

func skillsCmd() {
	if len(os.Args) < 3 {
		skillsHelp()
		return
	}

	subcommand := os.Args[2]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	workspace := cfg.WorkspacePath()
	installer := skills.NewSkillInstaller(workspace)
	skillsLoader := skills.NewSkillsLoader(workspace, "")

	switch subcommand {
	case "list":
		skillsListCmd(skillsLoader)
	case "install":
		skillsInstallCmd(installer)
	case "remove", "uninstall":
		if len(os.Args) < 4 {
			fmt.Println("Usage: marubot skills remove <skill-name>")
			return
		}
		skillsRemoveCmd(installer, os.Args[3])
	case "search":
		skillsSearchCmd(installer)
	case "show":
		if len(os.Args) < 4 {
			fmt.Println("Usage: marubot skills show <skill-name>")
			return
		}
		skillsShowCmd(skillsLoader, os.Args[3])
	case "list-builtin":
		skillsListBuiltinCmd()
	case "install-builtin":
		skillsInstallBuiltinCmd(workspace)
	default:
		fmt.Printf("Unknown skills command: %s\n", subcommand)
		skillsHelp()
	}
}

func skillsHelp() {
	fmt.Println("\nSkills commands:")
	fmt.Println("  install <repo>          Install skill from GitHub")
	fmt.Println("  install-builtin          Install all builtin skills to workspace")
	fmt.Println("  list                    List installed skills")
	fmt.Println("  list-builtin             List available builtin skills")
	fmt.Println("  remove <name>           Remove installed skill")
	fmt.Println("  search                  Search available skills")
	fmt.Println("  show <name>             Show skill details")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  marubot skills list")
	fmt.Println("  marubot skills install sipeed/marubot-skills/weather")
	fmt.Println("  marubot skills install-builtin")
	fmt.Println("  marubot skills list-builtin")
	fmt.Println("  marubot skills remove weather")
}

func skillsListCmd(loader *skills.SkillsLoader) {
	allSkills := loader.ListSkills(false)

	if len(allSkills) == 0 {
		fmt.Println("No skills installed.")
		return
	}

	fmt.Println("\nInstalled Skills:")
	fmt.Println("------------------")
	for _, skill := range allSkills {
		status := "??
		if !skill.Available {
			status = "??
		}
		fmt.Printf("  %s %s (%s)\n", status, skill.Name, skill.Source)
		if skill.Description != "" {
			fmt.Printf("    %s\n", skill.Description)
		}
		if !skill.Available {
			fmt.Printf("    Missing: %s\n", skill.Missing)
		}
	}
}

func skillsInstallCmd(installer *skills.SkillInstaller) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: marubot skills install <github-repo>")
		fmt.Println("Example: marubot skills install sipeed/marubot-skills/weather")
		return
	}

	repo := os.Args[3]
	fmt.Printf("Installing skill from %s...\n", repo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := installer.InstallFromGitHub(ctx, repo); err != nil {
		fmt.Printf("??Failed to install skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("??Skill '%s' installed successfully!\n", filepath.Base(repo))
}

func skillsRemoveCmd(installer *skills.SkillInstaller, skillName string) {
	fmt.Printf("Removing skill '%s'...\n", skillName)

	if err := installer.Uninstall(skillName); err != nil {
		fmt.Printf("??Failed to remove skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("??Skill '%s' removed successfully!\n", skillName)
}

func skillsInstallBuiltinCmd(workspace string) {
	builtinSkillsDir := filepath.Join(getResourceDir(), "skills")
	// If not found in resource dir, fallback to local dev path for backward compatibility
	if _, err := os.Stat(builtinSkillsDir); os.IsNotExist(err) {
		builtinSkillsDir = "./skills"
	}

	workspaceSkillsDir := filepath.Join(workspace, "skills")

	fmt.Printf("Copying builtin skills from %s to workspace...\n", builtinSkillsDir)

	skillsToInstall := []string{
		"weather",
		"news",
		"stock",
		"calculator",
	}

	for _, skillName := range skillsToInstall {
		builtinPath := filepath.Join(builtinSkillsDir, skillName)
		workspacePath := filepath.Join(workspaceSkillsDir, skillName)

		if _, err := os.Stat(builtinPath); err != nil {
			fmt.Printf("??Builtin skill '%s' not found: %v\n", skillName, err)
			continue
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			fmt.Printf("??Failed to create directory for %s: %v\n", skillName, err)
			continue
		}

		if err := copyDirectory(builtinPath, workspacePath); err != nil {
			fmt.Printf("??Failed to copy %s: %v\n", skillName, err)
		}
	}

	fmt.Println("\n??All builtin skills installed!")
	fmt.Println("Now you can use them in your workspace.")
}

func skillsListBuiltinCmd() {
	builtinSkillsDir := filepath.Join(getResourceDir(), "skills")
	// If not found in resource dir, fallback to local dev path for backward compatibility
	if _, err := os.Stat(builtinSkillsDir); os.IsNotExist(err) {
		builtinSkillsDir = "./skills"
	}

	fmt.Println("\nAvailable Builtin Skills:")
	fmt.Println("-----------------------")

	entries, err := os.ReadDir(builtinSkillsDir)
	if err != nil {
		fmt.Printf("Error reading builtin skills: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No builtin skills available.")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillName := entry.Name()
			manifestFile := filepath.Join(builtinSkillsDir, skillName, "manifest.json")

			description := "No description"
			if _, err := os.Stat(manifestFile); err == nil {
				data, err := os.ReadFile(manifestFile)
				if err == nil {
					var manifest map[string]interface{}
					if err := json.Unmarshal(data, &manifest); err == nil {
						if desc, ok := manifest["description"].(string); ok {
							description = desc
						}
					}
				}
			}
			status := "??
			fmt.Printf("  %s  %s\n", status, entry.Name())
			if description != "" {
				fmt.Printf("     %s\n", description)
			}
		}
	}
}

func skillsSearchCmd(installer *skills.SkillInstaller) {
	fmt.Println("Searching for available skills...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	availableSkills, err := installer.ListAvailableSkills(ctx)
	if err != nil {
		fmt.Printf("??Failed to fetch skills list: %v\n", err)
		return
	}

	if len(availableSkills) == 0 {
		fmt.Println("No skills available.")
		return
	}

	fmt.Printf("\nAvailable Skills (%d):\n", len(availableSkills))
	fmt.Println("--------------------")
	for _, skill := range availableSkills {
		fmt.Printf("  ?ì¶ %s\n", skill.Name)
		fmt.Printf("     %s\n", skill.Description)
		fmt.Printf("     Repo: %s\n", skill.Repository)
		if skill.Author != "" {
			fmt.Printf("     Author: %s\n", skill.Author)
		}
		if len(skill.Tags) > 0 {
			fmt.Printf("     Tags: %v\n", skill.Tags)
		}
		fmt.Println()
	}
}

func skillsShowCmd(loader *skills.SkillsLoader, skillName string) {
	content, ok := loader.LoadSkill(skillName)
	if !ok {
		fmt.Printf("??Skill '%s' not found\n", skillName)
		return
	}

	fmt.Printf("\n?ì¶ Skill: %s\n", skillName)
	fmt.Println("----------------------")
	fmt.Println(content)
}
func configCmd() {
	if len(os.Args) < 3 {
		configHelp()
		return
	}

	subcommand := os.Args[2]
	configPath := getConfigPath()

	switch subcommand {
	case "show":
		cfg, _ := loadConfig()
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Printf("Current Configuration (%s):\n%s\n", configPath, string(data))
	case "set":
		if len(os.Args) < 5 {
			fmt.Println("Usage: marubot config set <key> <value>")
			return
		}
		key := os.Args[3]
		value := os.Args[4]

		cfg, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// Update the key in config
		// For simple implementation of CLI 'set', we can either use a generic updater
		// or focus on common fields. Since we want to unify, we update the Config object.
		// Note: Detailed nested key update usually requires reflection or a dedicated updater.
		// For the CLI, we'll provide a simple message that we're updating the main config.

		// Map simple keys for CLI convenience
		if key == "admin_password" {
			cfg.AdminPassword = value
		} else if key == "language" {
			cfg.Language = value
		} else {
			fmt.Printf("?†Ô∏è  Key '%s' update via CLI is limited. Please use Web Admin for advanced settings.\n", key)
			return
		}

		if err := config.SaveConfig(configPath, cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}
		fmt.Printf("??Saved '%s' = %s directly to %s\n", key, value, configPath)
	case "reset":
		fmt.Println("Resetting to default config...")
		defaultCfg := config.DefaultConfig()
		if err := config.SaveConfig(configPath, defaultCfg); err != nil {
			fmt.Printf("Error resetting config: %v\n", err)
		} else {
			fmt.Println("??Configuration reset to defaults.")
		}
	default:
		configHelp()
	}
}

func configHelp() {
	fmt.Println("\nConfig commands:")
	fmt.Println("  reset             Reset config.json to defaults")
	fmt.Println("  set <key> <val>   Set a value in config.json (e.g. admin_password, language)")
	fmt.Println("  show              Show current configuration")
}

func statusCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	configPath := getConfigPath()
	workspace := cfg.WorkspacePath()

	fmt.Printf("%s Status\n\n", logo)

	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Workspace: %s\n", workspace)
	fmt.Printf("Model: %s\n", cfg.Agents.Defaults.Model)

	hasOpenRouter := len(cfg.Providers.OpenRouter.Models) > 0
	hasAnthropic := len(cfg.Providers.Anthropic.Models) > 0
	hasOpenAI := len(cfg.Providers.OpenAI.Models) > 0
	hasGemini := len(cfg.Providers.Gemini.Models) > 0
	hasZhipu := len(cfg.Providers.Zhipu.Models) > 0
	hasGroq := len(cfg.Providers.Groq.Models) > 0
	hasVLLM := len(cfg.Providers.VLLM.Models) > 0

	maskKey := func(key string) string {
		if key == "" {
			return "not set"
		}
		if len(key) <= 8 {
			return "(set)"
		}
		return fmt.Sprintf("%s...%s", key[:4], key[len(key)-4:])
	}

	status := func(enabled bool) string {
		if enabled {
			return "(OK)"
		}
		return "not set"
	}

	fmt.Printf("OpenRouter API: %s\n", status(hasOpenRouter))
	fmt.Printf("Anthropic API: %s\n", status(hasAnthropic))
	fmt.Printf("OpenAI API: %s\n", status(hasOpenAI))
	fmt.Printf("Gemini API: %s\n", status(hasGemini))
	fmt.Printf("Zhipu API: %s\n", status(hasZhipu))
	fmt.Printf("Groq API: %s\n", status(hasGroq))

	if hasVLLM {
		m := cfg.Providers.VLLM.Models[0]
		fmt.Printf("vLLM/Local API: (OK)\n")
		fmt.Printf("  - Base:  %s\n", m.APIBase)
		fmt.Printf("  - Model: %s\n", m.Model)
		fmt.Printf("  - Key:   %s\n", maskKey(m.APIKey))
	} else {
		fmt.Printf("vLLM/Local: not set\n")
	}
}

func getResourceDir() string {
	if overrideResourceDir != "" {
		return overrideResourceDir
	}
	if h := os.Getenv("MARUBOT_HOME"); h != "" {
		return h
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".marubot")
}

func getConfigPath() string {
	return filepath.Join(getResourceDir(), "config.json")
}

func installAndRunSystemdService(exePath string) error {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return fmt.Errorf("systemctl not found")
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	servicePath := filepath.Join(serviceDir, "marubot.service")
	serviceContent := fmt.Sprintf(`[Unit]
Description=MaruBot Service
After=network.target

[Service]
Type=simple
Environment="MARUBOT_DAEMON=1"
ExecStart=%s start --foreground
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
`, exePath)

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return err
	}

	exec.Command("loginctl", "enable-linger", u.Username).Run()

	cmds := [][]string{
		{"systemctl", "--user", "daemon-reload"},
		{"systemctl", "--user", "enable", "marubot.service"},
		{"systemctl", "--user", "start", "marubot.service"},
	}

	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if os.Getenv("XDG_RUNTIME_DIR") == "" {
			cmd.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%s", u.Uid))
		} else {
			cmd.Env = os.Environ()
		}
		cmd.Run() // ignore specific errors
	}

	return nil
}

func reloadCmd() {
	fmt.Println("Reloading MaruBot...")
	if runtime.GOOS == "linux" {
		_, err := exec.LookPath("systemctl")
		if err == nil {
			u, _ := user.Current()
			uid := ""
			if u != nil {
				uid = u.Uid
			}

			serviceDir := ""
			if u != nil {
				serviceDir = filepath.Join(u.HomeDir, ".config", "systemd", "user")
			}
			servicePath := filepath.Join(serviceDir, "marubot.service")
			if _, err := os.Stat(servicePath); err == nil {
				daemonReload := execHidden("systemctl", "--user", "daemon-reload")
				restart := execHidden("systemctl", "--user", "restart", "marubot.service")

				if os.Getenv("XDG_RUNTIME_DIR") == "" && uid != "" {
					daemonReload.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%s", uid))
					restart.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%s", uid))
				} else {
					daemonReload.Env = os.Environ()
					restart.Env = os.Environ()
				}

				daemonReload.Run()
				if err := restart.Run(); err == nil {
					fmt.Println("??Reloaded via systemd.")
					return
				}
			}
		}
	}

	// For macOS/Windows:
	os.Unsetenv("MARUBOT_DAEMON")

	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("??Executable path error: %v\n", err)
		return
	}

	cmd := execHidden(exe, "start")
	if err := cmd.Start(); err != nil {
		fmt.Printf("??Failed to start during reload: %v\n", err)
		return
	}
	fmt.Println("??Reload trigger sent.")
}

func reloadInternal() {
	backgroundMu.Lock()
	defer backgroundMu.Unlock()

	logger.InfoC("system", "Starting internal hot-reload...")

	// 1. Stop existing services
	if backgroundCancel != nil {
		backgroundCancel()
	}
	if currentCron != nil {
		currentCron.Stop()
	}
	if currentHeartbeat != nil {
		currentHeartbeat.Stop()
	}

	// Wait a bit for sockets/connections to clear
	time.Sleep(1 * time.Second)

	// 2. Refresh config
	cfg, err := config.LoadConfig(getConfigPath())
	if err != nil {
		logger.ErrorCF("system", "Failed to reload config during hot-reload", map[string]interface{}{"error": err.Error()})
		return
	}

	// 3. Re-initialize provider
	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		logger.ErrorCF("system", "Failed to recreate provider during hot-reload", map[string]interface{}{"error": err.Error()})
		return
	}

	// 4. Update Agent Loop with new provider
	if currentAgentLoop != nil {
		currentAgentLoop.SetProvider(provider)
	}

	// 5. Restart services
	backgroundCtx, backgroundCancel = context.WithCancel(context.Background())
	
	// Re-init channels with new config
	if currentAgentLoop != nil {
		newChanManager, err := channels.NewManager(cfg, currentAgentLoop.GetBus())
		if err == nil {
			currentChanManager = newChanManager
			currentAgentLoop.SetChannelManager(newChanManager)
			currentChanManager.StartAll(backgroundCtx)
		}
	}

	if currentCron != nil {
		currentCron.Start()
	}
	if currentHeartbeat != nil {
		currentHeartbeat.Start()
	}

	go currentAgentLoop.Run(backgroundCtx)

	logger.InfoC("system", "Internal hot-reload completed successfully")
}

func startCmd() {
	// Check for flags or service mode
	var runForeground bool
	if len(os.Args) > 2 && (os.Args[2] == "--foreground" || os.Args[2] == "-f") {
		runForeground = true
	}
	if os.Getenv("MARUBOT_SERVICE") == "1" {
		runForeground = true
	}

	// Double-fork / Detach logic
	if !runForeground && os.Getenv("MARUBOT_DAEMON") != "1" {
		// Clean up existing instance before starting a new one in background
		stopCmd()
		time.Sleep(500 * time.Millisecond)

		exe, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
			return
		}

		if runtime.GOOS == "linux" {
			err = installAndRunSystemdService(exe)
			if err == nil {
				fmt.Println("??MaruBot started as a systemd service.")
				fmt.Println("   URL: http://localhost:8080")
				return
			}
		}

		// Re-run with special env var
		cmd := execHidden(exe, "start")
		// Clean up inherited DAEMON env var if any
		newEnv := make([]string, 0)
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "MARUBOT_DAEMON=") {
				newEnv = append(newEnv, e)
			}
		}
		cmd.Env = append(newEnv, "MARUBOT_DAEMON=1")
		// Detach process
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start background process: %v\n", err)
			os.Exit(1)
		}
		pidFile := getPidFilePath()
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)

		fmt.Printf("??MaruBot Dashboard started in background (PID: %d)\n", cmd.Process.Pid)
		fmt.Println("   URL: http://localhost:8080")
		fmt.Println("   To stop: use 'marubot stop'")
		fmt.Println("   To reload config: use 'marubot reload'")
		fmt.Println("   Logs: ~/.marubot/dashboard.log")
		return
	}

	// Daemon Code Starts Here

	// Setup logging to file if daemon
	if os.Getenv("MARUBOT_DAEMON") == "1" {
		logFile := filepath.Join(getResourceDir(), "dashboard.log")
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			// Redirect stdout/stderr to log file
			os.Stdout = f
			os.Stderr = f
			fmt.Printf("\n--- MaruBot Log Started at %v ---\n", time.Now().Format(time.RFC3339))
		}
		// Note: we don't defer f.Close() here because it should stay open for the life of the process
	}

	if runForeground {
		fmt.Printf("%s Starting MaruBot Dashboard & API Server...\n", logo)
	}

	// Always write PID file so stopCmd/reloadCmd can find us
	pidFile := getPidFilePath()
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	cfg, err := loadConfig()
	if err != nil {
		if runForeground {
			fmt.Printf("Error loading config: %v\n", err)
		}
		return
	}

	// Ensure config file is in sync with latest struct (e.g. adding missing Slack fields)
	if err := config.SaveConfig(getConfigPath(), cfg); err != nil {
		fmt.Printf("Warning: Failed to sync config file: %v\n", err)
	}

	if cfg.AdminPassword == "" {
		ensureAdminPassword(cfg)
	}

	// Validate configuration: At least one AI provider must be enabled OR one channel must be enabled.
	// If password is also missing, we prioritize security setup.
	if !cfg.IsAIConfigured() && !cfg.IsChannelEnabled() {
		showGuideMessage(cfg)

		// Allow all platforms to enter Setup Mode so the user can finish configuration via Web-Admin.
		fmt.Println("Entering Setup Mode... (Server remains active for web configuration)")

		// Start dashboard server in a goroutine so user can configure
		port := cfg.Gateway.Port
		if port == 0 {
			port = 8080
		}
		dashAddr := fmt.Sprintf("0.0.0.0:%d", port)

		// ?í° Fix: Properly initialize dummyAgent components to prevent Nil Pointer Panic
		// even in Setup Mode.
		bus := bus.NewMessageBus()
		dummyAgent := agent.NewAgentLoop(cfg, bus, nil, Version)
		dashServer := dashboard.NewServer(dashAddr, dummyAgent, cfg, getConfigPath(), Version, reloadInternal)

		if runForeground {
			fmt.Printf("??Dashboard available at http://localhost:%d\n", port)
			fmt.Println("  Please complete your configuration in the Web Admin.")
		}

		go func() {
			if err := dashServer.Start(); err != nil {
				fmt.Printf("Dashboard failed to start: %v\n", err)
			}
		}()

		// Wait for interrupt to exit
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("Setup Mode exiting...")
		os.Exit(0)
	}

	workspace := cfg.WorkspacePath()
	createWorkspaceTemplates(workspace)

	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		if runForeground {
			fmt.Printf("Error creating provider: %v\n", err)
		}
		os.Exit(1)
	}

	bus := bus.NewMessageBus()

	// Enable logging to file for background service with daily rotation logic
	logDir := filepath.Join(getResourceDir(), "logs")
	if err := logger.EnableDailyRotation(logDir, "marubot"); err != nil {
		if runForeground {
			fmt.Printf("Warning: Failed to enable daily log rotation: %v\n", err)
		}
	}

	agentLoop := agent.NewAgentLoop(cfg, bus, provider, Version)
	currentAgentLoop = agentLoop

	gpioService := gpio.NewGPIOService(cfg, bus)
	gpioService.Start(context.Background())
	defer gpioService.Stop()

	// Background Services
	cronStorePath := filepath.Join(filepath.Dir(getConfigPath()), "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, func(job *cron.CronJob) (string, error) {
		return agentLoop.ProcessDirect(context.Background(), job.Payload.Message, "cron:"+job.ID)
	})
	currentCron = cronService

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		func(msg string) (string, error) {
			return agentLoop.ProcessDirect(context.Background(), msg, "heartbeat")
		},
		30*60,
		true,
	)
	currentHeartbeat = heartbeatService

	backgroundCtx, backgroundCancel = context.WithCancel(context.Background())
	defer backgroundCancel()

	if err := cronService.Start(); err != nil && runForeground {
		fmt.Printf("Error starting cron service: %v\n", err)
	}
	if err := heartbeatService.Start(); err != nil && runForeground {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	go agentLoop.Run(backgroundCtx)

	channelManager, err := channels.NewManager(cfg, bus)
	if err == nil {
		currentChanManager = channelManager
		agentLoop.SetChannelManager(channelManager)
		if err := channelManager.StartAll(backgroundCtx); err != nil && runForeground {
			fmt.Printf("Error starting channels: %v\n", err)
		}
		if runForeground {
			fmt.Println("??Background services started (Cron, Heartbeat, Channels)")
		}
	} else {
		if runForeground {
			fmt.Printf("Warning: Failed to initialize channel manager: %v\n", err)
			fmt.Println("??Background services started (Cron, Heartbeat)")
		}
	}

	// Initialize Dashboard Server
	port := "8080"
	server := dashboard.NewServer(":"+port, agentLoop, cfg, getConfigPath(), Version, reloadInternal)

	if runForeground {
		go func() {
			time.Sleep(1 * time.Second)
			fmt.Printf("??Dashboard available at http://localhost:%s\n", port)
		}()
	}

	if err := server.Start(); err != nil {
		if runForeground {
			fmt.Printf("Error starting dashboard server: %v\n", err)
		}
	}
}

func getPidFilePath() string {
	return filepath.Join(getResourceDir(), "marubot.pid")
}

func isMarubotProcessRunning() bool {
	pidFile := getPidFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pidStr := strings.TrimSpace(string(data))
	var pid int
	fmt.Sscanf(pidStr, "%d", &pid)

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds. Need to send signal 0 to check if it's actually alive.
	if runtime.GOOS != "windows" {
		err := proc.Signal(syscall.Signal(0))
		return err == nil
	}

	return true
}

func stopCmd() {
	stoppedViaSystemd := false
	if runtime.GOOS == "linux" {
		_, err := exec.LookPath("systemctl")
		if err == nil {
			u, _ := user.Current()
			uid := ""
			if u != nil {
				uid = u.Uid
			}

			serviceDir := ""
			if u != nil {
				serviceDir = filepath.Join(u.HomeDir, ".config", "systemd", "user")
			}
			servicePath := filepath.Join(serviceDir, "marubot.service")
			if _, err := os.Stat(servicePath); err == nil {
				cmd := exec.Command("systemctl", "--user", "stop", "marubot.service")
				if os.Getenv("XDG_RUNTIME_DIR") == "" && uid != "" {
					cmd.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%s", uid))
				} else {
					cmd.Env = os.Environ()
				}
				if err := cmd.Run(); err == nil {
					fmt.Println("??Stopped systemd service.")
					stoppedViaSystemd = true
				}
			}
		}
	}

	if runtime.GOOS == "windows" {
		fmt.Println("Stopping MaruBot processes...")
		// Use taskkill to ensure all tray and background processes are killed
		execHidden("taskkill", "/F", "/T", "/IM", "marubot.exe").Run()
		execHidden("taskkill", "/F", "/T", "/IM", "marubot-*.exe").Run()
	}

	pidFile := getPidFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if !stoppedViaSystemd && runtime.GOOS != "windows" {
			fmt.Println("No running marubot process found (pid file missing).")
		}
		return
	}

	pidStr := strings.TrimSpace(string(data))
	var pid int
	fmt.Sscanf(pidStr, "%d", &pid)

	proc, err := os.FindProcess(pid)
	if err != nil {
		if !stoppedViaSystemd {
			fmt.Printf("Process %d not found.\n", pid)
		}
		os.Remove(pidFile)
		return
	}

	fmt.Printf("Stopping marubot (PID: %d)...\n", pid)
	if err := proc.Signal(os.Interrupt); err != nil {
		proc.Kill()
	}

	os.Remove(pidFile)
	fmt.Println("??Stopped.")
}

func upgradeCmd() {
	autoConfirm := false
	if len(os.Args) > 2 && os.Args[2] == "--yes" {
		autoConfirm = true
	}

	fmt.Println("?ôÔ∏è  Checking for updates...")

	latest, err := config.CheckLatestVersion()
	if err != nil {
		fmt.Printf("?†Ô∏è  Failed to check latest Version: %v\n", err)
		fmt.Println("Proceeding with forced upgrade...")
	} else {
		if !config.IsNewVersionAvailable(latest) && !autoConfirm {
			fmt.Printf("??You are already using the latest Version (v%s).\n", config.Version)
			fmt.Print("Do you want to reinstall anyway? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				return
			}
		} else if config.IsNewVersionAvailable(latest) && !autoConfirm {
			fmt.Printf("??New Version available: v%s (Current: v%s)\n", latest, config.Version)
			fmt.Print("Do you want to upgrade? [Y/n]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response == "n" || response == "no" {
				return
			}
		}
	}

	// Stop existing process if running
	stopCmd()

	if runtime.GOOS == "windows" {
		fmt.Println("?? Windows native upgrade in progress...")
		latest, err := config.CheckLatestVersion()
		if err != nil {
			fmt.Printf("??Failed to get latest Version: %v\n", err)
			return
		}

		// Use the correct public repo maru-bot
		downloadUrl := fmt.Sprintf("https://github.com/dirmich/maru-bot/releases/download/v%s/marubot.exe", latest)
		fmt.Printf("?ì• Downloading from: %s\n", downloadUrl)

		resp, err := http.Get(downloadUrl)
		if err != nil {
			fmt.Printf("??Download failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("??Download failed: HTTP %s\n", resp.Status)
			return
		}

		exePath, _ := os.Executable()
		oldExePath := exePath + ".old"
		newExePath := exePath + ".new"

		// Download to .new file first
		out, err := os.Create(newExePath)
		if err != nil {
			fmt.Printf("??Failed to create temp file: %v\n", err)
			return
		}
		_, err = io.Copy(out, resp.Body)
		out.Close()
		if err != nil {
			fmt.Printf("??Download failed during copy: %v\n", err)
			return
		}

		// Rename current to .old and .new to current
		os.Remove(oldExePath) // Clean up any previous failed attempt
		err = os.Rename(exePath, oldExePath)
		if err != nil {
			fmt.Printf("??Failed to rename current binary: %v\n", err)
			os.Remove(newExePath)
			return
		}

		err = os.Rename(newExePath, exePath)
		if err != nil {
			fmt.Printf("??Failed to install new binary: %v\n", err)
			// Try to restore old one
			os.Rename(oldExePath, exePath)
			return
		}

		fmt.Println("??Upgrade complete! Please restart MaruBot.")
		// If running as service, it might be better to let the user or SCM restart it
		return
	}

	// For UNIX-like systems, keep using the install.sh bash script
	fmt.Println("?? Upgrading MaruBot to the latest Version...")
	cmd := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("??Upgrade failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("??Upgrade complete! Restarting MaruBot...")
	reloadCmd()
}
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = execHidden("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Please open your browser and go to: %s\n", url)
	}
}

func showGuideMessage(cfg *config.Config) {
	lang := cfg.Language
	port := cfg.Gateway.Port
	if port == 0 {
		port = 18790
	}
	adminURL := fmt.Sprintf("http://localhost:%d", port)

	isGUI := runtime.GOOS == "windows" || runtime.GOOS == "darwin"

	messages := map[string]map[string]string{
		"ko": {
			"title":    "?†Ô∏è Ï¥àÍ∏∞ ?§Ï†ï???ÑÏöî?©Îãà??",
			"sec_body": "Î≥¥Ïïà???ÑÌï¥ Í¥ÄÎ¶¨Ïûê ÎπÑÎ?Î≤àÌò∏Î•?Î®ºÏ? ?§Ï†ï?¥Ïïº ?©Îãà??",
			"cfg_body": "MaruBot???úÏûë?òÎ†§Î©?ÏµúÏÜå ?òÎÇò??AI Î™®Îç∏Í≥?Ï±ÑÎÑê???§Ï†ï?òÏñ¥???©Îãà??",
			"gui":      "?§Ï†ï ?òÏù¥ÏßÄ(Web-Admin)Î•?Î∏åÎùº?∞Ï????ÑÏõÅ?àÎã§. ?§Ï†ï???ÑÎ£å??Ï£ºÏÑ∏??",
			"cli":      "Î∏åÎùº?∞Ï??êÏÑú ?ÑÎûò Ï£ºÏÜåÎ°??ëÏÜç?òÏó¨ ?§Ï†ï???ÑÎ£å??Ï£ºÏÑ∏??",
			"restart":  "?§Ï†ï ?ÑÎ£å ???±ÏùÑ ?¨Ïãú?ëÌïòÎ©??úÎπÑ?§Í? ?ïÏÉÅ Í∏∞Îèô?©Îãà??",
		},
		"en": {
			"title":    "?†Ô∏è Initial Configuration Required!",
			"sec_body": "For security, you must set an administrator password first.",
			"cfg_body": "At least one AI model and one channel must be configured to start MaruBot.",
			"gui":      "Opening configuration page (Web-Admin) in your browser. Please complete setup.",
			"cli":      "Please access the following URL in your browser to complete setup:",
			"restart":  "After setup, restart the app to start the main service.",
		},
		"ja": {
			"title":    "?†Ô∏è ?ùÊúüË®?Æö?åÂøÖË¶Å„Åß?ôÔºÅ",
			"sec_body": "?ª„Ç≠?•„É™?Ü„Ç£??Åü?Å„ÄÅ„Åæ?öÁ??ÜËÄÖ„ÅÆ?ë„Çπ??Éº?â„ÇíË®?Æö?ô„ÇãÂøÖË¶Å?å„ÅÇ?ä„Åæ?ô„Ä?,
			"cfg_body": "MaruBot?íÈñãÂßã„Åô?ã„Å´??ÄÅÂ∞ë?™„Åè?®„ÇÇ1?§„ÅÆAI?¢„Éá?´„Å®?Å„É£?ç„É´?íË®≠ÂÆö„Åô?ãÂøÖË¶Å„Åå?Ç„Çä?æ„Åô??,
			"gui":      "?ñ„É©?¶„Ç∂?ßË®≠ÂÆö„Éö?º„Ç∏(Web-Admin)?íÈñã?ç„Åæ?ô„ÄÇË®≠ÂÆö„ÇíÂÆå‰∫Ü?ó„Å¶?è„Å†?ï„ÅÑ??,
			"cli":      "?ñ„É©?¶„Ç∂?ß‰ª•‰∏ã„ÅÆURL?´„Ç¢??Çª?π„Åó??®≠ÂÆö„ÇíÂÆå‰∫Ü?ó„Å¶?è„Å†?ï„ÅÑÔº?,
			"restart":  "Ë®?ÆöÂÆå‰∫ÜÂæå„ÄÅ„Ç¢?ó„É™?íÂÜçËµ∑Âãï?ô„Çã?®„Çµ?º„Éì?π„Åå?ãÂßã?ï„Çå?æ„Åô??,
		},
		"zh": {
			"title":    "?†Ô∏è ?ÄË¶ÅÂàùÂßãÈÖçÁΩ?ºÅ",
			"sec_body": "?∫‰∫éÂÆâÂÖ®?ÉËôëÔºåÊÇ®ÂøÖÈ°ª?àË?ÁΩ???ÜÂëòÂØÜÁ†Å??,
			"cfg_body": "??ä® MaruBot ?≥Â∞ë?ÄË¶ÅÈÖçÁΩ??‰∏?AI Ê®°Âûã?å‰?‰∏™È¢ë?ì„Ä?,
			"gui":      "Ê≠£Âú®ÊµèËßà?®‰∏≠?ìÂ??çÁΩÆÈ°µÈù¢ (Web-Admin)?ÇË?ÂÆåÊàê?çÁΩÆ??,
			"cli":      "ËØ∑Âú®ÊµèËßà?®‰∏≠ËÆøÈóÆ‰ª•‰∏ã?∞Â?ÂÆåÊàê?çÁΩÆÔº?,
			"restart":  "?çÁΩÆÂÆåÊàê?éÔºåËØ∑Èáç??®ãÂ∫è‰ª•??ä®‰∏ªÊúç?°„Ä?,
		},
	}

	msg, ok := messages[lang]
	if !ok {
		msg = messages["en"]
	}

	fmt.Println("\n" + msg["title"])
	if cfg.AdminPassword == "" {
		fmt.Println(msg["sec_body"])
	} else {
		fmt.Println(msg["cfg_body"])
	}
	fmt.Println()

	if isGUI {
		fmt.Println(msg["gui"])
		fmt.Println("URL:", adminURL)
		openBrowser(adminURL)
	} else {
		fmt.Println(msg["cli"])
		fmt.Println(adminURL)
	}
	fmt.Println("\n" + msg["restart"] + "\n")
}

// Windows Service Implementation
type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	// Service started by SCM
	go p.run()
	return nil
}

func (p *program) run() {
	// Core service logic
	os.Setenv("MARUBOT_SERVICE", "1")
	startCmd()
}

func (p *program) Stop(s service.Service) error {
	// Service stopped by SCM
	stopCmd()
	return nil
}

func serviceCmd() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: marubot service [install|uninstall|start|stop|restart|status]")
		return
	}

	sub := os.Args[2]

	if runtime.GOOS == "windows" {
		if sub == "install" || sub == "uninstall" || sub == "start" || sub == "stop" || sub == "restart" {
			if !isAdmin() {
				fmt.Println("Elevation required. Requesting administrator privileges...")
				runAsAdmin()
				return
			}
		}
	}

	svcConfig := &service.Config{
		Name:        "MaruBot",
		DisplayName: "MaruBot Service",
		Description: "Ultra-lightweight personal AI agent service.",
		Arguments:   []string{"--home", getResourceDir(), "service", "run"},
	}

	prg := &program{
		exit: make(chan struct{}),
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Printf("Service initialization failed: %v\n", err)
		return
	}

	switch sub {
	case "install":
		if runtime.GOOS == "windows" {
			targetPath, err := installBinary()
			if err == nil {
				svcConfig.Executable = targetPath
				fmt.Printf("Installing service pointing to: %s\n", targetPath)
			}
		}
		err = s.Install()
		if err == nil {
			fmt.Println("Service installed successfully. (Autostart enabled)")
		}
	case "uninstall":
		err = s.Uninstall()
		if runtime.GOOS == "windows" {
			// Extra cleanup for Windows to ensure it's removed from SCM immediately if possible
			execHidden("sc", "stop", svcConfig.Name).Run()
			execHidden("sc", "delete", svcConfig.Name).Run()
		}
		if err == nil {
			fmt.Println("Service uninstalled successfully.")
		}
	case "start":
		err = s.Start()
		if err == nil {
			fmt.Println("Service started.")
		}
	case "run":
		// This is called by Windows SCM to run the service
		err = s.Run()
		return
	case "stop":
		err = s.Stop()
		if err == nil {
			fmt.Println("Service stopped.")
		}
	case "restart":
		err = s.Restart()
		if err == nil {
			fmt.Println("Service restarted.")
		}
	case "status":
		status, _ := s.Status()
		fmt.Printf("Service status: %v\n", status)
		return
	default:
		fmt.Printf("Unknown service command: %s\n", sub)
		return
	}

	if err != nil {
		fmt.Printf("Service operation failed: %v\n", err)
	}
}

func isAdmin() bool {
	if runtime.GOOS == "windows" {
		// 'net session' command returns 0 if admin, 1 otherwise
		cmd := execHidden("net", "session")
		err := cmd.Run()
		return err == nil
	}
	return os.Geteuid() == 0
}

func runAsAdmin() {
	if runtime.GOOS != "windows" {
		return
	}
	// Check if already attempted to avoid cycles
	for _, arg := range os.Args {
		if arg == "--elevated" {
			fmt.Println("Already tried to elevate and failed. Please run manually as Administrator.")
			return
		}
	}

	exe, _ := os.Executable()
	args := strings.Join(append(os.Args[1:], "--elevated"), " ")

	// We use ' to wrap the path to handle spaces in exe path
	command := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs", exe, args)
	cmd := execHidden("powershell", "-Command", command)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to elevate: %v\n", err)
		fmt.Println("Please run this command in an Administrator terminal.")
	} else {
		fmt.Println("Elevated process requested. This process will now exit.")
	}
}
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func runAsAdminAction(action string) {
	exe, _ := os.Executable()
	// Split action into individual quoted arguments for PowerShell
	fields := strings.Fields(action)
	quotedArgs := make([]string, len(fields))
	isUninstall := false
	for i, f := range fields {
		quotedArgs[i] = fmt.Sprintf("'%s'", f)
		if strings.ToLower(f) == "uninstall" {
			isUninstall = true
		}
	}
	argList := strings.Join(quotedArgs, ", ")

	// For uninstallation, we want the window to be visible so the user can answer questions
	windowStyle := "Hidden"
	if isUninstall {
		windowStyle = "Normal"
	}

	// Use PowerShell Start-Process with -Verb RunAs to request elevation for the uninstall action
	psCmd := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList %s -Verb RunAs -WindowStyle %s", exe, argList, windowStyle)
	exec.Command("powershell", "-NoProfile", "-WindowStyle", windowStyle, "-Command", psCmd).Start()
}

func checkAndFixPort(cfg *config.Config) bool {
	// Standardize on 8080. If config has 18790, override it to 8080.
	if cfg.Gateway.Port == 18790 || cfg.Gateway.Port == 0 {
		cfg.Gateway.Port = 8080
	}

	port := cfg.Gateway.Port
	if isPortAvailable(port) {
		return true
	}

	if runtime.GOOS == "windows" {
		// Identify the process owning the port
		detectScript := fmt.Sprintf(`$conn = Get-NetTCPConnection -LocalPort %d -ErrorAction SilentlyContinue | Select-Object -First 1; if ($conn) { $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue; if ($proc) { $proc.ProcessName } else { "unknown" } } else { "none" }`, port)
		cmd := execHidden("powershell", "-NoProfile", "-Command", detectScript)
		out, _ := cmd.Output()
		ownerName := strings.TrimSpace(strings.ToLower(string(out)))
		isMarubot := strings.Contains(ownerName, "marubot")

		var promptScript string
		if isMarubot {
			promptScript = fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$title = "MaruBot Port Conflict"
$msg = "MaruBot is already running on port %d.` + "`n" + `Would you like to terminate the existing process and continue?"
$result = [System.Windows.Forms.MessageBox]::Show($msg, $title, "YesNo", "Warning")
if ($result -eq "Yes") { "kill" } else { "exit" }
`, port)
		} else {
			promptScript = fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$title = "MaruBot Port Conflict"
$msg = "Port %d is being used by '%s'.` + "`n" + `Would you like to use a different port?"
$result = [System.Windows.Forms.MessageBox]::Show($msg, $title, "YesNo", "Question")
if ($result -eq "Yes") { "newport" } else { "exit" }
`, port, ownerName)
		}

		cmd = exec.Command("powershell", "-NoProfile", "-Command", promptScript)
		out, _ = cmd.Output()
		action := strings.TrimSpace(string(out))

		if action == "kill" {
			killScript := fmt.Sprintf(`
$conn = Get-NetTCPConnection -LocalPort %d -ErrorAction SilentlyContinue | Select-Object -First 1
if ($conn) {
    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
    "ok"
} else { "notfound" }
`, port)
			exec.Command("powershell", "-NoProfile", "-Command", killScript).Run()
			time.Sleep(1 * time.Second)
			if isPortAvailable(port) {
				return true
			}
		}

		if action == "newport" {
			inputScript := fmt.Sprintf(`
Add-Type -AssemblyName Microsoft.VisualBasic
$newPort = [Microsoft.VisualBasic.Interaction]::InputBox("Enter a new port number for MaruBot:", "MaruBot", "8081")
if ($newPort) { $newPort } else { exit 1 }
`)
			cmd = exec.Command("powershell", "-NoProfile", "-Command", inputScript)
			out, err := cmd.Output()
			if err == nil {
				var newPort int
				fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &newPort)
				if newPort > 0 {
					cfg.Gateway.Port = newPort
					if err := config.SaveConfig(getConfigPath(), cfg); err != nil {
						fmt.Printf("Error saving config: %%v\n", err)
					}
					return true
				}
			}
		}
		return false
	} else if runtime.GOOS == "darwin" {
		// Ask for a new port via AppleScript (osascript)
		title := "MaruBot Port Conflict"
		msg := fmt.Sprintf("Port %d is already in use. Would you like to use a different port?", port)
		script := fmt.Sprintf(`display dialog "%s" with title "%s" buttons {"Cancel", "Yes"} default button "Yes" with icon caution`, msg, title)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			return false // User cancelled
		}

		// Show input dialog
		inputScript := fmt.Sprintf(`display dialog "Enter a new port number:" with title "%s" default answer "8081" buttons {"Cancel", "OK"} default button "OK"`, title)
		out, err := exec.Command("osascript", "-e", inputScript).Output()
		if err != nil {
			return false // User cancelled
		}

		// osascript returns: button returned:OK, text returned:8081
		outStr := string(out)
		if strings.Contains(outStr, "text returned:") {
			parts := strings.Split(outStr, "text returned:")
			newPortStr := strings.TrimSpace(parts[len(parts)-1])
			var newPort int
			fmt.Sscanf(newPortStr, "%d", &newPort)
			if newPort > 0 {
				cfg.Gateway.Port = newPort
				if err := config.SaveConfig(getConfigPath(), cfg); err != nil {
					fmt.Printf("Error saving config: %v\n", err)
				}
				return true
			}
		}
	} else {
		fmt.Printf("Warning: Port %d is in use. Please check your configuration.\n", port)
	}
	return false
}

func setupWorkspace() error {
	baseDir := getResourceDir()
	dirs := []string{
		"bin",
		"config",
		"skills",
		"tools",
		"workspace",
		"db",
	}

	for _, d := range dirs {
		path := filepath.Join(baseDir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}

	// Double check IDENTITY or initial config if needed
	configPath := filepath.Join(baseDir, "config", "maru-config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If not exists, we'll let LoadConfig handle it or we could copy a default here
		// For now, at least the folders are there
	}

	return nil
}

func installBinary() (string, error) {
	exe, _ := os.Executable()
	installDir := filepath.Join(getResourceDir(), "bin")
	os.MkdirAll(installDir, 0755)

	targetPath := filepath.Join(installDir, "marubot.exe")
	if runtime.GOOS != "windows" {
		targetPath = filepath.Join(installDir, "marubot")
	}

	// If already in target path, skip
	if exe == targetPath {
		return targetPath, nil
	}

	fmt.Printf("Installing binary to %s...\n", targetPath)
	src, err := os.Open(exe)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return targetPath, nil
}

func getTargetBinaryPath() (string, error) {
	installDir := filepath.Join(getResourceDir(), "bin")
	targetPath := filepath.Join(installDir, "marubot.exe")
	if runtime.GOOS != "windows" {
		targetPath = filepath.Join(installDir, "marubot")
	}
	return targetPath, nil
}

func serviceCmdInternalPath(sub string, exePath string) {
	if exePath == "" {
		exePath, _ = os.Executable()
	}
	cmd := exec.Command(exePath, "service", sub)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func migratePathsCmd() {
	oldPath := `C:\WINDOWS\system32\config\systemprofile\.marubot`
	newPath := getResourceDir()

	fmt.Printf("Migrating hardcoded paths from %s to %s...\n", oldPath, newPath)

	count := 0
	err := filepath.Walk(newPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".json" || ext == ".md" || ext == ".txt" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			content := string(data)
			if strings.Contains(content, oldPath) || strings.Contains(content, strings.ReplaceAll(oldPath, "\\", "/")) {
				fmt.Printf("  Processing: %s\n", path)
				newContent := strings.ReplaceAll(content, oldPath, newPath)
				newContent = strings.ReplaceAll(newContent, strings.ReplaceAll(oldPath, "\\", "/"), strings.ReplaceAll(newPath, "\\", "/"))

				err = os.WriteFile(path, []byte(newContent), info.Mode())
				if err != nil {
					fmt.Printf("    Error saving %s: %v\n", path, err)
				} else {
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("\n??Migration failed: %v\n", err)
	} else {
		fmt.Printf("\n??Migration complete! Updated %d files.\n", count)
		fmt.Println("Please restart MaruBot to apply changes.")
	}
}

func checkServiceUpgrade(s service.Service) bool {
	// We need to find the path of the existing service
	// On Windows, SC query or checking service config is needed.
	// kardianos/service doesn't expose the path easily, let's use sc.exe
	out, err := exec.Command("sc", "qc", "MaruBot").Output()
	if err != nil {
		return false
	}

	// Extract BINARY_PATH_NAME (handles optional quotes and potential arguments)
	// Example: BINARY_PATH_NAME   : "C:\path\to\marubot.exe" start
	// or: BINARY_PATH_NAME   : C:\path\to\marubot.exe start
	re := regexp.MustCompile(`BINARY_PATH_NAME\s*:\s*("([^"]+)"|([^\s]+))`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return false
	}

	var svcPath string
	if matches[2] != "" {
		svcPath = matches[2] // Quoted path
	} else {
		svcPath = matches[3] // Unquoted path
	}
	if svcPath == "" {
		return false
	}

	// Get version of that binary
	cmd := exec.Command(svcPath, "--version-only")
	svcVerOut, err := cmd.Output()
	if err != nil {
		return false
	}

	svcVer := strings.TrimSpace(string(svcVerOut))
	currentVer := Version

	// Simple string compare for now
	return svcVer != currentVer && svcVer != ""
}



func removeWindowsShortcuts() {
	if runtime.GOOS != "windows" {
		return
	}
	psScript := `
$shell = New-Object -ComObject WScript.Shell
$desktop = [System.Environment]::GetFolderPath('Desktop')
$startMenu = [System.Environment]::GetFolderPath('StartMenu')
$programs = Join-Path $startMenu "Programs"

$targets = @(Join-Path $desktop "MaruBot.lnk", Join-Path $programs "MaruBot.lnk")

foreach ($t in $targets) {
    if (Test-Path $t) {
        Remove-Item $t -Force
    }
}
`
	exec.Command("powershell", "-NoProfile", "-Command", psScript).Run()
}
