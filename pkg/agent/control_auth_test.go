package agent

import (
	"github.com/dirmich/marubot/pkg/config"
	"testing"
)

func TestControlSenderRequiresExplicitOwner(t *testing.T) {
	cfg := config.DefaultConfig()
	if controlSenderAllowed(cfg, "telegram", "any") {
		t.Fatal("공개 채널에서 소유자 장비 접근 허용")
	}
	cfg.Channels.Telegram.AllowFrom = []string{"owner"}
	if !controlSenderAllowed(cfg, "telegram", "owner") || controlSenderAllowed(cfg, "telegram", "other") {
		t.Fatal("채널 소유자 검증 실패")
	}
	if !controlSenderAllowed(cfg, "cli", "user") {
		t.Fatal("인증된 로컬 진입 차단")
	}
}
