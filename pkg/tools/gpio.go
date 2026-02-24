package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"

	"github.com/dirmich/marubot/pkg/config"
)

type GPIOTool struct {
	cfg     *config.Config
	actions map[string]map[string]int
}

func NewGPIOTool(cfg *config.Config, actions map[string]map[string]int) *GPIOTool {
	host.Init()
	return &GPIOTool{
		cfg:     cfg,
		actions: actions,
	}
}

func (t *GPIOTool) Name() string {
	return "gpio_control"
}

func (t *GPIOTool) Description() string {
	return "Control GPIO pins for reading/writing values and executing grouped actions (e.g., controlling motors/wings)"
}

func (t *GPIOTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Operation to perform: 'read', 'write', 'execute_action'",
				"enum":        []string{"read", "write", "execute_action"},
			},
			"pin": map[string]interface{}{
				"type":        "string",
				"description": "Pin name (from config) or number (e.g., '18', 'LED')",
			},
			"value": map[string]interface{}{
				"type":        "integer",
				"description": "Value to write (0 or 1)",
				"enum":        []int{0, 1},
			},
			"action_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the grouped action to execute (e.g., 'move_left', 'flap_wings')",
			},
		},
		"required": []string{"action"},
	}
}

func (t *GPIOTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)

	switch action {
	case "read":
		pinName, ok := args["pin"].(string)
		if !ok {
			return "", fmt.Errorf("pin is required for read")
		}
		return t.readPin(pinName)
	case "write":
		pinName, ok := args["pin"].(string)
		if !ok {
			return "", fmt.Errorf("pin is required for write")
		}
		val, ok := args["value"].(float64) // JSON numbers are float64
		if !ok {
			return "", fmt.Errorf("value is required for write")
		}
		return t.writePin(pinName, int(val))
	case "execute_action":
		actionName, ok := args["action_name"].(string)
		if !ok {
			return "", fmt.Errorf("action_name is required for execute_action")
		}
		return t.executeAction(actionName)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *GPIOTool) resolvePin(pinIdentifier string) gpio.PinIO {
	// 1. Check if it's a configured name in t.cfg.Hardware.GPIO.Pins
	if val, ok := t.cfg.Hardware.GPIO.Pins[pinIdentifier]; ok {
		// It might be a number, string, or map[string]interface{}
		switch v := val.(type) {
		case int:
			return gpioreg.ByName(fmt.Sprintf("%d", v))
		case float64:
			return gpioreg.ByName(fmt.Sprintf("%d", int(v)))
		case string:
			return gpioreg.ByName(v)
		case map[string]interface{}:
			if pin, ok := v["pin"]; ok {
				return gpioreg.ByName(fmt.Sprintf("%v", pin))
			}
		}
	}
	// 2. Try direct lookup (e.g. "GPIO18", "18")
	return gpioreg.ByName(pinIdentifier)
}

func (t *GPIOTool) readPin(pinIdentifier string) (string, error) {
	p := t.resolvePin(pinIdentifier)
	if p == nil {
		return "", fmt.Errorf("pin '%s' not found", pinIdentifier)
	}

	// Default to no pull unless specified or previously set
	pull := gpio.PullNoChange

	// Check if we have a specific mode in config
	if val, ok := t.cfg.Hardware.GPIO.Pins[pinIdentifier]; ok {
		if v, ok := val.(map[string]interface{}); ok {
			if mode, ok := v["mode"].(string); ok {
				if strings.ToUpper(mode) == "IN" || strings.ToUpper(mode) == "INPUT" {
					pull = gpio.PullUp // Default to pull-up for generic inputs
				}
			}
		}
	}

	// Set as input
	if err := p.In(pull, gpio.NoEdge); err != nil {
		return "", fmt.Errorf("failed to set read mode: %v", err)
	}

	val := p.Read()
	return fmt.Sprintf("Pin %s level is %v", pinIdentifier, val), nil
}

func (t *GPIOTool) writePin(pinIdentifier string, value int) (string, error) {
	p := t.resolvePin(pinIdentifier)
	if p == nil {
		return "", fmt.Errorf("pin '%s' not found", pinIdentifier)
	}

	l := gpio.Low
	if value > 0 {
		l = gpio.High
	}

	if err := p.Out(l); err != nil {
		return "", fmt.Errorf("failed to write to pin: %v", err)
	}

	return fmt.Sprintf("Set pin %s to %v", pinIdentifier, l), nil
}

func (t *GPIOTool) executeAction(actionName string) (string, error) {
	pinSettings, ok := t.actions[actionName]
	if !ok {
		return "", fmt.Errorf("action '%s' not defined in config", actionName)
	}

	var specificErrors []string

	for pinName, level := range pinSettings {
		_, err := t.writePin(pinName, level)
		if err != nil {
			specificErrors = append(specificErrors, fmt.Sprintf("%s: %v", pinName, err))
		} else {
			// Small delay effectively batches operations if run sequentially quickly enough for simple devices
			// Ideally we'd set them truly simultaneously if hardware supports it, but periph.io abstracts this.
			time.Sleep(10 * time.Millisecond)
		}
	}

	if len(specificErrors) > 0 {
		return fmt.Sprintf("Executed action '%s' with errors: %s", actionName, strings.Join(specificErrors, ", ")), nil
	}

	return fmt.Sprintf("Successfully executed action '%s'", actionName), nil
}
