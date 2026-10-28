package orchestration

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	Store        *Store
	NodeID       string
	Token        string
	AllowedTools []string
	ExecuteTool  func(context.Context, Task) (string, error)
	mu           sync.Mutex
	active       map[string]context.CancelFunc
	MaxParallel  int
}
type receipt struct {
	Fingerprint string `json:"fingerprint"`
	Result      Result `json:"result"`
}

func NewWorker(s *Store, node, token string, allow []string, execute func(context.Context, Task) (string, error)) (*Worker, error) {
	if len(token) < 32 || node == "" || execute == nil {
		return nil, errors.New("worker 식별자·32자 이상 토큰·실행기가 필요합니다")
	}
	w := &Worker{Store: s, NodeID: node, Token: token, AllowedTools: allow, ExecuteTool: execute, active: map[string]context.CancelFunc{}, MaxParallel: 4}
	var rs []struct {
		ID      string  `json:"id"`
		Receipt receipt `json:"receipt"`
	}
	if err := s.List("execution", &rs); err != nil {
		return nil, err
	}
	for _, r := range rs {
		if r.Receipt.Result.Status == "running" {
			r.Receipt.Result = Result{Status: "unknown", Error: "worker 재시작: 실제 실행 결과 확인 필요"}
			if err := s.Put("execution", r.ID, r); err != nil {
				return nil, err
			}
		}
	}
	return w, nil
}
func (w *Worker) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tasks", w.submit)
	mux.HandleFunc("GET /v1/tasks/{id}", w.status)
	mux.HandleFunc("POST /v1/tasks/{id}/cancel", w.cancel)
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		token := Bearer(r)
		if subtle.ConstantTimeCompare([]byte(token), []byte(w.Token)) != 1 {
			http.Error(rw, "Unauthorized", 401)
			return
		}
		mux.ServeHTTP(rw, r)
	})
}
func workerJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func (w *Worker) submit(rw http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(rw, r.Body, 262144)
	var call Invocation
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&call); err != nil {
		http.Error(rw, "Invalid request", 400)
		return
	}
	if call.Protocol != Protocol || call.OwnerID != w.Store.Owner || call.Task.NodeID != w.NodeID || call.RunID == "" || call.Task.ID == "" || strings.ContainsAny(call.RunID+call.Task.ID, "/\\") || call.IdempotencyKey != call.RunID+"/"+call.Task.ID || r.Header.Get("Idempotency-Key") != call.IdempotencyKey || !call.Deadline.After(time.Now()) || call.Deadline.After(time.Now().Add(24*time.Hour)) {
		workerJSON(rw, 403, Result{Status: "failed", Error: "신원·프로토콜·만료 검증 실패"})
		return
	}
	if !Contains(w.AllowedTools, call.Task.Tool) || Contains([]string{"orchestration_submit", "orchestration_status", "shared_memory"}, call.Task.Tool) {
		w.Store.Event(call.RunID, call.Task.ID, "policy.deny", call.Task.Tool)
		workerJSON(rw, 200, Result{Status: "failed", Error: "worker 로컬 정책에서 도구를 거부했습니다"})
		return
	}
	if !ReadOnlyTool(call.Task.Tool) && (call.Task.ApprovedAt == nil || call.Task.ApprovedAt.After(time.Now()) || call.Task.ApprovalLevel == "read-only") {
		workerJSON(rw, 200, Result{Status: "failed", Error: "변경 작업에는 개별 승인이 필요합니다"})
		return
	}
	b, _ := json.Marshal(call)
	hash := sha256.Sum256(b)
	fingerprint := hex.EncodeToString(hash[:])
	id := call.RunID + "-" + call.Task.ID
	w.mu.Lock()
	var existing struct {
		ID      string  `json:"id"`
		Receipt receipt `json:"receipt"`
	}
	err := w.Store.Get("execution", id, &existing)
	if err == nil {
		w.mu.Unlock()
		if existing.Receipt.Fingerprint != fingerprint {
			workerJSON(rw, 409, Result{Status: "failed", Error: "멱등 키 충돌"})
			return
		}
		workerJSON(rw, 200, existing.Receipt.Result)
		return
	}
	if !errors.Is(err, ErrNotFound) {
		w.mu.Unlock()
		http.Error(rw, "Storage error", 500)
		return
	}
	if len(w.active) >= w.MaxParallel {
		w.mu.Unlock()
		http.Error(rw, "Worker busy", 429)
		return
	}
	ctx, cancel := context.WithDeadline(context.Background(), call.Deadline)
	w.active[id] = cancel
	existing.ID = id
	existing.Receipt = receipt{Fingerprint: fingerprint, Result: Result{Status: "running"}}
	if err = w.Store.Put("execution", id, existing); err != nil {
		delete(w.active, id)
		cancel()
		w.mu.Unlock()
		http.Error(rw, "Storage error", 500)
		return
	}
	w.Store.Event(call.RunID, call.Task.ID, "policy.allow", call.Task.Tool)
	w.mu.Unlock()
	done := make(chan Result, 1)
	go func() {
		result := Result{Status: "failed"}
		defer func() {
			if recover() != nil {
				result = Result{Status: "unknown", Error: "도구 실행 중 비정상 종료"}
			}
			cancel()
			w.mu.Lock()
			delete(w.active, id)
			existing.Receipt.Result = result
			saveErr := w.Store.Put("execution", id, existing)
			w.Store.Event(call.RunID, call.Task.ID, "execution."+result.Status, call.Task.Tool)
			w.mu.Unlock()
			if saveErr != nil {
				result = Result{Status: "unknown", Error: "실행 결과 저장 실패"}
			}
			done <- result
		}()
		output, err := w.ExecuteTool(ctx, call.Task)
		result = Result{Status: "succeeded", Output: output}
		if len(result.Output) > 65536 {
			result.Output = result.Output[:65536] + "\n[출력 생략]"
		}
		if err != nil {
			result = Result{Status: "failed", Error: "도구 실행 실패: " + err.Error()}
			if ctx.Err() != nil {
				result.Status = "cancelled"
			}
		}
	}()
	select {
	case result := <-done:
		workerJSON(rw, 200, result)
	case <-r.Context().Done():
	}
}
func (w *Worker) status(rw http.ResponseWriter, r *http.Request) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var item struct {
		ID      string  `json:"id"`
		Receipt receipt `json:"receipt"`
	}
	if err := w.Store.Get("execution", r.PathValue("id"), &item); err != nil {
		http.NotFound(rw, r)
		return
	}
	workerJSON(rw, 200, item.Receipt.Result)
}
func (w *Worker) cancel(rw http.ResponseWriter, r *http.Request) {
	w.mu.Lock()
	defer w.mu.Unlock()
	id := r.PathValue("id")
	if c, ok := w.active[id]; ok {
		c()
		workerJSON(rw, 202, Result{Status: "cancelling"})
		return
	}
	var item struct {
		ID      string  `json:"id"`
		Receipt receipt `json:"receipt"`
	}
	if w.Store.Get("execution", id, &item) != nil {
		http.NotFound(rw, r)
		return
	}
	workerJSON(rw, 200, item.Receipt.Result)
}
