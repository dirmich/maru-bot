// MaruBot - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 MaruBot contributors

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/dirmich/marubot/cmd/marubot/dashboard"
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
	"github.com/dirmich/marubot/pkg/voice"

	"github.com/chzyer/readline"
	"github.com/getlantern/systray"
	"github.com/kardianos/service"
)

// 0.4.55: Fix Windows elevation infinite loop, improve runAsAdmin with MARUBOT_ELEVATED
// 0.4.54: Force 8080 port, Complete uninstall cleanup, Fix lint warnings
// 0.4.49: Fix Windows binary corruption by keeping releases clean, add service elevation (Admin check)

var Version = config.Version

const logo = "[MaruBot]"

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
	if len(os.Args) < 2 {
		if runtime.GOOS == "windows" {
			handleWindowsGUIMode()
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
	case "start":
		startCmd()
	case "reload":
		reloadCmd()
	case "skills":
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
		case "install-builtin":
			skillsInstallBuiltinCmd(workspace)
		case "list-builtin":
			skillsListBuiltinCmd()
		case "search":
			skillsSearchCmd(installer)
		case "show":
			if len(os.Args) < 4 {
				fmt.Println("Usage: marubot skills show <skill-name>")
				return
			}
			skillsShowCmd(skillsLoader, os.Args[3])
		default:
			fmt.Printf("Unknown skills command: %s\n", subcommand)
			skillsHelp()
		}
	case "version", "--version", "-v":
		fmt.Printf("%s marubot v%s\n", logo, Version)
	case "uninstall":
		uninstallCmd()
	case "stop":
		stopCmd()
	case "upgrade":
		upgradeCmd()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func uninstallCmd() {
	if runtime.GOOS == "windows" {
		if !isAdmin() {
			fmt.Println("Elevation required for uninstallation. Requesting administrator privileges...")
			runAsAdmin()
			return
		}
	}

	// Ensure all processes are stopped before uninstallation
	stopCmd()

	fmt.Printf("%s MaruBot Uninstaller\n", logo)
	fmt.Println("WARNING: This will remove MaruBot and its resources from your system.")

	fmt.Print("Are you sure you want to continue? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Aborted.")
		return
	}

	// 0. Remove Windows Service (if applicable)
	if runtime.GOOS == "windows" {
		svcConfig := &service.Config{
			Name: "MaruBot",
		}
		s, err := service.New(nil, svcConfig)
		if err == nil {
			fmt.Println("Removing Windows service...")
			s.Stop()
			if err := s.Uninstall(); err == nil {
				fmt.Println("✓ Windows service removed.")
			} else {
				// Fallback to sc delete
				fmt.Printf("Standard uninstall failed (%v), trying sc delete...\n", err)
				exec.Command("sc", "stop", "MaruBot").Run()
				if err := exec.Command("sc", "delete", "MaruBot").Run(); err == nil {
					fmt.Println("✓ Windows service removed via sc delete.")
				} else {
					fmt.Println("✗ Failed to remove Windows service. You may need to run 'sc delete MaruBot' manually.")
				}
			}
		}
	}

	fmt.Print("Do you want to keep your user data (config, memory, workspace)? (Y/n): ")
	var keepData string
	fmt.Scanln(&keepData)
	keep := true
	if strings.ToLower(keepData) == "n" {
		keep = false
	}

	// 1. Remove resources
	resourceDir := getResourceDir()
	if _, err := os.Stat(resourceDir); err == nil {
		if keep {
			fmt.Println("Cleaning system resources (keeping user data)...")
			os.RemoveAll(filepath.Join(resourceDir, "skills"))
			os.RemoveAll(filepath.Join(resourceDir, "tools"))
			os.RemoveAll(filepath.Join(resourceDir, "web-admin"))
			fmt.Printf("✓ System resources removed. User data kept in %s\n", resourceDir)
		} else {
			fmt.Println("Removing all data...")
			if err := os.RemoveAll(resourceDir); err != nil {
				fmt.Printf("Error removing %s: %v\n", resourceDir, err)
			} else {
				fmt.Printf("✓ %s removed\n", resourceDir)
			}
		}
	}

	// 2. Remove binary
	// Try to remove self using os.Executable
	exePath, err := os.Executable()
	if err == nil {
		fmt.Printf("Removing executable: %s\n", exePath)
		if runtime.GOOS == "windows" {
			// Windows cannot delete a running executable. 
			// Use a PowerShell trick to delete after exit.
			destDir := filepath.Dir(exePath)
			exeName := filepath.Base(exePath)
			script := fmt.Sprintf("Start-Sleep -Seconds 2; Remove-Item -Path '%s' -Force", exeName)
			cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Start-Process powershell -ArgumentList \"-Command %s\" -WindowStyle Hidden -WorkingDirectory '%s'", script, destDir))
			if err := cmd.Start(); err != nil {
				fmt.Printf("Error scheduling self-deletion: %v\n", err)
			} else {
				fmt.Println("✓ Executable scheduled for deletion after exit.")
			}
		} else {
			if err := os.Remove(exePath); err != nil {
				fmt.Printf("Error removing executable: %v\n", err)
				fmt.Println("Hint: You may need to run this command with sudo: 'sudo marubot uninstall'")
			} else {
				fmt.Println("✓ Executable removed")
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
MaruBot 🦞

## Description
Ultra-lightweight personal AI assistant written in Go, inspired by nanobot.

## Version
" + config.Version + "

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
- SSH & Remote System Access (Automated execution via shell)
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
https://marubot

## Contact
Issues: https://marubot/issues
Discussions: https://marubot/discussions

---

"Every bit helps, every bit matters."
- Picoclaw
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

	cronStorePath := filepath.Join(filepath.Dir(getConfigPath()), "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, func(job *cron.CronJob) (string, error) {
		return agentLoop.ProcessDirect(context.Background(), job.Payload.Message, "cron:"+job.ID)
	})

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		func(msg string) (string, error) {
			return agentLoop.ProcessDirect(context.Background(), msg, "heartbeat")
		},
		30*60,
		true,
	)

	channelManager, err := channels.NewManager(cfg, bus)
	if err != nil {
		fmt.Printf("Error creating channel manager: %v\n", err)
		os.Exit(1)
	}

	var transcriber *voice.GroqTranscriber
	if cfg.Providers.Groq.APIKey != "" {
		transcriber = voice.NewGroqTranscriber(cfg.Providers.Groq.APIKey)
		logger.InfoC("voice", "Groq voice transcription enabled")
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
		fmt.Printf("✓ Channels enabled: %s\n", enabledChannels)
	} else {
		fmt.Println("⚠ Warning: No channels enabled")
	}

	fmt.Printf("✓ Gateway started on %s:%d\n", cfg.Gateway.Host, cfg.Gateway.Port)
	fmt.Println("Press Ctrl+C to stop")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := cronService.Start(); err != nil {
		fmt.Printf("Error starting cron service: %v\n", err)
	}
	fmt.Println("✓ Cron service started")

	if err := heartbeatService.Start(); err != nil {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	fmt.Println("✓ Heartbeat service started")

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
	fmt.Println("✓ Gateway stopped")
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

	fmt.Printf("✓ Added job '%s' (%s)\n", job.Name, job.ID)
}

