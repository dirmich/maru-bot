package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/dirmich/marubot/pkg/utils"
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
	Mu            sync.RWMutex    `json:"-"`
}

type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
}

type AgentDefaults struct {
	Workspace         string   `json:"workspace" env:"MARUBOT_AGENTS_DEFAULTS_WORKSPACE"`
	Provider          string   `json:"provider" env:"MARUBOT_AGENTS_DEFAULTS_PROVIDER"`
	Model             string   `json:"model" env:"MARUBOT_AGENTS_DEFAULTS_MODEL"`
	FallbackModels    []string `json:"fallback_models" env:"MARUBOT_AGENTS_DEFAULTS_FALLBACK_MODELS"`
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
	Ollama     []ProviderConfig `json:"ollama"`
}

type ProviderConfig struct {
	APIKey  string        `json:"api_key,omitempty"`
	APIBase string        `json:"api_base,omitempty"`
	Models  []ModelConfig `json:"models"`
}

type ModelConfig struct {
	Model             string  `json:"model"`
	APIKey            string  `json:"api_key"`
	APIBase           string  `json:"api_base"`
	MaxTokens         int     `json:"max_tokens"`
	Temperature       float64 `json:"temperature"`
	MaxToolIterations int     `json:"max_tool_iterations"`
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
	IsRaspberryPi *bool      `json:"is_raspberry_pi,omitempty" env:"MARUBOT_HARDWARE_IS_RASPBERRY_PI"`
	GPIOTestMode  bool       `json:"gpio_test_mode" env:"MARUBOT_HARDWARE_GPIO_TEST_MODE"`
	GPIO          GPIOConfig `json:"gpio"`
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
				Workspace:      "~/.marubot/workspace",
				Provider:       "vllm",
				Model:          "openai/gpt-oss-20b",
				FallbackModels: []string{"gpt-4o", "claude-3-5-sonnet-20241022", "gemini-2.0-flash"},
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
			Anthropic: ProviderConfig{
				Models: []ModelConfig{},
			},
			OpenAI: ProviderConfig{
				Models: []ModelConfig{},
			},
			OpenRouter: ProviderConfig{
				Models: []ModelConfig{},
			},
			Groq: ProviderConfig{
				Models: []ModelConfig{},
			},
			Zhipu: ProviderConfig{
				Models: []ModelConfig{},
			},
			VLLM: ProviderConfig{
				Models: []ModelConfig{
					{
						Model:             "openai/gpt-oss-20b",
						APIKey:            "vllm",
						APIBase:           "http://192.168.0.20:8000/v1",
						MaxTokens:         8192,
						Temperature:       0.7,
						MaxToolIterations: 20,
					},
				},
			},
			Gemini: ProviderConfig{
				Models: []ModelConfig{},
			},
		},
		Gateway: GatewayConfig{
			Host: "0.0.0.0",
			Port: 8080,
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
			GPIOTestMode: false,
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
		// Temporary struct to capture old format fields for migration
		var oldCfg struct {
			Agents struct {
				Defaults struct {
					MaxTokens         int     `json:"max_tokens"`
					Temperature       float64 `json:"temperature"`
					MaxToolIterations int     `json:"max_tool_iterations"`
				} `json:"defaults"`
			} `json:"agents"`
			Providers map[string]struct {
				APIKey  string `json:"api_key"`
				APIBase string `json:"api_base"`
			} `json:"providers"`
		}
		json.Unmarshal(data, &oldCfg)

		// Special handling for GPIO pins in main config to avoid merging with defaults
		var rawCfg map[string]interface{}
		json.Unmarshal(data, &rawCfg)
		if hw, ok := rawCfg["hardware"].(map[string]interface{}); ok {
			if gp, ok := hw["gpio"].(map[string]interface{}); ok {
				if _, ok := gp["pins"].(map[string]interface{}); ok {
					// Clear default pins if pins are provided in the config file
					cfg.Hardware.GPIO.Pins = make(map[string]interface{})
				}
			}
		}

		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}

		// Migrate old format to new format if needed
		migrateProvider := func(name string, p *ProviderConfig) {
			if old, ok := oldCfg.Providers[name]; ok && old.APIKey != "" {
				// If Models is empty, migrate the old direct fields
				if len(p.Models) == 0 {
					p.Models = append(p.Models, ModelConfig{
						Model:             cfg.Agents.Defaults.Model, // Use current default model
						APIKey:            old.APIKey,
						APIBase:           old.APIBase,
						MaxTokens:         oldCfg.Agents.Defaults.MaxTokens,
						Temperature:       oldCfg.Agents.Defaults.Temperature,
						MaxToolIterations: oldCfg.Agents.Defaults.MaxToolIterations,
					})
				}
			}
		}

		migrateProvider("anthropic", &cfg.Providers.Anthropic)
		migrateProvider("openai", &cfg.Providers.OpenAI)
		migrateProvider("openrouter", &cfg.Providers.OpenRouter)
		migrateProvider("groq", &cfg.Providers.Groq)
		migrateProvider("zhipu", &cfg.Providers.Zhipu)
		migrateProvider("vllm", &cfg.Providers.VLLM)
		migrateProvider("gemini", &cfg.Providers.Gemini)

	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Unify: Check for usersetting.json and migrate to config.json if it exists
	userSettingsPath := filepath.Join(filepath.Dir(path), "usersetting.json")
	if _, err := os.Stat(userSettingsPath); err == nil {
		if err := MigrateUserSettings(path, userSettingsPath, cfg); err == nil {
			log.Printf("Successfully migrated usersetting.json to %s and removed it", path)
		} else {
			log.Printf("Migration of usersetting.json failed: %v", err)
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Auto-encrypt password if it's plaintext
	if cfg.AdminPassword != "" && !utils.IsPasswordHashed(cfg.AdminPassword) {
		log.Printf("Plaintext admin password detected. Encrypting...")
		cfg.AdminPassword = utils.HashPassword(cfg.AdminPassword)
		if err := SaveConfig(path, cfg); err != nil {
			log.Printf("Failed to save encrypted password: %v", err)
		} else {
			log.Printf("Admin password successfully encrypted.")
		}
	}

	return cfg, nil
}

// MigrateUserSettings merges all data from usersetting.json into Config and saves it to config.json
func MigrateUserSettings(configPath, userSettingsPath string, cfg *Config) error {
	userData, err := os.ReadFile(userSettingsPath)
	if err != nil {
		return err
	}

	// 1. First, decode into a map to log what's being migrated (optional, for debug)
	var userCfg map[string]interface{}
	if err := json.Unmarshal(userData, &userCfg); err != nil {
		return err
	}

	// 2. Unmarshal directly into the existing Config struct to merge all fields automatically.
	// Since cfg already has defaults/existing data, this will overwrite only what's in usersetting.json.
	if err := json.Unmarshal(userData, cfg); err != nil {
		return err
	}

	// 3. Save merged config back to config.json
	if err := SaveConfig(configPath, cfg); err != nil {
		return err
	}

	// 4. Rename usersetting.json to usersetting.json.bak for safety
	bakPath := userSettingsPath + ".bak"
	os.Remove(bakPath) // Remove old backup if exists
	return os.Rename(userSettingsPath, bakPath)
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
	cfg.Mu.RLock()
	defer cfg.Mu.RUnlock()

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
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if newCfg.Language != "" {
		c.Language = newCfg.Language
	}
	if newCfg.AdminPassword != "" {
		c.AdminPassword = newCfg.AdminPassword
	}
	
	// Selective Agents update
	if newCfg.Agents.Defaults.Model != "" {
		c.Agents = newCfg.Agents
	}

	// Merge Channels config
	c.Channels.Telegram = newCfg.Channels.Telegram
	c.Channels.WhatsApp = newCfg.Channels.WhatsApp
	c.Channels.Discord = newCfg.Channels.Discord
	c.Channels.Feishu = newCfg.Channels.Feishu
	c.Channels.MaixCam = newCfg.Channels.MaixCam
	c.Channels.Webhook = newCfg.Channels.Webhook

	// Merge Providers (assuming the UI sends the full provider model list it wants to update)
	if len(newCfg.Providers.OpenAI.Models) > 0 || newCfg.Providers.OpenAI.APIKey != "" { c.Providers.OpenAI = newCfg.Providers.OpenAI }
	if len(newCfg.Providers.Anthropic.Models) > 0 || newCfg.Providers.Anthropic.APIKey != "" { c.Providers.Anthropic = newCfg.Providers.Anthropic }
	if len(newCfg.Providers.Gemini.Models) > 0 || newCfg.Providers.Gemini.APIKey != "" { c.Providers.Gemini = newCfg.Providers.Gemini }
	if len(newCfg.Providers.Zhipu.Models) > 0 || newCfg.Providers.Zhipu.APIKey != "" { c.Providers.Zhipu = newCfg.Providers.Zhipu }
	if len(newCfg.Providers.Groq.Models) > 0 || newCfg.Providers.Groq.APIKey != "" { c.Providers.Groq = newCfg.Providers.Groq }
	if len(newCfg.Providers.VLLM.Models) > 0 || newCfg.Providers.VLLM.APIKey != "" { c.Providers.VLLM = newCfg.Providers.VLLM }
	if len(newCfg.Providers.Ollama) > 0 { c.Providers.Ollama = newCfg.Providers.Ollama }

	c.Gateway = newCfg.Gateway
	c.Tools = newCfg.Tools

	// Selective hardware update to avoid losing GPIO pins when saving general settings
	if newCfg.Hardware.GPIO.Enabled {
		c.Hardware.GPIO.Enabled = true
		if len(newCfg.Hardware.GPIO.Pins) > 0 {
			c.Hardware.GPIO.Pins = newCfg.Hardware.GPIO.Pins
		}
	}
	c.Hardware.GPIOTestMode = newCfg.Hardware.GPIOTestMode

	if newCfg.Drone.Connection != "" || newCfg.Drone.Enabled {
		c.Drone = newCfg.Drone
	}
	if newCfg.GPS.Device != "" || newCfg.GPS.Enabled {
		c.GPS = newCfg.GPS
	}
}

func (c *Config) UpdateGPIO(pins map[string]interface{}) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Hardware.GPIO.Pins = pins
}

