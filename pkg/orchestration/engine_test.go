package orchestration

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type execFunc func(context.Context, Node, Invocation) (Result, error)

func (f execFunc) Execute(c context.Context, n Node, i Invocation) (Result, error) { return f(c, n, i) }
func testStore(t *testing.T, owner string) *Store {
	t.Helper()
	s, err := OpenStore(t.TempDir(), owner)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func waitRun(t *testing.T, e *Engine, id, status string) Run {
	t.Helper()
	end := time.Now().Add(5 * time.Second)
	for time.Now().Before(end) {
		r, err := e.Run(id)
		if err != nil {
			t.Fatal(err)
		}
		if r.Status == status {
			return r
		}
		time.Sleep(10 * time.Millisecond)
	}
	r, _ := e.Run(id)
	t.Fatalf("원하는 상태 %s, 실제 %+v", status, r)
	return r
}
func TestParallelDAGApprovalAndIdempotency(t *testing.T) {
	s := testStore(t, "owner-a")
	var concurrent, max, calls atomic.Int32
	gate := make(chan struct{})
	e, err := NewEngine(s, execFunc(func(ctx context.Context, n Node, i Invocation) (Result, error) {
		calls.Add(1)
		v := concurrent.Add(1)
		for {
			old := max.Load()
			if old >= v || max.CompareAndSwap(old, v) {
				break
			}
		}
		if i.Task.ID != "c" {
			select {
			case <-gate:
			case <-ctx.Done():
			}
		}
		concurrent.Add(-1)
		return Result{Status: "succeeded", Output: n.ID}, nil
	}), 4)
	if err != nil {
		t.Fatal(err)
	}
	e.SaveNode(Node{ID: "one", Protocol: Protocol, Tools: []string{"read_file", "shell"}, LastSeen: time.Now()})
	e.SaveNode(Node{ID: "two", Protocol: Protocol, Tools: []string{"read_file", "shell"}, LastSeen: time.Now()})
	input := Run{Title: "병렬", IdempotencyKey: "same", Tasks: []Task{{ID: "a", NodeID: "one", Tool: "read_file", ApprovalLevel: "read-only"}, {ID: "b", NodeID: "two", Tool: "read_file", ApprovalLevel: "read-only"}, {ID: "c", Tool: "shell", DependsOn: []string{"a", "b"}}}}
	r, err := e.Submit(input)
	if err != nil {
		t.Fatal(err)
	}
	again, err := e.Submit(input)
	if err != nil || again.ID != r.ID {
		t.Fatalf("멱등성: %+v %v", again, err)
	}
	changed := input
	changed.Title = "다름"
	if _, err = e.Submit(changed); err == nil {
		t.Fatal("다른 본문 충돌 허용")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go e.Serve(ctx)
	end := time.Now().Add(3 * time.Second)
	for max.Load() < 2 && time.Now().Before(end) {
		time.Sleep(time.Millisecond)
	}
	if max.Load() != 2 {
		t.Fatal("두 장비 동시 실행 실패")
	}
	if calls.Load() != 2 {
		t.Fatal("의존성/승인 이전 실행")
	}
	close(gate)
	waitRun(t, e, r.ID, "need-approval")
	if err = e.Approve(r.ID, "c", true); err != nil {
		t.Fatal(err)
	}
	waitRun(t, e, r.ID, "succeeded")
	if calls.Load() != 3 {
		t.Fatal("중복 실행")
	}
	if err = e.Approve(r.ID, "c", true); !errors.Is(err, ErrConflict) {
		t.Fatal("승인 재사용 허용")
	}
}
func TestIsolationMemoryRevisionExpiry(t *testing.T) {
	root := t.TempDir()
	a, _ := OpenStore(root, "../owner-a")
	defer a.Close()
	b, _ := OpenStore(root, "owner-b")
	defer b.Close()
	m, err := a.SaveMemory(Memory{Title: "설정", Content: "장비 맥락", Kind: "short", Visibility: "private", NodeID: "one"})
	if err != nil {
		t.Fatal(err)
	}
	if m.ExpiresAt == nil {
		t.Fatal("단기 만료 누락")
	}
	out, _ := b.Memories("", "one", "", false)
	if len(out) != 0 {
		t.Fatal("사용자 격리 실패")
	}
	out, _ = a.Memories("", "two", "", true)
	if len(out) != 0 {
		t.Fatal("다른 장비 전용 기억 노출")
	}
	out, _ = a.Memories("", "one", "", true)
	if len(out) != 1 {
		t.Fatal("소유 장비 조회 실패")
	}
	old := m
	m.Content = "수정"
	m, err = a.SaveMemory(m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.SaveMemory(old); !errors.Is(err, ErrConflict) {
		t.Fatal("revision 충돌 누락")
	}
	m, err = a.Promote(m.ID, m.Revision)
	if err != nil || m.Kind != "wiki" || m.ExpiresAt != nil {
		t.Fatal("Wiki 승격 실패", err)
	}
	if err = a.DeleteMemory(m.ID, m.Revision); err != nil {
		t.Fatal(err)
	}
	h, _ := a.MemoryHistory(m.ID)
	if len(h) != 0 {
		t.Fatal("삭제된 기억 이력 잔존")
	}
}
func TestRecoveryCycleAndUnknown(t *testing.T) {
	s := testStore(t, "a")
	e, _ := NewEngine(s, nil, 1)
	_, err := e.Submit(Run{Title: "순환", IdempotencyKey: "cycle", Tasks: []Task{{ID: "a", Tool: "shell", DependsOn: []string{"b"}}, {ID: "b", Tool: "shell", DependsOn: []string{"a"}}}})
	if err == nil {
		t.Fatal("순환 허용")
	}
	s.Put("run", "recover", Run{ID: "recover", Tasks: []Task{{ID: "a", Status: "running"}}})
	e, err = NewEngine(s, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := e.Run("recover")
	if r.Tasks[0].Status != "unknown" {
		t.Fatal("불명 작업 재실행 위험")
	}
}
func TestOfflinePinnedNodeAndDeniedApproval(t *testing.T) {
	s := testStore(t, "a")
	var calls atomic.Int32
	e, _ := NewEngine(s, execFunc(func(context.Context, Node, Invocation) (Result, error) {
		calls.Add(1)
		return Result{Status: "succeeded"}, nil
	}), 2)
	e.SaveNode(Node{ID: "online", Protocol: Protocol, Tools: []string{"shell", "read_file"}, LastSeen: time.Now()})
	e.SaveNode(Node{ID: "offline", Protocol: Protocol, Tools: []string{"read_file"}, LastSeen: time.Now().Add(-time.Hour)})
	r, _ := e.Submit(Run{Title: "지명", IdempotencyKey: "pin", Tasks: []Task{{ID: "a", NodeID: "offline", Tool: "read_file", ApprovalLevel: "read-only"}, {ID: "b", Tool: "shell"}}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go e.Serve(ctx)
	if err := e.Approve(r.ID, "b", false); err != nil {
		t.Fatal(err)
	}
	time.Sleep(80 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatal("지명 노드 대체 또는 거부 작업 실행")
	}
}
