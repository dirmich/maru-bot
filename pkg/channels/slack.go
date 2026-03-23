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
					logger.DebugCF("slack", "EventsAPI event received", map[string]interface{}{"type": eventsAPIEvent.Type})

					switch eventsAPIEvent.Type {
					case slackevents.CallbackEvent:
						innerEvent := eventsAPIEvent.InnerEvent
						switch ev := innerEvent.Data.(type) {
						case *slackevents.MessageEvent:
							logger.InfoCF("slack", "Received message event", map[string]interface{}{
								"user":    ev.User,
								"channel": ev.Channel,
								"bot_id":  ev.BotID,
								"text":    ev.Text,
							})
							// Ignore messages from the bot itself
							if ev.BotID != "" {
								logger.DebugC("slack", "Ignoring message from bot itself")
								continue
							}
							if ev.User == "" {
								logger.DebugC("slack", "Ignoring message with empty user (system message?)")
								continue
							}
							c.handleMessage(ev)
						case *slackevents.AppMentionEvent:
							logger.InfoCF("slack", "Received app_mention event", map[string]interface{}{
								"user":    ev.User,
								"channel": ev.Channel,
								"text":    ev.Text,
							})
							c.handleAppMention(ev)
						default:
							logger.DebugCF("slack", "Received unhandled callback event type", map[string]interface{}{
								"type": fmt.Sprintf("%T", innerEvent.Data),
							})
						}
					default:
						logger.DebugCF("slack", "Received unhandled EventsAPIEvent type", map[string]interface{}{"type": eventsAPIEvent.Type})
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
		logger.DebugC("slack", "Calling socket.RunContext...")
		if err := c.socket.RunContext(ctx); err != nil {
			logger.ErrorCF("slack", "Slack socket mode runtime error (RunContext exited)", map[string]interface{}{
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
	options := []slack.MsgOption{
		slack.MsgOptionText(msg.Content, false),
	}

	// Support threading
	if threadTS, ok := msg.Metadata["ts"]; ok && threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}

	_, _, err := c.api.PostMessageContext(ctx, msg.ChatID, options...)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	return nil
}

func (c *SlackChannel) handleMessage(ev *slackevents.MessageEvent) {
	logger.InfoCF("slack", "Handling Slack message", map[string]interface{}{
		"user":    ev.User,
		"channel": ev.Channel,
		"text":    ev.Text,
		"subtype": ev.SubType,
		"thread":  ev.ThreadTimeStamp,
	})
	
	// Ignore messages from the bot itself (redundant check but safe)
	if ev.BotID != "" {
		return
	}

	metadata := map[string]string{
		"user":    ev.User,
		"channel": ev.Channel,
	}

	// If message is already part of a thread, continue in that thread
	if ev.ThreadTimeStamp != "" {
		metadata["ts"] = ev.ThreadTimeStamp
	}

	c.HandleMessage(ev.User, ev.Channel, ev.Text, nil, metadata)
}

func (c *SlackChannel) handleAppMention(ev *slackevents.AppMentionEvent) {
	logger.InfoCF("slack", "Handling Slack app_mention", map[string]interface{}{
		"user":    ev.User,
		"channel": ev.Channel,
		"text":    ev.Text,
	})

	metadata := map[string]string{
		"ts":      ev.EventTimeStamp, // Reply in a thread for mentions
		"user":    ev.User,
		"channel": ev.Channel,
	}

	c.HandleMessage(ev.User, ev.Channel, ev.Text, nil, metadata)
}
