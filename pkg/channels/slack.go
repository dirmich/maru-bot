package channels

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/logger"
)

type SlackChannel struct {
	*BaseChannel
	api    *slack.Client
	socket *socketmode.Client
	config config.SlackConfig
}

func NewSlackChannel(cfg config.SlackConfig, bus *bus.MessageBus) (*SlackChannel, error) {
	api := slack.New(
		cfg.Token,
		slack.OptionAppLevelToken(cfg.AppToken),
	)

	socket := socketmode.New(
		api,
		// socketmode.OptionLog(log.Default()), // Disable internal logging to avoid console clutter
	)

	base := NewBaseChannel("slack", cfg, bus, cfg.AllowFrom)

	return &SlackChannel{
		BaseChannel: base,
		api:         api,
		socket:      socket,
		config:      cfg,
	}, nil
}

func (c *SlackChannel) Start(ctx context.Context) error {
	logger.InfoC("slack", "Starting Slack channel (Socket Mode)...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-c.socket.Events:
				switch evt.Type {
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
					if !ok {
						logger.DebugC("slack", "Failed to cast event to EventsAPIEvent")
						continue
					}
					c.socket.Ack(*evt.Request)

					switch eventsAPIEvent.Type {
					case slackevents.CallbackEvent:
						innerEvent := eventsAPIEvent.InnerEvent
						switch ev := innerEvent.Data.(type) {
						case *slackevents.MessageEvent:
							logger.DebugCF("slack", "Received message event", map[string]interface{}{
								"user":    ev.User,
								"channel": ev.Channel,
								"bot_id":  ev.BotID,
							})
							// Ignore messages from the bot itself
							if ev.BotID != "" || ev.User == "" {
								continue
							}
							c.handleMessage(ev)
						default:
							logger.DebugCF("slack", "Received unhandled callback event type", map[string]interface{}{
								"type": fmt.Sprintf("%T", innerEvent.Data),
							})
						}
					}
				case socketmode.EventTypeConnected:
					logger.InfoC("slack", "Slack Socket Mode connected successfully")
					c.setRunning(true)
				case socketmode.EventTypeHello:
					logger.DebugC("slack", "Slack hello event received")
				case socketmode.EventTypeInvalidAuth, socketmode.EventTypeConnectionError:
					logger.ErrorCF("slack", "Slack connection failed", map[string]interface{}{
						"type": evt.Type,
					})
					c.setRunning(false)
				}
			}
		}
	}()

	go func() {
		if err := c.socket.RunContext(ctx); err != nil {
			logger.ErrorCF("slack", "Slack socket mode runtime error", map[string]interface{}{
				"error": err.Error(),
			})
			c.setRunning(false)
		}
	}()

	c.setRunning(true)
	return nil
}

func (c *SlackChannel) Stop(ctx context.Context) error {
	c.setRunning(false)
	return nil
}

func (c *SlackChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	_, _, err := c.api.PostMessageContext(ctx, msg.ChatID, slack.MsgOptionText(msg.Content, false))
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	return nil
}

func (c *SlackChannel) handleMessage(ev *slackevents.MessageEvent) {
	// Basic implementation: pass to base handler
	metadata := map[string]string{
		"ts":      ev.TimeStamp,
		"user":    ev.User,
		"channel": ev.Channel,
	}

	logger.InfoCF("slack", "Handling Slack message", map[string]interface{}{
		"user":    ev.User,
		"channel": ev.Channel,
		"text":    ev.Text,
	})
	c.HandleMessage(ev.User, ev.Channel, ev.Text, nil, metadata)
}
