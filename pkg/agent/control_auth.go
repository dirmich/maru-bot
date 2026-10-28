package agent

import "github.com/dirmich/marubot/pkg/config"

// 분산 장비와 Wiki 자격 자료가 있는 프로세스는 공개 채널 입력을 허용하지 않는다.
func controlSenderAllowed(cfg *config.Config, channel, sender string) bool {
	cfg.Mu.RLock()
	defer cfg.Mu.RUnlock()
	if channel == "cli" {
		return true
	}
	var allowed []string
	switch channel {
	case "telegram":
		allowed = cfg.Channels.Telegram.AllowFrom
	case "discord":
		allowed = cfg.Channels.Discord.AllowFrom
	case "slack":
		allowed = cfg.Channels.Slack.AllowFrom
	case "whatsapp":
		allowed = cfg.Channels.WhatsApp.AllowFrom
	case "webhook":
		allowed = cfg.Channels.Webhook.AllowFrom
	default:
		return false
	}
	for _, id := range allowed {
		if id == sender && id != "" {
			return true
		}
	}
	return false
}
