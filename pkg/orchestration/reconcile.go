package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func (x HTTPExecutor) Status(ctx context.Context, owner string, n Node, run, task string) (Result, error) {
	var result Result
	token, err := x.Transport.Token(owner, n.ID)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", n.URL+"/v1/tasks/"+url.PathEscape(run+"-"+task), nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := x.Transport.Client().Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return result, fmt.Errorf("worker 기록 조회: HTTP %d", resp.StatusCode)
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 131072)).Decode(&result)
	return result, err
}

func (e *Engine) Reconcile(ctx context.Context, id string) error {
	r, err := e.Run(id)
	if err != nil {
		return err
	}
	querier, ok := e.executor.(interface {
		Status(context.Context, string, Node, string, string) (Result, error)
	})
	if !ok {
		return fmt.Errorf("worker 기록 조회를 지원하지 않는 전송입니다")
	}
	for _, t := range r.Tasks {
		if t.Status != "unknown" {
			continue
		}
		var node Node
		if err = e.Store.Get("node", t.NodeID, &node); err != nil {
			return err
		}
		result, queryErr := querier.Status(ctx, e.Store.Owner, node, r.ID, t.ID)
		if queryErr != nil {
			continue
		}
		if !terminal(result.Status) || result.Status == "unknown" {
			continue
		}
		e.mu.Lock()
		var current Run
		if err = e.Store.Get("run", id, &current); err != nil {
			e.mu.Unlock()
			return err
		}
		for i := range current.Tasks {
			task := &current.Tasks[i]
			if task.ID == t.ID && task.Status == "unknown" {
				task.Status = result.Status
				task.Output = result.Output
				task.Error = result.Error
				now := time.Now().UTC()
				task.FinishedAt = &now
			}
		}
		aggregate(&current)
		err = e.Store.Put("run", id, current)
		if err == nil {
			e.Store.Event(id, t.ID, "task.reconciled", result.Status+" @ "+node.ID)
		}
		e.mu.Unlock()
		if err != nil {
			return err
		}
	}
	e.notify()
	return nil
}
