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
	"time"

	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/session"
	"github.com/dirmich/marubot/pkg/tools"
)

type AgentLoop struct {
	bus            *bus.MessageBus
	provider       providers.LLMProvider
	workspace      string
	model          string
	maxIterations  int
	sessions       *session.SessionManager
	contextBuilder *ContextBuilder
	tools          *tools.ToolRegistry
	running        bool
}

func NewAgentLoop(cfg *config.Config, bus *bus.MessageBus, provider providers.LLMProvider) *AgentLoop {
	workspace := cfg.WorkspacePath()
	os.MkdirAll(workspace, 0755)

	home, _ := os.UserHomeDir()
	marubotHome := filepath.Join(home, ".marubot")

	toolsRegistry := tools.NewToolRegistry()
	toolsRegistry.Register(&tools.ReadFileTool{})
	toolsRegistry.Register(&tools.WriteFileTool{})
	toolsRegistry.Register(&tools.ListDirTool{})
	configPath := filepath.Join(marubotHome, "config", "config.json")
	toolsRegistry.Register(tools.NewConfigTool(configPath, cfg))
	toolsRegistry.Register(tools.NewExecTool(workspace))

	braveAPIKey := cfg.Tools.Web.Search.APIKey
	toolsRegistry.Register(tools.NewWebSearchTool(braveAPIKey, cfg.Tools.Web.Search.MaxResults))
	toolsRegistry.Register(tools.NewWebFetchTool(50000))
	toolsRegistry.Register(tools.NewCameraTool(workspace))
	toolsRegistry.Register(tools.NewMotorTool(cfg.Hardware.GPIO.Pins))
	toolsRegistry.Register(tools.NewUltrasonicTool(cfg.Hardware.GPIO.Pins))
	toolsRegistry.Register(tools.NewIMUTool())
	toolsRegistry.Register(tools.NewVisionTool(workspace))
	toolsRegistry.Register(tools.NewGPIOTool(cfg.Hardware.GPIO.Pins, cfg.Hardware.GPIO.Actions))

	// Ensure extensions directory is under .marubot
	extensionDir := filepath.Join(marubotHome, "extensions")
	os.MkdirAll(extensionDir, 0755)

	toolsRegistry.Register(tools.NewCreateToolTool(toolsRegistry, extensionDir))
	tools.LoadDynamicTools(toolsRegistry, extensionDir)
	if cfg.Drone.Enabled {
		toolsRegistry.Register(tools.NewDroneTool(cfg.Drone.Connection, cfg.Drone.SysID, cfg.Drone.CompID))
	}
	if cfg.GPS.Enabled {
		toolsRegistry.Register(tools.NewGPSTool(cfg.GPS.Device, cfg.GPS.Baud))
	}

	// Ensure sessions directory is under .marubot
	sessionsDir := filepath.Join(marubotHome, "sessions")
	os.MkdirAll(sessionsDir, 0755)
	sessionsManager := session.NewSessionManager(sessionsDir)

	return &AgentLoop{
		bus:            bus,
		provider:       provider,
		workspace:      workspace,
		model:          cfg.Agents.Defaults.Model,
		maxIterations:  cfg.Agents.Defaults.MaxToolIterations,
		sessions:       sessionsManager,
		contextBuilder: NewContextBuilder(workspace),
		tools:          toolsRegistry,
		running:        false,
	}
}

func (al *AgentLoop) Run(ctx context.Context) error {
	al.running = true

	for al.running {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, ok := al.bus.ConsumeInbound(ctx)
			if !ok {
				continue
			}

			response, err := al.processMessage(ctx, msg)
			if err != nil {
				response = fmt.Sprintf("Error processing message: %v", err)
			}

			if response != "" {
				al.bus.PublishOutbound(bus.OutboundMessage{
					Channel: msg.Channel,
					ChatID:  msg.ChatID,
					Content: response,
				})
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
	messages := al.contextBuilder.BuildMessages(
		al.sessions.GetHistory(msg.SessionKey),
		msg.Content,
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

		response, err := al.provider.Chat(ctx, messages, providerToolDefs, al.model, map[string]interface{}{
			"max_tokens":  8192,
			"temperature": 0.7,
		})

		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		if len(response.ToolCalls) == 0 {
			finalContent = response.Content
			break
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
		finalContent = "I've completed processing but have no response to give."
	}

	al.sessions.AddMessage(msg.SessionKey, "user", msg.Content)
	al.sessions.AddMessage(msg.SessionKey, "assistant", finalContent)
	al.sessions.Save(al.sessions.GetOrCreate(msg.SessionKey))

	return finalContent, nil
}
