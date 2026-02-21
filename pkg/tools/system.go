package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/logger"
)

type SystemTool struct {
	config       *config.Config
	workspace    string
	marubotHome  string
}

func NewSystemTool(cfg *config.Config, workspace string) *SystemTool {
	home, _ := os.UserHomeDir()
	marubotHome := filepath.Join(home, ".marubot")
	return &SystemTool{
		config:      cfg,
		workspace:   workspace,
		marubotHome: marubotHome,
	}
}

func (t *SystemTool) Name() string {
	return "system_control"
}

func (t *SystemTool) Description() string {
	return "Control MaruBot system: install/list skills and reload/restart the agent. Use this after creating custom tools or changing configurations to apply changes."
}

func (t *SystemTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The action to perform: 'reload', 'skill_install', 'skill_list', 'status'",
				"enum":        []string{"reload", "skill_install", "skill_list", "status"},
			},
			"skill_repo": map[string]interface{}{
				"type":        "string",
				"description": "GitHub repository for the skill (e.g., 'sipeed/marubot-skills/weather'). Required for 'skill_install'.",
			},
		},
		"required": []string{"action"},
	}
}

func (t *SystemTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)

	switch action {
	case "reload":
		return t.handleReload()
	case "skill_install":
		repo, _ := args["skill_repo"].(string)
		if repo == "" {
			return "Error: skill_repo is required for skill_install", nil
		}
		return t.runMarubotCmd("skills", "install", repo)
	case "skill_list":
		return t.runMarubotCmd("skills", "list")
	case "status":
		return t.runMarubotCmd("status")
	default:
		return fmt.Sprintf("Unknown action: %s", action), nil
	}
}

func (t *SystemTool) handleReload() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// AI should know it will be disconnected
	logger.InfoC("system", "AI requested system reload")

	// We use an external command to perform the reload to avoid killing ourselves instantly
	// if we're not running under systemd.
	// But the most reliable way in MaruBot is calling 'marubot reload' via exec.
	
	go func() {
		// Wait a bit to let the AI finish sending the response
		select {
		case <-context.Background().Done():
		case <-(func() chan struct{} {
			c := make(chan struct{})
			go func() {
				// Artificial delay
				for i := 0; i < 2; i++ {
					// wait loop
				}
				close(c)
			}()
			return c
		}()):
		}
		
		// This is a bit hacky but it's how MaruBot manages its lifecycle
		cmd := exec.Command(exe, "reload")
		cmd.Start() 
	}()

	return "MaruBot is reloading... I will be offline for a few seconds. Please wait before sending the next message.", nil
}

func (t *SystemTool) runMarubotCmd(args ...string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
