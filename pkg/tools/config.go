package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dirmich/marubot/pkg/config"
)

type ConfigTool struct {
	configPath string
	cfg        *config.Config
}

func NewConfigTool(configPath string, cfg *config.Config) *ConfigTool {
	return &ConfigTool{
		configPath: configPath,
		cfg:        cfg,
	}
}

func (t *ConfigTool) Name() string {
	return "config"
}

func (t *ConfigTool) Description() string {
	return "Check or modify marubot configuration like GPIO pins, sensor settings, etc."
}

func (t *ConfigTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: 'get' or 'set'",
				"enum":        []string{"get", "set"},
			},
			"key": map[string]interface{}{
				"type":        "string",
				"description": "Configuration key (e.g., 'hardware.gpio.pins.led_status', 'drone.connection')",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Value to set (only for 'set' action). Note: Internal values will be parsed from string.",
			},
		},
		"required": []string{"action"},
	}
}

func (t *ConfigTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	key, _ := args["key"].(string)

	if action == "get" {
		if key == "" {
			// Return full config (sensitive info removed or masked?)
			data, _ := json.MarshalIndent(t.cfg, "", "  ")
			return string(data), nil
		}
		// In a real implementation, we would use reflection to get exact key.
		// For now, let's just return the full config if key is provided but not found,
		// or implement a simple switch for common keys.
		return t.handleGet(key)
	}

	if action == "set" {
		value, ok := args["value"]
		if !ok {
			return "", fmt.Errorf("value is required for 'set' action")
		}
		return t.handleSet(key, value)
	}

	return "", fmt.Errorf("invalid action: %s", action)
}

func (t *ConfigTool) handleGet(key string) (string, error) {
	// Simple implementation: return the whole thing for now,
	// or we can implement a map-based access if we convert config to map.
	data, _ := json.MarshalIndent(t.cfg, "", "  ")
	return fmt.Sprintf("Configuration (Key: %s):\n%s", key, string(data)), nil
}

func (t *ConfigTool) handleSet(key string, value interface{}) (string, error) {
	// Update the live config object
	// For simple implementation, we update the Config object based on common keys
	// This ensures config.json remains the single source of truth

	// Create a temporary config to apply updates
	updateCfg := config.DefaultConfig()
	
	// Map simple keys (expandable as needed)
	valStr := fmt.Sprintf("%v", value)
	if key == "admin_password" {
		updateCfg.AdminPassword = valStr
	} else if key == "language" {
		updateCfg.Language = valStr
	} else {
		return "", fmt.Errorf("setting key '%s' via tool is currently restricted. Please use Web Admin", key)
	}

	// Apply and Save
	t.cfg.Update(updateCfg)
	if err := config.SaveConfig(t.configPath, t.cfg); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully saved setting '%s' to %s. Changes applied.", key, t.configPath), nil
}