func cronRemoveCmd(storePath, jobID string) {
	cs := cron.NewCronService(storePath, nil)
	if cs.RemoveJob(jobID) {
		fmt.Printf("✓ Removed job %s\n", jobID)
	} else {
		fmt.Printf("✗ Job %s not found\n", jobID)
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
		fmt.Printf("✓ Job '%s' %s\n", job.Name, status)
	} else {
		fmt.Printf("✗ Job %s not found\n", jobID)
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
		status := "✓"
		if !skill.Available {
			status = "✗"
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
		fmt.Printf("✗ Failed to install skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Skill '%s' installed successfully!\n", filepath.Base(repo))
}

func skillsRemoveCmd(installer *skills.SkillInstaller, skillName string) {
	fmt.Printf("Removing skill '%s'...\n", skillName)

	if err := installer.Uninstall(skillName); err != nil {
		fmt.Printf("✗ Failed to remove skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Skill '%s' removed successfully!\n", skillName)
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
			fmt.Printf("⊘ Builtin skill '%s' not found: %v\n", skillName, err)
			continue
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			fmt.Printf("✗ Failed to create directory for %s: %v\n", skillName, err)
			continue
		}

		if err := copyDirectory(builtinPath, workspacePath); err != nil {
			fmt.Printf("✗ Failed to copy %s: %v\n", skillName, err)
		}
	}

	fmt.Println("\n✓ All builtin skills installed!")
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
			status := "✓"
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
		fmt.Printf("✗ Failed to fetch skills list: %v\n", err)
		return
	}

	if len(availableSkills) == 0 {
		fmt.Println("No skills available.")
		return
	}

	fmt.Printf("\nAvailable Skills (%d):\n", len(availableSkills))
	fmt.Println("--------------------")
	for _, skill := range availableSkills {
		fmt.Printf("  📦 %s\n", skill.Name)
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
		fmt.Printf("✗ Skill '%s' not found\n", skillName)
		return
	}

	fmt.Printf("\n📦 Skill: %s\n", skillName)
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
	userSettingsPath := filepath.Join(filepath.Dir(configPath), "usersetting.json")

	switch subcommand {
	case "show":
		cfg, _ := loadConfig()
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Printf("Current Configuration (including usersetting.json):\n%s\n", string(data))
	case "set":
		if len(os.Args) < 5 {
			fmt.Println("Usage: marubot config set <key> <value>")
			return
		}
		key := os.Args[3]
		value := os.Args[4]

		var settings map[string]interface{}
		data, err := os.ReadFile(userSettingsPath)
		if err == nil {
			json.Unmarshal(data, &settings)
		} else {
			settings = make(map[string]interface{})
		}

		// Try to parse as JSON if it looks like one, otherwise keep as string
		var val interface{}
		if err := json.Unmarshal([]byte(value), &val); err != nil {
			val = value // stay as string
		}
		settings[key] = val

		newData, _ := json.MarshalIndent(settings, "", "  ")
		os.WriteFile(userSettingsPath, newData, 0644)
		fmt.Printf("✓ Saved '%s' = %s to %s\n", key, value, userSettingsPath)
	case "reset":
		os.Remove(userSettingsPath)
		fmt.Println("✓ User settings reset to defaults.")
	default:
		configHelp()
	}
}

