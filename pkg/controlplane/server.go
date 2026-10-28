package controlplane

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dirmich/marubot/pkg/orchestration"
)

type Options struct {
	Root               string
	PublicURL          string
	GoogleClientID     string
	GoogleClientSecret string
	CAFile             string
	Parallelism        int
	Assets             fs.FS
	Plan               func(context.Context, string, []orchestration.Node) (orchestration.Run, error)
}
type Server struct {
	jwksMu      sync.Mutex
	jwks        map[string]*rsa.PublicKey
	jwksExpires time.Time
	opts        Options
	auth        *authStore
	client      *http.Client
	oauthClient *http.Client
	mu          sync.Mutex
	engines     map[string]*orchestration.Engine
	ctx         context.Context
	cancel      context.CancelFunc
}

func New(opts Options) (*Server, error) {
	u, err := url.Parse(opts.PublicURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || u.Scheme != "https" && (u.Scheme != "http" || u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" && u.Hostname() != "::1") {
		return nil, errors.New("Admin 공개 주소는 HTTPS 또는 loopback HTTP여야 합니다")
	}
	a, err := openAuth(opts.Root)
	if err != nil {
		return nil, err
	}
	client, err := orchestration.NewPrivateClient(opts.CAFile)
	if err != nil {
		a.db.Close()
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{opts: opts, auth: a, client: client, oauthClient: &http.Client{Timeout: 15 * time.Second}, engines: map[string]*orchestration.Engine{}, ctx: ctx, cancel: cancel}
	rows, err := a.db.Query(`SELECT DISTINCT owner FROM credentials`)
	if err != nil {
		s.Close()
		return nil, err
	}
	owners := []string{}
	for rows.Next() {
		var owner string
		if err = rows.Scan(&owner); err != nil {
			rows.Close()
			s.Close()
			return nil, err
		}
		owners = append(owners, owner)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		s.Close()
		return nil, err
	}
	for _, owner := range owners {
		if _, err = s.engine(owner); err != nil {
			s.Close()
			return nil, err
		}
	}
	return s, nil
}
func (s *Server) Client() *http.Client                     { return s.client }
func (s *Server) Token(owner, node string) (string, error) { return s.auth.Token(owner, node) }
func (s *Server) Close() error {
	s.cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.engines {
		e.Wait()
		e.Store.Close()
	}
	return s.auth.db.Close()
}
func (s *Server) engine(owner string) (*orchestration.Engine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.engines[owner]; ok {
		return e, nil
	}
	store, err := orchestration.OpenStore(s.opts.Root, owner)
	if err != nil {
		return nil, err
	}
	engine, err := orchestration.NewEngine(store, orchestration.HTTPExecutor{Transport: s}, s.opts.Parallelism)
	if err != nil {
		store.Close()
		return nil, err
	}
	s.engines[owner] = engine
	go engine.Serve(s.ctx)
	return engine, nil
}
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func apiError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{"error": map[string]interface{}{"code": code, "message": message, "retryable": false}})
}
func decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 262144)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return errors.New("JSON 필드와 요청 크기를 확인하세요")
	}
	var extra interface{}
	if d.Decode(&extra) != io.EOF {
		return errors.New("하나의 JSON 객체만 허용됩니다")
	}
	return nil
}
func handleErr(w http.ResponseWriter, err error) {
	if errors.Is(err, orchestration.ErrNotFound) {
		apiError(w, 404, "not_found", "항목을 찾을 수 없습니다")
	} else if errors.Is(err, orchestration.ErrConflict) {
		apiError(w, 409, "conflict", err.Error())
	} else {
		apiError(w, 400, "invalid_request", err.Error())
	}
}

type principalKey struct{}

