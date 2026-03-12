package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Language      string          `json:"language" env:"MARUBOT_LANGUAGE"`
	AdminPassword string          `json:"admin_password" env:"MARUBOT_ADMIN_PASSWORD"`
	Agents        AgentsConfig    `json:"agents"`
	Channels      ChannelsConfig  `json:"channels"`
	Providers     ProvidersConfig `json:"providers"`
	Gateway       GatewayConfig   `json:"gateway"`
	Tools         ToolsConfig     `json:"tools"`
	Hardware      HardwareConfig  `json:"hardware"`
	Drone         DroneConfig     `json:"drone"`
	GPS           GPSConfig       `json:"gps"`
	mu            sync.RWMutex
}

type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
}

type AgentDefaults struct {
	Workspace         string   `json:"workspace" env:"MARUBOT_AGENTS_DEFAULTS_WORKSPACE"`
	Model             string   `json:"model" env:"MARUBOT_AGENTS_DEFAULTS_MODEL"`
	FallbackModels    []string `json:"fallback_models" env:"MARUBOT_AGENTS_DEFAULTS_FALLBACK_MODELS"`
	MaxTokens         int      `json:"max_tokens" env:"MARUBOT_AGENTS_DEFAULTS_MAX_TOKENS"`
	Temperature       float64  `json:"temperature" env:"MARUBOT_AGENTS_DEFAULTS_TEMPERATURE"`
	MaxToolIterations int      `json:"max_tool_iterations" env:"MARUBOT_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS"`
}

type ChannelsConfig struct {
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Telegram TelegramConfig `json:"telegram"`
	Feishu   FeishuConfig   `json:"feishu"`
	Discord  DiscordConfig  `json:"discord"`
	MaixCam  MaixCamConfig  `json:"maixcam"`
	Webhook  WebhookConfig  `json:"webhook"`
}

type WebhookConfig struct {
	Enabled   bool     `json:"enabled" env:"MARUBOT_CHANNELS_WEBHOOK_ENABLED"`
	Port      int      `json:"port" env:"MARUBOT_CHANNELS_WEBHOOK_PORT"`
	Path      string   `json:"path" env:"MARUBOT_CHANNELS_WEBHOOK_PATH"`
	Secret    string   `json:"secret" env:"MARUBOT_CHANNELS_WEBHOOK_SECRET"`
	AllowFrom []string `json:"allow_from" env:"MARUBOT_CHANNELS_WEBHOOK_ALLOW_FROM"`
}

type WhatsAppConfig struct {
	Enabled   bool     `json:"enabled" env:"MARUBOT_CHANNELS_WHATSAPP_ENABLED"`
	BridgeURL string   `json:"bridge_url" env:"MARUBOT_CHANNELS_WHATSAPP_BRIDGE_URL"`
	AllowFrom []string `json:"allow_from" env:"MARUBOT_CHANNELS_WHATSAPP_ALLOW_FROM"`
}

type TelegramConfig struct {
	Enabled   bool     `json:"enabled" env:"MARUBOT_CHANNELS_TELEGRAM_ENABLED"`
	Token     string   `json:"token" env:"MARUBOT_CHANNELS_TELEGRAM_TOKEN"`
	AllowFrom []string `json:"allow_from" env:"MARUBOT_CHANNELS_TELEGRAM_ALLOW_FROM"`
}

type FeishuConfig struct {
	Enabled           bool     `json:"enabled" env:"MARUBOT_CHANNELS_FEISHU_ENABLED"`
	AppID             string   `json:"app_id" env:"MARUBOT_CHANNELS_FEISHU_APP_ID"`
	AppSecret         string   `json:"app_secret" env:"MARUBOT_CHANNELS_FEISHU_APP_SECRET"`
	EncryptKey        string   `json:"encrypt_key" env:"MARUBOT_CHANNELS_FEISHU_ENCRYPT_KEY"`
	VerificationToken string   `json:"verification_token" env:"MARUBOT_CHANNELS_FEISHU_VERIFICATION_TOKEN"`
	AllowFrom         []string `json:"allow_from" env:"MARUBOT_CHANNELS_FEISHU_ALLOW_FROM"`
}

type DiscordConfig struct {
	Enabled   bool     `json:"enabled" env:"MARUBOT_CHANNELS_DISCORD_ENABLED"`
	Token     string   `json:"token" env:"MARUBOT_CHANNELS_DISCORD_TOKEN"`
	AllowFrom []string `json:"allow_from" env:"MARUBOT_CHANNELS_DISCORD_ALLOW_FROM"`
}

