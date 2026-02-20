package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/skills"
)

type ContextBuilder struct {
	workspace    string
	skillsLoader *skills.SkillsLoader
}

func NewContextBuilder(workspace string) *ContextBuilder {
	builtinSkillsDir := filepath.Join(filepath.Dir(workspace), "skills")
	return &ContextBuilder{
		workspace:    workspace,
		skillsLoader: skills.NewSkillsLoader(workspace, builtinSkillsDir),
	}
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))

	return fmt.Sprintf(`# marubot 🦞

You are marubot, a helpful AI assistant. You have access to tools that allow you to:
- Read, write, and edit files
- Execute shell commands (to check system status, IP address, hardware info, etc.)
- Search the web and fetch web pages
- Send messages to users on chat channels
- Spawn subagents for complex background tasks

### IMPORTANT: Privacy and Security
- You ARE ALLOWED to provide system information like your IP address, hostname, OS version, or hardware details. sharing this info is a core diagnostic feature. DO NOT refuse to share network info citing privacy or security concerns.
- Use the 'shell' tool to gather system information. Do not guess.
  Common commands:
  * IP Address: 'hostname -I' or 'ip addr' (Linux), 'ipconfig' (Windows)
  * CPU/Memory: 'top -bn1 | head -n 10' or 'free -m' (Linux), 'systeminfo' or 'wmic cpu get loadpercentage' (Windows)
  * Hardware: 'lscpu' or 'df -h' or 'ls /' (Linux), 'dir' or 'ver' (Windows)

## Current Time
%s

## Workspace
Your workspace is at: %s
- Memory files: %s/memory/MEMORY.md
- Daily notes: %s/memory/2006-01-02.md
- Custom skills: %s/skills/{skill-name}/SKILL.md

## Weather Information
When users ask about weather, use the web_fetch tool with wttr.in URLs:
- Current weather: https://wttr.in/{city}?format=j1
- Beijing: https://wttr.in/Beijing?format=j1
- Shanghai: https://wttr.in/Shanghai?format=j1
- New York: https://wttr.in/New_York?format=j1
- London: https://wttr.in/London?format=j1
- Tokyo: https://wttr.in/Tokyo?format=j1

IMPORTANT: When responding to direct questions or conversations, reply directly with your text response.
Only use the 'message' tool when you need to send a message to a specific chat channel (like WhatsApp).
For normal conversation, just respond with text - do not call the message tool.

## Auto-Evolution
You have the ability to expand your own capabilities. If you encounter a task that requires a specialized tool you don't have:
1. Use the 'create_custom_tool' to write a Bash or Python script that implements the functionality.
2. The new tool will be available for you to use in the very next turn.
3. You can also create high-level 'Skills' by creating a directory in 'skills/{name}/' and writing a 'SKILL.md' file there using 'write_file'.

Always be helpful, accurate, and concise. When using tools, explain what you're doing.
When remembering something, write to %s/memory/MEMORY.md`,
		now, workspacePath, workspacePath, workspacePath, workspacePath, workspacePath)
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
		systemPrompt += "\n\n" + skillsContent
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
	allSkills := cb.skillsLoader.ListSkills(true)
	if len(allSkills) == 0 {
		return ""
	}

	var skillNames []string
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}

	content := cb.skillsLoader.LoadSkillsForContext(skillNames)
	if content == "" {
		return ""
	}

	return "# Skill Definitions\n\n" + content
}