func (c *Config) WorkspacePath() string {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return expandHome(c.Agents.Defaults.Workspace)
}

func (c *Config) GetAPIKey() string {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	
	// Use default model from defaults if configured
	model := c.Agents.Defaults.Model
	provider := c.Agents.Defaults.Provider
	
	if provider != "" {
		mCfg := c.findModelConfig(provider, model)
		if mCfg != nil {
			return mCfg.APIKey
		}
	}
	
	// Fallback global search
	providers := []ProviderConfig{
		c.Providers.OpenRouter,
		c.Providers.Anthropic,
		c.Providers.OpenAI,
		c.Providers.Gemini,
		c.Providers.Zhipu,
		c.Providers.Groq,
		c.Providers.VLLM,
	}
	for _, p := range providers {
		for _, m := range p.Models {
			if m.APIKey != "" {
				return m.APIKey
			}
		}
	}

	// Check Ollama list
	for _, p := range c.Providers.Ollama {
		for _, m := range p.Models {
			if m.APIKey != "" {
				return m.APIKey
			}
		}
	}
	return ""
}

func (c *Config) findModelConfig(providerName, modelName string) *ModelConfig {
	var provider ProviderConfig
	switch strings.ToLower(providerName) {
	case "anthropic": provider = c.Providers.Anthropic
	case "openai": provider = c.Providers.OpenAI
	case "openrouter": provider = c.Providers.OpenRouter
	case "groq": provider = c.Providers.Groq
	case "zhipu": provider = c.Providers.Zhipu
	case "vllm": provider = c.Providers.VLLM
	case "gemini": provider = c.Providers.Gemini
	case "ollama":
		// For Ollama, we search through the slice of providers
		for _, p := range c.Providers.Ollama {
			for _, m := range p.Models {
				if strings.EqualFold(m.Model, modelName) {
					return &m
				}
			}
		}
		return nil
	default: return nil
	}
	for _, m := range provider.Models {
		if strings.EqualFold(m.Model, modelName) {
			return &m
		}
	}
	return nil
}


