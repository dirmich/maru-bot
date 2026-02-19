package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/dirmich/maruminibot/pkg/agent"
	"github.com/dirmich/maruminibot/pkg/config"
	"github.com/dirmich/maruminibot/pkg/skills"
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
	// Serve static files from embedded FS (web-admin/dist)

	// dist, _ := fs.Sub(webAdminAssets, "dist")
	// The embed directive includes the directory prefix "dist", so we need to access it properly.
	// However, if we're running from the root of the repo during dev, this might be tricky if build fails.
	// For production binary, "dist" must exist inside web-admin before building marubot.
	// But since the embed directive is in this package, the "dist" folder must be relative to THIS file (cmd/marubot/dashboard).
	// This means we need to COPY web-admin/dist to cmd/marubot/dashboard/dist before building the Go binary.

	// For now, let's assume successful copy.
	distFS, err := fs.Sub(webAdminAssets, "dist")
	if err != nil {
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

		// If API path wasn't caught above (shouldn't happen with exact matches, but for safety)
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
				// Serve existing file
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

		// Copy index.html content to response
		http.ServeContent(w, r, "index.html", s.getFileModTime(index), index)
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

// API Handlers (Simplified for now)

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return chat history (mock)
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

		// TODO: Send to Agent
		// resp, err := s.agent.Chat(req.Message)
		resp := "Echo: " + req.Message // Placeholder

		json.NewEncoder(w).Encode(map[string]string{"response": resp})
	}
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(s.config)
	}
	// POST implementation for saving config...
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		skills := s.skillLoad.ListSkills(false)
		// Convert to simple string output for compatibility with current UI
		var output strings.Builder
		for _, sk := range skills {
			output.WriteString(fmt.Sprintf("- %s (%s)\n", sk.Name, sk.Source))
		}
		json.NewEncoder(w).Encode(map[string]string{"output": output.String()})
	}
}
