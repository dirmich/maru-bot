package orchestration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Executor interface {
	Execute(context.Context, Node, Invocation) (Result, error)
}
type Engine struct {
	Stopped    chan struct{}
	executions sync.WaitGroup
	Store      *Store
	executor   Executor
	mu         sync.Mutex
	active     map[string]context.CancelFunc
	limit      int
	wake       chan struct{}
}

func NewEngine(s *Store, x Executor, limit int) (*Engine, error) {
	if limit < 1 || limit > 32 {
		limit = 4
	}
	e := &Engine{Stopped: make(chan struct{}), Store: s, executor: x, limit: limit, active: map[string]context.CancelFunc{}, wake: make(chan struct{}, 1)}
	var runs []Run
	if err := s.List("run", &runs); err != nil {
		return nil, err
	}
	for _, r := range runs {
		changed := false
		for i := range r.Tasks {
			if r.Tasks[i].Status == "running" || r.Tasks[i].Status == "cancelling" {
				r.Tasks[i].Status = "unknown"
				r.Tasks[i].Error = "프로세스가 재시작되었습니다. 실제 장비의 실행 결과를 확인하세요"
				changed = true
			}
		}
		if changed {
			aggregate(&r)
			if err := s.Put("run", r.ID, r); err != nil {
				return nil, err
			}
			s.Event(r.ID, "", "run.recovered", "실행 여부 확인이 필요한 작업을 보존했습니다")
		}
	}
	return e, nil
}
func (e *Engine) notify() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}
func (e *Engine) Serve(ctx context.Context) {
	defer close(e.Stopped)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	cleanup := time.NewTicker(time.Hour)
	defer cleanup.Stop()
	e.schedule(ctx)
	for {
		select {
		case <-ctx.Done():
			e.mu.Lock()
			for _, cancel := range e.active {
				cancel()
			}
			e.mu.Unlock()
			return
		case <-e.wake:
			e.schedule(ctx)
		case <-tick.C:
			e.schedule(ctx)
		case <-cleanup.C:
			e.Store.PruneMemories()
		}
	}
}

