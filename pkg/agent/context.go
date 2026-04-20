package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/skills"
)

type ContextBuilder struct {
	workspace    string
	version      string
	webhookInfo  string
	gpioInfo     string
	language     string
	platform     string
	skillsLoader *skills.SkillsLoader
}

func NewContextBuilder(workspace, version string, cfg *config.Config) *ContextBuilder {
	builtinSkillsDir := filepath.Join(filepath.Dir(workspace), "skills")

	webhookInfo := "Webhook: Disabled"
	if cfg.Channels.Webhook.Enabled {
		webhookInfo = fmt.Sprintf("Webhook: Enabled (Port: %d, Path: %s)", cfg.Channels.Webhook.Port, cfg.Channels.Webhook.Path)
	}

	gpioInfo := "GPIO: Disabled"
	if cfg.Hardware.GPIO.Enabled {
		var pinDetails []string
		for name, val := range cfg.Hardware.GPIO.Pins {
			direction := "Output"
			if config.IsInputPin(name) {
				direction = "Input (Monitoring Enabled)"
			}
			pinDetails = append(pinDetails, fmt.Sprintf("- %s: Pin %v (%s)", name, val, direction))
		}
		gpioInfo = fmt.Sprintf("GPIO: Enabled\n%s", strings.Join(pinDetails, "\n"))
	}

	return &ContextBuilder{
		workspace:    workspace,
		version:      version,
		webhookInfo:  webhookInfo,
		gpioInfo:     gpioInfo,
		language:     cfg.Language,
		platform:     runtime.GOOS,
		skillsLoader: skills.NewSkillsLoader(workspace, builtinSkillsDir),
	}
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))

	langContext := "Respond in the same language as the user."
	if cb.language != "" {
		langContext = fmt.Sprintf("Respond STRICTLY in %s. All explanations and messages must be in %s.", cb.language, cb.language)
	}

	platformType := "Linux/Unix"
	if cb.platform == "windows" {
		platformType = "Windows"
	}

	mandatoryCommands := `- Hostname: 'hostname'
- IP Address: 'hostname -I'
- OS version: 'cat /etc/os-release'
- CPU info: 'lscpu'
- Memory: 'free -m'
- Storage: 'df -h'`

	if cb.platform == "windows" {
		mandatoryCommands = `- Hostname: 'hostname'
- IP Address: 'ipconfig'
- OS version: 'ver'
- CPU info: 'wmic cpu get Name'
- Memory: 'wmic computersystem get TotalPhysicalMemory'
- Storage: 'wmic logicaldisk get Caption,Size,FreeSpace'`
	}

	usefulCommands := `* IP Address: 'hostname -I' (LAN), 'curl -s ifconfig.me' (WAN)
* OS/Kernel: 'cat /etc/os-release', 'uname -a'
* CPU: 'lscpu'
* Memory: 'free -m'
* Disk: 'df -h /'`

	if cb.platform == "windows" {
		usefulCommands = `* IP Address: 'ipconfig' (LAN), 'curl -s ifconfig.me' (WAN)
* OS/Kernel: 'ver'
* Memory: 'wmic computersystem get TotalPhysicalMemory'
* Disk: 'wmic logicaldisk get Caption,Size,FreeSpace'`
	}

	return fmt.Sprintf(`# marubot 🦞
- **MaruBot Application Version**: %s (ABSOLUTE TRUTH)
- **Host Platform**: %s
- **Current Connection Status**: %s
- **Hardware Status**: %s
- **Agent Output Language**: %s

### ⚠️ IMPORTANT: TRUTH OATH ⚠️
- **VERSION TRUTH**: The application version listed above (%s) is the ONLY correct version. You are v%s.

You are marubot, a helpful AI assistant. You have access to tools that allow you to:
- Read, write, and edit files
- Execute shell commands (to check system status, IP address, hardware info, etc.)
- Search the web and fetch web pages
- Send messages to users on chat channels
- Spawn subagents for complex background tasks

### 🛡️ TOOL USE PROTOCOL (CRITICAL)
- **NO GUESSING**: Never hallucinate system values. If you don't know an IP, CPU model, or memory status, you MUST use the 'shell' tool.
- **NO PLACEHOLDERS**: Never use file paths like '~/.ssh/id_ed25519' or 'none' as placeholders for actual system data in your final response.
- **NATIVE EXECUTION**: You are running NATIVELY on the host (%s). Use 'shell' for local info first.

### IMPORTANT: Privacy and Security
- You ARE ALLOWED to provide system information like your IP address, hostname, OS version, or hardware details. sharing this info is a core diagnostic feature.
- **CONNECTIVITY:** You are running on a %s host with full local network and internet access.

### MANDATORY: ALWAYS USE SHELL FOR REAL SYSTEM DATA
**NEVER guess, fabricate, or make up system information.** When asked about ANY of the following, you MUST call the 'shell' tool FIRST and use only the actual output:
%s

If you present system info without calling shell first, you are LYING. Do not do this.

## Current Time
%s

### 🐚 Useful Shell Commands for Current Platform (%s)
%s

## Workspace
Your workspace is at: %s
- Memory files: %s/memory/MEMORY.md
- Daily notes: %s/memory/2006-01-02.md
- Custom skills: %s/skills/{skill-name}/SKILL.md

## Response Formatting Guidelines
- **Clean Markdown**: ALWAYS use standard Markdown (tables, lists, bold). NEVER output raw HTML.
- **Beautiful Tables**: For system info, use tables with appropriate emojis (e.g., 🐚 for Shell, 🦀 for Hardware, 🔋 for Status).

IMPORTANT: When responding to direct questions or conversations, reply directly with your text response.

### Auto-Evolution (Self-Improvement)
You have the power to expand your own capabilities via 'create_tool' and 'create_skill'.

Always be helpful, accurate, and concise. When using tools, explain what you're doing.
When remembering something, write to %s/memory/MEMORY.md`,
		cb.version, platformType, cb.webhookInfo, cb.gpioInfo, langContext, cb.version, cb.version, platformType, platformType, 
		mandatoryCommands, now, platformType, usefulCommands, workspacePath, workspacePath, workspacePath, workspacePath, workspacePath)
}