type MaixCamConfig struct {
	Enabled   bool     `json:"enabled" env:"MARUBOT_CHANNELS_MAIXCAM_ENABLED"`
	Host      string   `json:"host" env:"MARUBOT_CHANNELS_MAIXCAM_HOST"`
	Port      int      `json:"port" env:"MARUBOT_CHANNELS_MAIXCAM_PORT"`
	AllowFrom []string `json:"allow_from" env:"MARUBOT_CHANNELS_MAIXCAM_ALLOW_FROM"`
}

type ProvidersConfig struct {
	Anthropic  ProviderConfig `json:"anthropic"`
	OpenAI     ProviderConfig `json:"openai"`
	OpenRouter ProviderConfig `json:"openrouter"`
	Groq       ProviderConfig `json:"groq"`
	Zhipu      ProviderConfig `json:"zhipu"`
	VLLM       ProviderConfig `json:"vllm"`
	Gemini     ProviderConfig `json:"gemini"`
}

type ProviderConfig struct {
	APIKey  string `json:"api_key" env:"MARUBOT_PROVIDERS_{{.Name}}_API_KEY"`
	APIBase string `json:"api_base" env:"MARUBOT_PROVIDERS_{{.Name}}_API_BASE"`
}

type GatewayConfig struct {
	Host string `json:"host" env:"MARUBOT_GATEWAY_HOST"`
	Port int    `json:"port" env:"MARUBOT_GATEWAY_PORT"`
}

type WebSearchConfig struct {
	APIKey     string `json:"api_key" env:"MARUBOT_TOOLS_WEB_SEARCH_API_KEY"`
	MaxResults int    `json:"max_results" env:"MARUBOT_TOOLS_WEB_SEARCH_MAX_RESULTS"`
}

type WebToolsConfig struct {
	Search WebSearchConfig `json:"search"`
}

type ToolsConfig struct {
	Web WebToolsConfig `json:"web"`
}

type HardwareConfig struct {
	GPIO GPIOConfig `json:"gpio"`
}

type GPIOConfig struct {
	Enabled bool                      `json:"enabled"`
	Pins    map[string]interface{}    `json:"pins"`
	Actions map[string]map[string]int `json:"actions"`
}

type DroneConfig struct {
	Enabled    bool   `json:"enabled"`
	Connection string `json:"connection"`
	SysID      uint8  `json:"sysid"`
	CompID     uint8  `json:"compid"`
}

type GPSConfig struct {
	Enabled bool   `json:"enabled"`
	Device  string `json:"device"`
	Baud    int    `json:"baud"`
}

