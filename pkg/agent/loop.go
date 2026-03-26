// MaruBot - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 MaruBot contributors

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/logger" // Added logger
	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/session"
	"github.com/dirmich/marubot/pkg/tools"
)

type AgentLoop struct {
	bus            *bus.MessageBus
	provider       providers.LLMProvider
	workspace      string
	maxIterations  int
	sessions       *session.SessionManager
	contextBuilder *ContextBuilder
	tools          *tools.ToolRegistry
	version        string
	config         *config.Config
	channelManager bus.ChannelManager // Add this interface for channel tools
	running        bool
	mu             sync.RWMutex
}

func NewAgentLoop(cfg *config.Config, bus *bus.MessageBus, provider providers.LLMProvider, version string) *AgentLoop {
	workspace := cfg.WorkspacePath()
	os.MkdirAll(workspace, 0755)

	marubotHome := os.Getenv("MARUBOT_HOME")
	if marubotHome == "" {
		home, _ := os.UserHomeDir()
		marubotHome = filepath.Join(home, ".marubot")
	}

	toolsRegistry := tools.NewToolRegistry()
	toolsRegistry.Register(&tools.ReadFileTool{})
	toolsRegistry.Register(&tools.WriteFileTool{})
	toolsRegistry.Register(&tools.ListDirTool{})
	configPath := filepath.Join(marubotHome, "config.json")
	toolsRegistry.Register(tools.NewConfigTool(configPath, cfg))
	toolsRegistry.Register(tools.NewExecTool(workspace))
	toolsRegistry.Register(tools.NewSSHTool())

	cronStorePath := filepath.Join(marubotHome, "cron", "jobs.json")
	toolsRegistry.Register(tools.NewCronTool(cronStorePath))

	braveAPIKey := cfg.Tools.Web.Search.APIKey
	toolsRegistry.Register(tools.NewWebSearchTool(braveAPIKey, cfg.Tools.Web.Search.MaxResults))
	toolsRegistry.Register(tools.NewWebFetchTool(50000))

	// Hardware tools registration based on platform
	isLinux := runtime.GOOS == "linux"
	isARM := runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"

	if isLinux && isARM {
		toolsRegistry.Register(tools.NewCameraTool(workspace))
		toolsRegistry.Register(tools.NewMotorTool(cfg))
		toolsRegistry.Register(tools.NewUltrasonicTool(cfg))
		toolsRegistry.Register(tools.NewIMUTool())
		toolsRegistry.Register(tools.NewVisionTool(workspace))
		toolsRegistry.Register(tools.NewGPIOTool(cfg, cfg.Hardware.GPIO.Actions))
	} else if isLinux {
		// Generic Linux (EC2, etc) - Support Camera if USB webcam might be available
		toolsRegistry.Register(tools.NewCameraTool(workspace))
		toolsRegistry.Register(tools.NewVisionTool(workspace))
	}

	toolsRegistry.Register(tools.NewSystemTool(cfg, workspace))

	// Ensure extensions directory is under .marubot
	extensionDir := filepath.Join(marubotHome, "extensions")
	os.MkdirAll(extensionDir, 0755)

	toolsRegistry.Register(tools.NewCreateToolTool(toolsRegistry, extensionDir))
	toolsRegistry.Register(tools.NewCreateSkillTool(workspace))
	tools.LoadDynamicTools(toolsRegistry, extensionDir)
	if cfg.Drone.Enabled {
		toolsRegistry.Register(tools.NewDroneTool(cfg.Drone.Connection, cfg.Drone.SysID, cfg.Drone.CompID))
	}
	if cfg.GPS.Enabled {
		toolsRegistry.Register(tools.NewGPSTool(cfg.GPS.Device, cfg.GPS.Baud))
	}

	toolsRegistry.Register(tools.NewBrowserTool())

	// Ensure sessions directory is under .marubot
	sessionsDir := filepath.Join(marubotHome, "sessions")
	os.MkdirAll(sessionsDir, 0755)
	sessionsManager := session.NewSessionManager(sessionsDir)
	// Auto migrate old JSON sessions to SQLite
	sessionsManager.MigrateJSONToSQLite()

	// Check if version changed to prune stale system facts
	versionFile := filepath.Join(marubotHome, ".version_stamp")
	lastVersion, _ := os.ReadFile(versionFile)
	if string(lastVersion) != version {
		fmt.Printf("🚀 Version change detected (%s -> %s). Pruning stale system facts...\n", string(lastVersion), version)
		sessionsManager.PruneStaleFacts()
		os.WriteFile(versionFile, []byte(version), 0644)
	}

	al := &AgentLoop{
		bus:            bus,
		provider:       provider,
		workspace:      workspace,
		maxIterations:  20,
		sessions:       sessionsManager,
		contextBuilder: NewContextBuilder(workspace, version, cfg),
		tools:          toolsRegistry,
		version:        version,
		config:         cfg,
		running:        false,
	}
	
	// Set initial values from model config if possible
	if mCfg := al.findCurrentModelConfig(); mCfg != nil {
		if mCfg.MaxToolIterations > 0 {
			al.maxIterations = mCfg.MaxToolIterations
		}
	}
	return al
}

