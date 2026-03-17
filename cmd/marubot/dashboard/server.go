package dashboard

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dirmich/marubot/pkg/agent"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/skills"
)

//go:embed dist
var webAdminAssets embed.FS

// Server handles the MaruBot dashboard web interface and API
type Server struct {
	addr      string
	agent     *agent.AgentLoop
	config    *config.Config
	skillMgr  *skills.SkillInstaller
	skillLoad *skills.SkillsLoader
	version   string
}

// NewServer creates a new dashboard server instance
func NewServer(addr string, agent *agent.AgentLoop, cfg *config.Config, version string) *Server {
	return &Server{
		addr:      addr,
		agent:     agent,
		config:    cfg,
		skillMgr:  skills.NewSkillInstaller(cfg.WorkspacePath()),
		skillLoad: skills.NewSkillsLoader(cfg.WorkspacePath(), ""),
		version:   version,
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
	mux.Handle("/api/skills", s.authMiddleware(http.HandlerFunc(s.handleSkills)))
	s.registerGPIORoutes(mux)
	mux.Handle("/api/logs", s.authMiddleware(http.HandlerFunc(s.handleLogs)))
	mux.Handle("/api/system/stats", s.authMiddleware(http.HandlerFunc(s.handleSystemStats)))
	mux.Handle("/api/upgrade", s.authMiddleware(http.HandlerFunc(s.handleUpgrade)))

	// Static File Serving (SPA Fallback)
	fileServer := http.FileServer(http.FS(distFS))
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Check if file exists in FS
		f, err := distFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			defer f.Close()
			stat, _ := f.Stat()
			if !stat.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA routing
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
			fmt.Printf("Auth failed: session cookie mismatch.\n")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
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

	s.config.AdminPassword = req.Password

	// Save to usersetting.json
	home, _ := os.UserHomeDir()
	userSettingsPath := filepath.Join(home, ".marubot", "usersetting.json")
	
	// Try to read existing setting if any to merge
	var settings map[string]interface{}
	data, err := os.ReadFile(userSettingsPath)
	if err == nil {
		json.Unmarshal(data, &settings)
	} else {
		settings = make(map[string]interface{})
	}
	settings["admin_password"] = req.Password

	newData, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(userSettingsPath, newData, 0644)

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

	if req.Password == s.config.AdminPassword {
		http.SetCookie(w, &http.Cookie{
			Name:     "marubot_session",
			Value:    s.config.AdminPassword,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400 * 30, // 30 days
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	} else {
		fmt.Printf("Login failed: password mismatch. expected=%s, got=%s\n", s.config.AdminPassword, req.Password)
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

		// Save to usersetting.json for persistence
		home, _ := os.UserHomeDir()
		userSettingsPath := filepath.Join(home, ".marubot", "usersetting.json")

		// Use a temporary copy for marshaling to avoid lock contention during I/O
		data, _ := json.MarshalIndent(newCfg, "", "  ")
		if err := os.WriteFile(userSettingsPath, data, 0644); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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

	home, _ := os.UserHomeDir()
	logFile := filepath.Join(home, ".marubot", "dashboard.log")

	data, err := os.ReadFile(logFile)
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
	
	// User config override takes precedence, otherwise use platform detection
	if s.config.Hardware.IsRaspberryPi != nil {
		stats["is_rpi"] = *s.config.Hardware.IsRaspberryPi
	} else {
		stats["is_rpi"] = isRPi()
	}

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
