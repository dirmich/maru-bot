package tools

import (
	"context"
	"fmt"
	"strings"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

type MotorTool struct {
	pins map[string]interface{}
}

func NewMotorTool(pins map[string]interface{}) *MotorTool {
	host.Init()
	return &MotorTool{pins: pins}
}

func (t *MotorTool) Name() string {
	return "motor_control"
}

func (t *MotorTool) Description() string {
	return "Control DC motors for movement (forward, backward, left, right, stop)"
}

func (t *MotorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Movement command: 'forward', 'backward', 'left', 'right', 'stop'",
				"enum":        []string{"forward", "backward", "left", "right", "stop"},
			},
			"speed": map[string]interface{}{
				"type":        "number",
				"description": "Speed from 0.0 to 1.0",
				"minimum":     0.0,
				"maximum":     1.0,
			},
		},
		"required": []string{"command"},
	}
}

func (t *MotorTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	command, _ := args["command"].(string)
	speed, ok := args["speed"].(float64)
	if !ok {
		speed = 0.5
	}

	motorA, _ := t.pins["motor_a"].(map[string]interface{})
	motorB, _ := t.pins["motor_b"].(map[string]interface{})

	if motorA == nil || motorB == nil {
		return "", fmt.Errorf("motor pins not configured")
	}

	switch strings.ToLower(command) {
	case "forward":
		t.setMotor(motorA, speed, true)
		t.setMotor(motorB, speed, true)
	case "backward":
		t.setMotor(motorA, speed, false)
		t.setMotor(motorB, speed, false)
	case "left":
		t.setMotor(motorA, speed, false)
		t.setMotor(motorB, speed, true)
	case "right":
		t.setMotor(motorA, speed, true)
		t.setMotor(motorB, speed, false)
	case "stop":
		t.setMotor(motorA, 0, true)
		t.setMotor(motorB, 0, true)
	default:
		return "", fmt.Errorf("unknown command: %s", command)
	}

	return fmt.Sprintf("Executed motor command: %s at speed %.2f", command, speed), nil
}

func (t *MotorTool) setMotor(config map[string]interface{}, speed float64, forward bool) {
	enPin := gpioreg.ByName(fmt.Sprintf("%v", config["en"]))
	in1Pin := gpioreg.ByName(fmt.Sprintf("%v", config["in1"]))
	in2Pin := gpioreg.ByName(fmt.Sprintf("%v", config["in2"]))

	if enPin == nil || in1Pin == nil || in2Pin == nil {
		return
	}

	if speed <= 0 {
		in1Pin.Out(gpio.Low)
		in2Pin.Out(gpio.Low)
		enPin.Out(gpio.Low)
		return
	}

	if forward {
		in1Pin.Out(gpio.High)
		in2Pin.Out(gpio.Low)
	} else {
		in1Pin.Out(gpio.Low)
		in2Pin.Out(gpio.High)
	}

	// Simple PWM via periph.io (if supported by hardware)
	// For RPi, some pins support hardware PWM. Others might need software PWM which periph handles.
	enPin.PWM(gpio.Duty(float64(gpio.DutyMax)*speed), 1000*physic.Hertz)
}