func (al *AgentLoop) SetProvider(p providers.LLMProvider) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.provider = p
}

func (al *AgentLoop) SetChannelManager(m bus.ChannelManager) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.channelManager = m
	
	// Register channel-related tools once we have the manager
	if al.tools != nil {
		al.tools.Register(tools.NewChannelTool(m))
	}
}

func (al *AgentLoop) Run(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			logger.ErrorCF("agent", "Agent loop panicked", map[string]interface{}{"error": r})
			// Restart loop if context is not done
			if ctx.Err() == nil {
				go al.Run(ctx)
			}
		}
	}()

	al.running = true
	logger.InfoC("agent", "Agent loop started successfully")

	for al.running {
		select {
		case <-ctx.Done():
			logger.InfoC("agent", "Agent loop stopping (context done)")
			return nil
		default:
			msg, ok := al.bus.ConsumeInbound(ctx)
			if !ok {
				continue
			}
			logger.InfoCF("agent", "Consuming inbound message", map[string]interface{}{
				"channel": msg.Channel,
				"sender":  msg.SenderID,
				"session": msg.SessionKey,
			})

			response, err := al.processMessage(ctx, msg)
			if err != nil {
				response = fmt.Sprintf("Error processing message: %v", err)
			}

			if response != "" {
				outMsg := bus.OutboundMessage{
					Channel: msg.Channel,
					ChatID:  msg.ChatID,
					Content: response,
				}
				// Copy relevant metadata for threading (e.g. thread_ts for Slack)
				if msg.Metadata != nil {
					outMsg.Metadata = make(map[string]string)
					for k, v := range msg.Metadata {
						outMsg.Metadata[k] = v
					}
				}
				logger.InfoCF("agent", "Publishing outbound message", map[string]interface{}{
					"channel": outMsg.Channel,
					"chatID":  outMsg.ChatID,
					"content_len": len(outMsg.Content),
				})
				al.bus.PublishOutbound(outMsg)
			}
		}
	}

	return nil
}

func (al *AgentLoop) Stop() {
	al.running = false
}

func (al *AgentLoop) ProcessDirect(ctx context.Context, content, sessionKey string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   "user",
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
	}

	return al.processMessage(ctx, msg)
}