func (cb *ContextBuilder) LoadBootstrapFiles() string {
	bootstrapFiles := []string{
		"AGENTS.md",
		"SOUL.md",
		"USER.md",
		"TOOLS.md",
		"IDENTITY.md",
		"MEMORY.md",
	}

	var result string
	for _, filename := range bootstrapFiles {
		filePath := filepath.Join(cb.workspace, filename)
		if data, err := os.ReadFile(filePath); err == nil {
			result += fmt.Sprintf("## %s\n\n%s\n\n", filename, string(data))
		}
	}

	return result
}

func (cb *ContextBuilder) BuildMessages(history []providers.Message, currentMessage string, media []string) []providers.Message {
	messages := []providers.Message{}

	systemPrompt := cb.BuildSystemPrompt()
	bootstrapContent := cb.LoadBootstrapFiles()
	if bootstrapContent != "" {
		systemPrompt += "\n\n" + bootstrapContent
	}

	skillsSummary := cb.skillsLoader.BuildSkillsSummary()
	if skillsSummary != "" {
		systemPrompt += "\n\n## Available Skills\n\n" + skillsSummary
	}

	skillsContent := cb.loadSkills()
	if skillsContent != "" {
		// Only add skill content if it's not already summarized or to provide full details for 'always' skills
		systemPrompt += "\n\n## Skill Details (Always Loaded)\n\n" + skillsContent
	}

	messages = append(messages, providers.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	messages = append(messages, history...)

	messages = append(messages, providers.Message{
		Role:    "user",
		Content: currentMessage,
	})

	return messages
}

func (cb *ContextBuilder) AddToolResult(messages []providers.Message, toolCallID, toolName, result string) []providers.Message {
	messages = append(messages, providers.Message{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
	})
	return messages
}

func (cb *ContextBuilder) AddAssistantMessage(messages []providers.Message, content string, toolCalls []map[string]interface{}) []providers.Message {
	msg := providers.Message{
		Role:    "assistant",
		Content: content,
	}
	if len(toolCalls) > 0 {
		messages = append(messages, msg)
	}
	return messages
}

func (cb *ContextBuilder) loadSkills() string {
	alwaysSkills := cb.skillsLoader.GetAlwaysSkills()
	if len(alwaysSkills) == 0 {
		return ""
	}

	content := cb.skillsLoader.LoadSkillsForContext(alwaysSkills)
	if content == "" {
		return ""
	}

	return "# Skill Definitions\n\n" + content
}
