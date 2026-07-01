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

func TestIsSystemStatusRequest(t *testing.T) {
	if !isSystemStatusRequest("시스템 현황 알려줘") {
		t.Fatal("expected system status request")
	}
	if !isSystemStatusRequest("server status") {
		t.Fatal("expected English system status request")
	}
	if !isSystemStatusRequest("오늘 날짜와 현재 시간 알려줘") {
		t.Fatal("expected date/time request to use verified harness")
	}
	if !isSystemStatusRequest("what time is it?") {
		t.Fatal("expected English current time request")
	}
	if isSystemStatusRequest("현재 gpio상태를 알려줘") {
		t.Fatal("gpio status should use gpio flow, not system status harness")
	}
	if isSystemStatusRequest("check update") {
		t.Fatal("update should not be treated as a date request")
	}
}
