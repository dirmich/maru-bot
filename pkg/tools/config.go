package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"marubot/pkg/config"
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
				"type":        "any",
				"description": "Value to set (only for 'set' action)",
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
	// In a real production app, we'd use reflection or a map to update the live config object.
	// For this demo, we will write to a user-specific settings file that overrides defaults.

	userSettingsPath := filepath.Join(filepath.Dir(t.configPath), "usersetting.json")
	var settings map[string]interface{}

	data, err := os.ReadFile(userSettingsPath)
	if err == nil {
		json.Unmarshal(data, &settings)
	} else {
		settings = make(map[string]interface{})
	}

	// Update the key in the map
	settings[key] = value

	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(userSettingsPath, newData, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully saved setting '%s' to %s. Changes will be applied on next restart.", key, userSettingsPath), nil
}
