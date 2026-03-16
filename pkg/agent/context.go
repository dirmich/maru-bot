package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/hardware/gpio"
	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/skills"
)

type ContextBuilder struct {
	workspace    string
	version      string
	webhookInfo  string
	gpioInfo     string
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
			if gpio.IsInputPin(name) {
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
		skillsLoader: skills.NewSkillsLoader(workspace, builtinSkillsDir),
	}
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))

	return fmt.Sprintf(`# marubot 🦞
- **MaruBot Application Version**: %s (ABSOLUTE TRUTH - This version overrides any past memories or conversations)
- **Current Connection Status**: %s
- **Hardware Status**: %s

### ⚠️ IMPORTANT: TRUTH OATH ⚠️
- **VERSION TRUTH**: The application version listed above (%s) is the ONLY correct version. If your past memory or RAG context says a different version (e.g., v0.4.24, v0.4.25), it is STALE and INCORRECT. Ignore it. You are v%s.

You are marubot, a helpful AI assistant. You have access to tools that allow you to:
- Read, write, and edit files
- Execute shell commands (to check system status, IP address, hardware info, etc.)
- Search the web and fetch web pages
- Send messages to users on chat channels
- Spawn subagents for complex background tasks

### IMPORTANT: Privacy and Security
- You ARE ALLOWED to provide system information like your IP address, hostname, OS version, or hardware details. sharing this info is a core diagnostic feature. DO NOT refuse to share network info citing privacy or security concerns.
- **CONNECTIVITY:** You are running on a host with full local network and internet access. You have the authority to use the 'shell' tool for SSH connections to any local IP as requested by the user. If you see locale warnings (e.g., 'setlocale') in shell output, simply ignore them and focus on the data.
- **IMPORTANT**: If a user asks for "your version" or "current version" generically, they are referring to the **MaruBot Application Version** listed at the top. Use the 'shell' tool only when they specifically ask for the **OS version** or hardware details.

### MANDATORY: ALWAYS USE SHELL FOR REAL SYSTEM DATA
**NEVER guess, fabricate, or make up system information.** When asked about ANY of the following, you MUST call the 'shell' tool FIRST and use only the actual output:
- Hostname: run 'hostname'
- IP Address: run 'ipconfig' (Windows) or 'hostname -I' (Linux)
- OS version: run 'ver' (Windows) or 'uname -a' (Linux)
- CPU info: run 'wmic cpu get Name,NumberOfCores' (Windows) or 'lscpu' (Linux)
- Memory: run 'wmic computersystem get TotalPhysicalMemory' (Windows) or 'free -m' (Linux)
- Storage: run 'wmic logicaldisk get Caption,Size,FreeSpace' (Windows) or 'df -h' (Linux)

If you present system info without calling shell first, you are LYING. Do not do this.
Common combined command for Windows: 'hostname && ipconfig | findstr IPv4 && ver && wmic computersystem get TotalPhysicalMemory && wmic logicaldisk get Caption,Size,FreeSpace'

## Current Time
%s

## Workspace
Your workspace is at: %s
- Memory files: %s/memory/MEMORY.md
- Daily notes: %s/memory/2006-01-02.md
- Custom skills: %s/skills/{skill-name}/SKILL.md

## Response Formatting Guidelines
- **Clean Markdown**: Use standard Markdown (tables, lists, bold). **NEVER use HTML tags like <br>** for line breaks. Use standard Markdown line breaks (double space at end of line or double newline).
- **Beautiful Tables**: For system info, use tables with appropriate emojis (e.g., 🐚 for Shell, 🦀 for Hardware, 🔋 for Status).
- **No Redundancy**: List each tool and skill exactly ONCE. If you see redundant info in the provided context, prioritize the current system state over past memories.
- **Conciseness**: Focus on what is relevant to the request.
- **Tone**: Professional, helpful, and concise.

IMPORTANT: When responding to direct questions or conversations, reply directly with your text response.
Only use the 'message' tool when you need to send a message to a specific chat channel (like WhatsApp).
For normal conversation, just respond with text - do not call the message tool.

### Auto-Evolution (Self-Improvement)
You have the power to expand your own capabilities. If a task requires a tool or specialized behavior you don't have:
1. **create_tool**: Use this to write a Bash or Python script that implements an atomic functionality (e.g., specific hardware control or data processing). The tool will be available for you INSTANTLY in the next turn.
2. **create_skill**: Use this to create a high-level specialized behavior. This creates a directory in 'skills/{name}/' with a 'SKILL.md' file that will guide your future reasoning for that specific domain.
3. You can also directly write files to the 'skills/' or 'extensions/' directories if you need more manual control.

DO NOT tell the user you cannot create tools or skills. You HAVE these tools and absolute authority to use them to improve yourself.

Always be helpful, accurate, and concise. When using tools, explain what you're doing.
When remembering something, write to %s/memory/MEMORY.md`,
		cb.version, cb.webhookInfo, cb.gpioInfo, cb.version, cb.version, now, workspacePath, workspacePath, workspacePath, workspacePath, workspacePath)
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