func validateRun(r *Run) error {
	if len(strings.TrimSpace(r.Title)) == 0 || len(r.Title) > 500 || len(r.Tasks) == 0 || len(r.Tasks) > 64 {
		return errors.New("제목과 1~64개 작업이 필요합니다")
	}
	if r.IdempotencyKey == "" || len(r.IdempotencyKey) > 200 {
		return errors.New("멱등 키가 필요합니다 (200자 이내)")
	}
	if r.Deadline.IsZero() {
		r.Deadline = time.Now().Add(time.Hour)
	}
	if !r.Deadline.After(time.Now()) || r.Deadline.After(time.Now().Add(24*time.Hour)) {
		return errors.New("작업 기한은 현재부터 24시간 이내여야 합니다")
	}
	ids := map[string]int{}
	for i := range r.Tasks {
		t := &r.Tasks[i]
		if t.ID == "" {
			t.ID = fmt.Sprintf("step-%d", i+1)
		}
		if len(t.ID) > 100 || strings.ContainsAny(t.ID, "/\\") {
			return errors.New("작업 ID 형식을 확인하세요")
		}
		if _, ok := ids[t.ID]; ok {
			return errors.New("중복 작업 ID")
		}
		ids[t.ID] = i
		if t.Tool == "" {
			return errors.New("도구를 지정하세요")
		}
		if t.ApprovalLevel == "" {
			t.ApprovalLevel = "local-write"
		}
		if !Contains([]string{"read-only", "local-write", "external-send", "destructive"}, t.ApprovalLevel) {
			return errors.New("알 수 없는 승인 수준")
		}
		if !ReadOnlyTool(t.Tool) && t.ApprovalLevel == "read-only" {
			t.ApprovalLevel = "local-write"
		}
		t.Status = "queued"
		if t.ApprovalLevel != "read-only" {
			t.Status = "need-approval"
		}
		t.Output = ""
		t.Error = ""
		t.ApprovedAt = nil
		t.StartedAt = nil
		t.FinishedAt = nil
	}
	visiting := map[string]bool{}
	done := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return errors.New("작업 의존성에 순환이 있습니다")
		}
		if done[id] {
			return nil
		}
		i, ok := ids[id]
		if !ok {
			return errors.New("존재하지 않는 의존 작업: " + id)
		}
		visiting[id] = true
		for _, dep := range r.Tasks[i].DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		done[id] = true
		return nil
	}
	for id := range ids {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
func ReadOnlyTool(tool string) bool {
	return Contains([]string{"read_file", "list_dir", "web_search", "web_fetch", "system_status"}, tool)
}
func (e *Engine) Submit(r Run) (Run, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	r.ID = ""
	r.Status = ""
	r.CreatedAt = time.Time{}
	b, _ := json.Marshal(r)
	h := sha256.Sum256(b)
	fingerprint := hex.EncodeToString(h[:])
	var existing struct {
		ID   string
		Hash string
	}
	err := e.Store.Get("key", r.IdempotencyKey, &existing)
	if err == nil {
		if fingerprint != existing.Hash {
			return r, errors.New("같은 멱등 키에 다른 작업을 제출할 수 없습니다")
		}
		err = e.Store.Get("run", existing.ID, &r)
		return r, err
	}
	if !errors.Is(err, ErrNotFound) {
		return r, err
	}
	r = Run{}
	if err = json.Unmarshal(b, &r); err != nil {
		return r, err
	}
	if err = validateRun(&r); err != nil {
		return r, err
	}
	r.ID = NewID()
	r.CreatedAt = time.Now().UTC()
	r.Status = "queued"
	tx, err := e.Store.db.Begin()
	if err != nil {
		return r, err
	}
	defer tx.Rollback()
	rb, _ := json.Marshal(r)
	kb, _ := json.Marshal(struct {
		ID   string
		Hash string
	}{r.ID, fingerprint})
	if _, err = tx.Exec(`INSERT INTO records(kind,id,data) VALUES('run',?,?)`, r.ID, rb); err != nil {
		return r, err
	}
	if _, err = tx.Exec(`INSERT INTO records(kind,id,data) VALUES('key',?,?)`, r.IdempotencyKey, kb); err != nil {
		return r, err
	}
	if err = tx.Commit(); err != nil {
		return r, err
	}
	e.Store.Event(r.ID, "", "run.created", "작업을 접수했습니다")
	e.notify()
	return r, nil
}
func (e *Engine) Runs() ([]Run, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var rs []Run
	err := e.Store.List("run", &rs)
	return rs, err
}
func (e *Engine) Run(id string) (Run, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var r Run
	err := e.Store.Get("run", id, &r)
	return r, err
}
func (e *Engine) Approve(run, task string, allow bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	var r Run
	if err := e.Store.Get("run", run, &r); err != nil {
		return err
	}
	if !r.Deadline.After(time.Now()) {
		return errors.New("승인 기한이 지났습니다")
	}
	for i := range r.Tasks {
		t := &r.Tasks[i]
		if t.ID == task {
			if t.Status != "need-approval" {
				return ErrConflict
			}
			now := time.Now().UTC()
			if allow {
				t.Status = "queued"
				t.ApprovedAt = &now
			} else {
				t.Status = "cancelled"
				t.Error = "사용자가 승인을 거부했습니다"
				t.FinishedAt = &now
			}
			aggregate(&r)
			if err := e.Store.Put("run", r.ID, r); err != nil {
				return err
			}
			e.Store.Event(run, task, "task.approval", fmt.Sprintf("개별 작업 승인: %t", allow))
			e.notify()
			return nil
		}
	}
	return ErrNotFound
}
func (e *Engine) Cancel(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	var r Run
	if err := e.Store.Get("run", id, &r); err != nil {
		return err
	}
	for i := range r.Tasks {
		t := &r.Tasks[i]
		if terminal(t.Status) {
			continue
		}
		if cancel, ok := e.active[id+"/"+t.ID]; ok {
			cancel()
			t.Status = "cancelling"
		} else {
			t.Status = "cancelled"
			now := time.Now().UTC()
			t.FinishedAt = &now
		}
	}
	aggregate(&r)
	if err := e.Store.Put("run", id, r); err != nil {
		return err
	}
	return e.Store.Event(id, "", "run.cancel_requested", "취소를 요청했습니다")
}
func (e *Engine) Nodes() ([]Node, error) {
	var nodes []Node
	err := e.Store.List("node", &nodes)
	for i := range nodes {
		nodes[i].Online = !nodes[i].Disabled && nodes[i].Protocol == Protocol && time.Since(nodes[i].LastSeen) < 45*time.Second
	}
	return nodes, err
}
func (e *Engine) SaveNode(n Node) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if n.ID == "" {
		return errors.New("장비 ID가 필요합니다")
	}
	if err := e.Store.Put("node", n.ID, n); err != nil {
		return err
	}
	e.notify()
	return nil
}
func (e *Engine) Heartbeat(n Node) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	var current Node
	if err := e.Store.Get("node", n.ID, &current); err != nil {
		return err
	}
	if current.Disabled {
		return errors.New("장비 등록이 폐기되었습니다")
	}
	current.Platform = n.Platform
	current.Arch = n.Arch
	current.Version = n.Version
	current.Protocol = n.Protocol
	current.Tools = n.Tools
	current.CPU = n.CPU
	current.Memory = n.Memory
	current.LastSeen = time.Now().UTC()
	return e.Store.Put("node", n.ID, current)
}
func aggregate(r *Run) {
	pending := false
	running := false
	approval := false
	bad := ""
	for _, t := range r.Tasks {
		switch t.Status {
		case "running", "cancelling":
			running = true
		case "queued":
			pending = true
		case "need-approval":
			approval = true
		case "unknown":
			bad = "unknown"
		case "failed", "blocked":
			if bad != "unknown" {
				bad = "failed"
			}
		case "cancelled":
			if bad == "" {
				bad = "cancelled"
			}
		}
	}
	switch {
	case running:
		r.Status = "running"
	case approval:
		r.Status = "need-approval"
	case pending:
		r.Status = "queued"
	case bad != "":
		r.Status = bad
	default:
		r.Status = "succeeded"
	}
}
func (e *Engine) schedule(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var runs []Run
	if err := e.Store.List("run", &runs); err != nil {
		return
	}
	nodes, _ := e.Nodes()
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	loads := map[string]int{}
	for _, run := range runs {
		for _, task := range run.Tasks {
			if task.Status == "running" || task.Status == "cancelling" {
				loads[task.NodeID]++
			}
		}
	}
	for _, r := range runs {
		changed := false
		for i := range r.Tasks {
			t := &r.Tasks[i]
			if terminal(t.Status) || t.Status == "running" || t.Status == "cancelling" {
				continue
			}
			if !r.Deadline.After(time.Now()) {
				t.Status = "cancelled"
				t.Error = "작업 또는 승인 기한 만료"
				changed = true
				continue
			}
			ready := true
			for _, dep := range t.DependsOn {
				for _, dt := range r.Tasks {
					if dt.ID == dep && dt.Status != "succeeded" {
						ready = false
						if terminal(dt.Status) {
							t.Status = "blocked"
							t.Error = "의존 작업이 완료되지 않았습니다"
							changed = true
						}
					}
				}
			}
			if !ready || t.Status == "need-approval" || len(e.active) >= e.limit {
				continue
			}
			var selected *Node
			bestScore := 1e12
			for j := range nodes {
				n := &nodes[j]
				if n.Online && Contains(n.Tools, t.Tool) && (t.NodeID == "" || t.NodeID == n.ID) {
					score := float64(loads[n.ID])*100 + n.CPU + n.Memory*.1
					if score < bestScore {
						selected = n
						bestScore = score
					}
				}
			}
			if selected == nil {
				continue
			}
			t.NodeID = selected.ID
			loads[selected.ID]++
			t.Status = "running"
			now := time.Now().UTC()
			t.StartedAt = &now
			changed = true
			taskCtx, cancel := context.WithDeadline(ctx, r.Deadline)
			key := r.ID + "/" + t.ID
			e.active[key] = cancel
			call := Invocation{Protocol: Protocol, OwnerID: e.Store.Owner, RunID: r.ID, Task: *t, Deadline: r.Deadline, IdempotencyKey: key}
			e.executions.Add(1)
			go func(n Node, c Invocation) { defer e.executions.Done(); e.execute(taskCtx, n, c) }(*selected, call)
		}
		if changed {
			aggregate(&r)
			if err := e.Store.Put("run", r.ID, r); err != nil {
				for _, t := range r.Tasks {
					if cancel, ok := e.active[r.ID+"/"+t.ID]; ok {
						cancel()
					}
				}
			}
			e.Store.Event(r.ID, "", "run.updated", r.Status)
		}
	}
}