func configHelp() {
	fmt.Println("\nConfig commands:")
	fmt.Println("  reset             Remove all user overrides")
	fmt.Println("  set <key> <val>   Set an override in usersetting.json")
	fmt.Println("  show              Show merged configuration")
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

	userSettingsPath := filepath.Join(filepath.Dir(configPath), "usersetting.json")
	if _, err := os.Stat(userSettingsPath); err == nil {
		fmt.Printf("User Settings: %s (OK)\n", userSettingsPath)
	}

	hasOpenRouter := cfg.Providers.OpenRouter.APIKey != ""
	hasAnthropic := cfg.Providers.Anthropic.APIKey != ""
	hasOpenAI := cfg.Providers.OpenAI.APIKey != ""
	hasGemini := cfg.Providers.Gemini.APIKey != ""
	hasZhipu := cfg.Providers.Zhipu.APIKey != ""
	hasGroq := cfg.Providers.Groq.APIKey != ""
	hasVLLM := cfg.Providers.VLLM.APIBase != ""

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
		fmt.Printf("vLLM/Local API: (OK)\n")
		fmt.Printf("  - Base: %s\n", cfg.Providers.VLLM.APIBase)
		fmt.Printf("  - Key:  %s\n", maskKey(cfg.Providers.VLLM.APIKey))
	} else {
		fmt.Printf("vLLM/Local: not set\n")
	}
}

