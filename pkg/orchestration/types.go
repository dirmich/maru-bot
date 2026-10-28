package orchestration

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const Protocol = 1

type Node struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Platform string    `json:"platform"`
	Arch     string    `json:"arch"`
	Version  string    `json:"version"`
	Protocol int       `json:"protocol"`
	Tools    []string  `json:"tools"`
	LastSeen time.Time `json:"last_seen"`
	Disabled bool      `json:"disabled"`
	Paused   bool      `json:"paused"`
	Online   bool      `json:"online"`
	CPU      float64   `json:"cpu"`
	Memory   float64   `json:"memory"`
}

type Task struct {
	ID            string                 `json:"id"`
	Instruction   string                 `json:"instruction"`
	NodeID        string                 `json:"node_id"`
	Tool          string                 `json:"tool"`
	Args          map[string]interface{} `json:"args"`
	DependsOn     []string               `json:"depends_on"`
	ApprovalLevel string                 `json:"approval_level"`
	Status        string                 `json:"status"`
	Output        string                 `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	FinishedAt    *time.Time             `json:"finished_at,omitempty"`
	ApprovedAt    *time.Time             `json:"approved_at,omitempty"`
}

type Run struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	IdempotencyKey string    `json:"idempotency_key"`
	Tasks          []Task    `json:"tasks"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	Deadline       time.Time `json:"deadline"`
}

type Memory struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Kind       string     `json:"kind"`
	Visibility string     `json:"visibility"`
	NodeID     string     `json:"node_id,omitempty"`
	SessionID  string     `json:"session_id,omitempty"`
	Source     string     `json:"source"`
	Revision   int        `json:"revision"`
	Verified   bool       `json:"verified"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type Event struct {
	ID      int64     `json:"id"`
	RunID   string    `json:"run_id,omitempty"`
	TaskID  string    `json:"task_id,omitempty"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type Invocation struct {
	Protocol       int       `json:"protocol"`
	OwnerID        string    `json:"owner_id"`
	RunID          string    `json:"run_id"`
	Task           Task      `json:"task"`
	Deadline       time.Time `json:"deadline"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type Result struct {
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func terminal(s string) bool {
	switch s {
	case "succeeded", "failed", "cancelled", "blocked", "unknown":
		return true
	}
	return false
}
func Contains(items []string, value string) bool {
	for _, s := range items {
		if s == value {
			return true
		}
	}
	return false
}