func (e *Engine) Wait() { <-e.Stopped; e.executions.Wait() }
func (e *Engine) execute(ctx context.Context, node Node, call Invocation) {
	// schedule의 저장이 끝난 후 원격 실행을 시작한다.
	e.mu.Lock()
	e.mu.Unlock()
	var result Result
	var err error
	if ctx.Err() != nil {
		result = Result{Status: "cancelled"}
	} else {
		result, err = e.executor.Execute(ctx, node, call)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	key := call.IdempotencyKey
	if cancel, ok := e.active[key]; ok {
		cancel()
		delete(e.active, key)
	}
	var r Run
	if e.Store.Get("run", call.RunID, &r) != nil {
		return
	}
	for i := range r.Tasks {
		t := &r.Tasks[i]
		if t.ID != call.Task.ID {
			continue
		}
		if err != nil {
			t.Status = "unknown"
			t.Error = "원격 결과를 확인하지 못했습니다. 장비 실행 기록을 확인하세요"
		} else {
			t.Status = result.Status
			if !terminal(t.Status) {
				t.Status = "unknown"
			}
			t.Output = result.Output
			t.Error = result.Error
		}
		now := time.Now().UTC()
		t.FinishedAt = &now
		e.Store.Event(r.ID, t.ID, "task."+t.Status, t.Status+" @ "+node.ID)
		if t.Status == "succeeded" && t.Output != "" {
			content := t.Output
			if len(content) > 16000 {
				content = content[:16000]
			}
			e.Store.SaveMemory(Memory{Title: r.Title + " · " + t.ID, Content: content, Kind: "short", Visibility: "private", NodeID: node.ID, Source: "run:" + r.ID + "/" + t.ID})
		}
	}
	aggregate(&r)
	e.Store.Put("run", r.ID, r)
	e.notify()
}