func getResourceDir() string {
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
				// Ensure systemd knows about potential binary or service file changes
				daemonReload := exec.Command("systemctl", "--user", "daemon-reload")
				restart := exec.Command("systemctl", "--user", "restart", "marubot.service")
				
				if os.Getenv("XDG_RUNTIME_DIR") == "" && uid != "" {
					daemonReload.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%s", uid))
					restart.Env = append(os.Environ(), fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%s", uid))
				} else {
					daemonReload.Env = os.Environ()
					restart.Env = os.Environ()
				}
				
				daemonReload.Run()
				if err := restart.Run(); err == nil {
					fmt.Println("✓ Reloaded via systemd.")
					return
				}
			}
		}
	}

	stopCmd()
	time.Sleep(1 * time.Second)

	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("✗ Executable path error: %v\n", err)
		return
	}
	cmd := exec.Command(exe, "start")
	if err := cmd.Start(); err != nil {
		fmt.Printf("✗ Failed to start during reload: %v\n", err)
		return
	}
	fmt.Println("✓ Reload complete.")
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
		exe, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
			return
		}

		if runtime.GOOS == "linux" {
			err = installAndRunSystemdService(exe)
			if err == nil {
				fmt.Println("✨ MaruBot started as a systemd service.")
				fmt.Println("   It will auto-restart on reboot and continue working.")
				fmt.Println("   URL: http://localhost:8080")
				fmt.Println("   To stop: use 'marubot stop'")
				fmt.Println("   To reload config: use 'marubot reload'")
				return
			}
			fmt.Printf("Systemd service setup failed: %v. Falling back to simple daemon...\n", err)
		}

		// Re-run with special env var
		cmd := exec.Command(exe, "start")
		cmd.Env = append(os.Environ(), "MARUBOT_DAEMON=1")
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

		fmt.Printf("✨ MaruBot Dashboard started in background (PID: %d)\n", cmd.Process.Pid)
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
			// Note: This only redirects Go's fmt.Print output if we assign it,
			// but for true redirection we'd need syscalls which are OS specific.
			// Ideally just use a logger.
			// For simplicity in this cross-platform Go app without syscalls:
			// We will just let it run. Stdout/Stderr are discarded by the parent anyway.
			// But creating a log file is good practice.
		}
		defer f.Close()
	}

	if runForeground {
		fmt.Printf("%s Starting MaruBot Dashboard & API Server...\n", logo)
	}

	cfg, err := loadConfig()
	if err != nil {
		if runForeground {
			fmt.Printf("Error loading config: %v\n", err)
		}
		return
	}

	// Validate configuration: At least one AI provider and one channel must be enabled
	// If password is also missing, we prioritize security setup.
	if cfg.AdminPassword == "" || !cfg.IsAIConfigured() || !cfg.IsChannelEnabled() {
		showGuideMessage(cfg)
		
		// In GUI/Desktop environments, we might want to keep the process alive
		// so the user can finish setup via Web-Admin.
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" || os.Getenv("MARUBOT_GUI") == "1" {
			fmt.Println("Entering Setup Mode... (Server remains active for web configuration)")
			
			// Start dashboard server in a goroutine so user can configure
			port := cfg.Gateway.Port
			if port == 0 { port = 8080 }
			dashAddr := fmt.Sprintf("0.0.0.0:%d", port)
			
			// We need a dummy agent for the dashboard to start, but it won't be able to do much
			// until providers are configured.
			dummyAgent := &agent.AgentLoop{} 
			dashServer := dashboard.NewServer(dashAddr, dummyAgent, cfg, Version)
			
			go func() {
				if err := dashServer.Start(); err != nil {
					fmt.Printf("Dashboard failed to start: %v\n", err)
				}
			}()

			// Wait for interrupt to exit
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt)
			<-sig
			fmt.Println("Setup Mode exiting...")
			os.Exit(0)
		} else {
			// On Linux/RPi CLI, we might just show message and exit as instructions suggest
			os.Exit(0)
		}
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

	// Enable logging to file for background service
	logFile := filepath.Join(getResourceDir(), "dashboard.log")
	if err := logger.EnableFileLogging(logFile); err != nil {
		if runForeground {
			fmt.Printf("Warning: Failed to enable file logging: %v\n", err)
		}
	}

	agentLoop := agent.NewAgentLoop(cfg, bus, provider, Version)

	gpioService := gpio.NewGPIOService(cfg, bus)
	gpioService.Start(context.Background())
	defer gpioService.Stop()

	// Background Services
	cronStorePath := filepath.Join(filepath.Dir(getConfigPath()), "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, func(job *cron.CronJob) (string, error) {
		return agentLoop.ProcessDirect(context.Background(), job.Payload.Message, "cron:"+job.ID)
	})

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		func(msg string) (string, error) {
			return agentLoop.ProcessDirect(context.Background(), msg, "heartbeat")
		},
		30*60,
		true,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := cronService.Start(); err != nil && runForeground {
		fmt.Printf("Error starting cron service: %v\n", err)
	}
	if err := heartbeatService.Start(); err != nil && runForeground {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	go agentLoop.Run(ctx)

	channelManager, err := channels.NewManager(cfg, bus)
	if err == nil {
		if err := channelManager.StartAll(ctx); err != nil && runForeground {
			fmt.Printf("Error starting channels: %v\n", err)
		}
		if runForeground {
			fmt.Println("✓ Background services started (Cron, Heartbeat, Channels)")
		}
	} else {
		if runForeground {
			fmt.Printf("Warning: Failed to initialize channel manager: %v\n", err)
			fmt.Println("✓ Background services started (Cron, Heartbeat)")
		}
	}

	// Initialize Dashboard Server
	port := "8080"
	server := dashboard.NewServer(":"+port, agentLoop, cfg, Version)

	if runForeground {
		go func() {
			time.Sleep(1 * time.Second)
			fmt.Printf("✓ Dashboard available at http://localhost:%s\n", port)
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
					fmt.Println("✓ Stopped systemd service.")
					stoppedViaSystemd = true
				}
			}
		}
	}

	pidFile := getPidFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if !stoppedViaSystemd {
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

	fmt.Printf("Stopping marubot daemon (PID: %d)...\n", pid)
	if err := proc.Signal(os.Interrupt); err != nil {
		proc.Kill()
	}

	os.Remove(pidFile)
	fmt.Println("✓ Stopped.")
}