func principal(r *http.Request) Principal { return r.Context().Value(principalKey{}).(Principal) }
func (s *Server) protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := s.auth.principal(r)
		if err != nil {
			apiError(w, 401, "unauthorized", "로그인이 필요합니다")
			return
		}
		if p.Role == "human" && r.Method != "GET" {
			if r.Header.Get("Origin") != s.opts.PublicURL || r.Header.Get("X-CSRF-Token") != p.CSRF {
				apiError(w, 403, "csrf", "화면을 새로고침한 후 다시 시도하세요")
				return
			}
		}
		if p.Role != "human" {
			path := r.URL.Path
			allowed := r.Method == "POST" && (path == "/api/nodes/heartbeat" || path == "/api/memory-proposals") || r.Method == "GET" && (path == "/api/memories" || path == "/api/me")
			if p.Role == "orchestrator" {
				allowed = allowed || r.Method == "GET" && (path == "/api/nodes" || strings.HasPrefix(path, "/api/runs")) || r.Method == "POST" && (path == "/api/runs" || path == "/api/plan" || strings.HasSuffix(path, "/cancel"))
			}
			if !allowed {
				apiError(w, 403, "forbidden", "이 장비에는 해당 권한이 없습니다")
				return
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, p)))
	})
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth/google/login", s.login)
	mux.HandleFunc("GET /auth/google/callback", s.callback)
	mux.HandleFunc("GET /api/auth/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]interface{}{"configured": s.opts.GoogleClientID != "" && s.opts.GoogleClientSecret != ""})
	})
	api := http.NewServeMux()
	api.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, principal(r)) })
	api.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("maru_control_session"); err == nil {
			s.auth.db.Exec(`DELETE FROM sessions WHERE hash=?`, digest(c.Value))
		}
		s.cookie(w, "maru_control_session", "", -1)
		writeJSON(w, 200, map[string]bool{"ok": true})
	})
	api.HandleFunc("GET /api/overview", s.overview)
	api.HandleFunc("GET /api/nodes", s.nodes)
	api.HandleFunc("POST /api/nodes", s.createNode)
	api.HandleFunc("POST /api/nodes/heartbeat", s.heartbeat)
	api.HandleFunc("POST /api/nodes/{id}/revoke", s.revokeNode)
	api.HandleFunc("POST /api/nodes/{id}/rotate", s.rotateNode)
	api.HandleFunc("GET /api/runs", s.runs)
	api.HandleFunc("POST /api/runs", s.submit)
	api.HandleFunc("POST /api/plan", s.plan)
	api.HandleFunc("GET /api/runs/{id}", s.run)
	api.HandleFunc("POST /api/runs/{id}/cancel", s.cancelRun)
	api.HandleFunc("POST /api/runs/{id}/reconcile", func(w http.ResponseWriter, r *http.Request) {
		e := s.getEngine(w, r)
		if e == nil {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		if err := e.Reconcile(ctx, r.PathValue("id")); err != nil {
			handleErr(w, err)
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	})
	api.HandleFunc("POST /api/runs/{id}/tasks/{task}/approval", s.approval)
	api.HandleFunc("GET /api/memories", s.memories)
	api.HandleFunc("POST /api/memories", s.saveMemory)
	api.HandleFunc("POST /api/memory-proposals", s.proposeMemory)
	api.HandleFunc("DELETE /api/memories/{id}", s.deleteMemory)
	api.HandleFunc("POST /api/memories/{id}/promote", s.promoteMemory)
	api.HandleFunc("GET /api/memories/{id}/history", s.memoryHistory)
	api.HandleFunc("GET /api/events", s.events)
	api.HandleFunc("GET /api/runs/{id}/events", s.events)
	mux.Handle("/api/", s.protected(api))
	if s.opts.Assets != nil {
		mux.Handle("/", http.FileServer(http.FS(s.opts.Assets)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "MaruBot Admin") })
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) proposeMemory(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	if p.Node == "" {
		apiError(w, 403, "node_required", "장비 인증이 필요합니다")
		return
	}
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Source  string `json:"source"`
	}
	if err := decode(w, r, &input); err != nil {
		handleErr(w, err)
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	m, err := e.Store.SaveMemory(orchestration.Memory{Title: input.Title, Content: input.Content, Source: input.Source, Kind: "long", Visibility: "private", NodeID: p.Node})
	if err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "memory.proposed", m.ID+" @ "+p.Node)
	writeJSON(w, 201, m)
}
func (s *Server) getEngine(w http.ResponseWriter, r *http.Request) *orchestration.Engine {
	e, err := s.engine(principal(r).Owner)
	if err != nil {
		apiError(w, 500, "storage", "사용자 저장소를 열 수 없습니다")
		return nil
	}
	return e
}
func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	nodes, err := e.Nodes()
	if err != nil {
		handleErr(w, err)
		return
	}
	runs, err := e.Runs()
	if err != nil {
		handleErr(w, err)
		return
	}
	memories, err := e.Store.Memories("", "", "", false)
	if err != nil {
		handleErr(w, err)
		return
	}
	online, running, pending := 0, 0, 0
	for _, n := range nodes {
		if n.Online {
			online++
		}
	}
	for _, run := range runs {
		for _, t := range run.Tasks {
			if t.Status == "running" || t.Status == "cancelling" {
				running++
			}
			if t.Status == "need-approval" {
				pending++
			}
		}
	}
	writeJSON(w, 200, map[string]interface{}{"nodes": nodes, "runs": runs, "online": online, "running": running, "pending": pending, "memory_count": len(memories), "updated_at": time.Now().UTC()})
}
func (s *Server) nodes(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	items, err := e.Nodes()
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 200, items)
}
func (s *Server) createNode(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Role string `json:"role"`
	}
	if err := decode(w, r, &input); err != nil {
		handleErr(w, err)
		return
	}
	if input.Name == "" || len(input.Name) > 200 {
		handleErr(w, errors.New("장비 이름을 입력하세요"))
		return
	}
	if input.Role == "" {
		input.Role = "worker"
	}
	if !orchestration.Contains([]string{"worker", "orchestrator"}, input.Role) {
		handleErr(w, errors.New("장비 역할을 확인하세요"))
		return
	}
	if err := orchestration.ValidateNodeURL(input.URL); err != nil {
		handleErr(w, err)
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	n := orchestration.Node{ID: orchestration.NewID(), Name: input.Name, URL: strings.TrimRight(input.URL, "/")}
	if err := e.SaveNode(n); err != nil {
		handleErr(w, err)
		return
	}
	token, err := s.auth.issue(principal(r).Owner, n.ID, input.Role)
	if err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "node.enrolled", n.ID)
	writeJSON(w, 201, map[string]interface{}{"node": n, "token": token, "owner_id": principal(r).Owner, "expires_at": time.Now().Add(90 * 24 * time.Hour)})
}
func (s *Server) heartbeat(w http.ResponseWriter, r *http.Request) {
	p := principal(r)
	if p.Node == "" {
		apiError(w, 403, "node_required", "장비 인증이 필요합니다")
		return
	}
	var n orchestration.Node
	if err := decode(w, r, &n); err != nil {
		handleErr(w, err)
		return
	}
	if n.ID != p.Node {
		apiError(w, 403, "node_mismatch", "다른 장비의 상태는 변경할 수 없습니다")
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	if err := e.Heartbeat(n); err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) revokeNode(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	var n orchestration.Node
	if err := e.Store.Get("node", r.PathValue("id"), &n); err != nil {
		handleErr(w, err)
		return
	}
	n.Disabled = true
	if err := e.SaveNode(n); err != nil {
		handleErr(w, err)
		return
	}
	if _, err := s.auth.db.Exec(`DELETE FROM credentials WHERE owner=? AND node=?`, principal(r).Owner, n.ID); err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "node.revoked", n.ID)
	writeJSON(w, 200, n)
}
func (s *Server) rotateNode(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	var n orchestration.Node
	if err := e.Store.Get("node", r.PathValue("id"), &n); err != nil {
		handleErr(w, err)
		return
	}
	var role string
	if err := s.auth.db.QueryRow(`SELECT role FROM credentials WHERE owner=? AND node=?`, principal(r).Owner, n.ID).Scan(&role); err != nil {
		handleErr(w, orchestration.ErrNotFound)
		return
	}
	token, err := s.auth.issue(principal(r).Owner, n.ID, role)
	if err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "node.rotated", n.ID)
	writeJSON(w, 200, map[string]interface{}{"token": token, "expires_at": time.Now().Add(90 * 24 * time.Hour)})
}
func (s *Server) runs(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	items, err := e.Runs()
	if err != nil {
		handleErr(w, err)
		return
	}
	if items == nil {
		items = []orchestration.Run{}
	}
	writeJSON(w, 200, items)
}
func (s *Server) submit(w http.ResponseWriter, r *http.Request) {
	var input orchestration.Run
	if err := decode(w, r, &input); err != nil {
		handleErr(w, err)
		return
	}
	if key := r.Header.Get("Idempotency-Key"); key != "" && key != input.IdempotencyKey {
		handleErr(w, errors.New("멱등 키 헤더와 본문이 다릅니다"))
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	run, err := e.Submit(input)
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 202, run)
}
func (s *Server) run(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	run, err := e.Run(r.PathValue("id"))
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 200, run)
}
func (s *Server) cancelRun(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	if err := e.Cancel(r.PathValue("id")); err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 202, map[string]bool{"ok": true})
}
func (s *Server) approval(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Allow bool `json:"allow"`
	}
	if err := decode(w, r, &input); err != nil {
		handleErr(w, err)
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	if err := e.Approve(r.PathValue("id"), r.PathValue("task"), input.Allow); err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) plan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Instruction string `json:"instruction"`
	}
	if err := decode(w, r, &input); err != nil {
		handleErr(w, err)
		return
	}
	if len(input.Instruction) == 0 || len(input.Instruction) > 16000 {
		handleErr(w, errors.New("명령은 1~16000바이트로 입력하세요"))
		return
	}
	if s.opts.Plan == nil {
		apiError(w, 503, "planner_unconfigured", "backend LLM 설정이 필요합니다. 구조화된 작업 입력은 사용할 수 있습니다")
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	nodes, _ := e.Nodes()
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	run, err := s.opts.Plan(ctx, input.Instruction, nodes)
	if err != nil {
		apiError(w, 502, "plan_failed", "작업 계획 생성 실패. LLM 설정과 응답을 확인하세요")
		return
	}
	run.ID = ""
	run.Status = ""
	run.IdempotencyKey = orchestration.NewID()
	run.Deadline = time.Now().Add(time.Hour)
	writeJSON(w, 200, run)
}
func (s *Server) memories(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	p := principal(r)
	items, err := e.Store.Memories(r.URL.Query().Get("q"), p.Node, "", p.Role != "human")
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 200, items)
}
func (s *Server) saveMemory(w http.ResponseWriter, r *http.Request) {
	var m orchestration.Memory
	if err := decode(w, r, &m); err != nil {
		handleErr(w, err)
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	m, err := e.Store.SaveMemory(m)
	if err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "memory.saved", m.ID)
	writeJSON(w, 200, m)
}
func (s *Server) deleteMemory(w http.ResponseWriter, r *http.Request) {
	rev, err := strconv.Atoi(r.URL.Query().Get("revision"))
	if err != nil {
		handleErr(w, errors.New("revision이 필요합니다"))
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	if err = e.Store.DeleteMemory(r.PathValue("id"), rev); err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "memory.deleted", r.PathValue("id"))
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) promoteMemory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Revision int `json:"revision"`
	}
	if err := decode(w, r, &input); err != nil {
		handleErr(w, err)
		return
	}
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	m, err := e.Store.Promote(r.PathValue("id"), input.Revision)
	if err != nil {
		handleErr(w, err)
		return
	}
	e.Store.Event("", "", "memory.promoted", m.ID)
	writeJSON(w, 200, m)
}
func (s *Server) memoryHistory(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	items, err := e.Store.MemoryHistory(r.PathValue("id"))
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, 200, items)
}
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	e := s.getEngine(w, r)
	if e == nil {
		return
	}
	runID := r.PathValue("id")
	if runID != "" {
		if _, err := e.Run(runID); err != nil {
			handleErr(w, err)
			return
		}
	}
	after, _ := strconv.ParseInt(r.Header.Get("Last-Event-ID"), 10, 64)
	if q := r.URL.Query().Get("after"); q != "" {
		after, _ = strconv.ParseInt(q, 10, 64)
	}
	if !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		items, err := e.Store.Events(after, runID)
		if err != nil {
			handleErr(w, err)
			return
		}
		writeJSON(w, 200, items)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		apiError(w, 500, "stream_unavailable", "스트리밍을 사용할 수 없습니다")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	fmt.Fprint(w, "retry: 3000\n\n")
	flusher.Flush()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lifetime := time.NewTimer(5 * time.Minute)
	defer lifetime.Stop()
	for {
		items, err := e.Store.Events(after, runID)
		if err != nil {
			return
		}
		for _, item := range items {
			b, _ := json.Marshal(item)
			if _, err = fmt.Fprintf(w, "id: %d\nevent: update\ndata: %s\n\n", item.ID, b); err != nil {
				return
			}
			after = item.ID
		}
		fmt.Fprint(w, ": keepalive\n\n")
		flusher.Flush()
		select {
		case <-r.Context().Done():
			return
		case <-lifetime.C:
			return
		case <-ticker.C:
		}
	}
}
