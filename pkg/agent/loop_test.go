package agent

import "testing"

func TestTryParseCallStyleToolCall(t *testing.T) {
	al := &AgentLoop{}

	tc := al.tryParseToolCallFromContent(`call:gpio_control{action: "status"}`)
	if tc == nil {
		t.Fatal("expected tool call")
	}
	if tc.Name != "gpio_control" {
		t.Fatalf("expected gpio_control, got %q", tc.Name)
	}
	if got := tc.Arguments["action"]; got != "status" {
		t.Fatalf("expected status action, got %#v", got)
	}
}
