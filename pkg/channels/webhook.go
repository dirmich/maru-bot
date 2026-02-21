package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
)

type WebhookChannel struct {
	*BaseChannel
	config config.WebhookConfig
	server *http.Server
	// responseMap stores channels for synchronous responses
	// key is ChatID, value is a channel for the response content
	responseMap map[string]chan string
	mu          sync.RWMutex
}

type WebhookRequest struct {
	Message    string            `json:"message"`
	User       string            `json:"user"`
	ChatID     string            `json:"chat_id"`
	SessionKey string            `json:"session_key"`
	Metadata   map[string]string `json:"metadata"`
}

func NewWebhookChannel(cfg config.WebhookConfig, bus *bus.MessageBus) (*WebhookChannel, error) {
	base := NewBaseChannel("webhook", cfg, bus, cfg.AllowFrom)
	return &WebhookChannel{
		BaseChannel: base,
		config:      cfg,
		responseMap: make(map[string]chan string),
	}, nil
}

func (c *WebhookChannel) Start(ctx context.Context) error {
	if !c.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc(c.config.Path, c.handleWebhook)

	addr := fmt.Sprintf(":%d", c.config.Port)
	if c.config.Port == 0 {
		// If port is 0, we assume it's integrated elsewhere or we use a default
		// For standalone start, let's use a default if not specified
		addr = ":18791" 
		log.Printf("Webhook port not specified, using default :18791")
	}

	c.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	c.setRunning(true)
	log.Printf("Webhook channel listening on %s%s", addr, c.config.Path)

	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Webhook server error: %v", err)
			c.setRunning(false)
		}
	}()

	return nil
}

func (c *WebhookChannel) Stop(ctx context.Context) error {
	c.setRunning(false)
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

func (c *WebhookChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if msg.Action == "typing" {
		return nil // Webhook doesn't support typing indicator in a standard way
	}

	c.mu.Lock()
	respChan, ok := c.responseMap[msg.ChatID]
	if ok {
		delete(c.responseMap, msg.ChatID)
		c.mu.Unlock()
		respChan <- msg.Content
		return nil
	}
	c.mu.Unlock()

	log.Printf("Webhook outbound: channel not waiting for chatID %s", msg.ChatID)
	return nil
}

func (c *WebhookChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Secret validation
	if c.config.Secret != "" {
		secret := r.Header.Get("X-Maru-Secret")
		if secret != c.config.Secret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	senderID := req.User
	if senderID == "" {
		senderID = "webhook-user"
	}

	chatID := req.ChatID
	if chatID == "" {
		chatID = fmt.Sprintf("hook-%d", time.Now().UnixNano())
	}

	sessionKey := req.SessionKey
	if sessionKey == "" {
		sessionKey = "webhook:" + senderID
	}

	// Create a channel for the response
	respChan := make(chan string, 1)
	c.mu.Lock()
	c.responseMap[chatID] = respChan
	c.mu.Unlock()

	// Ensure cleanup if AI doesn't respond or takes too long
	defer func() {
		c.mu.Lock()
		delete(c.responseMap, chatID)
		c.mu.Unlock()
	}()

	c.HandleMessage(senderID, chatID, req.Message, nil, req.Metadata)

	// Wait for response or timeout
	select {
	case response := <-respChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": response,
			"chat_id":  chatID,
		})
	case <-time.After(60 * time.Second): // 60s timeout for AI response
		http.Error(w, "AI response timeout", http.StatusGatewayTimeout)
	case <-r.Context().Done():
		// Request cancelled by client
	}
}
