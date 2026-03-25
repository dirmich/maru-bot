package tools

import (
	"context"
	"fmt"

	"github.com/dirmich/marubot/pkg/bus"
)

type ChannelTool struct {
	manager bus.ChannelManager
}

func NewChannelTool(manager bus.ChannelManager) *ChannelTool {
	return &ChannelTool{manager: manager}
}

func (t *ChannelTool) Name() string {
	return "send_channel_message"
}

func (t *ChannelTool) Description() string {
	return "Send a message to a specific channel (e.g., telegram, slack, discord, whatsapp). Use this to relay messages between platforms."
}

func (t *ChannelTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "The target channel name (telegram, slack, discord, whatsapp).",
			},
			"chat_id": map[string]interface{}{
				"type":        "string",
				"description": "The target Chat ID or Channel ID.",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The message text to send.",
			},
		},
		"required": []string{"channel", "chat_id", "content"},
	}
}

func (t *ChannelTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.manager == nil {
		return "", fmt.Errorf("channel manager not initialized")
	}

	channel, _ := args["channel"].(string)
	chatID, _ := args["chat_id"].(string)
	content, _ := args["content"].(string)

	if channel == "" || chatID == "" || content == "" {
		return "", fmt.Errorf("missing required arguments")
	}

	err := t.manager.SendToChannel(ctx, channel, chatID, content)
	if err != nil {
		return fmt.Sprintf("Error sending message to %s: %v", channel, err), nil
	}

	return fmt.Sprintf("Successfully sent message to %s (%s)", channel, chatID), nil
}
