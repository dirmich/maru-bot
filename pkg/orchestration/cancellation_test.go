package orchestration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fixtureTransport struct {
	client *http.Client
	token  string
}

func (f fixtureTransport) Client() *http.Client                 { return f.client }
func (f fixtureTransport) Token(string, string) (string, error) { return f.token, nil }
func TestHTTPSCancelAndReconcileWithoutRerun(t *testing.T) {
	owner := "owner"
	token := strings.Repeat("t", 40)
	s := testStore(t, owner)
	started := make(chan struct{})
	w, err := NewWorker(s, "node", token, []string{"system_status"}, func(ctx context.Context, task Task) (string, error) {
		close(started)
		<-ctx.Done()
		return "", ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewTLSServer(w.Handler())
	defer ts.Close()
	store := testStore(t, owner)
	executor := HTTPExecutor{Transport: fixtureTransport{ts.Client(), token}}
	e, err := NewEngine(store, executor, 2)
	if err != nil {
		t.Fatal(err)
	}
	e.SaveNode(Node{ID: "node", URL: ts.URL, Tools: []string{"system_status"}, Protocol: Protocol, LastSeen: time.Now()})
	r, err := e.Submit(Run{Title: "취소", IdempotencyKey: "cancel", Tasks: []Task{{ID: "step", NodeID: "node", Tool: "system_status", ApprovalLevel: "read-only"}}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); e.Wait() }()
	go e.Serve(ctx)
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("worker 시작 실패")
	}
	if err = e.Cancel(r.ID); err != nil {
		t.Fatal(err)
	}
	waitRun(t, e, r.ID, "cancelled")
	r, _ = e.Run(r.ID)
	r.Tasks[0].Status = "unknown"
	r.Status = "unknown"
	e.Store.Put("run", r.ID, r)
	if err = e.Reconcile(context.Background(), r.ID); err != nil {
		t.Fatal(err)
	}
	waitRun(t, e, r.ID, "cancelled")
}
func TestApprovalTimeoutAndMemoryExpiry(t *testing.T) {
	s := testStore(t, "owner")
	e, _ := NewEngine(s, nil, 1)
	r, err := e.Submit(Run{Title: "승인 만료", IdempotencyKey: "expire", Deadline: time.Now().Add(100 * time.Millisecond), Tasks: []Task{{ID: "one", Tool: "shell"}}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); e.Wait() }()
	go e.Serve(ctx)
	waitRun(t, e, r.ID, "cancelled")
	if e.Approve(r.ID, "one", true) == nil {
		t.Fatal("기한 이후 승인 허용")
	}
	expiry := time.Now().Add(40 * time.Millisecond)
	m, err := s.SaveMemory(Memory{Title: "만료", Content: "단기", Kind: "short", Visibility: "shared", ExpiresAt: &expiry})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(60 * time.Millisecond)
	items, _ := s.Memories("", "", "", false)
	if len(items) != 0 {
		t.Fatal("만료 기억 조회 허용")
	}
	history, _ := s.MemoryHistory(m.ID)
	if len(history) != 0 {
		t.Fatal("만료 기억 이력 노출")
	}
}