func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	ctx = context.WithValue(ctx, tools.CtxKeyChannel, msg.Channel)
	ctx = context.WithValue(ctx, tools.CtxKeyChatID, msg.ChatID)

	// --- 🧠 STM & LTM Management (Enhanced RAG) ---
	// 🎯 1. Facts & Directives (Long-term persistent rules/preferences)
	facts, _ := al.sessions.GetActiveFacts("")
	factsContent := ""
	if len(facts) > 0 {
		factsContent = "\n\n### 🧘 Core Facts & Preferences:\n"
		for _, f := range facts {
			factsContent += fmt.Sprintf("- %s\n", f)
		}
	}

	// 🧵 2. STM (Short-term Memory): Get recent 20 messages
	history := al.sessions.GetHistory(msg.SessionKey)
	
	// 📚 3. LTM (Long-term Memory): Search past context for relevant info
	relevantContent := ""
	relevantMsgs := al.sessions.SearchRelevant(msg.Content, 5)
	if len(relevantMsgs) > 0 {
		seen := make(map[string]bool)
		uniqueMsgs := []providers.Message{}
		for _, rm := range relevantMsgs {
			// Basic deduplication based on content snippet to avoid repeating the same large info blocks
			contentKey := rm.Content
			if len(contentKey) > 100 {
				contentKey = contentKey[:100]
			}
			if !seen[contentKey] {
				seen[contentKey] = true
				uniqueMsgs = append(uniqueMsgs, rm)
			}
		}

		if len(uniqueMsgs) > 0 {
			relevantContent = "\n\n### 📚 Relevant Past Context (RAG):\n"
			for _, rm := range uniqueMsgs {
				relevantContent += fmt.Sprintf("- [%s]: %s\n", rm.Role, rm.Content)
			}
			relevantContent += "\nUse this information ONLY if it directly clarifies the user's intent."
		}
	}

	// Build messages with current history + injected Facts + LTM
	messages := al.contextBuilder.BuildMessages(
		history,
		msg.Content+factsContent+relevantContent,
		nil,
	)

	// Start a ticker to keep sending "typing" indicator while thinking
	typingCtx, typingCancel := context.WithCancel(ctx)
	defer typingCancel()
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-typingCtx.Done():
				return
			case <-ticker.C:
				al.bus.PublishOutbound(bus.OutboundMessage{
					Channel: msg.Channel,
					ChatID:  msg.ChatID,
					Action:  "typing",
				})
			}
		}
	}()

	// Initial typing indicator
	al.bus.PublishOutbound(bus.OutboundMessage{
		Channel: msg.Channel,
		ChatID:  msg.ChatID,
		Action:  "typing",
	})

	iteration := 0
	var finalContent string

	for iteration < al.maxIterations {
		iteration++

		toolDefs := al.tools.GetDefinitions()
		providerToolDefs := make([]providers.ToolDefinition, 0, len(toolDefs))
		for _, td := range toolDefs {
			providerToolDefs = append(providerToolDefs, providers.ToolDefinition{
				Type: td["type"].(string),
				Function: providers.ToolFunctionDefinition{
					Name:        td["function"].(map[string]interface{})["name"].(string),
					Description: td["function"].(map[string]interface{})["description"].(string),
					Parameters:  td["function"].(map[string]interface{})["parameters"].(map[string]interface{}),
				},
			})
		}

		// Find model config for parameters
		maxTokens := 8192
		temperature := 0.7
		
		// Search in the designated provider first
		mCfg := al.findCurrentModelConfig()
		if mCfg != nil {
			if mCfg.MaxTokens > 0 {
				maxTokens = mCfg.MaxTokens
			}
			if mCfg.Temperature > 0 {
				temperature = mCfg.Temperature
			}
		}

		al.mu.RLock()
		currentProvider := al.provider
		currentModel := al.config.Agents.Defaults.Model
		al.mu.RUnlock()

		if currentModel == "" {
			currentModel = currentProvider.GetDefaultModel()
		}

		response, err := currentProvider.Chat(ctx, messages, providerToolDefs, currentModel, map[string]interface{}{
			"max_tokens":  maxTokens,
			"temperature": temperature,
		})

		if err != nil {
			logger.ErrorC("agent", fmt.Sprintf("LLM call failed: %v", err))
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		logger.InfoC("agent", fmt.Sprintf("Iteration %d: LLM response content length: %d, tool calls: %d", iteration, len(response.Content), len(response.ToolCalls)))
		if len(response.Content) > 0 {
			logger.InfoC("agent", fmt.Sprintf("LLM Content: %s", response.Content))
		}

		if len(response.ToolCalls) == 0 {
			// Robust parsing: check if content contains a JSON tool call
			if tc := al.tryParseToolCallFromContent(response.Content); tc != nil {
				response.ToolCalls = []providers.ToolCall{*tc}
			} else {
				finalContent = response.Content
				break
			}
		}

		assistantMsg := providers.Message{
			Role:    "assistant",
			Content: response.Content,
		}

		for _, tc := range response.ToolCalls {
			argumentsJSON, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: &providers.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argumentsJSON),
				},
			})
		}
		messages = append(messages, assistantMsg)

		for _, tc := range response.ToolCalls {
			result, err := al.tools.Execute(ctx, tc.Name, tc.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			toolResultMsg := providers.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolResultMsg)
		}
	}

	if finalContent == "" {
		// Fallback: Try to find the last meaningful assistant message.
		// Exclude messages that look like raw tool call JSON (start with '{')
		for i := len(messages) - 1; i >= 0; i-- {
			c := strings.TrimSpace(messages[i].Content)
			if messages[i].Role == "assistant" && c != "" && !strings.HasPrefix(c, "{") {
				finalContent = c
				break
			}
		}

		if finalContent == "" {
			finalContent = "I've completed processing but have no response to give."
		}
	}

	al.sessions.AddMessage(msg.SessionKey, "user", msg.Content)
	al.sessions.AddMessage(msg.SessionKey, "assistant", finalContent)

	return finalContent, nil
}

