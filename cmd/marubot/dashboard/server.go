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

	// Protected API Routes
	mux.Handle("/api/chat", s.authMiddleware(http.HandlerFunc(s.handleChat)))
	mux.Handle("/api/config", s.authMiddleware(http.HandlerFunc(s.handleConfig)))
	mux.Handle("/api/skills", s.authMiddleware(http.HandlerFunc(s.handleSkills)))
	mux.Handle("/api/gpio", s.authMiddleware(http.HandlerFunc(s.handleGpio)))
	mux.Handle("/api/logs", s.authMiddleware(http.HandlerFunc(s.handleLogs)))
	mux.Handle("/api/system/stats", s.authMiddleware(http.HandlerFunc(s.handleSystemStats)))

	// Static File Serving (SPA Fallback)
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	fmt.Printf("Dashboard server listening on http://%s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
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
		cookie, err := r.Cookie("marubot_session")
		if err != nil || cookie.Value != s.config.AdminPassword {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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

		resp, err := s.agent.ProcessDirect(r.Context(), req.Message, "web-admin")
		if err != nil {
			http.Error(w, fmt.Sprintf("AI processing error: %v", err), http.StatusInternalServerError)
			return
		}

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

func (s *Server) handleGpio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		json.NewEncoder(w).Encode(s.config.Hardware.GPIO.Pins)
		return
	}

	if r.Method == "POST" {
		var pins map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&pins); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.config.Hardware.GPIO.Pins = pins

		// Save config
		home, _ := os.UserHomeDir()
		userSettingsPath := filepath.Join(home, ".marubot", "usersetting.json")
		data, _ := json.MarshalIndent(s.config, "", "  ")
		os.WriteFile(userSettingsPath, data, 0644)

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

	json.NewEncoder(w).Encode(stats)
}