func upgradeCmd() {
	autoConfirm := false
	if len(os.Args) > 2 && os.Args[2] == "--yes" {
		autoConfirm = true
	}

	fmt.Println("⚙️  Checking for updates...")

	latest, err := config.CheckLatestVersion()
	if err != nil {
		fmt.Printf("⚠️  Failed to check latest Version: %v\n", err)
		fmt.Println("Proceeding with forced upgrade...")
	} else {
		if !config.IsNewVersionAvailable(latest) && !autoConfirm {
			fmt.Printf("✅ You are already using the latest Version (v%s).\n", config.Version)
			fmt.Print("Do you want to reinstall anyway? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				return
			}
		} else if config.IsNewVersionAvailable(latest) && !autoConfirm {
			fmt.Printf("✨ New Version available: v%s (Current: v%s)\n", latest, config.Version)
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

	fmt.Println("🚀 Upgrading MaruBot to the latest Version...")

	// Use curl to download and run the install script
	// We use the same install script as it handles updates gracefully (git pull if exists)
	cmd := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash")

	// Connect pipes to let user interact (for language selection, sudo password, etc.)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Upgrade failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✨ Upgrade complete! Restarting MaruBot...")
	reloadCmd()
}
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
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
			"title":    "⚠️ 초기 설정이 필요합니다!",
			"sec_body": "보안을 위해 관리자 비밀번호를 먼저 설정해야 합니다.",
			"cfg_body": "MaruBot을 시작하려면 최소 하나의 AI 모델과 채널이 설정되어야 합니다.",
			"gui":      "설정 페이지(Web-Admin)를 브라우저에 띄웁니다. 설정을 완료해 주세요.",
			"cli":      "브라우저에서 아래 주소로 접속하여 설정을 완료해 주세요:",
			"restart":  "설정 완료 후 앱을 재시작하면 서비스가 정상 기동됩니다.",
		},
		"en": {
			"title":    "⚠️ Initial Configuration Required!",
			"sec_body": "For security, you must set an administrator password first.",
			"cfg_body": "At least one AI model and one channel must be configured to start MaruBot.",
			"gui":      "Opening configuration page (Web-Admin) in your browser. Please complete setup.",
			"cli":      "Please access the following URL in your browser to complete setup:",
			"restart":  "After setup, restart the app to start the main service.",
		},
		"ja": {
			"title":    "⚠️ 初期設定が必要です！",
			"sec_body": "セキュリティのため、まず管理者のパスワードを設定する必要があります。",
			"cfg_body": "MaruBotを開始するには、少なくとも1つのAIモデルとチャネルを設定する必要があります。",
			"gui":      "ブラウザで設定ページ(Web-Admin)を開きます。設定を完了してください。",
			"cli":      "ブラウザで以下のURLにアクセスして設定を完了してください：",
			"restart":  "設定完了後、アプリを再起動するとサービスが開始されます。",
		},
		"zh": {
			"title":    "⚠️ 需要初始配置！",
			"sec_body": "出于安全考虑，您必须先设置管理员密码。",
			"cfg_body": "启动 MaruBot 至少需要配置一个 AI 模型和一个频道。",
			"gui":      "正在浏览器中打开配置页面 (Web-Admin)。请完成配置。",
			"cli":      "请在浏览器中访问以下地址完成配置：",
			"restart":  "配置完成后，请重启程序以启动主服务。",
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
		Arguments:   []string{"service", "run"},
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
		err = s.Install()
		if err == nil {
			fmt.Println("Service installed successfully. (Autostart enabled)")
		}
	case "uninstall":
		err = s.Uninstall()
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
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	}
	return os.Geteuid() == 0
}