func (al *AgentLoop) findCurrentModelConfig() *config.ModelConfig {
	al.mu.RLock()
	providerName := al.config.Agents.Defaults.Provider
	modelName := al.config.Agents.Defaults.Model
	al.mu.RUnlock()

	// Helper to search in a provider
	getInProvider := func(p config.ProviderConfig) *config.ModelConfig {
		for _, m := range p.Models {
			if strings.EqualFold(m.Model, modelName) {
				return &m
			}
		}
		return nil
	}

	if providerName != "" {
		var p config.ProviderConfig
		switch strings.ToLower(providerName) {
		case "anthropic": p = al.config.Providers.Anthropic
		case "openai": p = al.config.Providers.OpenAI
		case "openrouter": p = al.config.Providers.OpenRouter
		case "groq": p = al.config.Providers.Groq
		case "zhipu": p = al.config.Providers.Zhipu
		case "vllm": p = al.config.Providers.VLLM
		case "gemini": p = al.config.Providers.Gemini
		}
		if cfg := getInProvider(p); cfg != nil {
			return cfg
		}
	}

	// Global search if not found in specific provider
	allProviders := []config.ProviderConfig{
		al.config.Providers.VLLM,
		al.config.Providers.OpenAI,
		al.config.Providers.Anthropic,
		al.config.Providers.Gemini,
		al.config.Providers.Zhipu,
		al.config.Providers.Groq,
		al.config.Providers.OpenRouter,
	}
	for _, p := range allProviders {
		if cfg := getInProvider(p); cfg != nil {
			return cfg
		}
	}

	return nil
}

func (al *AgentLoop) tryParseToolCallFromContent(content string) *providers.ToolCall {
	content = strings.TrimSpace(content)
	
	// Fast path for pure JSON
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return al.parseJSONToolCall(content)
	}

	// Remove markdown code blocks if present
	if strings.Contains(content, "```") {
		re := regexp.MustCompile("(?s)```(?:json)?\n?(.*?)\n?```")
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			if tc := al.parseJSONToolCall(strings.TrimSpace(match[1])); tc != nil {
				return tc
			}
		}
	}

	// Try extracting the first valid JSON object from the text
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		if tc := al.parseJSONToolCall(content[start : end+1]); tc != nil {
			return tc
		}
	}

	return nil
}

func (al *AgentLoop) parseJSONToolCall(content string) *providers.ToolCall {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil
	}

	// Case 1: Nanobot/MaruBot direct command style {"command": "..."} -> shell
	if cmd, ok := data["command"].(string); ok && cmd != "" {
		return &providers.ToolCall{
			ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
			Name:      "shell",
			Arguments: data,
		}
	}

	// Case 1.1: Direct config action style {"action": "get", "key": "..."} -> config
	if action, ok := data["action"].(string); ok && (action == "get" || action == "set") {
		return &providers.ToolCall{
			ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
			Name:      "config",
			Arguments: data,
		}
	}

	// Case 2: OpenAI-like structure embedded in content {"name": "...", "arguments": {...}}
	if name, ok := data["name"].(string); ok && name != "" {
		argsMap := make(map[string]interface{})
		if args, exists := data["arguments"]; exists {
			switch v := args.(type) {
			case map[string]interface{}:
				argsMap = v
			case string:
				json.Unmarshal([]byte(v), &argsMap)
			}
		}
		return &providers.ToolCall{
			ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
			Name:      name,
			Arguments: argsMap,
		}
	}

	return nil
}
