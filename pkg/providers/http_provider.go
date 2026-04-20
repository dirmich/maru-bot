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
	"strconv"
	"strings"

	"github.com/dirmich/marubot/pkg/config"
)

type HTTPProvider struct {
	apiKey       string
	apiBase      string
	providerType string
	httpClient   *http.Client
}

func NewHTTPProvider(apiKey, apiBase, providerType string) *HTTPProvider {
	return &HTTPProvider{
		apiKey:       apiKey,
		apiBase:      apiBase,
		providerType: providerType,
		httpClient: &http.Client{
			Timeout: 0,
		},
	}
}

func (p *HTTPProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	if p.apiBase == "" {
		return nil, fmt.Errorf("API base not configured")
	}

	effectiveModel := model
	fmt.Printf("[Debug] Sending request with model: %s (original: %s) to %s\n", effectiveModel, model, p.apiBase)

	requestBody := map[string]interface{}{
		"model":    effectiveModel,
		"messages": messages,
	}

	if len(tools) > 0 {
		effectiveTools := tools
		if p.providerType == "llamacpp" {
			effectiveTools = p.simplifyTools(tools)
		}
		requestBody["tools"] = effectiveTools
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

	if p.providerType == "llamacpp" {
		fmt.Printf("[Debug] Llama.cpp request body: %s\n", string(jsonData))
	}

	url := p.apiBase
	if !strings.HasSuffix(url, "/chat/completions") && !strings.HasSuffix(url, "/completions") {
		// If base ends in /v1, just append /chat/completions
		// If it's an Ollama URL without /v1, append /v1/chat/completions
		url = strings.TrimSuffix(url, "/")
		if strings.Contains(url, "ollama") && !strings.HasSuffix(url, "/v1") && !strings.HasSuffix(url, "/api") {
			url += "/v1"
		}
		url += "/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
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

func splitProviderModelRef(ref string) (string, string, bool) {
	parts := strings.SplitN(ref, "::", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func parseProviderRef(providerName string) (string, *int) {
	lower := strings.ToLower(strings.TrimSpace(providerName))
	if strings.HasPrefix(lower, "ollama#") {
		if idx, err := strconv.Atoi(strings.TrimPrefix(lower, "ollama#")); err == nil {
			return "ollama", &idx
		}
	}
	return lower, nil
}

func providerEnabled(providerName string, cfg *config.Config) bool {
	base, idx := parseProviderRef(providerName)
	switch base {
	case "anthropic":
		return cfg.Providers.Anthropic.Enabled
	case "openai":
		return cfg.Providers.OpenAI.Enabled
	case "openrouter":
		return cfg.Providers.OpenRouter.Enabled
	case "groq":
		return cfg.Providers.Groq.Enabled
	case "zhipu":
		return cfg.Providers.Zhipu.Enabled
	case "vllm":
		return cfg.Providers.VLLM.Enabled
	case "gemini":
		return cfg.Providers.Gemini.Enabled
	case "llamacpp":
		return cfg.Providers.LlamaCPP.Enabled
	case "ollama":
		if idx != nil && *idx >= 0 && *idx < len(cfg.Providers.Ollama) {
			return cfg.Providers.Ollama[*idx].Enabled
		}
		for _, p := range cfg.Providers.Ollama {
			if p.Enabled {
				return true
			}
		}
	}
	return false
}

func findModelConfig(providerName, modelName string, cfg *config.Config) (*config.ModelConfig, error) {
	baseProvider, idx := parseProviderRef(providerName)
	var provider config.ProviderConfig
	switch baseProvider {
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
	case "llamacpp":
		provider = cfg.Providers.LlamaCPP
	case "ollama":
		if idx != nil {
			if *idx < 0 || *idx >= len(cfg.Providers.Ollama) {
				return nil, fmt.Errorf("ollama provider index %d out of range", *idx)
			}
			for _, m := range cfg.Providers.Ollama[*idx].Models {
				if strings.EqualFold(m.Model, modelName) {
					return &m, nil
				}
			}
			return nil, fmt.Errorf("model %s not found in provider %s", modelName, providerName)
		}
		for _, p := range cfg.Providers.Ollama {
			for _, m := range p.Models {
				if strings.EqualFold(m.Model, modelName) {
					return &m, nil
				}
			}
		}
		return nil, fmt.Errorf("model %s not found in any ollama provider", modelName)
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

func createSingleProvider(providerName, model string, cfg *config.Config) (LLMProvider, error) {
	// If explicit provider is specified, use it first.
	if providerName != "" {
		mCfg, err := findModelConfig(providerName, model, cfg)
		if err == nil {
			apiBase := mCfg.APIBase
			if apiBase == "" {
				apiBase = config.GetDefaultBase(parseProviderOnly(providerName))
			}
			baseProvider, _ := parseProviderRef(providerName)
			return NewHTTPProvider(mCfg.APIKey, apiBase, baseProvider), nil
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
		{"llamacpp", cfg.Providers.LlamaCPP},
		{"openrouter", cfg.Providers.OpenRouter},
	}

	for _, p := range providersList {
		if providerName == "" && !p.cfg.Enabled {
			continue
		}
		if providerName != "" && p.name != parseProviderOnly(providerName) {
			continue
		}
		for _, m := range p.cfg.Models {
			if strings.EqualFold(m.Model, model) {
				apiKey := m.APIKey
				apiBase := m.APIBase
				if apiBase == "" {
					apiBase = config.GetDefaultBase(p.name)
				}
				return NewHTTPProvider(apiKey, apiBase, p.name), nil
			}
		}
	}

	// Search in Ollama list
	for i, p := range cfg.Providers.Ollama {
		if providerName == "" && !p.Enabled {
			continue
		}
		if providerName != "" {
			baseProvider, idx := parseProviderRef(providerName)
			if baseProvider != "ollama" {
				continue
			}
			if idx != nil && *idx != i {
				continue
			}
		}
		for _, m := range p.Models {
			if strings.EqualFold(m.Model, model) {
				return NewHTTPProvider(m.APIKey, m.APIBase, "ollama"), nil
			}
		}
	}

	// Legacy/Prefix-based fallback for direct model strings
	lowerModel := strings.ToLower(model)
	if strings.Contains(lowerModel, "gpt-") || strings.Contains(lowerModel, "dall-e") {
		apiKey := ""
		if len(cfg.Providers.OpenAI.Models) > 0 {
			apiKey = cfg.Providers.OpenAI.Models[0].APIKey
		}
		return NewHTTPProvider(apiKey, config.GetDefaultBase("openai"), "openai"), nil
	} else if strings.Contains(lowerModel, "claude-") {
		apiKey := ""
		if len(cfg.Providers.Anthropic.Models) > 0 {
			apiKey = cfg.Providers.Anthropic.Models[0].APIKey
		}
		return NewHTTPProvider(apiKey, config.GetDefaultBase("anthropic"), "anthropic"), nil
	} else if strings.Contains(lowerModel, "gemini-") {
		apiKey := ""
		if len(cfg.Providers.Gemini.Models) > 0 {
			apiKey = cfg.Providers.Gemini.Models[0].APIKey
		}
		return NewHTTPProvider(apiKey, config.GetDefaultBase("gemini"), "gemini"), nil
	}

	return nil, fmt.Errorf("no configuration found for model: %s", model)
}

func parseProviderOnly(providerName string) string {
	baseProvider, _ := parseProviderRef(providerName)
	return baseProvider
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
			apiBase := mCfg.APIBase
			if apiBase == "" {
				apiBase = config.GetDefaultBase(parseProviderOnly(primaryProviderName))
			}
			primaryProvider = NewHTTPProvider(mCfg.APIKey, apiBase, parseProviderOnly(primaryProviderName))
		} else {
			// If provider is explicitly specified but model configuration not found,
			// still attempt to use that provider with its default base.
			apiBase := config.GetDefaultBase(parseProviderOnly(primaryProviderName))
			if apiBase != "" {
				primaryProvider = NewHTTPProvider("", apiBase, parseProviderOnly(primaryProviderName))
			}
		}
	}

	if primaryProvider == nil {
		primaryProvider, err = createSingleProvider(primaryProviderName, primaryModel, cfg)
	}

	fallback := &FallbackProvider{entries: make([]fallbackEntry, 0)}

	if err == nil && primaryProvider != nil {
		fallback.entries = append(fallback.entries, fallbackEntry{provider: primaryProvider, model: primaryModel})
	}

	// Use configured fallback models if available
	if len(cfg.Agents.Defaults.FallbackModels) > 0 {
		for _, entry := range cfg.Agents.Defaults.FallbackModels {
			fallbackProviderName, fallbackModel, ok := splitProviderModelRef(entry)
			if !ok {
				// Backward compatibility for legacy model-only entries.
				fallbackProviderName = ""
				fallbackModel = entry
			}

			if fallbackProviderName != "" && !providerEnabled(fallbackProviderName, cfg) {
				continue
			}

			// Avoid adding the same model as primary
			if strings.EqualFold(fallbackModel, primaryModel) && strings.EqualFold(parseProviderOnly(fallbackProviderName), parseProviderOnly(primaryProviderName)) {
				continue
			}
			if p, _ := createSingleProvider(fallbackProviderName, fallbackModel, cfg); p != nil {
				fallback.entries = append(fallback.entries, fallbackEntry{provider: p, model: fallbackModel})
			}
		}
	}

	if len(fallback.entries) == 0 {
		return nil, fmt.Errorf("no valid AI providers configured")
	}

	return fallback, nil
}

func (p *HTTPProvider) simplifyTools(tools []ToolDefinition) []ToolDefinition {
	simplified := make([]ToolDefinition, len(tools))
	for i, td := range tools {
		simTool := td
		// Create a deep-ish copy of the parameters to avoid modifying the original
		params, _ := json.Marshal(td.Function.Parameters)
		var newParams map[string]interface{}
		json.Unmarshal(params, &newParams)

		// llama.cpp simplification: remove 'required' at the parameters level
		// Some models/versions fail to parse this correctly
		delete(newParams, "required")

		// Ensure all properties have a type and are simple
		if props, ok := newParams["properties"].(map[string]interface{}); ok {
			for name, prop := range props {
				if pMap, ok := prop.(map[string]interface{}); ok {
					// Remove any nested required/additionalProperties
					delete(pMap, "required")
					delete(pMap, "additionalProperties")
					props[name] = pMap
				}
			}
		}

		simTool.Function.Parameters = newParams
		simplified[i] = simTool
	}
	return simplified
}