func (c *Config) GetAPIBase() string {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	
	model := c.Agents.Defaults.Model
	provider := c.Agents.Defaults.Provider
	
	if provider != "" {
		mCfg := c.findModelConfig(provider, model)
		if mCfg != nil {
			return mCfg.APIBase
		}
	}
	return ""
}

func (c *Config) IsAIConfigured() bool {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	
	allProviders := []ProviderConfig{
		c.Providers.Anthropic,
		c.Providers.OpenAI,
		c.Providers.OpenRouter,
		c.Providers.Groq,
		c.Providers.Zhipu,
		c.Providers.Gemini,
		c.Providers.VLLM,
	}
	for _, p := range allProviders {
		if len(p.Models) > 0 {
			return true
		}
	}

	// Check Ollama list
	for _, p := range c.Providers.Ollama {
		if len(p.Models) > 0 {
			return true
		}
	}
	return false
}

func (c *Config) IsChannelEnabled() bool {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
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

func IsInputPin(name string) bool {
	n := strings.ToLower(name)
	return (n == "button" || n == "sensor" ||
		(len(n) > 6 && n[:6] == "button") ||
		(len(n) > 6 && n[:6] == "sensor") ||
		(len(n) > 7 && n[len(n)-7:] == "_button") ||
		(len(n) > 7 && n[len(n)-7:] == "_sensor"))
}
