package channels

import (
	"context"
	"fmt"
	"log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
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
		socketmode.OptionLog(log.Default()),
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
	log.Println("Starting Slack channel (Socket Mode)...")

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
						continue
					}
					c.socket.Ack(*evt.Request)

					switch eventsAPIEvent.Type {
					case slackevents.CallbackEvent:
						innerEvent := eventsAPIEvent.InnerEvent
						switch ev := innerEvent.Data.(type) {
						case *slackevents.MessageEvent:
							// Ignore messages from the bot itself
							if ev.BotID != "" || ev.User == "" {
								continue
							}
							c.handleMessage(ev)
						}
					}
				}
			}
		}
	}()

	go func() {
		if err := c.socket.RunContext(ctx); err != nil {
			log.Printf("Slack socket mode error: %v", err)
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

	log.Printf("Slack message from %s in %s: %s", ev.User, ev.Channel, ev.Text)
	c.HandleMessage(ev.User, ev.Channel, ev.Text, nil, metadata)
}