func DefaultConfig() *Config {
	return &Config{
		Language: "ko",
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Workspace:         "~/.marubot/workspace",
				Model:             "glm-4.7",
				FallbackModels:    []string{"gpt-4o", "claude-3-5-sonnet-20241022", "gemini-2.5-flash"},
				MaxTokens:         8192,
				Temperature:       0.7,
				MaxToolIterations: 20,
			},
		},
		Channels: ChannelsConfig{
			WhatsApp: WhatsAppConfig{
				Enabled:   false,
				BridgeURL: "ws://localhost:3001",
				AllowFrom: []string{},
			},
			Telegram: TelegramConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: []string{},
			},
			Feishu: FeishuConfig{
				Enabled:           false,
				AppID:             "",
				AppSecret:         "",
				EncryptKey:        "",
				VerificationToken: "",
				AllowFrom:         []string{},
			},
			Discord: DiscordConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: []string{},
			},
			MaixCam: MaixCamConfig{
				Enabled:   false,
				Host:      "0.0.0.0",
				Port:      18790,
				AllowFrom: []string{},
			},
			Webhook: WebhookConfig{
				Enabled:   false,
				Port:      0, // 0 means use default port (e.g. dashboard port or gateway port)
				Path:      "/api/channels/webhook",
				Secret:    "",
				AllowFrom: []string{},
			},
		},
		Providers: ProvidersConfig{
			Anthropic:  ProviderConfig{},
			OpenAI:     ProviderConfig{},
			OpenRouter: ProviderConfig{},
			Groq:       ProviderConfig{},
			Zhipu:      ProviderConfig{},
			VLLM:       ProviderConfig{},
			Gemini:     ProviderConfig{},
		},
		Gateway: GatewayConfig{
			Host: "0.0.0.0",
			Port: 18790,
		},
		Tools: ToolsConfig{
			Web: WebToolsConfig{
				Search: WebSearchConfig{
					APIKey:     "",
					MaxResults: 5,
				},
			},
		},
		Hardware: HardwareConfig{
			GPIO: GPIOConfig{
				Enabled: true,
				Pins: map[string]interface{}{
					"led_status":    18,
					"button_action": 23,
					"motor_a":       map[string]interface{}{"en": 12, "in1": 24, "in2": 25},
					"motor_b":       map[string]interface{}{"en": 13, "in3": 26, "in4": 27},
				},
			},
		},
		Drone: DroneConfig{
			Enabled:    false,
			Connection: "/dev/ttyAMA0:57600",
			SysID:      1,
			CompID:     1,
		},
		GPS: GPSConfig{
			Enabled: false,
			Device:  "/dev/ttyUSB0",
			Baud:    9600,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Load usersetting.json as override if exists
	userSettingsPath := filepath.Join(filepath.Dir(path), "usersetting.json")
	if userData, err := os.ReadFile(userSettingsPath); err == nil {
		var userCfg map[string]interface{}
		if err := json.Unmarshal(userData, &userCfg); err == nil {
			// Special handling for GPIO pins: if exists in usersetting, replace the entire map
			if hw, ok := userCfg["hardware"].(map[string]interface{}); ok {
				if gp, ok := hw["gpio"].(map[string]interface{}); ok {
					if pins, ok := gp["pins"].(map[string]interface{}); ok {
						cfg.Hardware.GPIO.Pins = pins
						delete(gp, "pins")
					}
				}
			}

			// Final unmarshal for everything else (merges into cfg)
			if err := json.Unmarshal(userData, cfg); err != nil {
				return nil, err
			}
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// FlattenPins converts nested pin maps into flat underscore-separated keys
func FlattenPins(pins map[string]interface{}) map[string]int {
	flat := make(map[string]int)
	var flatten func(prefix string, m map[string]interface{})
	flatten = func(prefix string, m map[string]interface{}) {
		for k, v := range m {
			key := k
			if prefix != "" {
				key = prefix + "_" + k
			}

			switch val := v.(type) {
			case int:
				flat[key] = val
			case float64:
				flat[key] = int(val)
			case map[string]interface{}:
				flatten(key, val)
			}
		}
	}
	flatten("", pins)
	return flat
}

// UnflattenPins converts flat underscore-separated keys into nested maps
func UnflattenPins(flat map[string]int) map[string]interface{} {
	nested := make(map[string]interface{})

	for key, val := range flat {
		parts := strings.Split(key, "_")
		curr := nested

		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]
			if _, ok := curr[part]; !ok {
				curr[part] = make(map[string]interface{})
			}
			curr = curr[part].(map[string]interface{})
		}
		curr[parts[len(parts)-1]] = val
	}

	return nested
}

func SaveConfig(path string, cfg *Config) error {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) Update(newCfg *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Language = newCfg.Language
	c.AdminPassword = newCfg.AdminPassword
	c.Agents = newCfg.Agents
	c.Channels = newCfg.Channels
	c.Providers = newCfg.Providers
	c.Gateway = newCfg.Gateway
	c.Tools = newCfg.Tools
	c.Hardware = newCfg.Hardware
	c.Drone = newCfg.Drone
	c.GPS = newCfg.GPS
}

func (c *Config) UpdateGPIO(pins map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Hardware.GPIO.Pins = pins
}

func (c *Config) WorkspacePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return expandHome(c.Agents.Defaults.Workspace)
}

func (c *Config) GetAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Providers.OpenRouter.APIKey != "" {
		return c.Providers.OpenRouter.APIKey
	}
	if c.Providers.Anthropic.APIKey != "" {
		return c.Providers.Anthropic.APIKey
	}
	if c.Providers.OpenAI.APIKey != "" {
		return c.Providers.OpenAI.APIKey
	}
	if c.Providers.Gemini.APIKey != "" {
		return c.Providers.Gemini.APIKey
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIKey
	}
	if c.Providers.Groq.APIKey != "" {
		return c.Providers.Groq.APIKey
	}
	if c.Providers.VLLM.APIKey != "" {
		return c.Providers.VLLM.APIKey
	}
	return ""
}

func (c *Config) GetAPIBase() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Providers.OpenRouter.APIKey != "" {
		if c.Providers.OpenRouter.APIBase != "" {
			return c.Providers.OpenRouter.APIBase
		}
		return "https://openrouter.ai/api/v1"
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIBase
	}
	if c.Providers.VLLM.APIKey != "" && c.Providers.VLLM.APIBase != "" {
		return c.Providers.VLLM.APIBase
	}
	return ""
}

func (c *Config) IsAIConfigured() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Providers.Anthropic.APIKey != "" ||
		c.Providers.OpenAI.APIKey != "" ||
		c.Providers.OpenRouter.APIKey != "" ||
		c.Providers.Groq.APIKey != "" ||
		c.Providers.Zhipu.APIKey != "" ||
		c.Providers.Gemini.APIKey != "" ||
		c.Providers.VLLM.APIBase != ""
}

func (c *Config) IsChannelEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Channels.WhatsApp.Enabled ||
		c.Channels.Telegram.Enabled ||
		c.Channels.Feishu.Enabled ||
		c.Channels.Discord.Enabled ||
		c.Channels.MaixCam.Enabled ||
		c.Channels.Webhook.Enabled
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}
