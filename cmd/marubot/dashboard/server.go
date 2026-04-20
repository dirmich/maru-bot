package dashboard

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dirmich/marubot/pkg/agent"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/providers"
	"github.com/dirmich/marubot/pkg/skills"
	"github.com/dirmich/marubot/pkg/utils"
)

//go:embed dist
var webAdminAssets embed.FS

// Server handles the MaruBot dashboard web interface and API
type Server struct {
	addr       string
	agent      *agent.AgentLoop
	config     *config.Config
	configPath string
	skillMgr   *skills.SkillInstaller
	skillLoad  *skills.SkillsLoader
	version    string
	onRestart  func()
}

// NewServer creates a new dashboard server instance
func NewServer(addr string, agent *agent.AgentLoop, cfg *config.Config, configPath string, version string, onRestart func()) *Server {
	return &Server{
		addr:       addr,
		agent:      agent,
		config:     cfg,
		configPath: configPath,
		skillMgr:   skills.NewSkillInstaller(cfg.WorkspacePath()),
		skillLoad:  skills.NewSkillsLoader(cfg.WorkspacePath(), ""),
		version:    version,
		onRestart:  onRestart,
	}
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	distFS, err := fs.Sub(webAdminAssets, "dist")
	if err != nil {
		// Should not happen if build environment is correct (folder exists)
		return fmt.Errorf("failed to load embedded assets: %w", err)
	}

	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/config/password", s.handleSetPassword)

	// Protected API Routes
	mux.Handle("/api/chat", s.authMiddleware(http.HandlerFunc(s.handleChat)))
	mux.Handle("/api/config", s.authMiddleware(http.HandlerFunc(s.handleConfig)))
	mux.Handle("/api/config/fetch-models", s.authMiddleware(http.HandlerFunc(s.handleFetchModels)))
	mux.Handle("/api/skills", s.authMiddleware(http.HandlerFunc(s.handleSkills)))
	s.registerGPIORoutes(mux)
	mux.Handle("/api/logs", s.authMiddleware(http.HandlerFunc(s.handleLogs)))
	mux.Handle("/api/system/stats", s.authMiddleware(http.HandlerFunc(s.handleSystemStats)))
	mux.Handle("/api/upgrade", s.authMiddleware(http.HandlerFunc(s.handleUpgrade)))

	// Register manual MIME types for environments without /etc/mime.types (e.g. minimal RPi/Docker)
	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".mjs", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".ico", "image/x-icon")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf", "font/ttf")

	// Static File Serving (SPA Fallback)
	fileServer := http.FileServer(http.FS(distFS))
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Check if file exists in FS
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		f, err := distFS.Open(cleanPath)
		if err == nil {
			defer f.Close()
			stat, _ := f.Stat()
			if !stat.IsDir() {
				// Manually set Content-Type based on extension to avoid system-dependent MIME issues
				ext := filepath.Ext(cleanPath)
				if contentType := mime.TypeByExtension(ext); contentType != "" {
					w.Header().Set("Content-Type", contentType)
				}

				// Add cache control for hashed assets
				if strings.Contains(cleanPath, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}

				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA routing
		// 🛡️ Improved: Don't fallback for static assets (JS, CSS, etc.)
		assetExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".webp"}
		lowerPath := strings.ToLower(path)
		for _, ext := range assetExts {
			if strings.HasSuffix(lowerPath, ext) {
				fmt.Printf("[Dashboard] Asset not found: %s\n", path)
				http.NotFound(w, r)
				return
			}
		}

		index, err := distFS.Open("index.html")
		if err != nil {
			http.Error(w, "Dashboard not found (build missing)", http.StatusNotFound)
			return
		}
		defer index.Close()

		// Use io.Copy since fs.File from embed doesn't implement io.ReadSeeker required for ServeContent
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := io.Copy(w, index); err != nil {
			http.Error(w, "Failed to serve index", http.StatusInternalServerError)
		}
	})
	mux.Handle("/", staticHandler)

	// Wrap everything in recovery middleware
	recoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("CRITICAL: Server Panic recovered: %v\n", err)
				http.Error(w, "Internal Server Error (Panic)", http.StatusInternalServerError)
			}
		}()
		mux.ServeHTTP(w, r)
	})

	fmt.Printf("Dashboard server listening on http://%s\n", s.addr)
	return http.ListenAndServe(s.addr, recoveryHandler)
}

