package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dirmich/marubot/pkg/orchestration"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(Options{Root: t.TempDir(), PublicURL: "http://localhost:8000"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

type userSession struct{ token, csrf string }

func user(t *testing.T, s *Server, owner string) userSession {
	t.Helper()
	token, err := s.auth.session(Principal{Owner: owner, Name: owner})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/api/me", nil)
	r.AddCookie(&http.Cookie{Name: "maru_control_session", Value: token})
	p, err := s.auth.principal(r)
	if err != nil {
		t.Fatal(err)
	}
	return userSession{token, p.CSRF}
}
func request(t *testing.T, s *Server, u userSession, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	r := httptest.NewRequest(method, path, reader)
	if u.token != "" {
		r.AddCookie(&http.Cookie{Name: "maru_control_session", Value: u.token})
	}
	r.Header.Set("Origin", s.opts.PublicURL)
	r.Header.Set("X-CSRF-Token", u.csrf)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func bearer(t *testing.T, s *Server, token, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	r := httptest.NewRequest(method, path, reader)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func TestUserIsolationCSRFAndWorkerPermissions(t *testing.T) {
	s := testServer(t)
	a := user(t, s, "google-a")
	b := user(t, s, "google-b")
	w := request(t, s, a, "POST", "/api/memories", orchestration.Memory{Title: "사용자 A", Content: "전용 데이터", Kind: "wiki", Visibility: "shared", Verified: true})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = request(t, s, b, "GET", "/api/memories", nil)
	if strings.Contains(w.Body.String(), "전용 데이터") {
		t.Fatal("다른 사용자 기억 노출")
	}
	w = request(t, s, userSession{token: a.token}, "POST", "/api/memories", nil)
	if w.Code != 403 {
		t.Fatal("CSRF 누락")
	}
	w = request(t, s, userSession{}, "GET", "/api/overview", nil)
	if w.Code != 401 {
		t.Fatal("미인증 허용")
	}
	w = request(t, s, a, "POST", "/api/nodes", map[string]string{"name": "A 장비", "url": "https://100.100.0.1:8443"})
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var provision struct {
		Node  orchestration.Node `json:"node"`
		Token string             `json:"token"`
	}
	json.Unmarshal(w.Body.Bytes(), &provision)
	w = bearer(t, s, provision.Token, "POST", "/api/runs", nil)
	if w.Code != 403 {
		t.Fatal("worker가 작업 제출")
	}
	w = bearer(t, s, provision.Token, "GET", "/api/memories", nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "전용 데이터") {
		t.Fatal("같은 소유자 공유 기억 조회 실패")
	}
	w = request(t, s, b, "POST", "/api/nodes/"+provision.Node.ID+"/revoke", nil)
	if w.Code != 404 {
		t.Fatal("타 사용자 장비 폐기 허용")
	}
	w = request(t, s, a, "POST", "/api/nodes/"+provision.Node.ID+"/revoke", nil)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if bearer(t, s, provision.Token, "GET", "/api/memories", nil).Code != 401 {
		t.Fatal("폐기 토큰 허용")
	}
}
func TestTLSWorkersDAGEventsAndCrossOwnerRuns(t *testing.T) {
	s := testServer(t)
	a := user(t, s, "a")
	b := user(t, s, "b")
	var started atomic.Int32
	gate := make(chan struct{})
	workers := []*orchestration.Worker{}
	nodes := []orchestration.Node{}
	for i := 0; i < 2; i++ {
		var worker *orchestration.Worker
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if worker == nil {
				http.Error(w, "not ready", 503)
				return
			}
			worker.Handler().ServeHTTP(w, r)
		}))
		defer ts.Close()
		ca := filepath.Join(t.TempDir(), "ca.pem")
		os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ts.Certificate().Raw}), 0600)
		client, err := orchestration.NewPrivateClient(ca)
		if err != nil {
			t.Fatal(err)
		}
		s.client = client
		w := request(t, s, a, "POST", "/api/nodes", map[string]string{"name": fmt.Sprintf("worker-%d", i), "url": ts.URL})
		if w.Code != 201 {
			t.Fatal(w.Body.String())
		}
		var p struct {
			Node  orchestration.Node `json:"node"`
			Token string             `json:"token"`
		}
		json.Unmarshal(w.Body.Bytes(), &p)
		store, err := orchestration.OpenStore(t.TempDir(), "a")
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()
		worker, err = orchestration.NewWorker(store, p.Node.ID, p.Token, []string{"system_status"}, func(ctx context.Context, task orchestration.Task) (string, error) {
			started.Add(1)
			select {
			case <-gate:
				return "검증된 결과", nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		workers = append(workers, worker)
		n := p.Node
		n.Protocol = orchestration.Protocol
		n.Tools = []string{"system_status"}
		if w = bearer(t, s, p.Token, "POST", "/api/nodes/heartbeat", n); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		nodes = append(nodes, n)
	}
	input := orchestration.Run{Title: "두 장비", IdempotencyKey: "tls-parallel", Tasks: []orchestration.Task{{ID: "one", NodeID: nodes[0].ID, Tool: "system_status", ApprovalLevel: "read-only"}, {ID: "two", NodeID: nodes[1].ID, Tool: "system_status", ApprovalLevel: "read-only"}}}
	w := request(t, s, a, "POST", "/api/runs", input)
	if w.Code != 202 {
		t.Fatal(w.Body.String())
	}
	var run orchestration.Run
	json.Unmarshal(w.Body.Bytes(), &run)
	end := time.Now().Add(5 * time.Second)
	for started.Load() < 2 && time.Now().Before(end) {
		time.Sleep(10 * time.Millisecond)
	}
	if started.Load() != 2 {
		t.Fatalf("TLS 병렬 실행 실패: %d", started.Load())
	}
	close(gate)
	e, _ := s.engine("a")
	for time.Now().Before(end) {
		run, _ = e.Run(run.ID)
		if run.Status == "succeeded" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if run.Status != "succeeded" {
		t.Fatalf("결과 %+v", run)
	}
	if request(t, s, b, "GET", "/api/runs/"+run.ID, nil).Code != 404 {
		t.Fatal("다른 사용자 작업 조회 허용")
	}
	if request(t, s, b, "GET", "/api/runs/"+run.ID+"/events", nil).Code != 404 {
		t.Fatal("다른 사용자 이벤트 조회 허용")
	}
	events, err := e.Store.Events(0, run.ID)
	if err != nil || len(events) < 3 {
		t.Fatal("영속 이벤트 누락")
	}
	tail, _ := e.Store.Events(events[len(events)-2].ID, run.ID)
	if len(tail) != 1 {
		t.Fatal("이벤트 cursor 재전송 실패")
	}
	for _, worker := range workers {
		_ = worker
	}
}
func TestGoogleStateAndUnconfigured(t *testing.T) {
	s := testServer(t)
	w := request(t, s, userSession{}, "GET", "/auth/google/login", nil)
	if w.Code != 503 {
		t.Fatal("설정 없는 OAuth 허용")
	}
	s.opts.GoogleClientID = "client"
	s.opts.GoogleClientSecret = "secret"
	w = request(t, s, userSession{}, "GET", "/auth/google/login", nil)
	if w.Code != 302 || !strings.Contains(w.Header().Get("Location"), "code_challenge=") {
		t.Fatal("PKCE 누락")
	}
	w = request(t, s, userSession{}, "GET", "/auth/google/callback?state=forged", nil)
	if w.Code != 400 {
		t.Fatal("state 위조 허용")
	}
}
