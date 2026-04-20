package tools

import (
	"context"
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"

	"github.com/dirmich/marubot/pkg/config"
)

type UltrasonicTool struct {
	cfg *config.Config
}

func NewUltrasonicTool(cfg *config.Config) *UltrasonicTool {
	host.Init()
	return &UltrasonicTool{cfg: cfg}
}

func (t *UltrasonicTool) Name() string {
	return "get_distance"
}

func (t *UltrasonicTool) Description() string {
	return "Measure distance to the nearest obstacle using HC-SR04 ultrasonic sensor"
}

func (t *UltrasonicTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *UltrasonicTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	config, _ := t.cfg.Hardware.GPIO.Pins["ultrasonic"].(map[string]interface{})
	if config == nil {
		return "", fmt.Errorf("ultrasonic pins not configured")
	}

	getPin := func(v interface{}) string {
		switch pv := v.(type) {
		case map[string]interface{}:
			if pin, ok := pv["pin"]; ok {
				return fmt.Sprintf("%v", pin)
			}
		default:
			return fmt.Sprintf("%v", v)
		}
		return ""
	}

	trigPin := gpioreg.ByName(getPin(config["trigger"]))
	echoPin := gpioreg.ByName(getPin(config["echo"]))

	if trigPin == nil || echoPin == nil {
		return "", fmt.Errorf("failed to find trigger or echo pins")
	}

	// Ensure trigger is low
	trigPin.Out(gpio.Low)
	time.Sleep(2 * time.Microsecond)

	// Sending trigger pulse
	trigPin.Out(gpio.High)
	time.Sleep(10 * time.Microsecond)
	trigPin.Out(gpio.Low)

	// Wait for echo to start
	timeout := time.Now().Add(10 * time.Millisecond)
	for echoPin.Read() == gpio.Low {
		if time.Now().After(timeout) {
			return "", fmt.Errorf("ultrasonic sensor timeout (failed to start)")
		}
	}
	start := time.Now()

	// Wait for echo to end
	timeout = time.Now().Add(30 * time.Millisecond) // Max distance (~5m)
	for echoPin.Read() == gpio.High {
		if time.Now().After(timeout) {
			break
		}
	}
	duration := time.Since(start)

	// distance = (time * speed of sound) / 2
	// Speed of sound = 343 m/s = 34300 cm/s
	distance := float64(duration.Seconds()*34300.0) / 2.0

	return fmt.Sprintf("Distance to nearest obstacle: %.2f cm", distance), nil
}