func (s *Server) getFileModTime(f fs.File) time.Time {
	stat, err := f.Stat()
	if err != nil {
		return time.Now()
	}
	return stat.ModTime()
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If password is not set, block everything EXCEPT /api/config/password
		if s.config.AdminPassword == "" {
			if r.URL.Path == "/api/config/password" {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Initial setup required: Password not set", http.StatusForbidden)
			return
		}

		cookie, err := r.Cookie("marubot_session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if cookie.Value != s.config.AdminPassword {
			// Compare with hashed password for extra safety if cookie value is plaintext (legacy)
			if utils.HashPassword(cookie.Value) != s.config.AdminPassword {
				fmt.Printf("Auth failed: session cookie mismatch.\n")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleSetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only allow if password is NOT set yet
	if s.config.AdminPassword != "" {
		http.Error(w, "Password already set", http.StatusForbidden)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 4 {
		http.Error(w, "Password too short", http.StatusBadRequest)
		return
	}

	s.config.AdminPassword = utils.HashPassword(req.Password)

	// Save updated config to config.json
	if err := config.SaveConfig(s.configPath, s.config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Password set successfully"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedInput := utils.HashPassword(req.Password)
	if hashedInput == s.config.AdminPassword || req.Password == s.config.AdminPassword {
		// Use hash for cookie value
		cookieValue := hashedInput
		if hashedInput != s.config.AdminPassword {
			// If it matched plaintext, use the hashed version for future
			cookieValue = s.config.AdminPassword
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "marubot_session",
			Value:    cookieValue,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400 * 30, // 30 days
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	} else {
		fmt.Printf("Login failed: password mismatch.\n")
		http.Error(w, "Invalid password", http.StatusUnauthorized)
	}
}

// API Handlers

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		json.NewEncoder(w).Encode([]map[string]string{})
		return
	}

	if r.Method == "POST" {
		var req struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("[Debug] Received chat request: %s\n", req.Message)

		if s.agent == nil {
			fmt.Println("[Error] Agent is nil in Server")
			http.Error(w, "Agent not initialized", http.StatusInternalServerError)
			return
		}

		fmt.Println("[Debug] Calling agent.ProcessDirect...")
		resp, err := s.agent.ProcessDirect(r.Context(), req.Message, "web-admin")
		if err != nil {
			fmt.Printf("[Error] Agent processing failed: %v\n", err)
			http.Error(w, fmt.Sprintf("AI processing error: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Println("[Debug] Chat successful, sending response.")
		json.NewEncoder(w).Encode(map[string]string{"response": resp})
	}
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		json.NewEncoder(w).Encode(s.config)
		return
	}

	if r.Method == "POST" {
		newCfg := &config.Config{}
		if err := json.NewDecoder(r.Body).Decode(newCfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update In-memory config fields selectively to avoid copying mutex
		s.config.Update(newCfg)

		// Proactively update AgentLoop provider if it's already running
		// This provides immediate "hot-reload" effect even before the full process restart
		if s.agent != nil {
			if p, err := providers.CreateProvider(s.config); err == nil {
				s.agent.SetProvider(p)
			}
		}

		// Save to config.json for persistence
		if err := config.SaveConfig(s.configPath, s.config); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

		if s.onRestart != nil {
			go func() {
				time.Sleep(800 * time.Millisecond) // Slightly longer sleep to ensure Save is flushed
				s.onRestart()
			}()
		}
	}
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		skList := s.skillLoad.ListSkills(false)
		var output strings.Builder
		output.WriteString("INSTALLED SKILLS:\n-----------------\n")
		for _, sk := range skList {
			status := "✓"
			if !sk.Available {
				status = "✗"
			}
			output.WriteString(fmt.Sprintf("%s %s (%s)\n", status, sk.Name, sk.Source))
			if sk.Description != "" {
				output.WriteString(fmt.Sprintf("  %s\n", sk.Description))
			}
		}
		json.NewEncoder(w).Encode(map[string]string{"output": output.String()})
		return
	}

	if r.Method == "POST" {
		var req struct {
			Action string `json:"action"` // "install", "remove"
			Skill  string `json:"skill"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Action == "install" {
			ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
			defer cancel()
			if err := s.skillMgr.InstallFromGitHub(ctx, req.Skill); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if req.Action == "remove" {
			if err := s.skillMgr.Uninstall(req.Skill); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	home := os.Getenv("MARUBOT_HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	logDir := filepath.Join(home, ".marubot", "logs")
	todayLog := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")

	data, err := os.ReadFile(todayLog)
	if err != nil {
		// Fallback: search for latest log in the directory
		files, readErr := os.ReadDir(logDir)
		if readErr == nil && len(files) > 0 {
			var latestFile string
			for i := len(files) - 1; i >= 0; i-- {
				if !files[i].IsDir() && strings.HasSuffix(files[i].Name(), ".log") {
					latestFile = filepath.Join(logDir, files[i].Name())
					break
				}
			}
			if latestFile != "" {
				data, err = os.ReadFile(latestFile)
			}
		}
	}

	if err != nil {
		if os.IsNotExist(err) {
			json.NewEncoder(w).Encode(map[string]string{"logs": "No logs found yet."})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return last 16000 characters
	logs := string(data)
	if len(logs) > 16000 {
		logs = logs[len(logs)-16000:]
	}

	json.NewEncoder(w).Encode(map[string]string{"logs": logs})
}

func (s *Server) handleSystemStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := getPlatformStats()
	stats["version"] = s.version

	latest, err := config.CheckLatestVersion()
	if err == nil {
		stats["latest_version"] = latest
		stats["is_update_available"] = config.IsNewVersionAvailable(latest)
	} else {
		stats["latest_version"] = ""
		stats["is_update_available"] = false
	}

	stats["is_ai_configured"] = s.config.IsAIConfigured()
	stats["is_channel_configured"] = s.config.IsChannelEnabled()

	// User config override takes precedence, otherwise use platform detection (considering test mode)
	if s.config.Hardware.IsRaspberryPi != nil {
		stats["is_raspberry_pi"] = *s.config.Hardware.IsRaspberryPi
	} else {
		stats["is_raspberry_pi"] = s.isRPi()
	}
	stats["is_rpi"] = stats["is_raspberry_pi"] // Maintain both for compatibility

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Trigger upgrade in background
	go func() {
		// Wait 1 second to let response finish
		time.Sleep(1 * time.Second)

		exe, _ := os.Executable()
		// Use --yes for non-interactive upgrade
		cmd := exec.Command(exe, "upgrade", "--yes")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Dashboard-triggered upgrade failed: %v\n", err)
			return
		}
	}()

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Upgrade started. The system will restart automatically.",
	})
}
func (s *Server) handleFetchModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		APIBase  string `json:"api_base"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	provider := strings.ToLower(req.Provider)
	if req.APIKey == "" && provider != "ollama" && provider != "llamacpp" {
		http.Error(w, "API Key is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var models []string
	var err error

	switch provider {
	case "openai", "groq", "openrouter", "vllm", "llamacpp":
		models, err = s.fetchOpenAIModels(ctx, req.APIKey, req.APIBase, provider)
	case "gemini":
		models, err = s.fetchGeminiModels(ctx, req.APIKey)
	case "anthropic":
		models, err = s.fetchAnthropicModels(ctx, req.APIKey)
	case "ollama":
		models, err = s.fetchOllamaModels(ctx, req.APIBase)
	default:
		err = fmt.Errorf("provider %s not supported for model fetching", req.Provider)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch models: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"models": models,
	})
}

func (s *Server) fetchOpenAIModels(ctx context.Context, apiKey, apiBase, provider string) ([]string, error) {
	baseUrl := apiBase
	if baseUrl == "" {
		switch provider {
		case "openai":
			baseUrl = "https://api.openai.com/v1"
		case "groq":
			baseUrl = "https://api.groq.com/openai/v1"
		case "openrouter":
			baseUrl = "https://openrouter.ai/api/v1"
		case "llamacpp":
			baseUrl = "http://localhost:8080/v1"
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/models", nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range data.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

func (s *Server) fetchGeminiModels(ctx context.Context, apiKey string) ([]string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range data.Models {
		// Name usually starts with "models/", strip it
		name := strings.TrimPrefix(m.Name, "models/")
		// Only include generative models
		if strings.Contains(name, "gemini") {
			models = append(models, name)
		}
	}
	return models, nil
}

func (s *Server) fetchAnthropicModels(ctx context.Context, apiKey string) ([]string, error) {
	// Anthropic's models list API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range data.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

func (s *Server) fetchOllamaModels(ctx context.Context, apiBase string) ([]string, error) {
	if apiBase == "" {
		apiBase = "http://localhost:11434"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	req, err := http.NewRequestWithContext(ctx, "GET", apiBase+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama error (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range data.Models {
		models = append(models, m.Name)
	}
	return models, nil
}
