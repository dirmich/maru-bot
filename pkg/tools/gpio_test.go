package tools

import "testing"

func TestGPIODirection(t *testing.T) {
	if got := gpioDirection("button_action"); got != "input" {
		t.Fatalf("button_action direction = %q, want input", got)
	}
	if got := gpioDirection("led_status"); got != "output" {
		t.Fatalf("led_status direction = %q, want output", got)
	}
}
