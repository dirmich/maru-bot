// MaruBot - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 MaruBot contributors

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dirmich/marubot/pkg/config"
)

type HTTPProvider struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

func NewHTTPProvider(apiKey, apiBase string) *HTTPProvider {
	return &HTTPProvider{
		apiKey:  apiKey,
		apiBase: apiBase,
		httpClient: &http.Client{
			Timeout: 0,
		},
	}
}

func (p *HTTPProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	if p.apiBase == "" {
		return nil, fmt.Errorf("API base not configured")
	}

	// Use the model name exactly as configured
	effectiveModel := model

	requestBody := map[string]interface{}{
		"model":    effectiveModel,
		"messages": messages,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	if maxTokens, ok := options["max_tokens"].(int); ok {
		requestBody["max_tokens"] = maxTokens
	}

	if temperature, ok := options["temperature"].(float64); ok {
		requestBody["temperature"] = temperature
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		authHeader := "Bearer " + p.apiKey
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return p.parseResponse(body)
}

func (p *HTTPProvider) parseResponse(body []byte) (*LLMResponse, error) {
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *UsageInfo `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return &LLMResponse{
			Content:      "",
			FinishReason: "stop",
		}, nil
	}

	choice := apiResponse.Choices[0]

	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		arguments := make(map[string]interface{})
		name := ""

		// Handle OpenAI format with nested function object
		if tc.Type == "function" && tc.Function != nil {
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		} else if tc.Function != nil {
			// Legacy format without type field
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		}

		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      name,
			Arguments: arguments,
		})
	}

	return &LLMResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
	}, nil
}

func (p *HTTPProvider) GetDefaultModel() string {
	return ""
}

func findModelConfig(providerName, modelName string, cfg *config.Config) (*config.ModelConfig, error) {
	var provider config.ProviderConfig
	switch strings.ToLower(providerName) {
	case "anthropic":
		provider = cfg.Providers.Anthropic
	case "openai":
		provider = cfg.Providers.OpenAI
	case "openrouter":
		provider = cfg.Providers.OpenRouter
	case "groq":
		provider = cfg.Providers.Groq
	case "zhipu":
		provider = cfg.Providers.Zhipu
	case "vllm":
		provider = cfg.Providers.VLLM
	case "gemini":
		provider = cfg.Providers.Gemini
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	for _, m := range provider.Models {
		if strings.EqualFold(m.Model, modelName) {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("model %s not found in provider %s", modelName, providerName)
}

func createSingleProvider(model string, cfg *config.Config) (LLMProvider, error) {
	// If explicit provider is specified in Agents.Defaults, use it
	if cfg.Agents.Defaults.Provider != "" && strings.EqualFold(cfg.Agents.Defaults.Model, model) {
		mCfg, err := findModelConfig(cfg.Agents.Defaults.Provider, model, cfg)
		if err == nil {
			return NewHTTPProvider(mCfg.APIKey, mCfg.APIBase), nil
		}
	}

	// Logic for finding model in any provider if not explicitly matched above
	providersList := []struct {
		name string
		cfg  config.ProviderConfig
	}{
		{"vllm", cfg.Providers.VLLM},
		{"openai", cfg.Providers.OpenAI},
		{"anthropic", cfg.Providers.Anthropic},
		{"gemini", cfg.Providers.Gemini},
		{"zhipu", cfg.Providers.Zhipu},
		{"groq", cfg.Providers.Groq},
		{"openrouter", cfg.Providers.OpenRouter},
	}

	for _, p := range providersList {
		for _, m := range p.cfg.Models {
			if strings.EqualFold(m.Model, model) {
				return NewHTTPProvider(m.APIKey, m.APIBase), nil
			}
		}
	}

	// Legacy/Prefix-based fallback for backward compatibility or direct model strings
	lowerModel := strings.ToLower(model)
	if strings.HasPrefix(lowerModel, "vllm/") || strings.HasPrefix(lowerModel, "ollama/") || strings.HasPrefix(lowerModel, "local/") {
		// Use first VLLM model if available
		if len(cfg.Providers.VLLM.Models) > 0 {
			m := cfg.Providers.VLLM.Models[0]
			return NewHTTPProvider(m.APIKey, m.APIBase), nil
		}
	}

	return nil, fmt.Errorf("no configuration found for model: %s", model)
}

type fallbackEntry struct {
	provider LLMProvider
	model    string
}

type FallbackProvider struct {
	entries []fallbackEntry
}

func (p *FallbackProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var lastErr error
	for i, entry := range p.entries {
		targetModel := model
		if i > 0 { // For fallback attempts, use the model explicitly associated with that provider
			targetModel = entry.model
		}

		resp, err := entry.provider.Chat(ctx, messages, tools, targetModel, options)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if i < len(p.entries)-1 {
			// Log fallback attempt to console
			fmt.Printf("⚠️ Provider '%s' failed (%v). Falling back to '%s'...\n", targetModel, err, p.entries[i+1].model)
		}
	}
	return nil, fmt.Errorf("all LLM providers failed. Last error: %w", lastErr)
}

func (p *FallbackProvider) GetDefaultModel() string {
	if len(p.entries) > 0 {
		return p.entries[0].model
	}
	return ""
}

func CreateProvider(cfg *config.Config) (LLMProvider, error) {
	primaryModel := cfg.Agents.Defaults.Model
	primaryProviderName := cfg.Agents.Defaults.Provider

	var primaryProvider LLMProvider
	var err error

	if primaryProviderName != "" {
		mCfg, err := findModelConfig(primaryProviderName, primaryModel, cfg)
		if err == nil {
			primaryProvider = NewHTTPProvider(mCfg.APIKey, mCfg.APIBase)
		}
	}

	if primaryProvider == nil {
		primaryProvider, err = createSingleProvider(primaryModel, cfg)
	}

	fallback := &FallbackProvider{entries: make([]fallbackEntry, 0)}

	if err == nil && primaryProvider != nil {
		fallback.entries = append(fallback.entries, fallbackEntry{provider: primaryProvider, model: primaryModel})
	}

	// Use configured fallback models if available
	if len(cfg.Agents.Defaults.FallbackModels) > 0 {
		for _, m := range cfg.Agents.Defaults.FallbackModels {
			// Avoid adding the same model as primary
			if strings.EqualFold(m, primaryModel) {
				continue
			}
			if p, _ := createSingleProvider(m, cfg); p != nil {
				fallback.entries = append(fallback.entries, fallbackEntry{provider: p, model: m})
			}
		}
	}

	if len(fallback.entries) == 0 {
		return nil, fmt.Errorf("no valid AI providers configured")
	}

	return fallback, nil
}
