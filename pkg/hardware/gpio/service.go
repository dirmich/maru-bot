package gpio

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type GPIOService struct {
	cfg     *config.Config
	mb      *bus.MessageBus
	running bool
}

func NewGPIOService(cfg *config.Config, mb *bus.MessageBus) *GPIOService {
	host.Init()
	return &GPIOService{
		cfg: cfg,
		mb:  mb,
	}
}

func (s *GPIOService) Start(ctx context.Context) error {
	if !s.cfg.Hardware.GPIO.Enabled {
		return nil
	}

	s.running = true
	log.Println("GPIO Monitoring Service started")

	// Find pins that look like inputs (containing "button", "sensor")
	for name, val := range s.cfg.Hardware.GPIO.Pins {
		pinName := fmt.Sprintf("%v", val)
		if IsInputPin(name) {
			go s.monitorPin(ctx, name, pinName)
		}
	}

	return nil
}

func IsInputPin(name string) bool {
	n := name
	return (n == "button" || n == "sensor" || 
			(len(n) > 6 && n[:6] == "button") || 
			(len(n) > 6 && n[:6] == "sensor") ||
			(len(n) > 7 && n[len(n)-7:] == "_button") ||
			(len(n) > 7 && n[len(n)-7:] == "_sensor"))
}

func (s *GPIOService) monitorPin(ctx context.Context, label, pinName string) {
	p := gpioreg.ByName(pinName)
	if p == nil {
		log.Printf("GPIO Service: Pin %s (%s) not found", label, pinName)
		return
	}

	// Configure as input with pull-up and edge detection
	if err := p.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Printf("GPIO Service: Failed to configure pin %s: %v", pinName, err)
		return
	}

	log.Printf("Monitoring GPIO pin: %s (%s)", label, pinName)

	lastLevel := p.Read()
	for s.running {
		select {
		case <-ctx.Done():
			return
		default:
			if p.WaitForEdge(time.Second) {
				newLevel := p.Read()
				if newLevel != lastLevel {
					event := "pressed"
					if newLevel == gpio.High {
						event = "released"
					}
					
					msg := fmt.Sprintf("[GPIO Event] %s (%s) has been %s", label, pinName, event)
					log.Println(msg)

					levelInt := 0
					if newLevel == gpio.High {
						levelInt = 1
					}

					s.mb.PublishInbound(bus.InboundMessage{
						Channel:    "gpio",
						SenderID:   "system",
						ChatID:     "gpio_event",
						Content:    msg,
						SessionKey: "cli:default", // Standard session for system events
						Metadata: map[string]string{
							"pin_label": label,
							"pin_name":  pinName,
							"event":     event,
							"level":     fmt.Sprintf("%d", levelInt),
						},
					})
					lastLevel = newLevel
				}
			}
		}
	}
}

func (s *GPIOService) Stop() {
	s.running = false
}
