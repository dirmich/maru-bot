package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPolicyIdentityDedupAndRecovery(t *testing.T) {
	s := testStore(t, "owner")
	token := strings.Repeat("x", 40)
	var calls atomic.Int32
	w, err := NewWorker(s, "node", token, []string{"system_status"}, func(context.Context, Task) (string, error) { calls.Add(1); return "ok", nil })
	if err != nil {
		t.Fatal(err)
	}
	call := Invocation{Protocol: Protocol, OwnerID: "owner", RunID: "run", Task: Task{ID: "step", NodeID: "node", Tool: "system_status", ApprovalLevel: "read-only"}, Deadline: time.Now().Add(time.Minute), IdempotencyKey: "run/step"}
	do := func(c Invocation) *httptest.ResponseRecorder {
		b, _ := json.Marshal(c)
		r := httptest.NewRequest("POST", "/v1/tasks", bytes.NewReader(b))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Idempotency-Key", c.IdempotencyKey)
		rec := httptest.NewRecorder()
		w.Handler().ServeHTTP(rec, r)
		return rec
	}
	if rec := do(call); rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	do(call)
	if calls.Load() != 1 {
		t.Fatal("중복 부작용 실행")
	}
	call.Task.Tool = "shell"
	do(call)
	if calls.Load() != 1 {
		t.Fatal("정책 거부 실패")
	}
	call.OwnerID = "other"
	if rec := do(call); rec.Code != 403 {
		t.Fatal("소유자 불일치 허용")
	}
	r := httptest.NewRequest(http.MethodGet, "/v1/tasks/run-step", nil)
	rec := httptest.NewRecorder()
	w.Handler().ServeHTTP(rec, r)
	if rec.Code != 401 {
		t.Fatal("미인증 조회 허용")
	}
}
