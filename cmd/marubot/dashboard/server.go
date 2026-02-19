package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
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
}

// NewServer creates a new dashboard server instance
func NewServer(addr string, agent *agent.AgentLoop, cfg *config.Config) *Server {
	return &Server{
		addr:      addr,
		agent:     agent,
		config:    cfg,
		skillMgr:  skills.NewSkillInstaller(cfg.WorkspacePath()),
		skillLoad: skills.NewSkillsLoader(cfg.WorkspacePath(), ""),
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

	// API Routes
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/skills", s.handleSkills)

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

// API Handlers (Simplified)

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

		// Mock response
		resp := "Echo: " + req.Message

		json.NewEncoder(w).Encode(map[string]string{"response": resp})
	}
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(s.config)
	}
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		skills := s.skillLoad.ListSkills(false)
		var output strings.Builder
		for _, sk := range skills {
			output.WriteString(fmt.Sprintf("- %s (%s)\n", sk.Name, sk.Source))
		}
		json.NewEncoder(w).Encode(map[string]string{"output": output.String()})
	}
}