func runAsAdmin() {
	if os.Getenv("MARUBOT_ELEVATED") == "1" {
		fmt.Println("Already tried to elevate and failed. Please run manually as Administrator.")
		return
	}

	exe, _ := os.Executable()
	args := strings.Join(os.Args[1:], " ")

	// PowerShell way which is more reliable for elevation
	cmd := exec.Command("powershell", "Start-Process", "-FilePath", exe, "-ArgumentList", args, "-Verb", "RunAs")
	cmd.Env = append(os.Environ(), "MARUBOT_ELEVATED=1")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to elevate: %v\n", err)
		fmt.Println("Please run this command in an Administrator terminal.")
	} else {
		fmt.Println("Elevated process started in a new window.")
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
		// Ask for a new port via PowerShell
		script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName Microsoft.VisualBasic
$title = "MaruBot Port Conflict"
$msg = "Port 8080 is already in use. Would you like to use a different port?"
$result = [System.Windows.Forms.MessageBox]::Show($msg, $title, "YesNo", "Warning")
if ($result -eq "Yes") {
    $newPort = [Microsoft.VisualBasic.Interaction]::InputBox("Enter a new port number:", $title, "8081")
    if ($newPort) { $newPort } else { exit 1 }
} else { exit 1 }
`)
		cmd := exec.Command("powershell", "-Command", script)
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		
		var newPort int
		fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &newPort)
		if newPort > 0 {
			cfg.Gateway.Port = newPort
			// Save to usersetting.json for persistence
			userSettingsPath := filepath.Join(filepath.Dir(getConfigPath()), "usersetting.json")
			var settings map[string]interface{}
			data, err := os.ReadFile(userSettingsPath)
			if err == nil {
				json.Unmarshal(data, &settings)
			} else {
				settings = make(map[string]interface{})
			}
			settings["gateway.port"] = newPort
			newData, _ := json.MarshalIndent(settings, "", "  ")
			os.WriteFile(userSettingsPath, newData, 0644)
			return true
		}
	} else {
		fmt.Printf("Warning: Port 8080 is in use. Please check your configuration.\n")
	}
	return false
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

func handleWindowsGUIMode() {
	// 1. Install binary to fixed location if not there
	// Before installing, check if we need admin rights
	if !isAdmin() && os.Getenv("MARUBOT_ELEVATED") != "1" {
		targetExe, _ := getTargetBinaryPath()
		// If binary doesn't exist or we don't have write access, try to elevate
		if _, err := os.Stat(targetExe); err != nil {
			fmt.Println("Elevation required for installation. Requesting administrator privileges...")
			runAsAdmin()
			os.Exit(0)
		}
	}

	targetExe, err := installBinary()
	if err != nil {
		fmt.Printf("Installation failed: %v\n", err)
		if os.Getenv("MARUBOT_ELEVATED") != "1" {
			runAsAdmin()
			os.Exit(0)
		} else {
			fmt.Println("Failed to install even with elevation. Please run marubot as Administrator manually.")
		}
	}

	// 2. Start Tray Icon (this blocks, so we run logic in onReady)
	systray.Run(func() { onTrayReady(targetExe) }, onTrayExit)
}

func getTargetBinaryPath() (string, error) {
	installDir := filepath.Join(getResourceDir(), "bin")
	targetPath := filepath.Join(installDir, "marubot.exe")
	if runtime.GOOS != "windows" {
		targetPath = filepath.Join(installDir, "marubot")
	}
	return targetPath, nil
}

func onTrayReady(targetExe string) {
	systray.SetTitle("MaruBot")
	systray.SetTooltip("MaruBot - AI Agent Service")
	// Use a simple dot as placeholder icon if needed, or just let it be blank for now
	// systray.SetIcon(iconBytes) 

	mDashboard := systray.AddMenuItem("Dashboard", "Open Web-Admin")
	systray.AddSeparator()
	mStart := systray.AddMenuItem("Start MaruBot", "Start the service")
	mStop := systray.AddMenuItem("Stop MaruBot", "Stop the service")
	systray.AddSeparator()
	mClose := systray.AddMenuItem("Close", "Uninstall and Exit")

	// 3. Initial check and install
	cfg, _ := loadConfig()
	if !checkAndFixPort(cfg) {
		systray.Quit()
		return
	}

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
				serviceCmdInternalPath("start", targetExe)
			case <-mStop.ClickedCh:
				serviceCmdInternalPath("stop", targetExe)
			case <-mClose.ClickedCh:
				serviceCmdInternalPath("stop", targetExe)
				serviceCmdInternalPath("uninstall", targetExe)
				systray.Quit()
			}
		}
	}()
}

func onTrayExit() {
	// Cleanup on exit
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

func getServiceBinaryPath() string {
	out, err := exec.Command("sc", "qc", "MaruBot").Output()
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`BINARY_PATH_NAME\s*:\s*("([^"]+)"|([^\s]+))`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return ""
	}
	if matches[2] != "" {
		return matches[2]
	}
	return matches[3]
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

func confirmUpgrade() bool {
	// Use PowerShell to show a Yes/No MessageBox
	script := `
Add-Type -AssemblyName System.Windows.Forms
$result = [System.Windows.Forms.MessageBox]::Show("A newer version of MaruBot is available. Would you like to upgrade the background service?", "MaruBot Upgrade", "YesNo", "Question")
if ($result -eq "Yes") { exit 0 } else { exit 1 }
`
	cmd := exec.Command("powershell", "-Command", script)
	err := cmd.Run()
	return err == nil
}
